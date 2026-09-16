package outbound

import (
	"context"
	"sage-backend/internal/admin/domain"
)

type AdminRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Admin, error)
	GetByEmail(ctx context.Context, email string) (*domain.Admin, error)
	Create(ctx context.Context, admin *domain.Admin) error
	Update(ctx context.Context, admin *domain.Admin) error
}
