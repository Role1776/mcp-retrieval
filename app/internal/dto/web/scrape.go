package web

import (
	"github.com/google/uuid"
)

type ScrapeRequest struct {
	URLs        []string `json:"urls" jsonschema:"1-10 absolute http(s) page URLs, fetched in parallel; one malformed URL rejects the whole call" validate:"required"`
	RobotsTxt   bool     `json:"robots_txt,omitempty" jsonschema:"if true, pages disallowed by the site robots.txt are not fetched; default false"`
	TimeoutMs   int64    `json:"timeout_ms,omitempty" jsonschema:"timeout for the whole call in milliseconds, shared by all URLs; default 5000, clamped to 1000-10000; unfinished pages get status 'timeout'"`
	RemoveLinks bool     `json:"remove_links,omitempty" jsonschema:"if true, markdown links are replaced by their text; default false"`
	MaxChars    int      `json:"max_chars,omitempty" jsonschema:"character limit applied to each page separately, not to the whole response; default and maximum 20000, so it can only lower the limit; a cut page ends with '[truncated]'"`
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
