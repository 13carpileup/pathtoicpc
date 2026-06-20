package backend

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	cf "pathtoicpc/backend/codeforces"
	"pathtoicpc/backend/db"
	cfjson "pathtoicpc/backend/json"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func InitializeSchema(ctx context.Context, dbs *sql.DB) error {
	return db.NewAuthService(dbs).InitializeSchema(ctx)
}

func NewHandler(dbs *sql.DB) http.Handler {
	auth := db.NewAuthService(dbs)

	mux := http.NewServeMux()

	// stupid endpoints
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/message", handleMessage)

	// auth endpoints
	mux.HandleFunc("POST /api/auth/register", auth.HandleRegister)
	mux.HandleFunc("POST /api/auth/login", auth.HandleLogin)
	mux.HandleFunc("GET /api/auth/me", auth.HandleMe)
	mux.HandleFunc("POST /api/auth/logout", auth.HandleLogout)

	// cf endpoints
	mux.HandleFunc("GET /api/cf", testAPI)
	mux.HandleFunc("GET /api/user.info", cf.GetUserInfo)
	mux.HandleFunc("GET /api/user.status", cf.GetUserStatus)
	mux.HandleFunc("GET /api/problemset.problems", cf.GetProblemsetProblems)
	mux.HandleFunc("POST /api/connect_cf", func(w http.ResponseWriter, r *http.Request) {
		HandleCodeforcesIntegration(dbs, *auth, w, r)
	})
	mux.HandleFunc("POST /api/verify_cf", func(w http.ResponseWriter, r *http.Request) {
		VerifyCodeforcesIntegration(dbs, *auth, w, r)
	})

	// actual challenge endpoints
	mux.HandleFunc("POST /api/chal", func(w http.ResponseWriter, r *http.Request) {
		GetChallenge(dbs, *auth, w, r)
	})
	mux.HandleFunc("POST /api/chal-update", func(w http.ResponseWriter, r *http.Request) {
		UpdateChallenge(dbs, *auth, w, r)
	})

	log.Printf("%f", NormalCDF(0, 1, 1, 0))

	return withCORS(mux)
}

func ProblemsByRating(ctx context.Context, dbs *sql.DB, rating int) ([]cf.CodeforcesProblem, error) {
	problems, err := db.NewAuthService(dbs).ProblemsByRating(ctx, rating)
	if err != nil {
		return nil, err
	}

	return codeforcesProblemsFromDB(problems), nil
}

func fetchCodeforcesProblems(ctx context.Context) ([]db.Problem, error) {
	problems, err := cf.GetProblemList(ctx)
	if err != nil {
		return nil, err
	}

	return dbProblemsFromCodeforces(problems), nil
}

func dbProblemsFromCodeforces(problems []cf.CodeforcesProblem) []db.Problem {
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

func codeforcesProblemsFromDB(problems []db.Problem) []cf.CodeforcesProblem {
	codeforcesProblems := make([]cf.CodeforcesProblem, 0, len(problems))
	for _, problem := range problems {
		codeforcesProblems = append(codeforcesProblems, cf.CodeforcesProblem{
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
	cfjson.WriteJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
		Time:   time.Now().UTC(),
	})
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	cfjson.WriteJSON(w, http.StatusOK, messageResponse{
		Message: "Hello from the Go backend.",
	})
}

func testAPI(w http.ResponseWriter, r *http.Request) {
	params := []Param{{key: "hu", value: "ho"}}

	cfjson.WriteJSON(w, http.StatusOK, messageResponse{
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
