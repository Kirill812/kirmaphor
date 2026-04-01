package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kgory/kirmaphor/internal/api"
	"github.com/kgory/kirmaphor/internal/config"
	"github.com/kgory/kirmaphor/internal/crypto"
	"github.com/kgory/kirmaphor/internal/db"
	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/kgory/kirmaphor/internal/scheduler"
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

	if err := db.RunMigrations(cfg.DBURL, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	masterKey, err := crypto.LoadMasterKey(cfg.MasterKey)
	if err != nil {
		log.Fatalf("master key: %v", err)
	}

	deps := execution.RunnerDeps{
		Pool: pool,
		Decrypt: func(encrypted, nonce []byte) ([]byte, error) {
			return crypto.Decrypt(masterKey, encrypted, nonce)
		},
	}

	taskPool := execution.NewTaskPool(10)
	taskPool.Start()
	defer taskPool.Stop()

	sched := scheduler.New(pool, taskPool, deps)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Run(ctx)

	router := api.NewRouter(cfg, pool, taskPool, deps)
	log.Printf("starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
