package scheduler

import (
	"context"
	"log"
	"time"

	// "sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/tasks"
)

type ProviderScheduler struct {
	taskClient     *tasks.TaskClient
	dataSourceRepo repositories.DataSourceRepositoryInt
	syncInterval   time.Duration
	stopCh         chan struct{}
}

func NewProviderScheduler(
	taskClient *tasks.TaskClient,
	dataSourceRepo repositories.DataSourceRepositoryInt,
	syncIntervalSec int,
) *ProviderScheduler {
	return &ProviderScheduler{
		taskClient:     taskClient,
		dataSourceRepo: dataSourceRepo,
		syncInterval:   time.Duration(syncIntervalSec) * time.Second,
		stopCh:         make(chan struct{}),
	}
}

// Start begins the periodic provider sync scheduler.
func (ps *ProviderScheduler) Start(ctx context.Context) {
	go ps.run(ctx)
	log.Println("Provider scheduler started")
}

// Stop gracefully stops the scheduler.
func (ps *ProviderScheduler) Stop() {
	close(ps.stopCh)
	log.Println("Provider scheduler stopped")
}

func (ps *ProviderScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(ps.syncInterval)
	defer ticker.Stop()

	// Run immediately on startup
	ps.syncAllActiveSources(ctx)

	for {
		select {
		case <-ticker.C:
			ps.syncAllActiveSources(ctx)
		case <-ps.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (ps *ProviderScheduler) syncAllActiveSources(ctx context.Context) {
	log.Println("Running scheduled provider sync for all active sources")

	// Get all organizations (we need to iterate through all orgs to find all active sources)
	// For now, we'll use a workaround: fetch sources by status across all orgs
	// In production, you'd want a repository method that returns all active sources globally
	// or iterate through known organizations

	// Fetch all active data sources (this is a simplified approach)
	// In a production system, you'd need to iterate through organizations or have a global query
	allSources, err := ps.dataSourceRepo.ListAllActiveDataSources(ctx)

	if err != nil {
		log.Printf("Failed to list active data sources for scheduled sync: %v", err)
		return
	}

	if len(allSources) == 0 {
		log.Println("No active data sources to sync")
		return
	}

	log.Printf("Found %d active data sources to sync", len(allSources))

	// Enqueue sync tasks for each active source
	for _, source := range allSources {
		if source.Provider == nil || *source.Provider == "" {
			log.Printf("Skipping source %s: no provider configured", source.ID)
			continue
		}

		if err := ps.taskClient.EnqueueProviderSync(ctx, source.OrganizationID, source.ID); err != nil {
			log.Printf("Failed to enqueue sync for source %s: %v", source.ID, err)
			continue
		}

		log.Printf("Enqueued sync for source %s (%s)", source.Name, *source.Provider)
	}

	log.Printf("Scheduled sync complete: %d sources enqueued", len(allSources))
}
