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

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func decodeIntegrationRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
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
