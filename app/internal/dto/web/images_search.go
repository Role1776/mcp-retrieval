package web

import (
	"github.com/google/uuid"
)

type ImagesSearchRequest struct {
	Queries   []string `json:"queries" jsonschema:"queries in English for image search" validate:"required"`
	MaxImages int      `json:"max_images,omitempty" jsonschema:"maximum number of images per query"`
	TimeoutMs int64    `json:"timeout_ms,omitempty" jsonschema:"timeout for the whole call in milliseconds"`
	Date      string   `json:"date,omitempty" jsonschema:"filter results by freshness: 'd' past day, 'w' past week, 'm' past month, 'y' past year; empty means all time"`
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
