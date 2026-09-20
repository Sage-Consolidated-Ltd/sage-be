package outbound

import (
	"context"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/usecase/dto"
	orgDomain "sage-backend/internal/organization/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest, hash string) error
	CreateUserWithOrganization(ctx context.Context, req *dto.CreateUserRequest, hash string) error
	OnboardUserWithTransaction(ctx context.Context, req *dto.OnboardingRequest, hash string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserOrganizations(ctx context.Context, userId string) (*[]orgDomain.Organization, error)
	UpdateUser(ctx context.Context, id string, req *dto.UpdateProfileRequest) error
	MarkEmailVerified(ctx context.Context, email string) error
	Enable2FA(ctx context.Context, secret string, userID string) error
	GetTOTPSecret(ctx context.Context, userID string) (string, error)
	UpdateUserPassword(ctx context.Context, email string, hash string) error
	UpdateUserContactInfo(ctx context.Context, id string, phoneNumber, backupEmail string) error
	UpdatePasswordHashByID(ctx context.Context, id string, hash string) error
	SoftDeleteUser(ctx context.Context, id string) error
}
