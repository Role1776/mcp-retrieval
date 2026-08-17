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

func (u *UseCase) SearchImages(ctx context.Context, req dto.ImagesSearchRequest) (dto.ImagesSearchResponse, error) {
	const op = "usecase.web.SearchImages"

	if err := u.validateImagesSearchRequest(req); err != nil {
		return dto.ImagesSearchResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	queries, err := web.NewQueries(req.Queries)
	if err != nil {
		return dto.ImagesSearchResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	timeoutMs := limits.ResolveMinMax(req.TimeoutMs, u.cfg.DefaultTimeoutMs, u.cfg.MinTimeoutMs, u.cfg.MaxTimeoutMs)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	start := time.Now()

	results, errs := parallel.Map(queries, func(query web.Query) (dto.ImagesResult, error) {
		return u.executeOneImagesQuery(ctx, query, req)
	})

	if len(errs) >= len(queries) && len(errs) > 0 {
		u.logger.Error("all queries failed", slog.String("op", op), slog.Any("errs", errs))

		return dto.ImagesSearchResponse{}, fmt.Errorf("%s: %w", op, domain.ErrAllQueriesFailed)
	}

	searchTime := time.Since(start)
	u.logger.Debug("images search completed", slog.String("op", op), slog.Duration("duration", searchTime))

	return dto.ImagesSearchResponse{
		Results: results,
		Metadata: dto.ImagesMetadata{
			TotalRequestTimeMs: searchTime.Milliseconds(),
			Date:               req.Date,
		},
	}, nil
}

func (u *UseCase) executeOneImagesQuery(ctx context.Context, query web.Query, req dto.ImagesSearchRequest) (dto.ImagesResult, error) {
	const op = "usecase.web.executeOneImagesQuery"
	start := time.Now()

	images, err := u.retriever.Images(ctx, query, req.Date)
	if err != nil {
		if errors.Is(err, domain.ErrNoRelevantImages) {
			return dto.ImagesResult{
				Query:       query.String(),
				Status:      statusNoRelevant,
				TotalTimeMs: time.Since(start).Milliseconds(),
			}, nil
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return dto.ImagesResult{
				Query:       query.String(),
				Status:      statusTimeout,
				TotalTimeMs: time.Since(start).Milliseconds(),
			}, nil
		}

		u.logger.Error("images search error", slog.String("op", op), slog.String("query", query.String()), slog.Any("err", err))

		return dto.ImagesResult{
			Query:       query.String(),
			Status:      statusFailed,
			TotalTimeMs: time.Since(start).Milliseconds(),
		}, err
	}

	resImages := images.Dedupe().Limit(limits.ResolveMax(req.MaxImages, u.cfg.DefaultImages, u.cfg.MaxImages))

	return dto.ImagesResult{
		Query:       query.String(),
		Status:      statusSuccess,
		Count:       resImages.Len(),
		Images:      toImages(resImages),
		TotalTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

func (u *UseCase) validateImagesSearchRequest(req dto.ImagesSearchRequest) error {
	if err := validator.Validate(req); err != nil {
		return domain.ErrInvalidRequest
	}

	if len(req.Queries) > u.cfg.MaxQueries {
		return domain.ErrTooManyQueries
	}

	return nil
}
