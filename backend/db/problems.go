package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

func (s *AuthService) ProblemsByRating(ctx context.Context, rating int) ([]Problem, error) {
	if rating < 0 {
		return nil, errors.New("rating must be non-negative")
	}

	problems, err := s.ProblemsByX(ctx,
		`SELECT id, contest, letter, rating, tags
		FROM problems
		WHERE rating = ?
		ORDER BY contest, letter`,
		[]any{rating},
	)

	if err != nil {
		return nil, err
	}

	return problems, nil
}

func (s *AuthService) ProblemsByX(ctx context.Context, query string, args []any) ([]Problem, error) {
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

	var problems []Problem
	for rows.Next() {
		var problem Problem
		var tagsJSON []byte

		if err := rows.Scan(&problem.ID, &problem.ContestID, &problem.Index, &problem.Rating, &tagsJSON); err != nil {
			return nil, err
		}

		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &problem.Tags); err != nil {
				return nil, fmt.Errorf("decode tags for problem %s: %w", problem.ID, err)
			}
		}

		problems = append(problems, problem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return problems, nil
}
