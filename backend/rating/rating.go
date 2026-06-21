package rating

import (
	"database/sql"
	"fmt"
	"maps"
	"math"
	"net/http"
	"pathtoicpc/backend/db"
	cfjson "pathtoicpc/backend/json"
	"time"
)

type RatingEstimate struct {
	RatingEstimate int `json:"rating_estimate"`
	Uncertainty    int `json:"uncertainty"`
}

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

	if len(problemList) == 0 {
		cfjson.WriteJSON(w, http.StatusUnauthorized, RatingEstimate{RatingEstimate: 1000, Uncertainty: 1000})
		return
	}

	statusByRating := getStatusByRating(problemList, time.Now().Add(-time.Hour*24*30))

	ratingEstimate, uncertainty := getRating(statusByRating)

	cfjson.WriteJSON(w, http.StatusUnauthorized, RatingEstimate{RatingEstimate: ratingEstimate, Uncertainty: uncertainty})
}

func getRating(statusMap map[int][]db.ProblemStatus) (int, int) {
	independentProbability := make(map[int]float64)
	precision := 10

	for i := range 23 {
		problemRating := 800 + i*100
		independentProbability[problemRating] = 0

		for j := range 23 * precision {
			userRating := 800 + j*100/precision

			independentProbability[problemRating] += ProbOfUserRating(userRating, precision) * ProbOfSolvingGivenRating(userRating, problemRating)
		}
	}

	probOfUserRating := make(map[int]float64)

	for j := range 23 * precision {
		userRating := 800 + j*100/precision

		probOfUserRating[userRating] = ProbOfUserRating(userRating, precision)
	}

	for rating := range maps.Keys(statusMap) {
		for _, problemStatus := range statusMap[rating] {
			for j := range 23 * precision {
				userRating := 800 + j*100/precision
				probSolved := ProbOfSolvingGivenRating(userRating, rating)

				if !problemStatus.Solved {
					probSolved = 1 - probSolved
				}

				// bayes
				probOfUserRating[userRating] = probSolved * probOfUserRating[userRating] / independentProbability[rating]
			}
		}
	}

	var expectedRating float64 = 0
	for rating := range probOfUserRating {
		fmt.Printf("RATING %d has prob %f\n", rating, probOfUserRating[rating])
		expectedRating += float64(rating) * probOfUserRating[rating]
	}

	return int(expectedRating), 500
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

// returns the probability of any given user having a rating between userRating and userRating + precision
func ProbOfUserRating(userRating int, precision int) float64 {
	var std float64 = 350
	var mean float64 = 1200

	return NormalCDF(float64(userRating), float64(userRating+precision), std, mean)
}

// returns the probability that a user with the given rating can correctly solve the problem with the given rating
func ProbOfSolvingGivenRating(userRating int, problemRating int) float64 {
	delta := float64(userRating - problemRating)

	L := 1.00
	K := 0.01

	return L / (1 + math.Exp(-K*delta))
}

func NormalCDF(left float64, right float64, std float64, mean float64) float64 {
	return 0.5*math.Erf(1/math.Sqrt(2)*(right-mean)/std) - 0.5*math.Erf(1/math.Sqrt(2)*(left-mean)/std)
}
