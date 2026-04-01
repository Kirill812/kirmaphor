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
