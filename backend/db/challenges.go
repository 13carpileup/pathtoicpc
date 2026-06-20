package db

import (
	"context"
	"errors"
	"time"
)

type Challenge struct {
	ChallengeID   int64     `json:"challenge_id"`
	UserID        int64     `json:"user_id"`
	ProblemID     string    `json:"problem_id"`
	Solved        bool      `json:"solved"`
	CreationTime  time.Time `json:"creation_time"`
	ExpiryTime    time.Time `json:"expiry_time"`
	ChallengeText string    `json:"challenge_text"`
}

type ProblemStatus struct {
	ProblemID    string
	UserID       int64
	Solved       bool
	Tracked      bool
	SecondsTaken int64
	SolvedAt     time.Time
	Rating       int64
}

// creates challenge in db and returns its id to you for safe keeping n shi
func (s *AuthService) InsertChallenge(ctx context.Context, challenge Challenge) (int64, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("mysql database is not configured")
	}

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO challenges (user_id, problem_id, solved, creation_time, expires_at)
		VALUES (?, ?, ?, ?, ?)`,
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

func (s *AuthService) InsertOrUpdateProblemStatus(ctx context.Context, problemStatus ProblemStatus) error {
	if s == nil {
		return errors.New("mysql database is not configured")
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO problem_status (problem_id, user_id, solved, tracked, seconds_taken, time_solved)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE solved=VALUES(solved), tracked=VALUES(tracked), seconds_taken=VALUES(seconds_taken)`,
		problemStatus.ProblemID, problemStatus.UserID, problemStatus.Solved, problemStatus.Tracked, problemStatus.SecondsTaken, problemStatus.SolvedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ProblemStatusByUser(ctx context.Context, userID int64) ([]ProblemStatus, error) {
	if s == nil {
		return nil, errors.New("mysql database is not configured")
	}

	challenges, err := s.ProblemStatusByX(ctx,
		`SELECT problem_status.problem_id, problem_status.user_id, problem_status.solved, problem_status.tracked, problem_status.seconds_taken, problems.rating
		FROM problem_status
		INNER JOIN problems ON problem_status.problem_id = problems.id
		WHERE problem_status.user_id = ?`,
		[]any{userID},
	)

	if err != nil {
		return nil, err
	}

	return challenges, nil
}

func (s *AuthService) ProblemStatusByX(ctx context.Context, query string, args []any) ([]ProblemStatus, error) {
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

	var problems []ProblemStatus
	for rows.Next() {
		var problem ProblemStatus

		if err := rows.Scan(&problem.ProblemID, &problem.UserID, &problem.Solved, &problem.Tracked, &problem.SecondsTaken, &problem.Rating); err != nil {
			return nil, err
		}

		problems = append(problems, problem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return problems, nil
}
