package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/handlers"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	apiMiddleware "github.com/kgory/kirmaphor/internal/api/middleware"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/config"
	"github.com/kgory/kirmaphor/internal/crypto"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	wa, err := auth.NewWebAuthn(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init webauthn: %v", err))
	}

	masterKey, err := crypto.LoadMasterKey(cfg.MasterKey)
	if err != nil {
		panic(fmt.Sprintf("failed to load master key: %v", err))
	}

	authHandler := handlers.NewAuthHandler(pool, wa, masterKey)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", getEnvOrDefault("ALLOWED_ORIGIN", "http://localhost:3000")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	requireAuth := apiMiddleware.RequireAuth(pool)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		// Auth routes (public)
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
			r.Delete("/users/me/sessions/{id}", handlers.RevokeUserSession(pool))

			r.Get("/projects", handlers.ListProjects(pool))
			r.Post("/projects", handlers.CreateProject(pool))
			r.Get("/projects/{id}", handlers.GetProject(pool))

			r.Post("/projects/{projectId}/secrets", handlers.CreateSecret(pool, masterKey))
			r.Get("/projects/{projectId}/secrets", handlers.ListSecrets(pool))
		})
	})

	return r
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
