package cf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	cfjson "pathtoicpc/backend/json"
)

const codeforcesAPIBaseURL = "https://codeforces.com/api/"

var codeforcesHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	proxyCodeforces(w, r, "user.info", "handles")
}

func GetUserStatus(w http.ResponseWriter, r *http.Request) {
	proxyCodeforces(w, r, "user.status", "handle")
}

func GetProblemsetProblems(w http.ResponseWriter, r *http.Request) {
	proxyCodeforces(w, r, "problemset.problems")
}

func GetProblemList(ctx context.Context) ([]CodeforcesProblem, error) {
	body, status, _, err := requestCodeforces(ctx, "problemset.problems", nil)
	if err != nil {
		return nil, err
	}

	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("codeforces problemset.problems returned status %d", status)
	}

	var response CodeforcesAPIResponse[CodeforcesProblemListResult]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode codeforces problem list: %w", err)
	}

	if response.Status != "OK" {
		if response.Comment != "" {
			return nil, fmt.Errorf("codeforces problemset.problems failed: %s", response.Comment)
		}

		return nil, fmt.Errorf("codeforces problemset.problems failed with status %q", response.Status)
	}

	for i := range response.Result.Problems {
		addID(&response.Result.Problems[i])
	}

	return response.Result.Problems, nil
}

func getRecommendedProblem(w http.ResponseWriter, r *http.Request) {

}

func proxyCodeforces(w http.ResponseWriter, r *http.Request, method string, requiredParams ...string) {
	query := cleanQuery(r.URL.Query())
	if err := requireQueryParams(query, requiredParams...); err != nil {
		cfjson.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	body, status, contentType, err := requestCodeforces(r.Context(), method, query)
	if err != nil {
		log.Printf("codeforces %s request failed: %v", method, err)
		cfjson.WriteJSON(w, http.StatusBadGateway, errorResponse{Error: "failed to call Codeforces API"})
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
