package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

type pendingUser struct {
	Email       string
	DisplayName string
	TempID      []byte
}

type challengeEntry struct {
	data      *webauthn.SessionData
	createdAt time.Time
	pending   *pendingUser // non-nil for unauthenticated registration
}

// pendingWebauthnUser satisfies webauthn.User for new (not-yet-created) users.
type pendingWebauthnUser struct {
	id          []byte
	email       string
	displayName string
}

func (u *pendingWebauthnUser) WebAuthnID() []byte                         { return u.id }
func (u *pendingWebauthnUser) WebAuthnName() string                       { return u.email }
func (u *pendingWebauthnUser) WebAuthnDisplayName() string                { return u.displayName }
func (u *pendingWebauthnUser) WebAuthnCredentials() []webauthn.Credential { return nil }

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	pool       *pgxpool.Pool
	wa         *webauthn.WebAuthn
	masterKey  []byte
	mu         sync.Mutex
	challenges map[string]*challengeEntry // userID -> challenge (in-memory, replace with Redis in prod)
}

func NewAuthHandler(pool *pgxpool.Pool, wa *webauthn.WebAuthn, masterKey []byte) *AuthHandler {
	h := &AuthHandler{
		pool:       pool,
		wa:         wa,
		masterKey:  masterKey,
		challenges: make(map[string]*challengeEntry),
	}
	go h.cleanupChallenges()
	return h
}

const challengeTTL = 5 * time.Minute

func (h *AuthHandler) cleanupChallenges() {
	ticker := time.NewTicker(challengeTTL)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-challengeTTL)
		h.mu.Lock()
		for k, v := range h.challenges {
			if v.createdAt.Before(cutoff) {
				delete(h.challenges, k)
			}
		}
		h.mu.Unlock()
	}
}

// webauthnUser adapts models.User to webauthn.User interface.
type webauthnUser struct {
	user  *models.User
	creds []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte                         { return []byte(u.user.ID.String()) }
func (u *webauthnUser) WebAuthnName() string                       { return u.user.Email }
func (u *webauthnUser) WebAuthnDisplayName() string                { return u.user.DisplayName }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.creds }

// POST /api/auth/register
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
	if req.Email == "" || req.DisplayName == "" || req.Password == "" {
		helpers.WriteError(w, http.StatusBadRequest, "email, display_name, and password are required")
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

// POST /api/auth/login
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
	if err != nil || user == nil || user.PasswordHash == nil {
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
	helpers.WriteJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
		},
	})
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session := helpers.GetSession(r)
	if session == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	_ = queries.RevokeSession(r.Context(), h.pool, session.ID, session.UserID)
	helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// POST /api/auth/passkey/register/begin
// - Unauthenticated: body must contain {email, display_name}; returns {pending_id, options}
// - Authenticated: body ignored; returns options directly
func (h *AuthHandler) PasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetUser(r)

	if user == nil {
		// New-user passkey registration: collect identity from body
		var req struct {
			Email       string `json:"email"`
			DisplayName string `json:"display_name"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Email == "" || req.DisplayName == "" {
			helpers.WriteError(w, http.StatusBadRequest, "email and display_name required")
			return
		}

		tempID := make([]byte, 16)
		if _, err := rand.Read(tempID); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		pendingKey := hex.EncodeToString(tempID)

		pu := &pendingWebauthnUser{id: tempID, email: req.Email, displayName: req.DisplayName}
		options, sessionData, err := h.wa.BeginRegistration(pu)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "webauthn error")
			return
		}

		h.mu.Lock()
		h.challenges[pendingKey] = &challengeEntry{
			data:      sessionData,
			createdAt: time.Now(),
			pending:   &pendingUser{Email: req.Email, DisplayName: req.DisplayName, TempID: tempID},
		}
		h.mu.Unlock()

		helpers.WriteJSON(w, http.StatusOK, map[string]any{
			"pending_id": pendingKey,
			"options":    options,
		})
		return
	}

	// Existing user: add passkey to account
	wu := &webauthnUser{user: user}
	options, sessionData, err := h.wa.BeginRegistration(wu)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "webauthn error")
		return
	}
	h.mu.Lock()
	h.challenges[user.ID.String()] = &challengeEntry{data: sessionData, createdAt: time.Now()}
	h.mu.Unlock()
	helpers.WriteJSON(w, http.StatusOK, options)
}

// POST /api/auth/passkey/register/finish
// - Unauthenticated: ?pending_id=<hex> — creates new user + session
// - Authenticated: adds passkey to existing account
func (h *AuthHandler) PasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetUser(r)

	if user == nil {
		// New-user path: finish unauthenticated registration
		pendingKey := r.URL.Query().Get("pending_id")
		if pendingKey == "" {
			helpers.WriteError(w, http.StatusBadRequest, "pending_id required")
			return
		}
		h.mu.Lock()
		entry := h.challenges[pendingKey]
		h.mu.Unlock()
		if entry == nil || entry.pending == nil {
			helpers.WriteError(w, http.StatusBadRequest, "no pending registration")
			return
		}

		pu := &pendingWebauthnUser{
			id:          entry.pending.TempID,
			email:       entry.pending.Email,
			displayName: entry.pending.DisplayName,
		}
		credential, err := h.wa.FinishRegistration(pu, *entry.data, r)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "registration failed")
			return
		}

		newUser, err := queries.CreateUser(r.Context(), h.pool, entry.pending.Email, entry.pending.DisplayName, nil)
		if err != nil {
			helpers.WriteError(w, http.StatusConflict, "email already registered")
			return
		}

		transports := make([]string, len(credential.Transport))
		for i, t := range credential.Transport {
			transports[i] = string(t)
		}
		deviceName := "My passkey"
		if dn := r.URL.Query().Get("device_name"); dn != "" {
			deviceName = dn
		}
		_, err = queries.CreatePasskeyCredential(r.Context(), h.pool,
			newUser.ID, credential.ID, credential.PublicKey, transports, deviceName)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "save credential failed")
			return
		}

		h.mu.Lock()
		delete(h.challenges, pendingKey)
		h.mu.Unlock()

		token, hash, err := queries.GenerateSessionToken()
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		ip := r.RemoteAddr
		ua := r.UserAgent()
		_, err = queries.CreateSession(r.Context(), h.pool, queries.CreateSessionParams{
			UserID:    newUser.ID,
			TokenHash: hash,
			IPAddress: &ip,
			UserAgent: &ua,
			ExpiresAt: time.Now().Add(time.Duration(newUser.SessionTimeoutMinutes) * time.Minute),
		})
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, map[string]any{
			"token": token,
			"user": map[string]any{
				"id":           newUser.ID,
				"email":        newUser.Email,
				"display_name": newUser.DisplayName,
			},
		})
		return
	}

	// Authenticated path: add passkey to existing account
	h.mu.Lock()
	entry := h.challenges[user.ID.String()]
	h.mu.Unlock()
	if entry == nil {
		helpers.WriteError(w, http.StatusBadRequest, "no pending registration")
		return
	}
	wu := &webauthnUser{user: user}
	credential, err := h.wa.FinishRegistration(wu, *entry.data, r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "registration failed")
		return
	}
	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}
	deviceName := "My passkey"
	if dn := r.URL.Query().Get("device_name"); dn != "" {
		deviceName = dn
	}
	_, err = queries.CreatePasskeyCredential(r.Context(), h.pool,
		user.ID, credential.ID, credential.PublicKey, transports, deviceName)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "save credential failed")
		return
	}
	h.mu.Lock()
	delete(h.challenges, user.ID.String())
	h.mu.Unlock()
	helpers.WriteJSON(w, http.StatusCreated, map[string]string{"status": "passkey registered"})
}

// POST /api/auth/passkey/login/begin
func (h *AuthHandler) PasskeyLoginBegin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	// Ignore parse errors — email is optional (discoverable login)
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Email == "" {
		// Discoverable login: browser picks the passkey, no email needed
		options, sessionData, err := h.wa.BeginDiscoverableLogin()
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "webauthn error")
			return
		}
		sessionIDBytes := make([]byte, 16)
		if _, err := rand.Read(sessionIDBytes); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		sessionID := hex.EncodeToString(sessionIDBytes)
		h.mu.Lock()
		h.challenges["disc:"+sessionID] = &challengeEntry{data: sessionData, createdAt: time.Now()}
		h.mu.Unlock()
		helpers.WriteJSON(w, http.StatusOK, map[string]any{
			"session_id": sessionID,
			"options":    options,
		})
		return
	}

	// Email-based login
	user, err := queries.GetUserByEmail(r.Context(), h.pool, req.Email)
	if err != nil || user == nil {
		helpers.WriteError(w, http.StatusBadRequest, "user not found")
		return
	}
	dbCreds, err := queries.GetPasskeyCredentialsByUserID(r.Context(), h.pool, user.ID)
	if err != nil || len(dbCreds) == 0 {
		helpers.WriteError(w, http.StatusBadRequest, "no passkeys registered")
		return
	}
	waCreds := make([]webauthn.Credential, len(dbCreds))
	for i, c := range dbCreds {
		waCreds[i] = webauthn.Credential{
			ID:        c.CredentialID,
			PublicKey: c.PublicKey,
			Authenticator: webauthn.Authenticator{
				SignCount: c.Counter,
			},
		}
	}
	wu := &webauthnUser{user: user, creds: waCreds}
	options, sessionData, err := h.wa.BeginLogin(wu)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "webauthn error")
		return
	}
	h.mu.Lock()
	h.challenges[user.ID.String()] = &challengeEntry{data: sessionData, createdAt: time.Now()}
	h.mu.Unlock()
	helpers.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": "",
		"options":    options,
	})
}

// POST /api/auth/passkey/login/finish
func (h *AuthHandler) PasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	var user *models.User
	var credential *webauthn.Credential

	if sessionID != "" {
		// Discoverable login path
		h.mu.Lock()
		entry := h.challenges["disc:"+sessionID]
		h.mu.Unlock()
		if entry == nil {
			helpers.WriteError(w, http.StatusBadRequest, "no pending login")
			return
		}
		var err error
		_, credential, err = h.wa.FinishPasskeyLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
			// Look up by credential ID (rawID), not userHandle.
			// The userHandle was set during registration as a random tempID,
			// not the user's UUID, so we can't rely on it for lookup.
			dbCred, dbErr := queries.GetPasskeyCredentialByCredentialID(r.Context(), h.pool, rawID)
			if dbErr != nil || dbCred == nil {
				return nil, fmt.Errorf("credential not found")
			}
			u, dbErr := queries.GetUserByID(r.Context(), h.pool, dbCred.UserID)
			if dbErr != nil || u == nil {
				return nil, fmt.Errorf("user not found")
			}
			user = u
			dbCreds, _ := queries.GetPasskeyCredentialsByUserID(r.Context(), h.pool, u.ID)
			waCreds := make([]webauthn.Credential, len(dbCreds))
			for i, c := range dbCreds {
				waCreds[i] = webauthn.Credential{
					ID:        c.CredentialID,
					PublicKey: c.PublicKey,
					Authenticator: webauthn.Authenticator{SignCount: c.Counter},
				}
			}
			return &webauthnUser{user: u, creds: waCreds}, nil
		}, *entry.data, r)
		if err != nil {
			helpers.WriteError(w, http.StatusUnauthorized, "passkey verification failed")
			return
		}
		h.mu.Lock()
		delete(h.challenges, "disc:"+sessionID)
		h.mu.Unlock()
	} else {
		// Email-based login path
		email := r.URL.Query().Get("email")
		if email == "" {
			helpers.WriteError(w, http.StatusBadRequest, "session_id or email required")
			return
		}
		var err error
		user, err = queries.GetUserByEmail(r.Context(), h.pool, email)
		if err != nil || user == nil {
			helpers.WriteError(w, http.StatusUnauthorized, "invalid")
			return
		}
		h.mu.Lock()
		loginEntry := h.challenges[user.ID.String()]
		h.mu.Unlock()
		if loginEntry == nil {
			helpers.WriteError(w, http.StatusBadRequest, "no pending login")
			return
		}
		sessionData := loginEntry.data
		dbCreds, _ := queries.GetPasskeyCredentialsByUserID(r.Context(), h.pool, user.ID)
		waCreds := make([]webauthn.Credential, len(dbCreds))
		for i, c := range dbCreds {
			waCreds[i] = webauthn.Credential{
				ID:        c.CredentialID,
				PublicKey: c.PublicKey,
				Authenticator: webauthn.Authenticator{SignCount: c.Counter},
			}
		}
		wu := &webauthnUser{user: user, creds: waCreds}
		credential, err = h.wa.FinishLogin(wu, *sessionData, r)
		if err != nil {
			helpers.WriteError(w, http.StatusUnauthorized, "passkey verification failed")
			return
		}
		h.mu.Lock()
		delete(h.challenges, user.ID.String())
		h.mu.Unlock()
	}
	_ = queries.UpdatePasskeyCounter(r.Context(), h.pool, credential.ID, credential.Authenticator.SignCount)

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
	helpers.WriteJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": user.ID, "email": user.Email},
	})
}
