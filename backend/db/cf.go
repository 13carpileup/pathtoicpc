package db

import (
	"context"
	"errors"
	"time"

	cf "pathtoicpc/backend/codeforces"
)

func (s *AuthService) InsertCodeforcesIntegration(ctx context.Context, userID int, codeforcesName string, problemID string) (time.Time, error) {
	if s == nil || s.db == nil {
		return time.Time{}, errors.New("mysql database is not configured")
	}

	nowUTC := time.Now().UTC()
	expiryTime := nowUTC.Add(10 * time.Minute)

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO codeforces_linking (user_id, cf_account, problem_id, expires_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			cf_account = VALUES(cf_account),
			problem_id = VALUES(problem_id),
			expires_at = VALUES(expires_at)`,
		userID,
		codeforcesName,
		problemID,
		expiryTime,
	)

	if err != nil {
		return time.Time{}, err
	}

	return expiryTime, nil
}

func (s *AuthService) GetCodeforcesIntegration(ctx context.Context, userID int) (cf.CodeforcesIntegration, error) {
	if s == nil || s.db == nil {
		return cf.CodeforcesIntegration{}, errors.New("mysql database is not configured")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT user_id, cf_account, problem_id, expires_at
		FROM codeforces_linking
		WHERE user_id = ?`,
		userID,
	)

	if err != nil {
		return cf.CodeforcesIntegration{}, err
	}

	defer rows.Close()

	var integrations []cf.CodeforcesIntegration
	for rows.Next() {
		var integration cf.CodeforcesIntegration

		if err := rows.Scan(&integration.UserID, &integration.CfAccount, &integration.ProblemID, &integration.ExpiryTime); err != nil {
			return cf.CodeforcesIntegration{}, err
		}

		integrations = append(integrations, integration)
	}

	if err := rows.Err(); err != nil {
		return cf.CodeforcesIntegration{}, err
	}

	if len(integrations) == 0 {
		return cf.CodeforcesIntegration{}, errors.New("No valid integrations to verify")
	}

	return integrations[0], nil
}
