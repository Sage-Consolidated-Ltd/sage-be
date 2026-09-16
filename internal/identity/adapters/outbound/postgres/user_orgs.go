package postgres

import (
	"context"
	orgDomain "sage-backend/internal/organization/domain"
)

func (r *UserRepository) GetUserOrganizations(ctx context.Context, userId string) (*[]orgDomain.Organization, error) {
	var models []organizationModel
	err := r.Executor(ctx).SelectContext(ctx, &models, GET_USER_ORGANIZATIONS, userId)
	if err != nil {
		return nil, err
	}
	orgs := make([]orgDomain.Organization, len(models))
	for i, m := range models {
		orgs[i] = m.ToDomain()
	}
	return &orgs, nil
}
