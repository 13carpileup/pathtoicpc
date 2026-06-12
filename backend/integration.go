package backend

import (
	"database/sql"
	"net/http"

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

type integrationErrorResponse struct {
	Error string `json:"error"`
}

func HandleCodeforcesIntegration(
	dbs *sql.DB,
	auth db.AuthService,
	w http.ResponseWriter,
	r *http.Request,
) {
	if dbs == nil {
		cfjson.WriteJSON(w, http.StatusServiceUnavailable, integrationErrorResponse{Error: "mysql database is not configured"})
		return
	}

	user, err := auth.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, integrationErrorResponse{Error: "authentication required"})
		return
	}

	_ = user

}
