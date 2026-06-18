package cfjson

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

const maxIntegrationJSONBodySize = 1 << 20

type ErrorResponse struct {
	Error string `json:"error"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type integrationRequest struct {
	CodeforcesUsername string `json:"codeforces_username"`
}

type challengeRequest struct {
	ChallengeType string `json:"challenge_type"` // ENUM: EASY, MEDIUM, or HARD
}

type challengeUpdate struct {
	ChallengeID int64 `json:"challenge_id"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func DecodeIntegrationRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxIntegrationJSONBodySize)

	var req integrationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON request"})
		return "", false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "request body must contain one JSON object"})
		return "", false
	}

	codeforcesUsername := strings.TrimSpace(req.CodeforcesUsername)
	if codeforcesUsername == "" {
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: "codeforces username not provided"})
		return "", false
	}

	return codeforcesUsername, true
}

func DecodeChallengeRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxIntegrationJSONBodySize)

	var req challengeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON request"})
		return "", false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "request body must contain one JSON object"})
		return "", false
	}

	challengeType := strings.TrimSpace(req.ChallengeType)
	if challengeType != "HARD" && challengeType != "MEDIUM" && challengeType != "EASY" {
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: "invalid challenge type provided"})
		return "", false
	}

	return challengeType, true
}

func DecodeChallengeUpdate(w http.ResponseWriter, r *http.Request) (int64, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxIntegrationJSONBodySize)

	var req challengeUpdate
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON request"})
		return 0, false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "request body must contain one JSON object"})
		return 0, false
	}

	return req.ChallengeID, true
}
