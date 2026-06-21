package db

import (
	"context"
	"errors"
	"time"
)

// id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
// 			user_id BIGINT NOT NULL,
// 			timestamp TIMESTAMP NOT NULL,
// 			rating_estimate int NOT NULL

type RatingUpdate struct {
	UserID         int64     `json:"user_id"`
	Timestamp      time.Time `json:"timestamp"`
	RatingEstimate int       `json:"rating_estimate"`
}

func (s *AuthService) UpdateUserRating(ctx context.Context, userID int64, newRating int) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE users SET rating_estimate = ? where user_id = ?`,
		newRating, userID,
	)

	if err != nil {
		err = s.InsertRatingEstimate(ctx, RatingUpdate{UserID: userID, Timestamp: time.Now(), RatingEstimate: newRating})
	}

	return err
}

func (s *AuthService) InsertRatingEstimate(ctx context.Context, update RatingUpdate) error {
	if s == nil || s.db == nil {
		return errors.New("mysql database is not configured")
	}
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO rating_updates (user_id, timestamp, rating_estimate) VALUES (?, ?, ?)",
		update.UserID, update.Timestamp, update.RatingEstimate,
	)

	return err
}

func (s *AuthService) GetRatingEstimatesByUser(ctx context.Context, userID int64) ([]RatingUpdate, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("mysql database is not configured")
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT user_id, timestamp, rating_estimate FROM rating_updates WHERE user_id = ?",
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var updates []RatingUpdate
	for rows.Next() {
		var update RatingUpdate

		if err := rows.Scan(&update.UserID, &update.Timestamp, &update.RatingEstimate); err != nil {
			return nil, err
		}

		updates = append(updates, update)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return updates, nil
}
