package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"sage-backend/internal/shared/errors/apperrors"
	"strings"
)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Email{}, apperrors.BadException("email cannot be empty")
	}
	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}
	return Email{value: strings.ToLower(trimmed)}, nil
}

func MustNewEmail(value string) Email {
	email, err := NewEmail(value)
	if err != nil {
		panic(err)
	}
	return email
}

func (e Email) String() string {
	return e.value
}

func (e Email) IsZero() bool {
	return e.value == ""
}

func (e Email) Value() (driver.Value, error) {
	return e.value, nil
}

func (e *Email) Scan(src interface{}) error {
	if src == nil {
		*e = Email{}
		return nil
	}
	switch s := src.(type) {
	case string:
		email, err := NewEmail(s)
		if err != nil {
			return err
		}
		*e = email
		return nil
	case []byte:
		email, err := NewEmail(string(s))
		if err != nil {
			return err
		}
		*e = email
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Email", src)
	}
}

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

func (e *Email) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	email, err := NewEmail(s)
	if err != nil {
		return err
	}
	*e = email
	return nil
}
