package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func InitializeSchema(ctx context.Context, db *sql.DB) error {
	return newAuthService(db).initializeSchema(ctx)
}

func NewHandler(db *sql.DB) http.Handler {
	auth := newAuthService(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/message", handleMessage)
	mux.HandleFunc("POST /api/auth/register", auth.handleRegister)
	mux.HandleFunc("POST /api/auth/login", auth.handleLogin)
	mux.HandleFunc("GET /api/auth/me", auth.handleMe)
	mux.HandleFunc("POST /api/auth/logout", auth.handleLogout)
	mux.HandleFunc("GET /api/cf", testAPI)
	mux.HandleFunc("GET /api/user.info", getUserInfo)
	mux.HandleFunc("GET /api/user.status", getUserStatus)
	mux.HandleFunc("GET /api/problemset.problems", getProblemsetProblems)

	return withCORS(mux)
}

func ProblemsByRating(ctx context.Context, db *sql.DB, rating int) ([]codeforcesProblem, error) {
	return newAuthService(db).getProblemsByRating(ctx, rating)
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

func testAPI(w http.ResponseWriter, r *http.Request) {
	params := []Param{{key: "hu", value: "ho"}}

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
