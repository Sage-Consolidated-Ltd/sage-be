package services

import (
	"context"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/providers/aws"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/requests"
	"sage-backend/pkg/crypto"
	"sage-backend/pkg/validator"
)

type IntegrationServiceInt interface {
	CreateIntegration(ctx context.Context, tenantID string, req requests.CreateIntegrationRequest) (*models.IntegrationResponse, error)
	ActivateIntegration(ctx context.Context, id string) error
}

type IntegrationService struct {
	IntegrationRepo repositories.IntegrationRepositoryInt
	encryptor       crypto.Encryptor
}

func NewIntegrationService(IntegrationRepo repositories.IntegrationRepositoryInt, encryptor crypto.Encryptor) IntegrationServiceInt {
	return &IntegrationService{
		IntegrationRepo: IntegrationRepo,
		encryptor:       encryptor,
	}
}

func (s *IntegrationService) CreateIntegration(ctx context.Context, tenantID string, req requests.CreateIntegrationRequest) (*models.IntegrationResponse, error) {
	_, ok := validator.Validators[req.Provider]
	if !ok {
		return nil, apperrors.BadException("unsupported provider")
	}

	integration := requests.MapCreateIntegration(req, tenantID)

	if err := s.IntegrationRepo.CreateIntegration(ctx, &integration); err != nil {
		return nil, err
	}

	for _, cred := range req.Credentials {
		enc, err := s.encryptor.Encrypt(cred.Value)
		if err != nil {
			return nil, err
		}

		c := models.IntegrationCredentials{
			IntegrationId:  integration.ID,
			Key:            cred.Key,
			EncryptedValue: enc,
		}

		if err := s.IntegrationRepo.CreateCredential(ctx, &c); err != nil {
			return nil, err
		}
	}

	return requests.MapIntegrationResponse(&integration), nil
}

func (s *IntegrationService) ActivateIntegration(ctx context.Context, id string) error {
	integration, err := s.IntegrationRepo.GetIntegrationById(ctx, id)
	if err != nil {
		return err
	}

	if integration.Status == "active" {
		return nil
	}

	creds, err := s.IntegrationRepo.GetCredentialsByIntegration(ctx, id)
	if err != nil {
		return err
	}

	decrypted := map[string]string{}
	for _, c := range creds {
		val, err := s.encryptor.Decrypt(c.EncryptedValue)
		if err != nil {
			return err
		}
		decrypted[c.Key] = val
	}

	if err := s.testConnection(ctx, *integration, decrypted); err != nil {
		_ = s.IntegrationRepo.UpdateIntegrationStatus(ctx, id, "error")
		return err
	}

	return s.IntegrationRepo.UpdateIntegrationStatus(ctx, id, "active")
}

func (s *IntegrationService) testConnection(ctx context.Context, integration models.Integration, creds map[string]string) error {
	switch integration.Provider {
	case "aws":
		cfg, err := aws.ParseConfig(integration.Config)
		if err != nil {
			return err
		}

		if accessKey, ok := creds["access_key"]; ok {
			cfg.AccessKey = accessKey
		}
		if secretKey, ok := creds["secret_key"]; ok {
			cfg.SecretKey = secretKey
		}

		if err := aws.Validate(cfg); err != nil {
			return err
		}

		return aws.TestConnection(ctx, cfg)
	default:
		return apperrors.BadException("unsupported provider")
	}
}
