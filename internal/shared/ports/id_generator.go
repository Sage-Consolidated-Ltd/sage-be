package ports

import "github.com/google/uuid"

type IdentifierGenerator interface {
	NewUUID() uuid.UUID
	NewUUIDString() string
}
