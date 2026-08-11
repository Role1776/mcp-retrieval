package web

import (
	"context"
	"log/slog"

	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
)

const (
	statusSuccess = "success"
	statusFailed  = "failed"
	statusTimeout = "timeout"
)

type retrieval interface {
	Search(ctx context.Context, query web.Query, date string) (web.Snippets, error)
	Scrape(ctx context.Context, link web.Link, robotsTxt bool) (web.Document, error)
	Images(ctx context.Context, query web.Query, date string) (web.Images, error)
}

type SearchConfig struct {
	MaxQueries           int   `env:"MAX_QUERIES" envDefault:"10" validate:"gt=0"`
	DefaultResults       int   `env:"DEFAULT_RESULTS" envDefault:"5" validate:"gt=0,ltefield=MaxResults"`
	MaxResults           int   `env:"MAX_RESULTS" envDefault:"20" validate:"gt=0"`
	DefaultTimeoutMs     int64 `env:"DEFAULT_TIMEOUT_MS" envDefault:"5000" validate:"gtefield=MinTimeoutMs,ltefield=MaxTimeoutMs"`
	MaxTimeoutMs         int64 `env:"MAX_TIMEOUT_MS" envDefault:"10000" validate:"gt=0"`
	MinTimeoutMs         int64 `env:"MIN_TIMEOUT_MS" envDefault:"1000" validate:"gt=0,ltefield=MaxTimeoutMs"`
	DefaultImages        int   `env:"DEFAULT_IMAGES" envDefault:"5" validate:"gt=0,ltefield=MaxImages"`
	MaxImages            int   `env:"MAX_IMAGES" envDefault:"10" validate:"gt=0"`
	DefaultDocumentChars int   `env:"DEFAULT_DOCUMENT_CHARS" envDefault:"20000" validate:"gt=0,ltefield=MaxDocumentChars"`
	MaxDocumentChars     int   `env:"MAX_DOCUMENT_CHARS" envDefault:"20000" validate:"gt=0"`
}

type UseCase struct {
	retriever retrieval
	logger    *slog.Logger
	cfg       *SearchConfig
}

func New(retriever retrieval, logger *slog.Logger, cfg *SearchConfig) *UseCase {
	return &UseCase{
		retriever: retriever,
		logger:    logger,
		cfg:       cfg,
	}
}
