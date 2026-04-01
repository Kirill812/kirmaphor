package config

import (
	"encoding/hex"
	"fmt"
	"os"
)

type Config struct {
	Port      string
	DBURL     string
	MasterKey string // 64 hex chars = 32 bytes for AES-256
	RPOrigin  string
	RPID      string
	RPName    string
}

func Load() (*Config, error) {
	dbURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	masterKey, err := requireEnv("MASTER_KEY")
	if err != nil {
		return nil, err
	}
	// Validate: must be 64 valid hex chars (32 bytes)
	decoded, err := hex.DecodeString(masterKey)
	if err != nil || len(decoded) != 32 {
		return nil, fmt.Errorf("MASTER_KEY must be 64 valid hex characters (32 bytes)")
	}
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBURL:     dbURL,
		MasterKey: masterKey,
		RPOrigin:  getEnv("RP_ORIGIN", "http://localhost:3000"),
		RPID:      getEnv("RP_ID", "localhost"),
		RPName:    getEnv("RP_NAME", "Kirmaphore"),
	}, nil
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return v, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
