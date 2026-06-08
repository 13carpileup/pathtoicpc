package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}
	if db != nil {
		defer db.Close()
	}

	auth := newAuthService(db)
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := auth.initializeSchema(ctx); err != nil {
			log.Fatalf("database schema setup failed: %v", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/message", handleMessage)
	mux.HandleFunc("POST /api/auth/register", auth.handleRegister)
	mux.HandleFunc("POST /api/auth/login", auth.handleLogin)
	mux.HandleFunc("GET /api/auth/me", auth.handleMe)
	mux.HandleFunc("POST /api/auth/logout", auth.handleLogout)
	mux.HandleFunc("GET /api/cf", testApi)
	mux.HandleFunc("GET /api/user.info", getUserInfo)
	mux.HandleFunc("GET /api/user.status", getUserStatus)
	mux.HandleFunc("GET /api/problemset.problems", getProblemsetProblems)

	addr := ":" + getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         addr,
		Handler:      withCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("backend listening on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
		Time:   time.Now().UTC(),
	})
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, messageResponse{
		Message: "Hello from the Go backend.",
	})
}

func testApi(w http.ResponseWriter, r *http.Request) {
	params := []Param{Param{key: "hu", value: "ho"}}

	writeJSON(w, http.StatusOK, messageResponse{
		Message: getSig("hello", params),
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
