package auth

import (
	"errors"
	"time"

	"github.com/kgory/kirmaphor/internal/db/models"
)

var ErrSecureSessionRequired = errors.New("secure_session_required")

const secureSessionWindow = 5 * time.Minute

// CheckSecureSession returns ErrSecureSessionRequired if the session
// has not been re-authenticated within the last 5 minutes.
// Use for destructive/sensitive operations: secret rotation, role changes,
// project deletion, cloud credential update.
func CheckSecureSession(s *models.UserSession) error {
	if s.SecureAt == nil {
		return ErrSecureSessionRequired
	}
	if time.Since(*s.SecureAt) > secureSessionWindow {
		return ErrSecureSessionRequired
	}
	return nil
}
