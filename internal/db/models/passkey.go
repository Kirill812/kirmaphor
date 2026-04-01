package models

import (
	"time"

	"github.com/google/uuid"
)

type PasskeyCredential struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CredentialID []byte
	PublicKey    []byte
	Counter      uint32
	Transports   []string
	DeviceName   string
	LastUsedAt   *time.Time
	CreatedAt    time.Time
}
