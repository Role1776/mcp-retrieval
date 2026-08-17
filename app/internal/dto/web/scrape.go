package web

import (
	"github.com/google/uuid"
)

type ScrapeRequest struct {
	URLs        []string `json:"urls" jsonschema:"links to pages, downloaded in parallel" validate:"required"`
	RobotsTxt   bool     `json:"robots_txt,omitempty" jsonschema:"respect the page robots.txt"`
	TimeoutMs   int64    `json:"timeout_ms,omitempty" jsonschema:"timeout for the whole call in milliseconds"`
	RemoveLinks bool     `json:"remove_links,omitempty" jsonschema:"strip markdown links from the text"`
	MaxChars    int      `json:"max_chars,omitempty" jsonschema:"truncate each page separately to N characters, default and maximum 20000; only lowers the limit, higher values are ignored"`
}

type ScrapeResponse struct {
	Results  []ScrapeResult `json:"results"`
	Metadata ScrapeMetadata `json:"metadata"`
}

type ScrapeResult struct {
	URL         string   `json:"url"`
	Status      string   `json:"status"`
	ScrapedData Document `json:"scraped_data"`
	TotalTimeMs int64    `json:"total_time_ms"`
}

type Document struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Byline    string    `json:"byline"`
	Markdown  string    `json:"markdown"`
	Length    int       `json:"length"`
	Excerpt   string    `json:"excerpt"`
	SiteName  string    `json:"site_name"`
	MainImage string    `json:"main_image"`
	AllImages []string  `json:"all_images"`
	Favicon   string    `json:"favicon"`
	Language  string    `json:"language"`
	Truncated bool      `json:"truncated"`
}

type ScrapeMetadata struct {
	TotalRequestTimeMs int64 `json:"total_request_time_ms"`
}
