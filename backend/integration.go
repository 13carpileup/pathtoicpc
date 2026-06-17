package backend

import (
	"database/sql"
	"math/rand/v2"
	"net/http"
	"time"

	cf "pathtoicpc/backend/codeforces"
	"pathtoicpc/backend/db"
	cfjson "pathtoicpc/backend/json"
)

// for reference
// type userRecord struct {
// 	ID           int64
// 	Email        string
// 	Username     string
// 	PasswordHash string
// 	CreatedAt    time.Time
// }

type integrationData struct {
	Problem string    `json:"problem"`
	Expiry  time.Time `json:"expiresAt"`
}

func HandleCodeforcesIntegration(
	dbs *sql.DB,
	auth db.AuthService,
	w http.ResponseWriter,
	r *http.Request,
) {
	if dbs == nil {
		cfjson.WriteJSON(w, http.StatusServiceUnavailable, cfjson.ErrorResponse{Error: "mysql database is not configured"})
		return
	}

	user, err := auth.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "authentication required"})
		return
	}

	codeforcesUsername, ok := cfjson.DecodeIntegrationRequest(w, r)
	if !ok {
		return
	}

	problems, err := auth.ProblemsByX(r.Context(),
		`SELECT id, contest, letter, rating, tags
		FROM problems`,
		[]any{},
	)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusServiceUnavailable, cfjson.ErrorResponse{Error: "error selecting problem"})
		return
	}

	randomIndex := rand.IntN(len(problems))
	randomProblem := problems[randomIndex]

	// actually insert data into integration db

	expiryTime, err := auth.InsertCodeforcesIntegration(r.Context(), user.ID, codeforcesUsername, randomProblem.ID)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnprocessableEntity, cfjson.ErrorResponse{Error: "failed to create integration listing in db, check logs"})
		return
	}

	// giving user the data they need
	cfjson.WriteJSON(w, http.StatusOK, integrationData{
		Problem: randomProblem.ID,
		Expiry:  expiryTime,
	})
}

func VerifyCodeforcesIntegration(
	dbs *sql.DB,
	auth db.AuthService,
	w http.ResponseWriter,
	r *http.Request,
) {
	if dbs == nil {
		cfjson.WriteJSON(w, http.StatusServiceUnavailable, cfjson.ErrorResponse{Error: "mysql database is not configured"})
		return
	}

	user, err := auth.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "authentication required"})
		return
	}

	integration, err := auth.GetCodeforcesIntegration(r.Context(), user.ID)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnprocessableEntity, cfjson.ErrorResponse{Error: "failed to load codeforces integration"})
		return
	}

	submissions, err := cf.GetRecentSubmissions(r.Context(), 1, integration.CfAccount)
	if len(submissions) == 0 {
		cfjson.WriteJSON(w, http.StatusUnprocessableEntity, cfjson.ErrorResponse{Error: "failed to get recent submissions"})
		return
	}

	firstSubmission := submissions[0]
	submissionTime := time.Unix(firstSubmission.CreationTime, 0).UTC()

	if integration.ExpiryTime.Compare(submissionTime) == -1 || integration.CreationTime.Compare(submissionTime) == 1 || firstSubmission.Problem.ID != integration.ProblemID {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "verification failed"})
		return
	}

	err = auth.UpdateCfAccount(r.Context(), user.ID, integration.CfAccount)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnprocessableEntity, cfjson.ErrorResponse{Error: "failed to update username"})
		return
	}

	// giving user the data they need
	cfjson.WriteJSON(w, http.StatusOK, messageResponse{Message: "updated cf username"})
}
