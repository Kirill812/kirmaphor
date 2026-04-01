package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	SessionTokenHash  string
	DeviceFingerprint *string
	DeviceLabel       *string
	GeoCity           *string
	GeoCountry        *string
	IPAddress         *string
	UserAgent         *string
	IsCurrent         bool
	SecureAt          *time.Time
	ExpiresAt         time.Time
	CreatedAt         time.Time
}
