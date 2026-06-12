package db

import (
	"context"
	"errors"

	cf "pathtoicpc/backend/codeforces"
)

func (s *AuthService) GetCodeforcesIntegration(ctx context.Context, userID int) (cf.CodeforcesIntegration, error) {
	if s == nil || s.db == nil {
		return cf.CodeforcesIntegration{}, errors.New("mysql database is not configured")
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT FROM codeforces_linking WHERE user_id = ?",
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
