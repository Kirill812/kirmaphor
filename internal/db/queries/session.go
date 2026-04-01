package queries

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func GenerateSessionToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	token = hex.EncodeToString(b)
	sum := sha256.Sum256(b)
	hash = hex.EncodeToString(sum[:])
	return
}

type CreateSessionParams struct {
	UserID            uuid.UUID
	TokenHash         string
	DeviceFingerprint *string
	IPAddress         *string
	UserAgent         *string
	ExpiresAt         time.Time
}

func CreateSession(ctx context.Context, pool *pgxpool.Pool, p CreateSessionParams) (*models.UserSession, error) {
	s := &models.UserSession{}
	err := pool.QueryRow(ctx,
		`INSERT INTO user_sessions
		   (user_id, session_token_hash, device_fingerprint, ip_address, user_agent, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, session_token_hash, device_fingerprint, device_label,
		           geo_city, geo_country, ip_address, user_agent, is_current,
		           secure_at, expires_at, created_at`,
		p.UserID, p.TokenHash, p.DeviceFingerprint, p.IPAddress, p.UserAgent, p.ExpiresAt,
	).Scan(&s.ID, &s.UserID, &s.SessionTokenHash, &s.DeviceFingerprint, &s.DeviceLabel,
		&s.GeoCity, &s.GeoCountry, &s.IPAddress, &s.UserAgent, &s.IsCurrent,
		&s.SecureAt, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func GetSessionByTokenHash(ctx context.Context, pool *pgxpool.Pool, hash string) (*models.UserSession, error) {
	s := &models.UserSession{}
	err := pool.QueryRow(ctx,
		`SELECT id, user_id, session_token_hash, device_fingerprint, device_label,
		        geo_city, geo_country, ip_address, user_agent, is_current,
		        secure_at, expires_at, created_at
		 FROM user_sessions
		 WHERE session_token_hash = $1 AND is_current = TRUE AND expires_at > NOW()`,
		hash,
	).Scan(&s.ID, &s.UserID, &s.SessionTokenHash, &s.DeviceFingerprint, &s.DeviceLabel,
		&s.GeoCity, &s.GeoCountry, &s.IPAddress, &s.UserAgent, &s.IsCurrent,
		&s.SecureAt, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func RevokeSession(ctx context.Context, pool *pgxpool.Pool, sessionID, userID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE user_sessions SET is_current = FALSE WHERE id = $1 AND user_id = $2`,
		sessionID, userID)
	return err
}

func UpdateSecureAt(ctx context.Context, pool *pgxpool.Pool, sessionID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE user_sessions SET secure_at = NOW() WHERE id = $1`, sessionID)
	return err
}
