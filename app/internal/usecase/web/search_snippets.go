package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/validator"
	"github.com/Role1776/mcp-retrieval/app/internal/usecase/web/webutils/limits"
	"github.com/Role1776/mcp-retrieval/app/internal/usecase/web/webutils/parallel"
)

func (u *UseCase) SearchSnippets(ctx context.Context, req dto.SearchRequest) (dto.SearchResponse, error) {
	const op = "usecase.web.SearchSnippets"

	if err := u.validateSearchRequest(req); err != nil {
		return dto.SearchResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	queries, err := web.NewQueries(req.Queries)
	if err != nil {
		return dto.SearchResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	timeoutMs := limits.ResolveMinMax(req.TimeoutMs, u.cfg.DefaultTimeoutMs, u.cfg.MinTimeoutMs, u.cfg.MaxTimeoutMs)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	start := time.Now()

	results, errs := parallel.Map(queries, func(query web.Query) (dto.Result, error) {
		return u.executeOneQuery(ctx, query, req)
	})

	if len(errs) >= len(queries) && len(errs) > 0 {
		u.logger.Error("all queries failed", slog.String("op", op), slog.Any("errs", errs))

		return dto.SearchResponse{}, fmt.Errorf("%s: %w", op, domain.ErrAllQueriesFailed)
	}

	searchTime := time.Since(start)
	u.logger.Debug("search completed", slog.String("op", op), slog.Duration("duration", searchTime))

	return dto.SearchResponse{
		Results: results,
		Metadata: dto.SearchMetadata{
			TotalRequestTimeMs: searchTime.Milliseconds(),
			Date:               req.Date,
		},
	}, nil
}

func (u *UseCase) executeOneQuery(ctx context.Context, query web.Query, req dto.SearchRequest) (dto.Result, error) {
	const op = "usecase.web.executeOneQuery"
	start := time.Now()
	status := statusSuccess

	snippets, err := u.retriever.Search(ctx, query, req.Date)
	if err != nil {
		status = statusFailed
		if errors.Is(err, context.DeadlineExceeded) {
			status = statusTimeout
		}
		u.logger.Error("search error", slog.String("op", op), slog.String("query", query.String()), slog.Any("err", err))
	}

	resSnippets := snippets.Dedupe().Limit(limits.ResolveMax(req.MaxResults, u.cfg.DefaultResults, u.cfg.MaxResults)).Rerank()

	return dto.Result{
		Query:       query.String(),
		Status:      status,
		Count:       resSnippets.Len(),
		Snippets:    resSnippets,
		TotalTimeMs: time.Since(start).Milliseconds(),
	}, err
}

func (u *UseCase) validateSearchRequest(req dto.SearchRequest) error {
	if err := validator.Validate(req); err != nil {
		return domain.ErrInvalidRequest
	}

	if len(req.Queries) > u.cfg.MaxQueries {
		return domain.ErrTooManyQueries
	}

	return nil
}
