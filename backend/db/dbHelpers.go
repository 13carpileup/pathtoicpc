package db

import (
	"context"
	"encoding/json"
	"fmt"
	cf "pathtoicpc/backend/codeforces"
)

type statement struct {
	query string
	args  []any
}

func (s *AuthService) InitializeSchema(ctx context.Context) error {
	if s == nil || s.db == nil {
		return nil
	}

	statements := []statement{
		{query: `DROP TABLE problem_status`},
		{query: `CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			email VARCHAR(255) NOT NULL,
			username VARCHAR(64) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			linked_cf BOOLEAN NOT NULL DEFAULT FALSE,
			cf_account VARCHAR(255),
			rating_estimate INT,

			PRIMARY KEY (id),
			UNIQUE KEY users_email_unique (email),
			UNIQUE KEY users_username_unique (username),
			UNIQUE KEY cf_account_unique (cf_account)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS user_sessions (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NOT NULL,
			token_hash CHAR(64) NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

			PRIMARY KEY (id),
			UNIQUE KEY user_sessions_token_hash_unique (token_hash),
			KEY user_sessions_user_id_index (user_id),
			KEY user_sessions_expires_at_index (expires_at),
			CONSTRAINT user_sessions_user_id_foreign
				FOREIGN KEY (user_id) REFERENCES users(id)
				ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS problems (
			id VARCHAR(255) NOT NULL PRIMARY KEY,
			contest BIGINT UNSIGNED NOT NULL,
			letter VARCHAR(16) NOT NULL,
			rating BIGINT UNSIGNED,
			tags JSON NOT NULL DEFAULT (JSON_ARRAY())
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS submissions (
			submission_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			problem_id VARCHAR(255) NOT NULL,
			solved BOOLEAN NOT NULL,
			status VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS problem_status (
			problem_id VARCHAR(255) NOT NULL,
			user_id BIGINT NOT NULL,
			solved BOOLEAN NOT NULL,
			tracked BOOLEAN NOT NULL,
			seconds_taken BIGINT,
			time_solved TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

			PRIMARY KEY (problem_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS codeforces_linking (
			user_id BIGINT NOT NULL PRIMARY KEY,
			cf_account VARCHAR(255) NOT NULL,
			problem_id VARCHAR(255) NOT NULL,
			creation_time TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS challenges (
			challenge_id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			problem_id VARCHAR(255) NOT NULL,
			solved BOOLEAN NOT NULL,
			creation_time TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},

		{query: `CREATE TABLE IF NOT EXISTS rating_updates (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			rating_estimate int NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},
	}

	problems, err := cf.GetProblemList(ctx)

	if err != nil {
		return err
	}

	for _, problem := range problems {
		tags, err := json.Marshal(problem.Tags)
		if err != nil {
			return fmt.Errorf("encode tags for problem %s: %w", problem.ID, err)
		}

		statements = append(statements, statement{
			query: `INSERT INTO problems (id, contest, letter, rating, tags) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE contest=contest`,
			args:  []any{problem.ID, problem.ContestID, problem.Index, problem.Rating, string(tags)},
		})
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			return err
		}
	}

	return nil
}
