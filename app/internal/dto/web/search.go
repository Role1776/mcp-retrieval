package web

import (
	"github.com/google/uuid"
)

type SearchRequest struct {
	Queries    []string `json:"queries" jsonschema:"search queries, executed in parallel" validate:"required"`
	MaxResults int      `json:"max_results,omitempty" jsonschema:"maximum number of snippets per query"`
	TimeoutMs  int64    `json:"timeout_ms,omitempty" jsonschema:"timeout for the whole call in milliseconds"`
	Date       string   `json:"date,omitempty" jsonschema:"filter results by freshness: 'd' past day, 'w' past week, 'm' past month, 'y' past year; empty means all time"`
}

type SearchResponse struct {
	Results  []Result       `json:"results"`
	Metadata SearchMetadata `json:"metadata"`
}

type Result struct {
	Query       string    `json:"query"`
	Status      string    `json:"status"`
	Count       int       `json:"count"`
	Snippets    []Snippet `json:"snippets"`
	TotalTimeMs int64     `json:"total_time_ms"`
}

type Snippet struct {
	ID      uuid.UUID `json:"id"`
	Link    string    `json:"link"`
	Title   string    `json:"title"`
	Rank    int       `json:"rank"`
	Source  string    `json:"source"`
	Snippet string    `json:"snippet"`
	Favicon string    `json:"favicon"`
}

type SearchMetadata struct {
	TotalRequestTimeMs int64  `json:"total_request_time_ms"`
	Date               string `json:"date"`
}
