package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
	"unicode"
)

type Password struct {
	value string
}

func NewPassword(value string) (Password, error) {
	if len(value) < 8 {
		return Password{}, apperrors.BadException("password must be at least 8 characters long")
	}
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)
	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return Password{}, apperrors.BadException("password must contain uppercase, lowercase, digit, and special character")
	}
	return Password{value: value}, nil
}

func (p Password) String() string {
	return p.value
}

type PasswordHash struct {
	value string
}

func NewPasswordHash(hash string) PasswordHash {
	return PasswordHash{value: hash}
}

func (ph PasswordHash) String() string {
	return ph.value
}

func (ph PasswordHash) IsZero() bool {
	return ph.value == ""
}

func (ph PasswordHash) Value() (driver.Value, error) {
	return ph.value, nil
}

func (ph *PasswordHash) Scan(src interface{}) error {
	if src == nil {
		*ph = PasswordHash{}
		return nil
	}
	switch s := src.(type) {
	case string:
		*ph = PasswordHash{value: s}
		return nil
	case []byte:
		*ph = PasswordHash{value: string(s)}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into PasswordHash", src)
	}
}

func (ph PasswordHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(ph.value)
}

func (ph *PasswordHash) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*ph = PasswordHash{value: s}
	return nil
}
