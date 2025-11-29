package main

import (
	"context"
	"log"
	"os"

	"pgweb-service/internal/migrate"
)

func main() {
	pgURL := os.Getenv("PGWEB_DATABASE_URL")
	if pgURL == "" {
		log.Fatal("PGWEB_DATABASE_URL must be set")
	}

	if err := migrate.Run(context.Background(), pgURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("migrations applied successfully")
}
