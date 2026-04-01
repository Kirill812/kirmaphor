package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreatePasskeyCredential(ctx context.Context, pool *pgxpool.Pool,
	userID uuid.UUID, credID, pubKey []byte, transports []string, deviceName string,
) (*models.PasskeyCredential, error) {
	c := &models.PasskeyCredential{}
	err := pool.QueryRow(ctx,
		`INSERT INTO passkey_credentials (user_id, credential_id, public_key, transports, device_name)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, credential_id, public_key, counter, transports, device_name, last_used_at, created_at`,
		userID, credID, pubKey, transports, deviceName,
	).Scan(&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.Counter,
		&c.Transports, &c.DeviceName, &c.LastUsedAt, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetPasskeyCredentialsByUserID(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]*models.PasskeyCredential, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, user_id, credential_id, public_key, counter, transports, device_name, last_used_at, created_at
		 FROM passkey_credentials WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []*models.PasskeyCredential
	for rows.Next() {
		c := &models.PasskeyCredential{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.Counter,
			&c.Transports, &c.DeviceName, &c.LastUsedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

func UpdatePasskeyCounter(ctx context.Context, pool *pgxpool.Pool, credID []byte, counter uint32) error {
	_, err := pool.Exec(ctx,
		`UPDATE passkey_credentials SET counter = $1, last_used_at = NOW() WHERE credential_id = $2`,
		counter, credID)
	return err
}
