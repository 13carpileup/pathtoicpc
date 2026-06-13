package cf

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const codeforcesProblemBaseURL = "https://codeforces.com/problemset/problem/"

var problemIDPattern = regexp.MustCompile(`^(\d+)/?([A-Za-z][A-Za-z0-9]*)$`)

func GetProblemText(ctx context.Context, problemID string) (string, error) {
	problemURL, err := codeforcesProblemURL(problemID)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, problemURL, nil)
	if err != nil {
		return "", err
	}

	setCodeforcesScrapingHeaders(req)

	resp, err := codeforcesHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("codeforces problem page returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	problem, err := doc.Find(".problem-statement").First().Html()
	if err != nil {
		return "", err
	}

	return string(problem), nil
}

func codeforcesProblemURL(problemID string) (string, error) {
	problemID = strings.TrimSpace(problemID)
	matches := problemIDPattern.FindStringSubmatch(problemID)
	if matches == nil {
		return "", fmt.Errorf("invalid codeforces problem id %q", problemID)
	}

	return codeforcesProblemBaseURL + matches[1] + "/" + matches[2], nil
}

func setCodeforcesScrapingHeaders(req *http.Request) {
	req.Host = "codeforces.com"
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="125", "Chromium";v="125", "Not.A/Brand";v="24"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
}
