package db

import (
	"context"
	"errors"
	"time"
)

type Challenge struct {
	ChallengeID  int64
	UserID       int64
	ProblemID    string
	Solved       bool
	CreationTime time.Time
	ExpiryTime   time.Time
}

// creates challenge in db and returns its id to you for safe keeping n shi
func (s *AuthService) InsertChallenge(ctx context.Context, challenge Challenge) (int64, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("mysql database is not configured")
	}

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO challenges (user_id, problem_id, solved, creation_time, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		challenge.UserID, challenge.ProblemID, false, challenge.CreationTime, challenge.ExpiryTime,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// returns the challenge with the given id if such a challenge exists, else returns false in 2nd return value
func (s *AuthService) ChallengeByID(ctx context.Context, challengeID int64) (Challenge, bool, error) {
	if s == nil {
		return Challenge{}, false, errors.New("mysql database is not configured")
	}

	challenges, err := s.ChallengesByX(ctx,
		`SELECT challenge_id, user_id, problem_id, solved, creation_time, expires_at
		FROM challenges
		WHERE challenge_id = ?`,
		[]any{challengeID},
	)

	if err != nil {
		return Challenge{}, false, err
	}

	if len(challenges) == 0 {
		return Challenge{}, false, nil
	}

	return challenges[0], true, nil
}

func (s *AuthService) ChallengesByUser(ctx context.Context, userID int64) ([]Challenge, error) {
	if s == nil {
		return nil, errors.New("mysql database is not configured")
	}

	challenges, err := s.ChallengesByX(ctx,
		`SELECT challenge_id, user_id, problem_id, solved, creation_time, expires_at
		FROM challenges
		WHERE user_id = ?`,
		[]any{userID},
	)

	if err != nil {
		return nil, err
	}

	return challenges, nil
}

func (s *AuthService) ChallengesByX(ctx context.Context, query string, args []any) ([]Challenge, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("mysql database is not configured")
	}

	rows, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var challenges []Challenge
	for rows.Next() {
		var challenge Challenge

		if err := rows.Scan(&challenge.ChallengeID, &challenge.UserID, &challenge.ProblemID, &challenge.Solved, &challenge.CreationTime, &challenge.ExpiryTime); err != nil {
			return nil, err
		}

		challenges = append(challenges, challenge)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return challenges, nil
}
