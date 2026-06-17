package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	cfjson "pathtoicpc/backend/json"
)

const maxJSONBodySize = 1 << 20

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type AuthService struct {
	db              *sql.DB
	sessionDuration time.Duration
}

type ProblemFetcher func(context.Context) ([]Problem, error)

type Problem struct {
	ID        string
	ContestID int
	Index     string
	Rating    int
	Tags      []string
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type userRecord struct {
	ID            int64
	Email         string
	Username      string
	PasswordHash  string
	CreatedAt     time.Time
	HasCodeforces bool
	Codeforces    string
}

type UserRecord = userRecord

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
}

type authResponse struct {
	User      userResponse `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		db:              db,
		sessionDuration: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if !s.ensureEnabled(w) {
		return
	}

	var req registerRequest
	if !decodeJSONRequest(w, r, &req) {
		return
	}

	email, username, err := normalizeRegistration(req)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to secure password"})
		return
	}

	result, err := s.db.ExecContext(
		r.Context(),
		`INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)`,
		email,
		username,
		string(passwordHash),
	)
	if err != nil {
		if isDuplicateEntry(err) {
			cfjson.WriteJSON(w, http.StatusConflict, errorResponse{Error: "email or username is already registered"})
			return
		}

		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create account"})
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load account"})
		return
	}

	user, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load account"})
		return
	}

	token, expiresAt, err := s.createSession(r.Context(), user.ID)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to start session"})
		return
	}

	cfjson.WriteJSON(w, http.StatusCreated, authResponse{
		User:      toUserResponse(user),
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (s *AuthService) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.ensureEnabled(w) {
		return
	}

	var req loginRequest
	if !decodeJSONRequest(w, r, &req) {
		return
	}

	identifier := normalizeLoginIdentifier(req)
	if identifier == "" || req.Password == "" {
		cfjson.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "identifier and password are required"})
		return
	}

	user, err := s.getUserByIdentifier(r.Context(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			cfjson.WriteJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load account"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}

	token, expiresAt, err := s.createSession(r.Context(), user.ID)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to start session"})
		return
	}

	cfjson.WriteJSON(w, http.StatusOK, authResponse{
		User:      toUserResponse(user),
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (s *AuthService) HandleMe(w http.ResponseWriter, r *http.Request) {
	if !s.ensureEnabled(w) {
		return
	}

	user, err := s.UserFromRequest(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	cfjson.WriteJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *AuthService) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if !s.ensureEnabled(w) {
		return
	}

	token, err := bearerToken(r)
	if err != nil {
		cfjson.WriteJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	if _, err := s.db.ExecContext(
		r.Context(),
		`DELETE FROM user_sessions WHERE token_hash = ?`,
		hashSessionToken(token),
	); err != nil {
		cfjson.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to end session"})
		return
	}

	cfjson.WriteJSON(w, http.StatusOK, messageResponse{Message: "Logged out."})
}

func (s *AuthService) ensureEnabled(w http.ResponseWriter) bool {
	if s != nil && s.db != nil {
		return true
	}

	cfjson.WriteJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "mysql database is not configured"})
	return false
}

func (s *AuthService) getUserByID(ctx context.Context, userID int64) (userRecord, error) {
	var user userRecord
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, email, username, password_hash, created_at FROM users WHERE id = ?`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

func (s *AuthService) getUserByIdentifier(ctx context.Context, identifier string) (userRecord, error) {
	var user userRecord
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE email = ? OR username = ?
		LIMIT 1`,
		strings.ToLower(identifier),
		identifier,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

func (s *AuthService) UserFromRequest(r *http.Request) (userRecord, error) {
	token, err := bearerToken(r)
	if err != nil {
		return userRecord{}, err
	}

	var user userRecord
	err = s.db.QueryRowContext(
		r.Context(),
		`SELECT users.id, users.email, users.username, users.password_hash, users.created_at, users.linked_cf, users.cf_account
		FROM user_sessions
		INNER JOIN users ON users.id = user_sessions.user_id
		WHERE user_sessions.token_hash = ? AND user_sessions.expires_at > ?
		LIMIT 1`,
		hashSessionToken(token),
		time.Now().UTC(),
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.HasCodeforces, &user.Codeforces)

	return user, err
}

func (s *AuthService) createSession(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().UTC().Add(s.sessionDuration)
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO user_sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID,
		hashSessionToken(token),
		expiresAt,
	); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func normalizeRegistration(req registerRequest) (string, string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	username := strings.TrimSpace(req.Username)

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return "", "", errors.New("a valid email is required")
	}

	if len(username) < 3 || len(username) > 64 || !usernamePattern.MatchString(username) {
		return "", "", errors.New("username must be 3-64 letters, numbers, or underscores")
	}

	if len(req.Password) < 8 {
		return "", "", errors.New("password must be at least 8 characters")
	}

	if len([]byte(req.Password)) > 72 {
		return "", "", errors.New("password must be 72 bytes or fewer")
	}

	return email, username, nil
}

func normalizeLoginIdentifier(req loginRequest) string {
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}
	return identifier
}

func toUserResponse(user userRecord) userResponse {
	return userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
}

func decodeJSONRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		cfjson.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request"})
		return false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		cfjson.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "request body must contain one JSON object"})
		return false
	}

	return true
}

func bearerToken(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", errors.New("missing bearer token")
	}

	return token, nil
}

func newSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("create session token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
