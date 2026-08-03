package clock

import (
	"time"
	"sage-backend/internal/shared/ports"
)

type SystemClock struct{}

func NewSystemClock() ports.Clock {
	return &SystemClock{}
}

func (c *SystemClock) Now() time.Time {
	return time.Now()
}
