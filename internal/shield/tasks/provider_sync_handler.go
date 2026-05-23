package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/providers"
	"sage-backend/internal/shield/repositories"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"sage-backend/internal/shared/types"
)

type ProviderSyncHandler struct {
	dataSourceRepo  repositories.DataSourceRepositoryInt
	integrationRepo repositories.IntegrationRepositoryInt
	eventRepo       repositories.SecurityEventRepositoryInt
	taskClient      *TaskClient
	client          *resty.Client
	encryptor       crypto.Encryptor
}

func NewProviderSyncHandler(
	dataSourceRepo repositories.DataSourceRepositoryInt,
	integrationRepo repositories.IntegrationRepositoryInt,
	eventRepo repositories.SecurityEventRepositoryInt,
	taskClient *TaskClient,
	client *resty.Client,
	encryptor crypto.Encryptor,
) *ProviderSyncHandler {
	return &ProviderSyncHandler{
		dataSourceRepo:  dataSourceRepo,
		integrationRepo: integrationRepo,
		eventRepo:       eventRepo,
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

	log.Printf("Starting provider sync task for org=%s source=%s", payload.OrganizationID, payload.SourceID)

	// get datasource
	source, err := h.dataSourceRepo.GetDataSourceByID(
		ctx,
		payload.SourceID,
		payload.OrganizationID,
	)

	if err != nil {
		return err
	}

	if source.Provider == nil || strings.TrimSpace(*source.Provider) == "" {
		return fmt.Errorf("source %s has no provider configured", source.ID)
	}

	providerName := strings.ToLower(strings.TrimSpace(*source.Provider))

	// get encrypted credentials
	creds, err := h.integrationRepo.GetCredentialsByIntegration(
		ctx,
		source.ID.String(),
	)

	if err != nil {
		return err
	}

	if len(creds) == 0 {
		return fmt.Errorf("no credentials found for source %s (provider=%s); verify integration_credentials links to data_sources.id and migration 000034 has been applied", source.ID, providerName)
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
	fmt.Printf("decrypted credentials: %v\n", decryptedCreds)

	checkpoint := &models.Checkpoint{
		LastCheckpoint:   source.LastCheckpoint,
		LastCheckpointAt: source.LastCheckpointAt,
	}

	// build provider
	provider, err :=
		providers.LaunchProviderSync(
			providerName,
			decryptedCreds,
			checkpoint,
			h.client,
		)

	if err != nil {
		log.Printf("Provider sync failed for source %s provider=%s: %v", source.ID, providerName, err)
		return err
	}

	// collect events
	events, err := provider.Collect(ctx)
	if err != nil {
		log.Printf("Error collecting logs: %v", err)
		return err
	}

	log.Printf("Collected %d events from %s", len(events), providerName)

	if len(events) == 0 {
		return nil
	}

	_, latest, err := h.PersistProviderEvents(ctx, events, source.OrganizationID, source.ID, providerName)
	if err != nil {
		log.Printf("failed to persist events for source %s: %w", source.ID, err)
		return fmt.Errorf("failed to persist events for source %s: %w", source.ID, err)
	}

	now := time.Now()
	if err := h.dataSourceRepo.UpdateHealthMetrics(ctx, payload.SourceID, int64(len(events)), int64(len(events)), 0, latest, &now); err != nil {
		log.Printf("Failed to update data source metrics: %v", err)
	}

	// if err := h.taskClient.EnqueueProviderEventBatch(ctx, payload.OrganizationID, source.ID, providerName, events); err != nil {
	// 	return err
	// }

	lastCheckpoint := latestEventCheckpoint(events)
	if err := h.dataSourceRepo.UpdateCheckpoint(ctx, source.ID, lastCheckpoint); err != nil {
		return fmt.Errorf("failed to persist checkpoint for source %s: %w", source.ID, err)
	}

	log.Printf("Persisted %d normalized events for source %s", len(events), source.ID)

	log.Printf("Persisted checkpoint %s for source %s", lastCheckpoint, source.ID)

	return nil
}

func latestEventCheckpoint(events []models.NormalizedEvent) string {
	var latest time.Time
	for _, event := range events {
		if event.Timestamp.After(latest) {
			latest = event.Timestamp
		}
	}

	if latest.IsZero() {
		latest = time.Now().UTC()
	}

	return latest.UTC().Format(time.RFC3339)
}

func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *ProviderSyncHandler) PersistProviderEvents(
	ctx context.Context,
	events []models.NormalizedEvent,
	orgID uuid.UUID,
	sourceID uuid.UUID,
	provider string,
) ([]uuid.UUID, *time.Time, error) {
	var securityEvents []*models.SecurityEvent
	var latest time.Time

	for _, ev := range events {
		if ev.Timestamp.After(latest) {
			latest = ev.Timestamp
		}
		idCopy := ev.ID
		ipCopy := ev.IPAddress
		userCopy := ev.UserName
		securityEvents = append(securityEvents, &models.SecurityEvent{
			OrganizationID:    orgID,
			SourceID:          sourceID,
			SourceEventID:     &idCopy,
			Source:            provider,
			EventType:         ev.EventType,
			IPAddress:         &ipCopy,
			ActorUsername:     &userCopy,
			RawPayload:        ev.Raw,
			NormalizedPayload: map[string]interface{}{"user_id": ev.UserID},
			ParseStatus:       types.ParseStatusPending,
			OccurredAt:        ev.Timestamp,
		})
	}

	eventIDs, err := h.eventRepo.BulkCreateEventsWithReturning(ctx, securityEvents)
	if err != nil {
		return nil, nil, err
	}
	return eventIDs, &latest, nil
}
