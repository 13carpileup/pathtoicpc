package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"pathtoicpc/backend"
	"pathtoicpc/backend/db"
)

func main() {
	database, err := db.OpenDatabase()
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}
	if database != nil {
		defer database.Close()
	}

	if database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		if err := backend.InitializeSchema(ctx, database); err != nil {
			log.Fatalf("database schema setup failed: %v", err)
		}
	}

	if database != nil {
		problems, err := backend.ProblemsByRating(context.Background(), database, 1500)
		if err != nil {
			log.Printf("failed to load problems rated 1500: %v", err)
		} else {
			log.Printf("Problems rated 1500: %d", len(problems))
			for _, element := range problems {
				log.Printf("%s", element.ID)
			}
		}
	}

	addr := ":" + getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         addr,
		Handler:      backend.NewHandler(database),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("backend listening on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
