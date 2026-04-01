// internal/config/config.go
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port      string
	DBURL     string
	MasterKey string // 32-byte hex for AES-256
	RPOrigin  string // WebAuthn: "https://kirmaphore.example.com"
	RPID      string // WebAuthn: "kirmaphore.example.com"
	RPName    string // WebAuthn: "Kirmaphore"
}

func Load() (*Config, error) {
	c := &Config{
		Port:      getEnv("PORT", "8080"),
		DBURL:     mustEnv("DATABASE_URL"),
		MasterKey: mustEnv("MASTER_KEY"),
		RPOrigin:  getEnv("RP_ORIGIN", "http://localhost:3000"),
		RPID:      getEnv("RP_ID", "localhost"),
		RPName:    getEnv("RP_NAME", "Kirmaphore"),
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
