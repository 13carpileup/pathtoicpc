package cf

import (
	"fmt"
	"strconv"
	"time"
)

func addID(p *CodeforcesProblem) {
	p.ID = fmt.Sprintf("%s%s", strconv.Itoa(p.ContestID), p.Index)
}

type errorResponse struct {
	Error string `json:"error"`
}

type CodeforcesAPIResponse[T any] struct {
	Status  string `json:"status"`
	Comment string `json:"comment,omitempty"`
	Result  T      `json:"result"`
}

type CodeforcesProblemListResult struct {
	Problems []CodeforcesProblem `json:"problems"`
}

type CodeforcesProblem struct {
	ID             string
	ContestID      int      `json:"contestId,omitempty"`
	ProblemsetName string   `json:"problemsetName,omitempty"`
	Index          string   `json:"index"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Points         float64  `json:"points,omitempty"`
	Rating         int      `json:"rating,omitempty"`
	Tags           []string `json:"tags"`
}

type CodeforcesSubmission struct {
	ID                  int               `json:"id"`
	ContestID           int               `json:"contestId,omitempty"`
	TimeConsumedMillis  int               `json:"timeConsumedMillis"`
	MemoryConsumedBytes int               `json:"memoryConsumedBytes"`
	Verdict             string            `json:"verdict,omitempty"`
	Problem             CodeforcesProblem `json:"problem"`
	User                string            `json:"author.members[0].handle"`
}

type CodeforcesIntegration struct {
	UserID     int
	CfAccount  string
	ProblemID  string
	ExpiryTime time.Time
}
