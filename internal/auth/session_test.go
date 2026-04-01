package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func TestSecureSessionValid(t *testing.T) {
	now := time.Now()
	s := &models.UserSession{
		ID:        uuid.New(),
		SecureAt:  &now,
		ExpiresAt: now.Add(8 * time.Hour),
	}
	if err := auth.CheckSecureSession(s); err != nil {
		t.Fatalf("expected valid secure session, got: %v", err)
	}
}

func TestSecureSessionExpired(t *testing.T) {
	old := time.Now().Add(-6 * time.Minute)
	s := &models.UserSession{
		ID:        uuid.New(),
		SecureAt:  &old,
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}
	if err := auth.CheckSecureSession(s); err == nil {
		t.Fatal("expected error for expired secure session")
	}
}

func TestSecureSessionNil(t *testing.T) {
	s := &models.UserSession{
		ID:        uuid.New(),
		SecureAt:  nil,
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}
	if err := auth.CheckSecureSession(s); err == nil {
		t.Fatal("expected error when SecureAt is nil")
	}
}
