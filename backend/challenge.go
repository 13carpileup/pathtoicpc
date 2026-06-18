package backend

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
	"net/http"
	cf "pathtoicpc/backend/codeforces"
	"pathtoicpc/backend/db"
	cfjson "pathtoicpc/backend/json"
	"slices"
	"time"
)

// generates a challenge for the user, updates db record, then gives challenge (problem + id) back to user
func GetChallenge(
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

	submissions, err := cf.GetRecentSubmissions(r.Context(), -1, user.Codeforces)

	challengeType, ok := cfjson.DecodeChallengeRequest(w, r)
	if !ok {
		return
	}

	ratingEstimate := 1000

	if user.RatingEstimate != 0 {
		ratingEstimate = user.RatingEstimate
	}

	switch challengeType {
	case "EASY":
		ratingEstimate -= 200
	case "HARD":
		ratingEstimate += 200
	}

	// actually choose rating of the problem

	chosenDifficulty := chooseRandomRating(ratingEstimate)
	chosenDifficulty = max(chosenDifficulty, 800)
	chosenDifficulty = min(chosenDifficulty, 3500)

	// randomly choose problem of said rating

	problem, err := chooseRandomProblem(r.Context(), dbs, submissions, chosenDifficulty)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "something weird happened!"})
		return
	}

	challenge := db.Challenge{
		UserID:       user.ID,
		ProblemID:    problem,
		Solved:       false,
		CreationTime: time.Now(),
		ExpiryTime:   time.Now().Add(time.Minute * 40),
	}

	id, err := auth.InsertChallenge(r.Context(), challenge)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "db insertion failed"})
		return
	}

	challenge.ChallengeID = id

	challengeText, err := cf.GetProblemText(r.Context(), problem)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "failed to get challenge text"})
		return
	}

	challenge.ChallengeText = challengeText

	cfjson.WriteJSON(w, http.StatusOK, challenge)
	return
}

func UpdateChallenge(
	dbs *sql.DB,
	auth db.AuthService,
	w http.ResponseWriter,
	r *http.Request,
) {
	if dbs == nil {
		cfjson.WriteJSON(w, http.StatusServiceUnavailable, cfjson.ErrorResponse{Error: "mysql database is not configured"})
		return
	}

	_, err := auth.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "authentication required"})
		return
	}
}

func chooseRandomRating(meanRating int) int {
	randomNum := rand.Float64()
	chosenRating := meanRating

	switch {
	case randomNum <= 0.05:
		chosenRating -= 200
	case randomNum <= 0.25:
		chosenRating -= 100
	case randomNum <= 0.75:
		chosenRating += 0
	case randomNum <= 0.95:
		chosenRating += 100
	case randomNum <= 1.00:
		chosenRating += 200
	}

	return chosenRating
}

func chooseRandomProblem(ctx context.Context, dbs *sql.DB, submissions []cf.CodeforcesSubmission, rating int) (string, error) {
	problems, err := ProblemsByRating(ctx, dbs, rating)

	if err != nil {
		return "", err
	}

	if len(problems) <= 0 {
		return "", errors.New("no problems with that rating")
	}

	solvedProblems := []string{}

	for _, submission := range submissions {
		solvedProblems = append(solvedProblems, submission.Problem.ID)
	}

	problem := problems[rand.IntN(len(problems))]
	c := 0

	for slices.Contains(solvedProblems, problem.ID) {
		problem = problems[rand.IntN(len(problems))]

		c += 1

		if c >= 100 {
			return "", errors.New("solved every problem")
		}
	}

	return problem.ID, nil
}
