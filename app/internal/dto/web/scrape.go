package web

import (
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
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
	URL         string       `json:"url"`
	Status      string       `json:"status"`
	ScrapedData web.Document `json:"scraped_data"`
	TotalTimeMs int64        `json:"total_time_ms"`
}

type ScrapeMetadata struct {
	TotalRequestTimeMs int64 `json:"total_request_time_ms"`
}
