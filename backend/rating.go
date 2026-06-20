package backend

import (
	"database/sql"
	"net/http"
	"pathtoicpc/backend/db"
)

func EstimateRating(
	dbs *sql.DB,
	auth db.AuthService,
	w http.ResponseWriter,
	r *http.Request,
) {

}
