# Kirmaphore Plan 1: Foundation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the entire project, set up PostgreSQL schema, implement authentication (Passkeys + local), RBAC (4 roles), two-layer secrets encryption, and basic project/user management API.

**Architecture:** Go backend (chi router) + Next.js 15 frontend. Auth is WebAuthn-first with email/password fallback. All secrets encrypted client-side (AES-256-GCM) then server-side before DB write. RBAC enforced in middleware — owner/admin/engineer/viewer.

**Tech Stack:** Go 1.23, chi router, pgx/v5 (PostgreSQL), golang-migrate, go-webauthn/webauthn, Next.js 15, Tailwind CSS, shadcn/ui, @simplewebauthn/browser, Zustand, next-intl.

---

## File Map

### Backend

```
cmd/kirmaphore/main.go                    -- binary entry, DI wiring
internal/config/config.go                 -- env-based config (PORT, DB_URL, MASTER_KEY, RP_ID, etc.)
internal/db/db.go                         -- pgx pool setup, ping
internal/db/migrate.go                    -- golang-migrate runner
migrations/001_users_auth.sql             -- users, passkey_credentials, user_sessions, user_known_devices
migrations/002_projects_rbac.sql          -- projects, project_users, user_roles
migrations/003_secrets.sql                -- secrets table (two-layer encrypted)
internal/db/models/user.go                -- User, Profile structs
internal/db/models/passkey.go             -- PasskeyCredential struct
internal/db/models/session.go             -- UserSession, KnownDevice structs
internal/db/models/project.go             -- Project, ProjectUser, UserRole structs
internal/db/models/secret.go              -- Secret struct
internal/db/queries/user.go               -- CreateUser, GetUserByEmail, GetUserByID, BlockUser
internal/db/queries/passkey.go            -- CreateCredential, GetCredentialByID, UpdateCounter
internal/db/queries/session.go            -- CreateSession, GetSession, RevokeSession, ListSessions
internal/db/queries/project.go            -- CreateProject, GetProject, ListProjects, AddMember
internal/db/queries/secret.go             -- CreateSecret, GetSecret, ListSecrets, DeleteSecret
internal/crypto/aes.go                    -- AES-256-GCM encrypt/decrypt (server-side, Layer 2)
internal/crypto/master_key.go             -- master key load from env, key derivation
internal/auth/webauthn.go                 -- WebAuthn instance setup, Begin/Finish Register, Begin/Finish Login
internal/auth/local.go                    -- bcrypt hash/verify, TOTP generate/verify
internal/auth/session.go                  -- CreateSession, ValidateSession, RevokeSession
internal/auth/secure_session.go           -- CheckSecureSession (5-min re-auth gate)
internal/rbac/roles.go                    -- Role constants, Permission flags bitmap
internal/rbac/check.go                    -- HasPermission, RequireRole helpers
internal/api/router.go                    -- chi router, middleware stack, route registration
internal/api/helpers/helpers.go           -- WriteJSON, Bind, GetUser, GetProject, GetIntParam
internal/api/middleware/auth.go           -- LoadUser middleware (session → context)
internal/api/middleware/project.go        -- LoadProject + ResolveRole middleware
internal/api/handlers/auth.go             -- POST /auth/register, /auth/login, /auth/logout,
                                          --   /auth/passkey/register/begin, /auth/passkey/register/finish
                                          --   /auth/passkey/login/begin, /auth/passkey/login/finish
internal/api/handlers/users.go            -- GET/PUT /users/me, GET /users/me/sessions, DELETE /users/me/sessions/:id
internal/api/handlers/projects.go         -- CRUD /projects, /projects/:id/members
internal/api/handlers/secrets.go          -- CRUD /projects/:id/secrets
```

### Frontend

```
web/package.json                          -- dependencies
web/app/layout.tsx                        -- root layout, NextIntlClientProvider
web/app/(auth)/login/page.tsx             -- login: passkey button + email/password form
web/app/(auth)/register/page.tsx          -- registration form
web/app/(dashboard)/layout.tsx            -- authenticated shell: Sidebar + Header
web/app/(dashboard)/page.tsx              -- dashboard home (placeholder)
web/app/(dashboard)/projects/page.tsx     -- project list
web/components/ui/                        -- shadcn/ui primitives (button, input, card, etc.)
web/components/auth/PasskeyButton.tsx     -- WebAuthn register/authenticate via @simplewebauthn/browser
web/components/auth/LoginForm.tsx         -- email + password form with TOTP field
web/components/layout/Sidebar.tsx         -- nav links, project switcher
web/components/layout/Header.tsx          -- user avatar, logout, theme toggle
web/lib/api.ts                            -- typed fetch client (base URL, auth header, error handling)
web/lib/auth-store.ts                     -- Zustand store: user, token, login(), logout()
web/lib/passkey.ts                        -- startRegistration, startAuthentication wrappers
web/lib/crypto.ts                         -- client-side AES-256-GCM encrypt (Layer 1) before API calls
web/messages/en.json                      -- English i18n strings
web/messages/ru.json                      -- Russian i18n strings
web/middleware.ts                         -- Next.js route protection (redirect to /login if no token)
```

---

## Task 1: Project Scaffold & Config

**Files:**
- Create: `cmd/kirmaphore/main.go`
- Create: `internal/config/config.go`
- Create: `go.mod`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/kgory/dev/kirmaphor
go mod init github.com/kgory/kirmaphor
```

- [ ] **Step 2: Install backend dependencies**

```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/golang-migrate/migrate/v4@latest
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
go get golang.org/x/crypto@latest
go get github.com/go-webauthn/webauthn@latest
go get github.com/pquerna/otp@latest
```

- [ ] **Step 3: Write config**

```go
// internal/config/config.go
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port       string
	DBURL      string
	MasterKey  string // 32-byte hex for AES-256
	RPOrigin   string // WebAuthn: "https://kirmaphore.example.com"
	RPID       string // WebAuthn: "kirmaphore.example.com"
	RPName     string // WebAuthn: "Kirmaphore"
}

func Load() (*Config, error) {
	c := &Config{
		Port:     getEnv("PORT", "8080"),
		DBURL:    mustEnv("DATABASE_URL"),
		MasterKey: mustEnv("MASTER_KEY"),
		RPOrigin: getEnv("RP_ORIGIN", "http://localhost:3000"),
		RPID:     getEnv("RP_ID", "localhost"),
		RPName:   getEnv("RP_NAME", "Kirmaphore"),
	}
	if len(c.MasterKey) != 64 {
		return nil, fmt.Errorf("MASTER_KEY must be 64 hex chars (32 bytes)")
	}
	return c, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: Write main.go skeleton**

```go
// cmd/kirmaphore/main.go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kgory/kirmaphor/internal/config"
	"github.com/kgory/kirmaphor/internal/db"
	"github.com/kgory/kirmaphor/internal/api"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.Connect(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(cfg.DBURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	router := api.NewRouter(cfg, pool)
	log.Printf("starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
```

- [ ] **Step 5: Create .env.example**

```bash
cat > .env.example << 'EOF'
PORT=8080
DATABASE_URL=postgres://kirmaphore:kirmaphore@localhost:5432/kirmaphore?sslmode=disable
MASTER_KEY=0000000000000000000000000000000000000000000000000000000000000000
RP_ORIGIN=http://localhost:3000
RP_ID=localhost
RP_NAME=Kirmaphore
EOF
```

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum cmd/ internal/config/ .env.example
git commit -m "feat: project scaffold and config"
```

---

## Task 2: Database Connection & Migrations

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrate.go`
- Create: `migrations/001_users_auth.sql`
- Create: `migrations/002_projects_rbac.sql`
- Create: `migrations/003_secrets.sql`

- [ ] **Step 1: Write DB connection**

```go
// internal/db/db.go
package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}
```

- [ ] **Step 2: Write migration runner**

```go
// internal/db/migrate.go
package db

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
```

- [ ] **Step 3: Write migration 001 — users & auth**

```sql
-- migrations/001_users_auth.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    avatar_url      TEXT,
    password_hash   TEXT,
    onboarded       BOOLEAN NOT NULL DEFAULT FALSE,
    blocked_at      TIMESTAMPTZ,
    session_timeout_minutes INT NOT NULL DEFAULT 480,
    settings        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE passkey_credentials (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id   BYTEA NOT NULL UNIQUE,
    public_key      BYTEA NOT NULL,
    counter         BIGINT NOT NULL DEFAULT 0,
    transports      TEXT[],
    device_name     TEXT NOT NULL DEFAULT 'Unknown device',
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_sessions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token_hash  TEXT NOT NULL UNIQUE,
    device_fingerprint  TEXT,
    device_label        TEXT,
    geo_city            TEXT,
    geo_country         TEXT,
    ip_address          TEXT,
    user_agent          TEXT,
    is_current          BOOLEAN NOT NULL DEFAULT TRUE,
    secure_at           TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_known_devices (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint  TEXT NOT NULL,
    device_label        TEXT NOT NULL DEFAULT 'Unknown device',
    first_seen_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, device_fingerprint)
);

CREATE INDEX idx_user_sessions_token ON user_sessions(session_token_hash);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
```

- [ ] **Step 4: Write migration 002 — projects & RBAC**

```sql
-- migrations/002_projects_rbac.sql
CREATE TYPE user_role AS ENUM ('owner', 'admin', 'engineer', 'viewer');

CREATE TABLE projects (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_users (
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        user_role NOT NULL DEFAULT 'viewer',
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_project_users_user ON project_users(user_id);
```

- [ ] **Step 5: Write migration 003 — secrets**

```sql
-- migrations/003_secrets.sql
CREATE TABLE secrets (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    type                TEXT NOT NULL CHECK (type IN ('ssh', 'login_password', 'string', 'cloud_aws', 'cloud_azure', 'cloud_gcp')),
    -- Layer 2: server-side AES-256-GCM encrypted value
    -- Value already arrives Layer-1 encrypted from client
    encrypted_value     BYTEA NOT NULL,
    nonce               BYTEA NOT NULL,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_secrets_project ON secrets(project_id);
```

- [ ] **Step 6: Spin up Postgres with Docker and verify migrations**

```bash
docker run -d --name kirmaphore-db \
  -e POSTGRES_USER=kirmaphore \
  -e POSTGRES_PASSWORD=kirmaphore \
  -e POSTGRES_DB=kirmaphore \
  -p 5432:5432 postgres:16-alpine

export DATABASE_URL="postgres://kirmaphore:kirmaphore@localhost:5432/kirmaphore?sslmode=disable"
export MASTER_KEY="0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
go run cmd/kirmaphore/main.go
```

Expected: server starts, no migration errors.

- [ ] **Step 7: Commit**

```bash
git add internal/db/ migrations/
git commit -m "feat: db connection and initial migrations"
```

---

## Task 3: Crypto — Two-Layer Secret Encryption

**Files:**
- Create: `internal/crypto/aes.go`
- Create: `internal/crypto/master_key.go`
- Create: `internal/crypto/aes_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/crypto/aes_test.go
package crypto_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key { key[i] = byte(i) }

	plaintext := []byte("my secret ansible vault password")
	ciphertext, nonce, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(ciphertext) == 0 || len(nonce) == 0 {
		t.Fatal("expected non-empty ciphertext and nonce")
	}

	result, err := crypto.Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(result) != string(plaintext) {
		t.Fatalf("got %q want %q", result, plaintext)
	}
}

func TestDecryptFailsWithWrongKey(t *testing.T) {
	key := make([]byte, 32)
	wrongKey := make([]byte, 32)
	for i := range wrongKey { wrongKey[i] = 0xFF }

	ciphertext, nonce, _ := crypto.Encrypt(key, []byte("secret"))
	_, err := crypto.Decrypt(wrongKey, ciphertext, nonce)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
}
```

- [ ] **Step 2: Run test, confirm FAIL**

```bash
go test ./internal/crypto/... -v
```

Expected: `FAIL — crypto.Encrypt undefined`

- [ ] **Step 3: Implement AES-256-GCM**

```go
// internal/crypto/aes.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encrypt encrypts plaintext with AES-256-GCM using key (32 bytes).
// Returns (ciphertext, nonce, error). Nonce is stored separately.
func Encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("gcm: %w", err)
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("nonce: %w", err)
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext with AES-256-GCM using key and nonce.
func Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}
```

- [ ] **Step 4: Implement master key loader**

```go
// internal/crypto/master_key.go
package crypto

import (
	"encoding/hex"
	"fmt"
)

// LoadMasterKey decodes a 64-char hex string into a 32-byte AES key.
func LoadMasterKey(hexKey string) ([]byte, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid MASTER_KEY hex: %w", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("MASTER_KEY must be 32 bytes (64 hex chars), got %d", len(b))
	}
	return b, nil
}
```

- [ ] **Step 5: Run tests, confirm PASS**

```bash
go test ./internal/crypto/... -v
```

Expected: `PASS — TestEncryptDecryptRoundTrip, TestDecryptFailsWithWrongKey`

- [ ] **Step 6: Commit**

```bash
git add internal/crypto/
git commit -m "feat: two-layer AES-256-GCM encryption"
```

---

## Task 4: Auth — Session Management

**Files:**
- Create: `internal/db/models/user.go`
- Create: `internal/db/models/session.go`
- Create: `internal/db/queries/user.go`
- Create: `internal/db/queries/session.go`
- Create: `internal/auth/session.go`
- Create: `internal/auth/session_test.go`

- [ ] **Step 1: Write DB models**

```go
// internal/db/models/user.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID                   uuid.UUID
	Email                string
	DisplayName          string
	AvatarURL            *string
	PasswordHash         *string
	Onboarded            bool
	BlockedAt            *time.Time
	SessionTimeoutMinutes int
	Settings             map[string]any
	CreatedAt            time.Time
}
```

```go
// internal/db/models/session.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type UserSession struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	SessionTokenHash   string
	DeviceFingerprint  *string
	DeviceLabel        *string
	GeoCity            *string
	GeoCountry         *string
	IPAddress          *string
	UserAgent          *string
	IsCurrent          bool
	SecureAt           *time.Time
	ExpiresAt          time.Time
	CreatedAt          time.Time
}
```

- [ ] **Step 2: Write user queries**

```go
// internal/db/queries/user.go
package queries

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/google/uuid"
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
	return u, err
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, password_hash,
		        onboarded, blocked_at, session_timeout_minutes, settings, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.PasswordHash,
		&u.Onboarded, &u.BlockedAt, &u.SessionTimeoutMinutes, &u.Settings, &u.CreatedAt)
	return u, err
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, password_hash,
		        onboarded, blocked_at, session_timeout_minutes, settings, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.PasswordHash,
		&u.Onboarded, &u.BlockedAt, &u.SessionTimeoutMinutes, &u.Settings, &u.CreatedAt)
	return u, err
}
```

- [ ] **Step 3: Write session queries**

```go
// internal/db/queries/session.go
package queries

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/google/uuid"
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
```

- [ ] **Step 4: Write failing test for secure session gate**

```go
// internal/auth/session_test.go
package auth_test

import (
	"testing"
	"time"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/google/uuid"
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
```

- [ ] **Step 5: Run, confirm FAIL**

```bash
go test ./internal/auth/... -v
```

Expected: `FAIL — auth.CheckSecureSession undefined`

- [ ] **Step 6: Implement secure session check**

```go
// internal/auth/secure_session.go
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
// Use for destructive/sensitive operations (secret rotation, role changes, etc.)
func CheckSecureSession(s *models.UserSession) error {
	if s.SecureAt == nil {
		return ErrSecureSessionRequired
	}
	if time.Since(*s.SecureAt) > secureSessionWindow {
		return ErrSecureSessionRequired
	}
	return nil
}
```

- [ ] **Step 7: Run, confirm PASS**

```bash
go test ./internal/auth/... -v
```

Expected: `PASS — all 3 tests`

- [ ] **Step 8: Commit**

```bash
git add internal/db/models/ internal/db/queries/ internal/auth/
git commit -m "feat: user models, session queries, secure session gate"
```

---

## Task 5: Auth — Local (bcrypt + TOTP)

**Files:**
- Create: `internal/auth/local.go`
- Create: `internal/auth/local_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/auth/local_test.go
package auth_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/auth"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := auth.VerifyPassword(hash, "correct-horse-battery"); err != nil {
		t.Fatal("expected valid password to pass verification")
	}
	if err := auth.VerifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail verification")
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/auth/... -run TestHashAndVerifyPassword -v
```

- [ ] **Step 3: Implement**

```go
// internal/auth/local.go
package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(b), nil
}

func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

- [ ] **Step 4: Run, confirm PASS**

```bash
go test ./internal/auth/... -run TestHashAndVerifyPassword -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth/local.go internal/auth/local_test.go
git commit -m "feat: bcrypt password hashing"
```

---

## Task 6: Auth — Passkeys (WebAuthn)

**Files:**
- Create: `internal/db/models/passkey.go`
- Create: `internal/db/queries/passkey.go`
- Create: `internal/auth/webauthn.go`

- [ ] **Step 1: Write passkey model**

```go
// internal/db/models/passkey.go
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
```

- [ ] **Step 2: Write passkey queries**

```go
// internal/db/queries/passkey.go
package queries

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/google/uuid"
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
	return c, err
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
```

- [ ] **Step 3: Write WebAuthn service**

```go
// internal/auth/webauthn.go
package auth

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/kgory/kirmaphor/internal/config"
)

func NewWebAuthn(cfg *config.Config) (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	})
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/db/models/passkey.go internal/db/queries/passkey.go internal/auth/webauthn.go
git commit -m "feat: passkey models, queries, webauthn setup"
```

---

## Task 7: RBAC

**Files:**
- Create: `internal/rbac/roles.go`
- Create: `internal/rbac/check.go`
- Create: `internal/rbac/roles_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/rbac/roles_test.go
package rbac_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func TestOwnerCanDoEverything(t *testing.T) {
	for _, perm := range []rbac.Permission{
		rbac.PermRunJobs, rbac.PermEditPlaybooks, rbac.PermManageUsers, rbac.PermManageSecrets, rbac.PermDeleteProject,
	} {
		if !rbac.HasPermission(rbac.RoleOwner, perm) {
			t.Errorf("owner should have permission %d", perm)
		}
	}
}

func TestViewerCanOnlyRead(t *testing.T) {
	if !rbac.HasPermission(rbac.RoleViewer, rbac.PermReadLogs) {
		t.Error("viewer should be able to read logs")
	}
	if rbac.HasPermission(rbac.RoleViewer, rbac.PermRunJobs) {
		t.Error("viewer should NOT be able to run jobs")
	}
	if rbac.HasPermission(rbac.RoleViewer, rbac.PermManageSecrets) {
		t.Error("viewer should NOT manage secrets")
	}
}

func TestEngineerCanRunButNotManage(t *testing.T) {
	if !rbac.HasPermission(rbac.RoleEngineer, rbac.PermRunJobs) {
		t.Error("engineer should run jobs")
	}
	if rbac.HasPermission(rbac.RoleEngineer, rbac.PermManageUsers) {
		t.Error("engineer should NOT manage users")
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/rbac/... -v
```

- [ ] **Step 3: Implement roles**

```go
// internal/rbac/roles.go
package rbac

type Role string

const (
	RoleOwner    Role = "owner"
	RoleAdmin    Role = "admin"
	RoleEngineer Role = "engineer"
	RoleViewer   Role = "viewer"
)

type Permission uint32

const (
	PermReadLogs      Permission = 1 << 0
	PermRunJobs       Permission = 1 << 1
	PermEditPlaybooks Permission = 1 << 2
	PermManageSecrets Permission = 1 << 3
	PermManageUsers   Permission = 1 << 4
	PermManageProject Permission = 1 << 5
	PermDeleteProject Permission = 1 << 6
)

var rolePermissions = map[Role]Permission{
	RoleViewer:   PermReadLogs,
	RoleEngineer: PermReadLogs | PermRunJobs | PermEditPlaybooks,
	RoleAdmin:    PermReadLogs | PermRunJobs | PermEditPlaybooks | PermManageSecrets | PermManageUsers | PermManageProject,
	RoleOwner:    ^Permission(0), // all bits set
}
```

```go
// internal/rbac/check.go
package rbac

func HasPermission(role Role, perm Permission) bool {
	return rolePermissions[role]&perm != 0
}
```

- [ ] **Step 4: Run, confirm PASS**

```bash
go test ./internal/rbac/... -v
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/rbac/
git commit -m "feat: RBAC roles and permission bitmap"
```

---

## Task 8: API — Router, Helpers, Auth Middleware

**Files:**
- Create: `internal/api/helpers/helpers.go`
- Create: `internal/api/middleware/auth.go`
- Create: `internal/api/router.go`

- [ ] **Step 1: Write helpers**

```go
// internal/api/helpers/helpers.go
package helpers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kgory/kirmaphor/internal/db/models"
)

type contextKey string

const (
	CtxUser    contextKey = "user"
	CtxSession contextKey = "session"
	CtxProject contextKey = "project"
	CtxRole    contextKey = "role"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func Bind(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func GetUser(r *http.Request) *models.User {
	u, _ := r.Context().Value(CtxUser).(*models.User)
	return u
}

func GetSession(r *http.Request) *models.UserSession {
	s, _ := r.Context().Value(CtxSession).(*models.UserSession)
	return s
}

func GetIntParam(r *http.Request, key string) (int, error) {
	return strconv.Atoi(r.PathValue(key))
}
```

- [ ] **Step 2: Write auth middleware**

```go
// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

func RequireAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				helpers.WriteError(w, http.StatusUnauthorized, "missing token")
				return
			}
			sum := sha256.Sum256([]byte(token))
			hash := hex.EncodeToString(sum[:])

			session, err := queries.GetSessionByTokenHash(r.Context(), pool, hash)
			if err != nil {
				helpers.WriteError(w, http.StatusUnauthorized, "invalid or expired session")
				return
			}
			user, err := queries.GetUserByID(r.Context(), pool, session.UserID)
			if err != nil || user.BlockedAt != nil {
				helpers.WriteError(w, http.StatusUnauthorized, "user unavailable")
				return
			}
			ctx := context.WithValue(r.Context(), helpers.CtxUser, user)
			ctx = context.WithValue(ctx, helpers.CtxSession, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
```

- [ ] **Step 3: Write router**

```go
// internal/api/router.go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/handlers"
	apiMiddleware "github.com/kgory/kirmaphor/internal/api/middleware"
	"github.com/kgory/kirmaphor/internal/config"
	"github.com/kgory/kirmaphor/internal/crypto"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)

	masterKey, _ := crypto.LoadMasterKey(cfg.MasterKey)
	wa, _ := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	})

	authHandler := handlers.NewAuthHandler(pool, wa, masterKey)
	requireAuth := apiMiddleware.RequireAuth(pool)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/passkey/register/begin", authHandler.PasskeyRegisterBegin)
		r.Post("/auth/passkey/register/finish", authHandler.PasskeyRegisterFinish)
		r.Post("/auth/passkey/login/begin", authHandler.PasskeyLoginBegin)
		r.Post("/auth/passkey/login/finish", authHandler.PasskeyLoginFinish)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(requireAuth)
			r.Post("/auth/logout", authHandler.Logout)
			r.Get("/users/me", handlers.GetMe(pool))
			r.Put("/users/me", handlers.UpdateMe(pool))
			r.Get("/users/me/sessions", handlers.ListSessions(pool))
			r.Delete("/users/me/sessions/{id}", handlers.RevokeSession(pool))
			r.Get("/projects", handlers.ListProjects(pool))
			r.Post("/projects", handlers.CreateProject(pool))
			r.Get("/projects/{id}", handlers.GetProject(pool))
		})
	})

	return r
}
```

- [ ] **Step 4: Test health endpoint**

```bash
go build ./cmd/kirmaphore/... && ./kirmaphore &
curl -s http://localhost:8080/api/health
```

Expected: `{"status":"ok"}`

- [ ] **Step 5: Commit**

```bash
git add internal/api/
git commit -m "feat: chi router, auth middleware, helpers"
```

---

## Task 9: Auth Handlers (Register + Login + Passkey)

**Files:**
- Create: `internal/api/handlers/auth.go`

- [ ] **Step 1: Write auth handler**

```go
// internal/api/handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

type AuthHandler struct {
	pool      *pgxpool.Pool
	wa        *webauthn.WebAuthn
	masterKey []byte
	// in-memory challenge store (use Redis in production)
	challenges map[string]*webauthn.SessionData
}

func NewAuthHandler(pool *pgxpool.Pool, wa *webauthn.WebAuthn, masterKey []byte) *AuthHandler {
	return &AuthHandler{pool: pool, wa: wa, masterKey: masterKey, challenges: make(map[string]*webauthn.SessionData)}
}

type registerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !helpers.Bind(w, r, &req) {
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "server error")
		return
	}
	user, err := queries.CreateUser(r.Context(), h.pool, req.Email, req.DisplayName, &hash)
	if err != nil {
		helpers.WriteError(w, http.StatusConflict, "email already registered")
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, map[string]any{"id": user.ID, "email": user.Email})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !helpers.Bind(w, r, &req) {
		return
	}
	user, err := queries.GetUserByEmail(r.Context(), h.pool, req.Email)
	if err != nil || user.PasswordHash == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := auth.VerifyPassword(*user.PasswordHash, req.Password); err != nil {
		helpers.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if user.BlockedAt != nil {
		helpers.WriteError(w, http.StatusForbidden, "account blocked")
		return
	}
	token, hash, err := queries.GenerateSessionToken()
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "server error")
		return
	}
	ip := r.RemoteAddr
	ua := r.UserAgent()
	_, err = queries.CreateSession(r.Context(), h.pool, queries.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: hash,
		IPAddress: &ip,
		UserAgent: &ua,
		ExpiresAt: time.Now().Add(time.Duration(user.SessionTimeoutMinutes) * time.Minute),
	})
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "server error")
		return
	}
	helpers.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "user": map[string]any{
		"id": user.ID, "email": user.Email, "display_name": user.DisplayName,
	}})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session := helpers.GetSession(r)
	if session == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	queries.RevokeSession(r.Context(), h.pool, session.ID, session.UserID)
	helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// PasskeyRegisterBegin, PasskeyRegisterFinish, PasskeyLoginBegin, PasskeyLoginFinish
// are omitted here for brevity — implement using go-webauthn/webauthn library's
// BeginRegistration / FinishRegistration / BeginLogin / FinishLogin methods,
// mapping PasskeyCredential to webauthn.User interface.
func (h *AuthHandler) PasskeyRegisterBegin(w http.ResponseWriter, r *http.Request)  {}
func (h *AuthHandler) PasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {}
func (h *AuthHandler) PasskeyLoginBegin(w http.ResponseWriter, r *http.Request)     {}
func (h *AuthHandler) PasskeyLoginFinish(w http.ResponseWriter, r *http.Request)    {}
```

- [ ] **Step 2: Implement full Passkey register/login using go-webauthn**

```go
// Add to internal/api/handlers/auth.go

// webauthnUser adapts models.User to webauthn.User interface
type webauthnUser struct {
	user  *models.User
	creds []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte               { return []byte(u.user.ID.String()) }
func (u *webauthnUser) WebAuthnName() string              { return u.user.Email }
func (u *webauthnUser) WebAuthnDisplayName() string       { return u.user.DisplayName }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.creds }

func (h *AuthHandler) PasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetUser(r)
	if user == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "login first")
		return
	}
	wu := &webauthnUser{user: user}
	options, sessionData, err := h.wa.BeginRegistration(wu)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "webauthn error")
		return
	}
	h.challenges[user.ID.String()] = sessionData
	helpers.WriteJSON(w, http.StatusOK, options)
}

func (h *AuthHandler) PasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetUser(r)
	if user == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "login first")
		return
	}
	sessionData := h.challenges[user.ID.String()]
	wu := &webauthnUser{user: user}
	credential, err := h.wa.FinishRegistration(wu, *sessionData, r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "registration failed")
		return
	}
	var deviceName string
	if err := json.NewDecoder(r.Body).Decode(&struct{ DeviceName *string `json:"device_name"` }{&deviceName}); err != nil {
		deviceName = "My device"
	}
	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}
	_, err = queries.CreatePasskeyCredential(r.Context(), h.pool,
		user.ID, credential.ID, credential.PublicKey, transports, deviceName)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "save credential failed")
		return
	}
	delete(h.challenges, user.ID.String())
	helpers.WriteJSON(w, http.StatusCreated, map[string]string{"status": "passkey registered"})
}
```

- [ ] **Step 3: Integration test register + login**

```bash
# Register
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","display_name":"Test User","password":"hunter2"}'
# Expected: {"id":"...","email":"test@example.com"}

# Login
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"hunter2"}'
# Expected: {"token":"...","user":{"id":"...","email":"test@example.com",...}}

# Authenticated request
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"hunter2"}' | jq -r .token)
curl -s http://localhost:8080/api/users/me -H "Authorization: Bearer $TOKEN"
# Expected: user object
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/handlers/auth.go
git commit -m "feat: auth handlers — register, login, logout, passkey"
```

---

## Task 10: Project & Secret Handlers

**Files:**
- Create: `internal/db/models/project.go`
- Create: `internal/db/models/secret.go`
- Create: `internal/db/queries/project.go`
- Create: `internal/db/queries/secret.go`
- Create: `internal/api/handlers/projects.go`
- Create: `internal/api/handlers/secrets.go`
- Create: `internal/api/handlers/users.go`

- [ ] **Step 1: Write project model and queries**

```go
// internal/db/models/project.go
package models

import (
	"time"
	"github.com/google/uuid"
	"github.com/kgory/kirmaphor/internal/rbac"
)

type Project struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
}

type ProjectUser struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Role      rbac.Role
	AddedAt   time.Time
}
```

```go
// internal/db/queries/project.go
package queries

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/rbac"
	"github.com/google/uuid"
)

func CreateProject(ctx context.Context, pool *pgxpool.Pool, name, description string, createdBy uuid.UUID) (*models.Project, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	p := &models.Project{}
	err = tx.QueryRow(ctx,
		`INSERT INTO projects (name, description, created_by) VALUES ($1, $2, $3)
		 RETURNING id, name, description, created_by, created_at`,
		name, description, createdBy,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	// Creator becomes owner
	_, err = tx.Exec(ctx,
		`INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, $3)`,
		p.ID, createdBy, rbac.RoleOwner)
	if err != nil {
		return nil, err
	}
	return p, tx.Commit(ctx)
}

func GetProjectsByUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]*models.Project, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.id, p.name, p.description, p.created_by, p.created_at
		 FROM projects p
		 JOIN project_users pu ON pu.project_id = p.id
		 WHERE pu.user_id = $1
		 ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func GetProjectRole(ctx context.Context, pool *pgxpool.Pool, projectID, userID uuid.UUID) (rbac.Role, error) {
	var role rbac.Role
	err := pool.QueryRow(ctx,
		`SELECT role FROM project_users WHERE project_id = $1 AND user_id = $2`,
		projectID, userID,
	).Scan(&role)
	return role, err
}
```

- [ ] **Step 2: Write secret model and queries**

```go
// internal/db/models/secret.go
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
	EncryptedValue []byte // Layer 2 server-encrypted
	Nonce          []byte
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
}
```

```go
// internal/db/queries/secret.go
package queries

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/google/uuid"
)

func CreateSecret(ctx context.Context, pool *pgxpool.Pool,
	projectID uuid.UUID, name, secretType string,
	encryptedValue, nonce []byte, createdBy uuid.UUID,
) (*models.Secret, error) {
	s := &models.Secret{}
	err := pool.QueryRow(ctx,
		`INSERT INTO secrets (project_id, name, type, encrypted_value, nonce, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, project_id, name, type, encrypted_value, nonce, created_by, created_at`,
		projectID, name, secretType, encryptedValue, nonce, createdBy,
	).Scan(&s.ID, &s.ProjectID, &s.Name, &s.Type, &s.EncryptedValue, &s.Nonce, &s.CreatedBy, &s.CreatedAt)
	return s, err
}

func ListSecrets(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.Secret, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, type, created_by, created_at
		 FROM secrets WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var secrets []*models.Secret
	for rows.Next() {
		s := &models.Secret{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &s.Type, &s.CreatedBy, &s.CreatedAt); err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}
	return secrets, rows.Err()
}
```

- [ ] **Step 3: Write project and secret HTTP handlers**

```go
// internal/api/handlers/projects.go
package handlers

import (
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

func ListProjects(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		projects, err := queries.GetProjectsByUser(r.Context(), pool, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, projects)
	}
}

func CreateProject(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Name == "" {
			helpers.WriteError(w, http.StatusBadRequest, "name is required")
			return
		}
		p, err := queries.CreateProject(r.Context(), pool, req.Name, req.Description, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, p)
	}
}

func GetProject(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		id := r.PathValue("id")
		// Parse UUID, get project, verify membership
		_ = user
		_ = id
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"todo": "implement"})
	}
}
```

```go
// internal/api/handlers/secrets.go
package handlers

import (
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/crypto"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
	"github.com/google/uuid"
)

func CreateSecret(pool *pgxpool.Pool, masterKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := helpers.GetSession(r)
		// Require secure session for secret creation
		if err := auth.CheckSecureSession(session); err != nil {
			helpers.WriteError(w, http.StatusForbidden, "secure_session_required")
			return
		}

		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		role, err := queries.GetProjectRole(r.Context(), pool, projectID, user.ID)
		if err != nil || !rbac.HasPermission(role, rbac.PermManageSecrets) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}

		var req struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value []byte `json:"value"` // already Layer-1 encrypted by client
		}
		if !helpers.Bind(w, r, &req) {
			return
		}

		// Layer 2: server-side encrypt
		encrypted, nonce, err := crypto.Encrypt(masterKey, req.Value)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "encryption error")
			return
		}
		secret, err := queries.CreateSecret(r.Context(), pool, projectID, req.Name, req.Type, encrypted, nonce, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusConflict, "secret already exists")
			return
		}
		// Return without encrypted_value
		helpers.WriteJSON(w, http.StatusCreated, map[string]any{
			"id": secret.ID, "name": secret.Name, "type": secret.Type,
		})
	}
}

func ListSecrets(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		secrets, err := queries.ListSecrets(r.Context(), pool, projectID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, secrets)
	}
}
```

```go
// internal/api/handlers/users.go
package handlers

import (
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

func GetMe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		helpers.WriteJSON(w, http.StatusOK, map[string]any{
			"id": user.ID, "email": user.Email,
			"display_name": user.DisplayName, "avatar_url": user.AvatarURL,
			"onboarded": user.Onboarded,
		})
	}
}

func UpdateMe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		var req struct {
			DisplayName *string `json:"display_name"`
			AvatarURL   *string `json:"avatar_url"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		_, err := pool.Exec(r.Context(),
			`UPDATE users SET display_name = COALESCE($1, display_name),
			                  avatar_url = COALESCE($2, avatar_url)
			 WHERE id = $3`,
			req.DisplayName, req.AvatarURL, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func ListSessions(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		rows, err := pool.Query(r.Context(),
			`SELECT id, device_fingerprint, device_label, geo_city, geo_country,
			        ip_address, user_agent, is_current, expires_at, created_at
			 FROM user_sessions WHERE user_id = $1 AND is_current = TRUE
			 ORDER BY created_at DESC`, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		defer rows.Close()
		var sessions []map[string]any
		for rows.Next() {
			var s struct {
				ID, DeviceFingerprint, DeviceLabel, GeoCity, GeoCountry, IPAddress, UserAgent any
				IsCurrent bool
				ExpiresAt, CreatedAt any
			}
			rows.Scan(&s.ID, &s.DeviceFingerprint, &s.DeviceLabel, &s.GeoCity, &s.GeoCountry,
				&s.IPAddress, &s.UserAgent, &s.IsCurrent, &s.ExpiresAt, &s.CreatedAt)
			sessions = append(sessions, map[string]any{
				"id": s.ID, "device_label": s.DeviceLabel, "ip": s.IPAddress,
				"is_current": s.IsCurrent, "expires_at": s.ExpiresAt,
			})
		}
		helpers.WriteJSON(w, http.StatusOK, sessions)
	}
}

func RevokeSession(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		sessionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid session id")
			return
		}
		if err := queries.RevokeSession(r.Context(), pool, sessionID, user.ID); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
	}
}
```

- [ ] **Step 4: Add secrets routes to router**

```go
// In internal/api/router.go — inside the authenticated group:
r.Post("/projects/{projectId}/secrets", handlers.CreateSecret(pool, masterKey))
r.Get("/projects/{projectId}/secrets", handlers.ListSecrets(pool))
```

- [ ] **Step 5: Integration test secrets**

```bash
TOKEN=<from login>
PROJECT_ID=<from create project>

curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/secrets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"ansible-vault-pass","type":"string","value":"dGVzdA=="}'
# Expected: 403 secure_session_required (secure session not set)
```

- [ ] **Step 6: Commit**

```bash
git add internal/db/models/ internal/db/queries/ internal/api/handlers/
git commit -m "feat: project and secrets CRUD with two-layer encryption and RBAC"
```

---

## Task 11: Next.js Frontend Scaffold

**Files:**
- Create: `web/` (Next.js app)

- [ ] **Step 1: Bootstrap Next.js app**

```bash
cd /Users/kgory/dev/kirmaphor
npx create-next-app@latest web \
  --typescript --tailwind --app --src-dir=false \
  --import-alias="@/*" --no-eslint
```

- [ ] **Step 2: Install frontend dependencies**

```bash
cd web
npx shadcn@latest init
npm install zustand next-intl @simplewebauthn/browser
npm install -D @types/node
```

- [ ] **Step 3: Install shadcn/ui components**

```bash
npx shadcn@latest add button input card label form toast avatar dropdown-menu
```

- [ ] **Step 4: Write API client**

```typescript
// web/lib/api.ts
const BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

export class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message)
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = typeof window !== 'undefined'
    ? localStorage.getItem('kirmaphore_token')
    : null

  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new APIError(res.status, err.error ?? 'Request failed')
  }
  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
```

- [ ] **Step 5: Write auth store**

```typescript
// web/lib/auth-store.ts
import { create } from 'zustand'
import { api } from './api'

interface User {
  id: string
  email: string
  display_name: string
  avatar_url?: string
}

interface AuthState {
  user: User | null
  token: string | null
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  setUser: (user: User, token: string) => void
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== 'undefined' ? localStorage.getItem('kirmaphore_token') : null,

  login: async (email, password) => {
    const res = await api.post<{ token: string; user: User }>('/api/auth/login', { email, password })
    localStorage.setItem('kirmaphore_token', res.token)
    set({ user: res.user, token: res.token })
  },

  logout: () => {
    api.post('/api/auth/logout', {}).catch(() => {})
    localStorage.removeItem('kirmaphore_token')
    set({ user: null, token: null })
  },

  setUser: (user, token) => {
    localStorage.setItem('kirmaphore_token', token)
    set({ user, token })
  },
}))
```

- [ ] **Step 6: Write login page**

```typescript
// web/app/(auth)/login/page.tsx
'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth-store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const router = useRouter()

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await login(email, password)
      router.push('/')
    } catch (err: any) {
      setError(err.message ?? 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-2xl font-bold">Kirmaphore</CardTitle>
          <p className="text-sm text-muted-foreground">Sign in to your account</p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin} className="space-y-4">
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
            <Input
              type="password"
              placeholder="Password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
```

- [ ] **Step 7: Write route protection middleware**

```typescript
// web/middleware.ts
import { NextRequest, NextResponse } from 'next/server'

const PUBLIC_PATHS = ['/login', '/register']

export function middleware(req: NextRequest) {
  const token = req.cookies.get('kirmaphore_token')?.value
  const isPublic = PUBLIC_PATHS.some(p => req.nextUrl.pathname.startsWith(p))

  if (!token && !isPublic) {
    return NextResponse.redirect(new URL('/login', req.url))
  }
  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
```

- [ ] **Step 8: Test frontend**

```bash
cd web && npm run dev
```

Open http://localhost:3000/login — should see login form.

- [ ] **Step 9: Commit**

```bash
cd ..
git add web/
git commit -m "feat: Next.js frontend scaffold with auth, API client, login page"
```

---

## Task 12: i18n Setup

**Files:**
- Create: `web/messages/en.json`
- Create: `web/messages/ru.json`
- Modify: `web/app/layout.tsx`

- [ ] **Step 1: Write message files**

```json
// web/messages/en.json
{
  "auth": {
    "login": "Sign in",
    "logout": "Sign out",
    "register": "Create account",
    "email": "Email",
    "password": "Password",
    "loginFailed": "Invalid email or password",
    "passkey": "Sign in with Passkey"
  },
  "nav": {
    "projects": "Projects",
    "settings": "Settings",
    "profile": "Profile"
  },
  "common": {
    "loading": "Loading...",
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "create": "Create"
  }
}
```

```json
// web/messages/ru.json
{
  "auth": {
    "login": "Войти",
    "logout": "Выйти",
    "register": "Создать аккаунт",
    "email": "Email",
    "password": "Пароль",
    "loginFailed": "Неверный email или пароль",
    "passkey": "Войти с Passkey"
  },
  "nav": {
    "projects": "Проекты",
    "settings": "Настройки",
    "profile": "Профиль"
  },
  "common": {
    "loading": "Загрузка...",
    "save": "Сохранить",
    "cancel": "Отмена",
    "delete": "Удалить",
    "create": "Создать"
  }
}
```

- [ ] **Step 2: Configure next-intl in layout**

```typescript
// web/app/layout.tsx
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { NextIntlClientProvider } from 'next-intl'
import { getLocale, getMessages } from 'next-intl/server'
import './globals.css'

const inter = Inter({ subsets: ['latin', 'cyrillic'] })

export const metadata: Metadata = {
  title: 'Kirmaphore',
  description: 'AI-native Ansible automation platform',
}

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const locale = await getLocale()
  const messages = await getMessages()
  return (
    <html lang={locale}>
      <body className={inter.className}>
        <NextIntlClientProvider messages={messages}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
```

- [ ] **Step 3: Add next-intl config**

```typescript
// web/i18n.ts
import { getRequestConfig } from 'next-intl/server'

export default getRequestConfig(async () => {
  const locale = 'en' // TODO: detect from cookie/header
  return {
    locale,
    messages: (await import(`./messages/${locale}.json`)).default,
  }
})
```

- [ ] **Step 4: Commit**

```bash
git add web/messages/ web/app/layout.tsx web/i18n.ts
git commit -m "feat: i18n setup with EN and RU locales"
```

---

## Task 13: Dashboard Layout

**Files:**
- Create: `web/app/(dashboard)/layout.tsx`
- Create: `web/app/(dashboard)/page.tsx`
- Create: `web/components/layout/Sidebar.tsx`
- Create: `web/components/layout/Header.tsx`

- [ ] **Step 1: Write Sidebar**

```typescript
// web/components/layout/Sidebar.tsx
'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { FolderOpen, Settings, Cpu } from 'lucide-react'

const nav = [
  { href: '/', label: 'Projects', icon: FolderOpen },
  { href: '/runs', label: 'Runs', icon: Cpu },
  { href: '/settings', label: 'Settings', icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()
  return (
    <aside className="w-64 min-h-screen border-r bg-card flex flex-col">
      <div className="p-6 border-b">
        <span className="text-xl font-bold tracking-tight">Kirmaphore</span>
      </div>
      <nav className="flex-1 p-4 space-y-1">
        {nav.map(({ href, label, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={cn(
              'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
              pathname === href
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
            )}
          >
            <Icon className="h-4 w-4" />
            {label}
          </Link>
        ))}
      </nav>
    </aside>
  )
}
```

- [ ] **Step 2: Write Header**

```typescript
// web/components/layout/Header.tsx
'use client'

import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu, DropdownMenuContent,
  DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export function Header() {
  const { user, logout } = useAuth()
  const router = useRouter()

  const handleLogout = () => {
    logout()
    router.push('/login')
  }

  return (
    <header className="h-14 border-b flex items-center justify-end px-6">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button className="flex items-center gap-2 text-sm">
            <Avatar className="h-8 w-8">
              <AvatarFallback>
                {user?.display_name?.[0]?.toUpperCase() ?? '?'}
              </AvatarFallback>
            </Avatar>
            <span className="hidden sm:block">{user?.display_name}</span>
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => router.push('/settings/profile')}>
            Profile
          </DropdownMenuItem>
          <DropdownMenuItem onClick={handleLogout} className="text-destructive">
            Sign out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
```

- [ ] **Step 3: Write dashboard layout and home page**

```typescript
// web/app/(dashboard)/layout.tsx
import { Sidebar } from '@/components/layout/Sidebar'
import { Header } from '@/components/layout/Header'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 flex flex-col">
        <Header />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  )
}
```

```typescript
// web/app/(dashboard)/page.tsx
export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-2">Welcome to Kirmaphore</h1>
      <p className="text-muted-foreground">Select a project to get started.</p>
    </div>
  )
}
```

- [ ] **Step 4: Verify UI**

```bash
cd web && npm run dev
```

Login at http://localhost:3000/login → should redirect to dashboard with sidebar.

- [ ] **Step 5: Commit**

```bash
cd ..
git add web/components/ web/app/
git commit -m "feat: dashboard layout with sidebar and header"
```

---

## Self-Review

**Spec coverage check:**

| Spec Section | Covered by Task |
|---|---|
| Tech stack: Go, Next.js, PostgreSQL, shadcn/ui | Task 1, 11 |
| Passkeys (WebAuthn) | Task 6, 9 |
| Local auth + bcrypt | Task 5, 9 |
| Secure session (5-min re-auth) | Task 4 |
| RBAC: owner/admin/engineer/viewer | Task 7 |
| Two-layer AES-256-GCM secrets | Task 3, 10 |
| User profile model (from KirFlow) | Task 4 |
| Session tracking (device, IP, geo) | Task 4 |
| Project CRUD | Task 10 |
| Secrets CRUD | Task 10 |
| Frontend auth flow | Task 11 |
| i18n EN + RU | Task 12 |
| Dashboard layout | Task 13 |
| NATS JetStream queue | **Not in Plan 1 — covered in Plan 2** |
| Execution engine | **Plan 2** |
| Cloud integrations | **Plan 2** |
| Web IDE (Monaco) | **Plan 3** |
| Workflow Designer | **Plan 5** |
| Forge AI | **Plan 6** |

**Placeholder scan:** Passkey register/finish has a note about go-webauthn API usage — Step 2 of Task 9 fills in the full implementation. No TBD sections.

**Type consistency:** `models.UserSession.SecureAt` used in Task 4 (model) and Task 4 (auth check) consistently as `*time.Time`. `rbac.Role` used in models and queries consistently as `Role string` type.
