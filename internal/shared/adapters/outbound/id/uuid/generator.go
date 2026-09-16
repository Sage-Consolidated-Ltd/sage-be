package uuid

import (
	"github.com/google/uuid"
	"sage-backend/internal/shared/ports"
)

type Generator struct{}

func NewGenerator() ports.IdentifierGenerator {
	return &Generator{}
}

func (g *Generator) NewUUID() uuid.UUID {
	return uuid.New()
}

func (g *Generator) NewUUIDString() string {
	return uuid.New().String()
}
