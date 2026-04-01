package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kgory/kirmaphor/internal/db"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://kirmaphore:kirmaphore@localhost:5432/kirmaphore?sslmode=disable"
	}
	fmt.Println("Running migrations against:", url)
	if err := db.RunMigrations(url, "migrations"); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("Migrations applied successfully.")
}
