package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"pathtoicpc/backend/db"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func InitializeSchema(ctx context.Context, dbs *sql.DB) error {
	return db.NewAuthService(dbs, fetchCodeforcesProblems).InitializeSchema(ctx)
}

func NewHandler(dbs *sql.DB) http.Handler {
	auth := db.NewAuthService(dbs)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/message", handleMessage)
	mux.HandleFunc("POST /api/auth/register", auth.HandleRegister)
	mux.HandleFunc("POST /api/auth/login", auth.HandleLogin)
	mux.HandleFunc("GET /api/auth/me", auth.HandleMe)
	mux.HandleFunc("POST /api/auth/logout", auth.HandleLogout)
	mux.HandleFunc("GET /api/cf", testAPI)
	mux.HandleFunc("GET /api/user.info", getUserInfo)
	mux.HandleFunc("GET /api/user.status", getUserStatus)
	mux.HandleFunc("GET /api/problemset.problems", getProblemsetProblems)

	return withCORS(mux)
}

func ProblemsByRating(ctx context.Context, dbs *sql.DB, rating int) ([]codeforcesProblem, error) {
	problems, err := db.NewAuthService(dbs).ProblemsByRating(ctx, rating)
	if err != nil {
		return nil, err
	}

	return codeforcesProblemsFromDB(problems), nil
}

func fetchCodeforcesProblems(ctx context.Context) ([]db.Problem, error) {
	problems, err := GetProblemList(ctx)
	if err != nil {
		return nil, err
	}

	return dbProblemsFromCodeforces(problems), nil
}

func dbProblemsFromCodeforces(problems []codeforcesProblem) []db.Problem {
	dbProblems := make([]db.Problem, 0, len(problems))
	for _, problem := range problems {
		dbProblems = append(dbProblems, db.Problem{
			ID:        problem.ID,
			ContestID: problem.ContestID,
			Index:     problem.Index,
			Rating:    problem.Rating,
			Tags:      problem.Tags,
		})
	}

	return dbProblems
}

func codeforcesProblemsFromDB(problems []db.Problem) []codeforcesProblem {
	codeforcesProblems := make([]codeforcesProblem, 0, len(problems))
	for _, problem := range problems {
		codeforcesProblems = append(codeforcesProblems, codeforcesProblem{
			ID:        problem.ID,
			ContestID: problem.ContestID,
			Index:     problem.Index,
			Rating:    problem.Rating,
			Tags:      problem.Tags,
		})
	}

	return codeforcesProblems
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
