package jwt

import (
	"sage-backend/internal/shared/types"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id    string     `json:"user_id"`
	Email string     `json:"email"`
	Role  types.Role `json:"role"`
	jwt.RegisteredClaims
}
