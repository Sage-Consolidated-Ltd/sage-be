package dto

import (
	"strings"
	"time"

	"sage-backend/internal/identity/domain"
	orgDomain "sage-backend/internal/organization/domain"
)

type GetUserResponse struct {
	ID               string                    `json:"id"`
	FirstName        string                    `json:"first_name"`
	LastName         string                    `json:"last_name"`
	AvatarURL        string                    `json:"avatar_url,omitempty"`
	Email            string                    `json:"email"`
	Role             string                    `json:"role"`
	TwoFactorEnabled bool                      `json:"two_factor_enabled"`
	CreatedAt        time.Time                 `json:"created_at"`
	Organization     []GetOrganizationResponse `json:"organization,omitempty"`
	TimeZone         *string                   `json:"time_zone,omitempty"`
	IsVerified       bool                      `json:"is_verified"`
}

type GetOrganizationResponse struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	OwnerID              string     `json:"owner_id"`
	Industry             string     `json:"industry"`
	Country              string     `json:"country,omitempty"`
	Timezone             string     `json:"timezone"`
	RiskThresholdDefault int        `json:"risk_threshold_default"`
	Role                 string     `json:"role"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

type ProfileResponse struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	FullName          string    `json:"full_name"`
	AvatarURL         string    `json:"avatar_url,omitempty"`
	Role              string    `json:"role"`
	JobTitle          string    `json:"job_title"`
	PhoneNumber       string    `json:"phone_number,omitempty"`
	BackupEmail       string    `json:"backup_email,omitempty"`
	PasswordChangedAt time.Time `json:"password_changed_at,omitempty"`
	OrganizationID    string    `json:"organization_id,omitempty"`
	OrganizationName  string    `json:"organization_name,omitempty"`
	LastLoginAt       time.Time `json:"last_login_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type UserSessionResponse struct {
	SessionID    string    `json:"session_id"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Location     string    `json:"location,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsCurrent    bool      `json:"is_current"`
}

func UserToResponse(u *domain.User, orgs *[]orgDomain.Organization) *GetUserResponse {
	var organizationsResp []GetOrganizationResponse
	if orgs != nil {
		for _, org := range *orgs {
			var countryStr string
			if org.Country.Valid {
				countryStr = org.Country.String
			}
			organizationsResp = append(organizationsResp, GetOrganizationResponse{
				ID:                   org.ID,
				Name:                 org.Name,
				Slug:                 org.Slug,
				OwnerID:              org.OwnerID,
				Industry:             org.Industry,
				Country:              countryStr,
				Timezone:             org.Timezone,
				RiskThresholdDefault: org.RiskThresholdDefault,
				Role:                 org.Role,
				Status:               org.Status,
				CreatedAt:            org.CreatedAt,
				UpdatedAt:            org.UpdatedAt,
			})
		}
	}
	resp := &GetUserResponse{
		ID:               u.ID(),
		FirstName:        u.FirstName(),
		LastName:         u.LastName(),
		Email:            u.Email().String(),
		Role:             u.Role().String(),
		TwoFactorEnabled: u.TwoFactorEnabled(),
		CreatedAt:        u.CreatedAt(),
		IsVerified:       u.IsVerified(),
		Organization:     organizationsResp,
	}

	if tz := u.TimeZone(); tz.Valid {
		resp.TimeZone = &tz.String
	}
	if av := u.AvatarURL(); av.Valid {
		resp.AvatarURL = av.String
	}
	return resp
}

func formatJobTitle(role string) string {
	switch strings.ToLower(role) {
	case "owner", "admin", "organization_admin":
		return "Organization Admin"
	case "analyst", "security_analyst":
		return "Security Analyst"
	case "viewer":
		return "Viewer"
	case "automation_admin":
		return "Automation Admin"
	case "billing_admin":
		return "Billing Admin"
	default:
		if role != "" {
			return strings.Title(strings.ReplaceAll(role, "_", " "))
		}
		return "Organization Member"
	}
}

func UserToProfileResponse(u *domain.User, org *orgDomain.Organization) *ProfileResponse {
	jobTitle := "Organization Member"
	if org != nil && org.Role != "" {
		jobTitle = formatJobTitle(org.Role)
	} else if u.Role().String() != "" {
		jobTitle = formatJobTitle(u.Role().String())
	}

	resp := &ProfileResponse{
		ID:        u.ID(),
		Email:     u.Email().String(),
		FullName:  u.FirstName() + " " + u.LastName(),
		Role:      u.Role().String(),
		JobTitle:  jobTitle,
		CreatedAt: u.CreatedAt(),
	}
	if av := u.AvatarURL(); av.Valid {
		resp.AvatarURL = av.String
	}
	if pn := u.PhoneNumber(); pn.Valid {
		resp.PhoneNumber = pn.String
	}
	if be := u.BackupEmail(); be.Valid {
		resp.BackupEmail = be.String
	}
	if pca := u.PasswordChangedAt(); pca.Valid {
		resp.PasswordChangedAt = pca.Time
	}
	if ll := u.LastLoginAt(); ll.Valid {
		resp.LastLoginAt = ll.Time
	}
	if org != nil {
		resp.OrganizationID = org.ID
		resp.OrganizationName = org.Name
	}
	return resp
}

func UserSessionToResponse(s *domain.UserSession, currentSessionID string) *UserSessionResponse {
	resp := &UserSessionResponse{
		SessionID:    s.ID(),
		CreatedAt:    s.CreatedAt(),
		LastActiveAt: s.LastActiveAt(),
		IsCurrent:    s.ID() == currentSessionID,
	}
	if ip := s.IPAddress(); ip.Valid {
		resp.IPAddress = ip.String
	}
	if ua := s.UserAgent(); ua.Valid {
		resp.UserAgent = ua.String
	}
	if loc := s.Location(); loc.Valid {
		resp.Location = loc.String
	}
	return resp
}
