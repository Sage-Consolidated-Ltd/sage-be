package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserService_Struct(t *testing.T) {
	service := &ParserService{}
	assert.NotNil(t, service)
}
