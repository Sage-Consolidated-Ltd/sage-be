package domain

import "errors"

type Status string

const (
	StatusActive            Status = "active"
	StatusSuspended         Status = "suspended"
	StatusTemporarilyLocked Status = "temporarily_locked"
	StatusPermanentlyLocked Status = "permanently_locked"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusTemporarilyLocked, StatusPermanentlyLocked:
		return true
	default:
		return false
	}
}

func ParseStatus(val string) (Status, error) {
	s := Status(val)
	if !s.IsValid() {
		return "", errors.New("invalid admin status")
	}
	return s, nil
}
