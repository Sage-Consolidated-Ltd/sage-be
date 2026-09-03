package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
)

// OrgMemberRole represents organization-level member roles ("owner", "admin", "analyst", "viewer", "automation_admin", "billing_admin")
type OrgMemberRole struct {
	value string
}

const (
	OrgRoleOwner           = "owner"
	OrgRoleAdmin           = "admin"
	OrgRoleAnalyst         = "analyst"
	OrgRoleViewer          = "viewer"
	OrgRoleAutomationAdmin = "automation_admin"
	OrgRoleBillingAdmin    = "billing_admin"
)

var validOrgRoles = []string{
	OrgRoleOwner,
	OrgRoleAdmin,
	OrgRoleAnalyst,
	OrgRoleViewer,
	OrgRoleAutomationAdmin,
	OrgRoleBillingAdmin,
}

func ValidOrganizationRoles() []string {
	return append([]string{}, validOrgRoles...)
}

func IsValidOrganizationRole(role string) bool {
	for _, r := range validOrgRoles {
		if r == role {
			return true
		}
	}
	return false
}

func NewOrgMemberRole(value string) (OrgMemberRole, error) {
	if !IsValidOrganizationRole(value) {
		return OrgMemberRole{}, apperrors.BadException("invalid organization role: " + value)
	}
	return OrgMemberRole{value: value}, nil
}

func MustNewOrgMemberRole(value string) OrgMemberRole {
	role, err := NewOrgMemberRole(value)
	if err != nil {
		panic(err)
	}
	return role
}

func (r OrgMemberRole) String() string {
	return r.value
}

func (r OrgMemberRole) Value() (driver.Value, error) {
	return r.value, nil
}

func (r *OrgMemberRole) Scan(src interface{}) error {
	if src == nil {
		*r = OrgMemberRole{}
		return nil
	}
	switch s := src.(type) {
	case string:
		role, err := NewOrgMemberRole(s)
		if err != nil {
			return err
		}
		*r = role
		return nil
	case []byte:
		role, err := NewOrgMemberRole(string(s))
		if err != nil {
			return err
		}
		*r = role
		return nil
	default:
		return fmt.Errorf("cannot scan %T into OrgMemberRole", src)
	}
}

func (r OrgMemberRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}

func (r *OrgMemberRole) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	role, err := NewOrgMemberRole(s)
	if err != nil {
		return err
	}
	*r = role
	return nil
}
