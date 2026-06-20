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

	problem, err := chooseRandomProblem(r.Context(), dbs, user.Codeforces, chosenDifficulty)

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

	problemStatus := db.ProblemStatus{
		ProblemID:    problem,
		UserID:       user.ID,
		Solved:       false,
		SecondsTaken: -1,
		Tracked:      true,
	}

	err = auth.InsertOrUpdateProblemStatus(r.Context(), problemStatus)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "failed to update database"})
		return
	}

	cfjson.WriteJSON(w, http.StatusOK, challenge)
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

	user, err := auth.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "authentication required"})
		return
	}

	id, b := cfjson.DecodeChallengeUpdate(w, r)
	if !b {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "You did not provide challenge data"})
		return
	}

	// obfuscate challenge data & user data
	challenge, exists, err := auth.ChallengeByID(r.Context(), id)
	if !exists || challenge.UserID != user.ID {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "Invalid challenge data"})
		return
	}

	// check if user solved challenge
	submissions, err := cf.GetRecentSubmissions(r.Context(), 5, user.Codeforces)

	if err != nil || len(submissions) == 0 {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "You have no submissions"})
		return
	}

	trackedSubmission := submissions[0]

	flag := false
	for _, submission := range submissions {
		if submission.Problem.ID == challenge.ProblemID && submission.Verdict == "OK" {
			flag = true
			trackedSubmission = submission
		}
	}

	if !flag {
		cfjson.WriteJSON(w, http.StatusOK, cfjson.ErrorResponse{Error: "You have not solved the challenge!"})
		return
	}

	submissionTime := time.Unix(trackedSubmission.CreationTime, 0).UTC()

	if submissionTime.After(challenge.ExpiryTime) {
		cfjson.WriteJSON(w, http.StatusOK, cfjson.ErrorResponse{Error: "Good job, but you were a little too slow!"})
		return
	}

	problemStatus := db.ProblemStatus{
		ProblemID:    challenge.ProblemID,
		UserID:       user.ID,
		Solved:       true,
		SecondsTaken: int64(submissionTime.Sub(challenge.CreationTime).Seconds()),
		Tracked:      true,
	}

	err = auth.InsertOrUpdateProblemStatus(r.Context(), problemStatus)

	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, cfjson.ErrorResponse{Error: "congrats, but I failed to update the database lol"})
		return
	}

	cfjson.WriteJSON(w, http.StatusOK, messageResponse{Message: "congrats!"})
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

func chooseRandomProblem(ctx context.Context, dbs *sql.DB, codeforcesUsername string, rating int) (string, error) {
	problems, err := ProblemsByRating(ctx, dbs, rating)

	if err != nil {
		return "", err
	}

	if len(problems) <= 0 {
		return "", errors.New("no problems with that rating")
	}

	solvedProblems := []string{}

	submissions, err := cf.GetRecentSubmissions(ctx, -1, codeforcesUsername)
	if err != nil {
		return "", err
	}

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
