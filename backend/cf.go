package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const codeforcesAPIBaseURL = "https://codeforces.com/api/"

var codeforcesHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

type errorResponse struct {
	Error string `json:"error"`
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	proxyCodeforces(w, r, "user.info", "handles")
}

func getUserStatus(w http.ResponseWriter, r *http.Request) {
	proxyCodeforces(w, r, "user.status", "handle")
}

func getProblemsetProblems(w http.ResponseWriter, r *http.Request) {
	proxyCodeforces(w, r, "problemset.problems")
}

func getRecommendedProblem(w http.ResponseWriter, r *http.Request) {

}

func proxyCodeforces(w http.ResponseWriter, r *http.Request, method string, requiredParams ...string) {
	query := cleanQuery(r.URL.Query())
	if err := requireQueryParams(query, requiredParams...); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	body, status, contentType, err := requestCodeforces(r.Context(), method, query)
	if err != nil {
		log.Printf("codeforces %s request failed: %v", method, err)
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "failed to call Codeforces API"})
		return
	}

	if contentType == "" {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write codeforces %s response: %v", method, err)
	}
}

func requestCodeforces(ctx context.Context, method string, query url.Values) ([]byte, int, string, error) {
	endpoint, err := url.Parse(codeforcesAPIBaseURL + method)
	if err != nil {
		return nil, 0, "", err
	}

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, 0, "", err
	}

	resp, err := codeforcesHTTPClient.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, "", err
	}

	return body, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func cleanQuery(query url.Values) url.Values {
	cleaned := make(url.Values, len(query))

	for key, values := range query {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				continue
			}

			cleaned.Add(key, value)
		}
	}

	return cleaned
}

func requireQueryParams(query url.Values, requiredParams ...string) error {
	for _, param := range requiredParams {
		if strings.TrimSpace(query.Get(param)) == "" {
			return fmt.Errorf("missing required query parameter %q", param)
		}
	}

	return nil
}
