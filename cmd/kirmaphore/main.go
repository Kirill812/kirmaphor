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
