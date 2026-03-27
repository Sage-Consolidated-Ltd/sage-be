package jwt

import (
	"errors"
	"log"
	"sage-backend/internal/shared/types"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserPayload struct {
	Id    string     `json:"user_id"`
	Email string     `json:"email"`
	Role  types.Role `json:"role"`
}

type JwtService struct {
	JwtSecret string
}

func (j *JwtService) GenerateToken(userPayload UserPayload) (string, error) {
	claims := Claims{
		Id:    userPayload.Id,
		Email: userPayload.Email,
		Role:  userPayload.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenValue, err := token.SignedString([]byte(j.JwtSecret))
	if err != nil {
		log.Println("Error sigining token: ", err)
		return "", errors.New("Error signing token")
	}

	return tokenValue, nil
}

func (j *JwtService) VerifyToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.JwtSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
