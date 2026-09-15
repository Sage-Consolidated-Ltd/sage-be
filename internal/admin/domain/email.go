package domain

import (
	"net/mail"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Email{}, ErrEmptyEmail
	}

	if len(trimmed) > 254 {
		return Email{}, ErrInvalidEmail
	}

	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 {
		return Email{}, ErrInvalidEmail
	}

	localPart, domainPart := parts[0], parts[1]
	if len(localPart) == 0 || len(localPart) > 64 {
		return Email{}, ErrInvalidEmail
	}

	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") || strings.Contains(localPart, "..") {
		return Email{}, ErrInvalidEmail
	}

	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") || strings.Contains(domainPart, "..") {
		return Email{}, ErrInvalidEmail
	}

	if !emailPattern.MatchString(trimmed) {
		return Email{}, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: strings.ToLower(trimmed)}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) IsZero() bool {
	return e.value == ""
}

func (e Email) Equals(other Email) bool {
	return e.value == other.value
}
