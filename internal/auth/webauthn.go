package auth

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/kgory/kirmaphor/internal/config"
)

// NewWebAuthn creates a configured WebAuthn instance from app config.
func NewWebAuthn(cfg *config.Config) (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	})
}
