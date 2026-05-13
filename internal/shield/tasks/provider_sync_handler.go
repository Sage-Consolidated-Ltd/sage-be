package tasks

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"sage-backend/internal/shared/types"
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
	eventRepo       repositories.SecurityEventRepositoryInt
	client          *resty.Client
	encryptor       crypto.Encryptor
}

func NewProviderSyncHandler(
	dataSourceRepo repositories.DataSourceRepositoryInt,
	integrationRepo repositories.IntegrationRepositoryInt,
	eventRepo repositories.SecurityEventRepositoryInt,
	client *resty.Client,
	encryptor crypto.Encryptor,
) *ProviderSyncHandler {
	return &ProviderSyncHandler{
		dataSourceRepo:  dataSourceRepo,
		integrationRepo: integrationRepo,
		eventRepo:       eventRepo,
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

	if err := json.Unmarshal(task.Payload(),&payload); err != nil {
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

	// build provider
	provider, err :=
		providers.LaunchProviderSync(
			*source.Provider,
			decryptedCreds,
			h.client,
		)

	if err != nil {
		return err
	}

	// collect events
	events, err := provider.Collect(ctx)
	if err != nil {
		return err
	}

	log.Printf("Collected %d events from %s", len(events), *source.Provider)

	if len(events) == 0 {
		return nil
	}

	// Convert to SecurityEvent models and bulk insert
	var toInsert []*models.SecurityEvent
	var latest time.Time
	for _, ev := range events {
		if ev.Timestamp.After(latest) {
			latest = ev.Timestamp
		}
		idCopy := ev.ID
		ipCopy := ev.IPAddress
		userCopy := ev.UserName
		se := &models.SecurityEvent{
			OrganizationID:    payload.OrganizationID,
			SourceID:          source.ID,
			SourceEventID:     &idCopy,
			Source:            ptrStringValue(source.Provider),
			EventType:         ev.EventType,
			IPAddress:         &ipCopy,
			ActorUsername:     &userCopy,
			OccurredAt:        ev.Timestamp,
			RawPayload:        ev.Raw,
			NormalizedPayload: map[string]interface{}{"user_id": ev.UserID},
			ParseStatus:       types.ParseStatusPending,
		}
		toInsert = append(toInsert, se)
	}

	if err := h.eventRepo.BulkCreateEvents(ctx, toInsert); err != nil {
		return err
	}

	// Update data source health metrics
	now := time.Now()
	if err := h.dataSourceRepo.UpdateHealthMetrics(ctx, source.ID, int64(len(toInsert)), int64(len(toInsert)), 0, &latest, &now); err != nil {
		log.Printf("Failed to update data source metrics: %v", err)
	}

	log.Printf("Persisted %d events for source %s", len(toInsert), source.ID)

	return nil
}

func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}