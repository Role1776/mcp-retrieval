package web

import (
	"github.com/google/uuid"
)

type SearchRequest struct {
	Queries    []string `json:"queries" jsonschema:"1-10 search queries in English, run in parallel; each must be non-empty and at most 512 characters, otherwise the whole call is rejected" validate:"required"`
	MaxResults int      `json:"max_results,omitempty" jsonschema:"maximum snippets per query, not in total; default 5, values above 20 are clamped to 20"`
	TimeoutMs  int64    `json:"timeout_ms,omitempty" jsonschema:"timeout for the whole call in milliseconds, shared by all queries; default 5000, clamped to 1000-10000; unfinished queries get status 'timeout'"`
	Date       string   `json:"date,omitempty" jsonschema:"freshness filter: 'd' past day, 'w' past week, 'm' past month, 'y' past year; empty means all time; use only these values"`
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
