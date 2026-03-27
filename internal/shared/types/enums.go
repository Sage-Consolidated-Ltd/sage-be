package types

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case AdminRole, UserRole:
		return true
	default:
		return false
	}
}
