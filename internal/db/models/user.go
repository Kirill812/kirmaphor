package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID
	Email                 string
	DisplayName           string
	AvatarURL             *string
	PasswordHash          *string
	Onboarded             bool
	BlockedAt             *time.Time
	SessionTimeoutMinutes int
	Settings              map[string]any
	CreatedAt             time.Time
}
