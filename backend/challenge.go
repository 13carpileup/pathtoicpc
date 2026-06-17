package backend

import (
	"database/sql"
	"net/http"
	"pathtoicpc/backend/db"
	cfjson "pathtoicpc/backend/json"
)

func GetChallenge(
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
}

func UpdateChallenge(
	dbs *sql.DB,
	auth db.AuthService,
	w http.ResponseWriter,
	r *http.Request,
) {
	if dbs == nil {
		cfjson.WriteJSON(w, http.StatusServiceUnavailable, integrationErrorResponse{Error: "mysql database is not configured"})
		return
	}

	_, err := auth.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, integrationErrorResponse{Error: "authentication required"})
		return
	}
}
