package cf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// gets all n recent submissions. set n = -1 to fetch all submissions.
func GetRecentSubmissions(ctx context.Context, numSubmissions int, user string) ([]CodeforcesSubmission, error) {
	query := url.Values{}
	query.Set("handle", user)

	if numSubmissions != -1 {
		query.Set("count", strconv.Itoa(numSubmissions))
	}

	body, status, _, err := requestCodeforces(ctx, "user.status", query)

	if err != nil {
		return nil, err
	}

	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("codeforces problemset.problems returned status %d", status)
	}

	var response CodeforcesAPIResponse[[]CodeforcesSubmission]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode codeforces problem list: %w", err)
	}

	if response.Status != "OK" {
		if response.Comment != "" {
			return nil, fmt.Errorf("codeforces problemset.problems failed: %s", response.Comment)
		}

		return nil, fmt.Errorf("codeforces problemset.problems failed with status %q", response.Status)
	}

	for i := range response.Result {
		addID(&response.Result[i].Problem)
	}

	return response.Result, nil
}
