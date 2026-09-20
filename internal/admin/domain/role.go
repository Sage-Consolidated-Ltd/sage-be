package domain

import "errors"

type Role string

const (
	RoleSuperAdmin   Role = "super_admin"
	RoleSupportAdmin Role = "support_admin"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return r == RoleSuperAdmin || r == RoleSupportAdmin
}

func ParseRole(val string) (Role, error) {
	r := Role(val)
	if !r.IsValid() {
		return "", errors.New("invalid admin role")
	}
	return r, nil
}
