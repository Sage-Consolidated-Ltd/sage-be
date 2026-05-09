package repositories

import (
	"context"
	"encoding/json"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/models"
	"time"

	"github.com/google/uuid"
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
	CreateCredential(ctx context.Context, c *models.IntegrationCredentials) error
	CreateDataSourceWithCredentialsBulk(ctx context.Context, creds *[]models.IntegrationCredentials, ds *models.DataSource) error
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

func (r *IntegrationRepository) CreateDataSourceWithCredentialsBulk(ctx context.Context, creds *[]models.IntegrationCredentials, ds *models.DataSource) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id uuid.UUID
	var createdAt, updatedAt time.Time
	metaJSON := ds.Metadata
	if metaJSON == nil {
		metaJSON = json.RawMessage{}
	}
	metaJSONMarshalled, err := json.Marshal(ds.Metadata)
	if err != nil {
		return err
	}
	err = tx.QueryRowContext(
		ctx, CREATE_DATA_SOURCE,
		ds.OrganizationID, ds.Name, ds.Description, ds.Type, ds.Provider,
		ds.Status, ds.LastEventAt, ds.LastSyncAt, metaJSONMarshalled,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}
	ds.ID = id
	ds.CreatedAt = createdAt
	ds.UpdatedAt = updatedAt
	ds.Metadata = metaJSON

	stmt, err := tx.PrepareContext(ctx, CREATE_INTEGRATION_CREDENTIALS)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range *creds {
		if _, err := stmt.ExecContext(ctx,
			c.ID,
			ds.ID,
			c.Key,
			c.EncryptedValue,
			c.ExpiresAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
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
