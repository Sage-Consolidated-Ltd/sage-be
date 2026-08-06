package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProfileHandler(t *testing.T) {
	handler := &ProfileHandler{}
	assert.NotNil(t, handler)
}

func TestProfileHandler_Struct(t *testing.T) {
	// Test that the handler struct is properly defined
	handler := NewProfileHandler(nil)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.userServ)
}
