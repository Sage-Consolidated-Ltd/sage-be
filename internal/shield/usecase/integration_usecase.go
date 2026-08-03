package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/adapters/outbound/providers"
	"sage-backend/internal/shield/ports/outbound"
	"sage-backend/internal/shield/usecase/dto"
	"sage-backend/pkg/crypto"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type IntegrationUseCase interface {
	CreateDataSource(ctx context.Context, orgID uuid.UUID, req dto.CreateIntegrationRequest) error
}

type DataSourceService struct {
	dataSourceRepo  outbound.DataSourceRepository
	integrationRepo outbound.IntegrationRepository
	encryptor       crypto.Encryptor
	client          *resty.Client
}

func NewDataSourceService(dataSourceRepo outbound.DataSourceRepository, integrationRepo outbound.IntegrationRepository, encryptor crypto.Encryptor, client *resty.Client) IntegrationUseCase {
	return &DataSourceService{
		dataSourceRepo:  dataSourceRepo,
		integrationRepo: integrationRepo,
		encryptor:       encryptor,
		client:          client,
	}
}

func (s *DataSourceService) CreateDataSource(ctx context.Context, orgID uuid.UUID, req dto.CreateIntegrationRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var (
		provider providers.Provider
		err      error

		credentialMap map[string]string
		metadata      map[string]interface{}
	)

	switch req.Provider {

	case "okta":

		if req.Okta == nil {
			return fmt.Errorf("okta config required")
		}

		provider, err = providers.NewProvider(
			"okta",
			providers.OktaCredentials{
				Domain: req.Okta.Domain,
				Token:  req.Okta.Token,
			},
			s.client,
		)
		if err != nil {
			return err
		}

		credentialMap = map[string]string{
			"token":  req.Okta.Token,
			"domain": req.Okta.Domain,
		}

		metadata = map[string]interface{}{
			"domain": req.Okta.Domain,
		}

	case "entra":

		if req.Entra == nil {
			return fmt.Errorf("entra config required")
		}

		provider, err = providers.NewProvider(
			"entra",
			providers.EntraCredentials{
				TenantID:     req.Entra.TenantID,
				ClientID:     req.Entra.ClientID,
				ClientSecret: req.Entra.ClientSecret,
			},
			s.client,
		)
		if err != nil {
			return err
		}

		credentialMap = map[string]string{
			"tenant_id":     req.Entra.TenantID,
			"client_id":     req.Entra.ClientID,
			"client_secret": req.Entra.ClientSecret,
		}

		metadata = map[string]interface{}{}

	default:
		return fmt.Errorf("unsupported provider")
	}

	if err := provider.Verify(ctx); err != nil {
		return err
	}

	encrypted := make(map[string]string)

	for k, v := range credentialMap {
		enc, err := s.encryptor.Encrypt(v)
		if err != nil {
			return err
		}

		encrypted[k] = enc
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	fmt.Println("encrypted credentials:", encrypted)

	ds := &domain.DataSource{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           req.Name,
		Provider:       &req.Provider,
		Type:           req.ConnectionType,
		Status:         domain.DataSourceStatusActive,
		Metadata:       metadataJSON,
	}

	// create credential rows
	var creds []domain.IntegrationCredentials

	now := time.Now().UTC()

	for k, v := range encrypted {
		creds = append(creds, domain.IntegrationCredentials{
			ID:             uuid.New(),
			Key:            k,
			EncryptedValue: v,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	return s.integrationRepo.CreateDataSourceWithCredentialsBulk(ctx, &creds, ds)
}

// import (
// 	"context"
// 	"sage-backend/internal/shared/errors/apperrors"
// 	"sage-backend/internal/shield/domain"
// 	"sage-backend/internal/shield/ports/outbound"
// 	"sage-backend/internal/shield/usecase/dto"
// 	"sage-backend/pkg/crypto"
// 	"sage-backend/pkg/validator"
// )

// type IntegrationServiceInt interface {
// 	CreateIntegration(ctx context.Context, tenantID string, req dto.CreateIntegrationRequest) (*domain.IntegrationResponse, error)
// 	ActivateIntegration(ctx context.Context, id string) error
// }

// type IntegrationService struct {
// 	IntegrationRepo outbound.IntegrationRepository
// 	encryptor       crypto.Encryptor
// }

// func NewIntegrationService(IntegrationRepo outbound.IntegrationRepository, encryptor crypto.Encryptor) IntegrationServiceInt {
// 	return &IntegrationService{
// 		IntegrationRepo: IntegrationRepo,
// 		encryptor:       encryptor,
// 	}
// }

// func (s *IntegrationService) CreateIntegration(ctx context.Context, tenantID string, req dto.CreateIntegrationRequest) (*domain.IntegrationResponse, error) {
// 	_, ok := validator.Validators[req.Provider]
// 	if !ok {
// 		return nil, apperrors.BadException("unsupported provider")
// 	}

// 	integration := dto.MapCreateIntegration(req, tenantID)

// 	if err := s.IntegrationRepo.CreateIntegration(ctx, &integration); err != nil {
// 		return nil, err
// 	}

// 	for _, cred := range req.Credentials {
// 		enc, err := s.encryptor.Encrypt(cred.Value)
// 		if err != nil {
// 			return nil, err
// 		}

// 		c := domain.IntegrationCredentials{
// 			IntegrationId:  integration.ID,
// 			Key:            cred.Key,
// 			EncryptedValue: enc,
// 		}

// 		if err := s.IntegrationRepo.CreateCredential(ctx, &c); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return dto.MapIntegrationResponse(&integration), nil
// }

// func (s *IntegrationService) ActivateIntegration(ctx context.Context, id string) error {
// 	integration, err := s.IntegrationRepo.GetIntegrationById(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	if integration.Status == "active" {
// 		return nil
// 	}

// 	creds, err := s.IntegrationRepo.GetCredentialsByIntegration(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	decrypted := map[string]string{}
// 	for _, c := range creds {
// 		val, err := s.encryptor.Decrypt(c.EncryptedValue)
// 		if err != nil {
// 			return err
// 		}
// 		decrypted[c.Key] = val
// 	}

// 	if err := s.testConnection(ctx, *integration, decrypted); err != nil {
// 		_ = s.IntegrationRepo.UpdateIntegrationStatus(ctx, id, "error")
// 		return err
// 	}

// 	return s.IntegrationRepo.UpdateIntegrationStatus(ctx, id, "active")
// }

// func (s *IntegrationService) testConnection(ctx context.Context, integration domain.Integration, creds map[string]string) error {
// 	switch integration.Provider {
// 	case "aws":
// 		cfg, err := aws.ParseConfig(integration.Config)
// 		if err != nil {
// 			return err
// 		}

// 		if accessKey, ok := creds["access_key"]; ok {
// 			cfg.AccessKey = accessKey
// 		}
// 		if secretKey, ok := creds["secret_key"]; ok {
// 			cfg.SecretKey = secretKey
// 		}

// 		if err := aws.Validate(cfg); err != nil {
// 			return err
// 		}

// 		return aws.TestConnection(ctx, cfg)
// 	default:
// 		return apperrors.BadException("unsupported provider")
// 	}
// }
