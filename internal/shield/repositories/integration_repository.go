package repositories

import (
	"context"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/models"
)

var (
	CREATE_INTEGRATION = `
	INSERT INTO integrations (
		id,
		tenant_id,
		name,
		provider,
		connection_type,
		status,
		config
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	GET_INTEGRATION_BY_ID = `
	SELECT id, tenant_id, name, provider, connection_type, status, config, created_at, updated_at
	FROM integrations
	WHERE id = $1
	`
	UPDATE_INTEGRATION_STATUS = `
	UPDATE integrations
	SET status = $1, updated_at = now()
	WHERE id = $2
	`
	CREATE_INTEGRATION_CREDENTIALS = `
	INSERT INTO integration_credentials (
		id, integration_id, key, encrypted_value, expires_at
	)
	VALUES ($1,$2,$3,$4,$5)
	`
	GET_CREDENTIALS_BY_INTEGRATION = `
	SELECT id, integration_id, key, encrypted_value, expires_at, created_at
	FROM integration_credentials
	WHERE integration_id = $1
	`
)

type IntegrationRepositoryInt interface {
	CreateIntegration(ctx context.Context, integration *models.Integration) error
	GetIntegrationById(ctx context.Context, id string) (*models.Integration, error)
	UpdateIntegrationStatus(ctx context.Context, id string, status string) error
	CreateCredential(ctx context.Context, c *models.IntegrationCredentials) error
	GetCredentialsByIntegration(ctx context.Context, integrationID string) ([]models.IntegrationCredentials, error)
}

type IntegrationRepository struct {
	db *db.DB
}

func NewIntegrationRepository(db *db.DB) IntegrationRepositoryInt {
	return &IntegrationRepository{
		db: db,
	}
}

func (r *IntegrationRepository) CreateIntegration(ctx context.Context, integration *models.Integration) error {
	_, err := r.db.ExecContext(
		ctx,
		CREATE_INTEGRATION,
		&integration.ID,
		&integration.TenantId,
		&integration.Name,
		&integration.Provider,
		&integration.ConnectionType,
		&integration.Status,
		&integration.Config,
	)
	if err != nil {
		return err
	}

	return nil
}
func (r *IntegrationRepository) GetIntegrationById(ctx context.Context, id string) (*models.Integration, error) {
	var integration models.Integration
	err := r.db.QueryRowContext(ctx, GET_INTEGRATION_BY_ID, id).Scan(
		&integration.ID,
		&integration.TenantId,
		&integration.Name,
		&integration.Provider,
		&integration.ConnectionType,
		&integration.Status,
		&integration.Config,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &integration, nil
}
func (r *IntegrationRepository) UpdateIntegrationStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx, UPDATE_INTEGRATION_STATUS, status, id)
	return err
}
func (r *IntegrationRepository) CreateCredential(ctx context.Context, c *models.IntegrationCredentials) error {
	_, err := r.db.ExecContext(ctx,
		CREATE_INTEGRATION_CREDENTIALS,
		c.ID,
		c.IntegrationId,
		c.Key,
		c.EncryptedValue,
		c.ExpiresAt,
	)

	return err
}
func (r *IntegrationRepository) GetCredentialsByIntegration(ctx context.Context, integrationID string) ([]models.IntegrationCredentials, error) {
	rows, err := r.db.QueryContext(ctx, GET_CREDENTIALS_BY_INTEGRATION, integrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []models.IntegrationCredentials

	for rows.Next() {
		var c models.IntegrationCredentials

		if err := rows.Scan(
			&c.ID,
			&c.IntegrationId,
			&c.Key,
			&c.EncryptedValue,
			&c.ExpiresAt,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		creds = append(creds, c)
	}

	return creds, nil
}
