package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"edu-platform/internal/db"
	"edu-platform/internal/server"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// create DB pool
	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer pool.Close()

	srv, err := server.New(ctx, pool)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	s := &http.Server{
		Addr:         addr,
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("listening %s", addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
