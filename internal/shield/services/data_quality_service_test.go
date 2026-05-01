package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataQualityService_Struct(t *testing.T) {
	service := &DataQualityService{}
	assert.NotNil(t, service)
}
