package backend

import (
	"database/sql"
	"maps"
	"net/http"
	"pathtoicpc/backend/db"
	cfjson "pathtoicpc/backend/json"
	"time"
)

func EstimateRating(
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

	problemList, err := auth.ProblemStatusByUser(r.Context(), user.ID)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "db pull failed for some rsn"})
		return
	}

	statusByRating := getStatusByRating(problemList, time.Now().Add(-time.Hour*24*30))
	ratio := getRatio(statusByRating)

}

func getStatusByRating(problemList []db.ProblemStatus, olderThan time.Time) map[int][]db.ProblemStatus {
	curMap := make(map[int][]db.ProblemStatus)

	for _, problem := range problemList {
		if olderThan.After(problem.SolvedAt) {
			continue
		}

		list, exists := curMap[int(problem.Rating)]

		if !exists {
			curMap[int(problem.Rating)] = []db.ProblemStatus{problem}
		} else {
			curMap[int(problem.Rating)] = append(list, problem)
		}
	}

	return curMap
}

func getRatio(problemMap map[int][]db.ProblemStatus) map[int]float64 {
	totalCount := make(map[int]int)
	totalSolved := make(map[int]int)

	for rating := range maps.Keys(problemMap) {
		for _, problem := range problemMap[rating] {
			count, exists := totalCount[rating]

			if !exists {
				totalCount[rating] = 1
			} else {
				totalCount[rating] = count + 1
			}

			if !problem.Solved {
				continue
			}

			count, exists = totalSolved[rating]

			if !exists {
				totalSolved[rating] = 1
			} else {
				totalSolved[rating] = count + 1
			}
		}
	}

	ratio := make(map[int]float64)
	for rating := range maps.Keys(problemMap) {
		ratio[rating] = float64(totalSolved[rating]) / float64(totalCount[rating])
	}

	return ratio
}
