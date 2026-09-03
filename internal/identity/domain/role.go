package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
)

// UserRole represents system-level user roles ("admin", "user")
type UserRole struct {
	value string
}

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

var validUserRoles = []string{
	UserRoleAdmin,
	UserRoleUser,
}

func NewUserRole(value string) (UserRole, error) {
	if value == "" {
		value = UserRoleUser
	}
	for _, r := range validUserRoles {
		if r == value {
			return UserRole{value: value}, nil
		}
	}
	return UserRole{}, apperrors.BadException("invalid user role: " + value)
}

func MustNewUserRole(value string) UserRole {
	role, err := NewUserRole(value)
	if err != nil {
		panic(err)
	}
	return role
}

func (r UserRole) String() string {
	return r.value
}

func (r UserRole) Value() (driver.Value, error) {
	return r.value, nil
}

func (r *UserRole) Scan(src interface{}) error {
	if src == nil {
		*r = UserRole{}
		return nil
	}
	switch s := src.(type) {
	case string:
		role, err := NewUserRole(s)
		if err != nil {
			return err
		}
		*r = role
		return nil
	case []byte:
		role, err := NewUserRole(string(s))
		if err != nil {
			return err
		}
		*r = role
		return nil
	default:
		return fmt.Errorf("cannot scan %T into UserRole", src)
	}
}

func (r UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}

func (r *UserRole) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	role, err := NewUserRole(s)
	if err != nil {
		return err
	}
	*r = role
	return nil
}
