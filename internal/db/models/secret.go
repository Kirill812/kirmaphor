package models

import (
	"time"

	"github.com/google/uuid"
)

type Secret struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	Type           string
	EncryptedValue []byte
	Nonce          []byte
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
