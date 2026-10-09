package web

import (
	"github.com/google/uuid"
)

type ImagesSearchRequest struct {
	Queries   []string `json:"queries" jsonschema:"1-10 image search queries in English, run in parallel; each must be non-empty and at most 512 characters, otherwise the whole call is rejected" validate:"required"`
	MaxImages int      `json:"max_images,omitempty" jsonschema:"maximum images per query, not in total; default 5, values above 10 are clamped to 10"`
	TimeoutMs int64    `json:"timeout_ms,omitempty" jsonschema:"timeout for the whole call in milliseconds, shared by all queries; default 5000, clamped to 1000-10000; unfinished queries get status 'timeout'"`
	Date      string   `json:"date,omitempty" jsonschema:"freshness filter: 'd' past day, 'w' past week, 'm' past month, 'y' past year; empty means all time; other values are ignored"`
}

type ImagesSearchResponse struct {
	Results  []ImagesResult `json:"results"`
	Metadata ImagesMetadata `json:"metadata"`
}

type ImagesMetadata struct {
	TotalRequestTimeMs int64  `json:"total_request_time_ms"`
	Date               string `json:"date"`
}

type ImagesResult struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Count       int     `json:"count"`
	Images      []Image `json:"images"`
	TotalTimeMs int64   `json:"total_time_ms"`
}

type Image struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	PageURL     string    `json:"page_url"`
	Description string    `json:"description"`
}
