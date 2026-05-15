package tasks

import (
	"context"
	"encoding/json"
	"log"

	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/providers"
	"sage-backend/internal/shield/repositories"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ProviderSyncHandler struct {
	dataSourceRepo  repositories.DataSourceRepositoryInt
	integrationRepo repositories.IntegrationRepositoryInt
	taskClient      *TaskClient
	client          *resty.Client
	encryptor       crypto.Encryptor
}

func NewProviderSyncHandler(
	dataSourceRepo repositories.DataSourceRepositoryInt,
	integrationRepo repositories.IntegrationRepositoryInt,
	taskClient *TaskClient,
	client *resty.Client,
	encryptor crypto.Encryptor,
) *ProviderSyncHandler {
	return &ProviderSyncHandler{
		dataSourceRepo:  dataSourceRepo,
		integrationRepo: integrationRepo,
		taskClient:      taskClient,
		client:          client,
		encryptor:       encryptor,
	}
}

func (h *ProviderSyncHandler) ProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload struct {
		OrganizationID uuid.UUID `json:"organization_id"`
		SourceID       uuid.UUID `json:"source_id"`
	}

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	// get datasource
	source, err := h.dataSourceRepo.GetDataSourceByID(
		ctx,
		payload.SourceID,
		payload.OrganizationID,
	)

	if err != nil {
		return err
	}

	// get encrypted credentials
	creds, err := h.integrationRepo.GetCredentialsByIntegration(
		ctx,
		source.ID.String(),
	)

	if err != nil {
		return err
	}

	// decrypt credentials
	decryptedCreds := make(map[string]string)

	for _, cred := range creds {
		value, err :=
			h.encryptor.Decrypt(
				cred.EncryptedValue,
			)

		if err != nil {
			return err
		}

		decryptedCreds[cred.Key] = value
	}

	checkpoint := &models.Checkpoint{
		LastCheckpoint:   source.LastCheckpoint,
		LastCheckpointAt: source.LastCheckpointAt,
	}

	// build provider
	provider, err :=
		providers.LaunchProviderSync(
			*source.Provider,
			decryptedCreds,
			checkpoint,
			h.client,
		)

	if err != nil {
		return err
	}

	// collect events
	events, err := provider.Collect(ctx)
	if err != nil {
		log.Printf("Error collecting logs: %v", err)
		return err
	}

	log.Printf("Collected %d events from %s", len(events), *source.Provider)

	if len(events) == 0 {
		return nil
	}

	if err := h.taskClient.EnqueueProviderEventBatch(ctx, payload.OrganizationID, source.ID, ptrStringValue(source.Provider), events); err != nil {
		return err
	}

	log.Printf("Queued %d normalized events for source %s", len(events), source.ID)

	return nil
}

func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
