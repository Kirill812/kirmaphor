package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateUser(ctx context.Context, pool *pgxpool.Pool, email, displayName string, passwordHash *string) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, display_name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, display_name, avatar_url, password_hash,
		           onboarded, blocked_at, session_timeout_minutes, settings, created_at`,
		email, displayName, passwordHash,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.PasswordHash,
		&u.Onboarded, &u.BlockedAt, &u.SessionTimeoutMinutes, &u.Settings, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, password_hash,
		        onboarded, blocked_at, session_timeout_minutes, settings, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.PasswordHash,
		&u.Onboarded, &u.BlockedAt, &u.SessionTimeoutMinutes, &u.Settings, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, password_hash,
		        onboarded, blocked_at, session_timeout_minutes, settings, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.PasswordHash,
		&u.Onboarded, &u.BlockedAt, &u.SessionTimeoutMinutes, &u.Settings, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
