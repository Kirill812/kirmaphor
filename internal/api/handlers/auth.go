package handlers

import (
	"encoding/json"
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

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	pool       *pgxpool.Pool
	wa         *webauthn.WebAuthn
	masterKey  []byte
	mu         sync.Mutex
	challenges map[string]*webauthn.SessionData // userID -> challenge (in-memory, replace with Redis in prod)
}

func NewAuthHandler(pool *pgxpool.Pool, wa *webauthn.WebAuthn, masterKey []byte) *AuthHandler {
	return &AuthHandler{
		pool:       pool,
		wa:         wa,
		masterKey:  masterKey,
		challenges: make(map[string]*webauthn.SessionData),
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

// POST /api/auth/passkey/register/begin (requires prior login)
func (h *AuthHandler) PasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetUser(r)
	if user == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "login required to register passkey")
		return
	}
	wu := &webauthnUser{user: user}
	options, sessionData, err := h.wa.BeginRegistration(wu)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "webauthn error")
		return
	}
	h.mu.Lock()
	h.challenges[user.ID.String()] = sessionData
	h.mu.Unlock()
	helpers.WriteJSON(w, http.StatusOK, options)
}

// POST /api/auth/passkey/register/finish
func (h *AuthHandler) PasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetUser(r)
	if user == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "login required")
		return
	}
	h.mu.Lock()
	sessionData := h.challenges[user.ID.String()]
	h.mu.Unlock()
	if sessionData == nil {
		helpers.WriteError(w, http.StatusBadRequest, "no pending registration")
		return
	}
	wu := &webauthnUser{user: user}
	credential, err := h.wa.FinishRegistration(wu, *sessionData, r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "registration failed")
		return
	}
	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}

	// Parse optional device_name from body
	deviceName := "My device"
	var deviceReq struct {
		DeviceName string `json:"device_name"`
	}
	// Body already consumed by FinishRegistration — device_name can come from query param
	if dn := r.URL.Query().Get("device_name"); dn != "" {
		deviceName = dn
	} else if err := json.NewDecoder(r.Body).Decode(&deviceReq); err == nil && deviceReq.DeviceName != "" {
		deviceName = deviceReq.DeviceName
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
	if !helpers.Bind(w, r, &req) {
		return
	}
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
	h.challenges[user.ID.String()] = sessionData
	h.mu.Unlock()
	helpers.WriteJSON(w, http.StatusOK, options)
}

// POST /api/auth/passkey/login/finish
func (h *AuthHandler) PasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		helpers.WriteError(w, http.StatusBadRequest, "email query param required")
		return
	}
	user, err := queries.GetUserByEmail(r.Context(), h.pool, email)
	if err != nil || user == nil {
		helpers.WriteError(w, http.StatusUnauthorized, "invalid")
		return
	}
	h.mu.Lock()
	sessionData := h.challenges[user.ID.String()]
	h.mu.Unlock()
	if sessionData == nil {
		helpers.WriteError(w, http.StatusBadRequest, "no pending login")
		return
	}
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
	credential, err := h.wa.FinishLogin(wu, *sessionData, r)
	if err != nil {
		helpers.WriteError(w, http.StatusUnauthorized, "passkey verification failed")
		return
	}
	_ = queries.UpdatePasskeyCounter(r.Context(), h.pool, credential.ID, credential.Authenticator.SignCount)
	h.mu.Lock()
	delete(h.challenges, user.ID.String())
	h.mu.Unlock()

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
