package web

import (
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
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
	Query       string       `json:"query"`
	Status      string       `json:"status"`
	Count       int          `json:"count"`
	Snippets    web.Snippets `json:"snippets"`
	TotalTimeMs int64        `json:"total_time_ms"`
}

type SearchMetadata struct {
	TotalRequestTimeMs int64  `json:"total_request_time_ms"`
	Date               string `json:"date"`
}
