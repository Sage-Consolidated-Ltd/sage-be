package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"sage-backend/internal/shared/errors/apperrors"
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Email{}, apperrors.BadException("email cannot be empty")
	}

	if len(trimmed) > 254 {
		return Email{}, apperrors.BadException("email exceeds maximum allowed length of 254 characters")
	}

	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}

	localPart, domainPart := parts[0], parts[1]
	if len(localPart) == 0 || len(localPart) > 64 {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}

	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") || strings.Contains(localPart, "..") {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}

	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") || strings.Contains(domainPart, "..") {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}

	if !emailPattern.MatchString(trimmed) {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return Email{}, apperrors.BadException("invalid email format: " + value)
	}

	return Email{value: strings.ToLower(trimmed)}, nil
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
		if strings.TrimSpace(s) == "" {
			*e = Email{}
			return nil
		}
		email, err := NewEmail(s)
		if err != nil {
			return err
		}
		*e = email
		return nil
	case []byte:
		str := string(s)
		if strings.TrimSpace(str) == "" {
			*e = Email{}
			return nil
		}
		email, err := NewEmail(str)
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
	if strings.TrimSpace(s) == "" {
		*e = Email{}
		return nil
	}
	email, err := NewEmail(s)
	if err != nil {
		return err
	}
	*e = email
	return nil
}
