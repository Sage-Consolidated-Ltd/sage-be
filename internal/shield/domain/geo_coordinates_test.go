package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveCountryCoordinates(t *testing.T) {
	tests := []struct {
		input   string
		wantLat float64
		wantLng float64
	}{
		{"Russia", 55.7558, 37.6173},
		{"RU", 55.7558, 37.6173},
		{"China", 39.9042, 116.4074},
		{"cn", 39.9042, 116.4074},
		{"North Korea", 39.0392, 125.7625},
		{"kp", 39.0392, 125.7625},
		{"United States", 37.0902, -95.7129},
		{"USA", 37.0902, -95.7129},
		{"UnknownCountryXYZ", 20.0, 0.0},
	}

	for _, tt := range tests {
		lat, lng := ResolveCountryCoordinates(tt.input)
		assert.InDelta(t, tt.wantLat, lat, 0.0001, "lat for %s", tt.input)
		assert.InDelta(t, tt.wantLng, lng, 0.0001, "lng for %s", tt.input)
	}
}
