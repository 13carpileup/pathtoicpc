package rating

import (
	"fmt"
	"maps"
	"pathtoicpc/backend/db"
	"time"
)

func TestProblemEstimate() {
	// rating : [total problems, successful problems]
	problems := map[int][]int{
		1300: {10, 6},
		1900: {3, 3},
		2100: {2, 2},
	}

	problemList := generateProblemList(problems)

	for _, problem := range problemList {
		fmt.Printf("problem status: %b\n", problem.Solved)
	}

	statusByRating := getStatusByRating(problemList, time.Now().Add(-time.Hour*24*30))

	ratingEstimate, uncertainty := getRating(statusByRating)

	fmt.Printf("\nRATING TEST: %d, %d\n", ratingEstimate, uncertainty)
}

func generateProblemList(ratingsMap map[int][]int) []db.ProblemStatus {
	var currentProblemList []db.ProblemStatus

	for rating := range maps.Keys(ratingsMap) {
		for i := range ratingsMap[rating][0] {
			success := true

			if i >= ratingsMap[rating][1] {
				success = false
			}

			currentProblemList = append(currentProblemList, db.ProblemStatus{
				ProblemID:    "x",
				UserID:       1,
				Solved:       success,
				Tracked:      true,
				SecondsTaken: 1,
				SolvedAt:     time.Now(),
				Rating:       int64(rating),
			})
		}
	}

	return currentProblemList
}
