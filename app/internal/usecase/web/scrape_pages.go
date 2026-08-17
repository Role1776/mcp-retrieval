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

func (u *UseCase) ScrapePages(ctx context.Context, req dto.ScrapeRequest) (dto.ScrapeResponse, error) {
	const op = "usecase.web.ScrapePages"

	if err := u.validateScrapeRequest(req); err != nil {
		return dto.ScrapeResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	links, err := web.NewLinks(req.URLs)
	if err != nil {
		return dto.ScrapeResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	timeoutMs := limits.ResolveMinMax(req.TimeoutMs, u.cfg.DefaultTimeoutMs, u.cfg.MinTimeoutMs, u.cfg.MaxTimeoutMs)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	start := time.Now()

	results, errs := parallel.Map(links, func(link web.Link) (dto.ScrapeResult, error) {
		return u.executeOneScrape(ctx, link, req)
	})

	if len(errs) >= len(links) && len(errs) > 0 {
		u.logger.Error("all urls failed", slog.String("op", op), slog.Any("errs", errs))

		return dto.ScrapeResponse{}, fmt.Errorf("%s: %w", op, domain.ErrAllURLsFailed)
	}

	scrapeTime := time.Since(start)
	u.logger.Debug("scrape completed", slog.String("op", op), slog.Duration("duration", scrapeTime))

	return dto.ScrapeResponse{
		Results: results,
		Metadata: dto.ScrapeMetadata{
			TotalRequestTimeMs: scrapeTime.Milliseconds(),
		},
	}, nil
}

func (u *UseCase) executeOneScrape(ctx context.Context, link web.Link, req dto.ScrapeRequest) (dto.ScrapeResult, error) {
	const op = "usecase.web.executeOneScrape"
	start := time.Now()
	status := statusSuccess

	doc, err := u.retriever.Scrape(ctx, link, req.RobotsTxt)
	if err != nil {
		status = statusFailed
		if errors.Is(err, context.DeadlineExceeded) {
			status = statusTimeout
		}
		u.logger.Error("scrape error", slog.String("op", op), slog.String("url", link.String()), slog.Any("err", err))
	}

	if req.RemoveLinks {
		doc.RemoveAllLinks()
	}

	if truncErr := doc.Truncate(limits.ResolveMax(req.MaxChars, u.cfg.DefaultDocumentChars, u.cfg.MaxDocumentChars)); truncErr != nil {
		u.logger.Error("truncate error", slog.String("op", op), slog.String("url", link.String()), slog.Any("err", truncErr))
	}

	return dto.ScrapeResult{
		URL:         link.String(),
		Status:      status,
		ScrapedData: toDocument(doc),
		TotalTimeMs: time.Since(start).Milliseconds(),
	}, err
}

func (u *UseCase) validateScrapeRequest(req dto.ScrapeRequest) error {
	if err := validator.Validate(req); err != nil {
		return domain.ErrInvalidRequest
	}

	if len(req.URLs) > u.cfg.MaxQueries {
		return domain.ErrTooManyURLs
	}

	return nil
}
