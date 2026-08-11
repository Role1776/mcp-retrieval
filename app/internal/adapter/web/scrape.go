package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"

	"github.com/free-llms-foundation/retrieval-go"
)

func (r *Retrieval) Scrape(ctx context.Context, link web.Link, robotsTxt bool) (web.Document, error) {
	const op = "retrieval.ParseContentFromLink"

	doc, err := r.client.ParseContentFromLink(ctx, link.String(), robotsTxt)
	if err != nil {
		if errors.Is(err, retrieval.ErrRobotsDenied) {
			return web.Document{}, fmt.Errorf("%s: %w", op, domain.ErrRobotsDenied)
		}

		if errors.Is(err, retrieval.ErrUnexpectedStatusCode) {
			return web.Document{}, fmt.Errorf("%s: %w", op, domain.ErrUnexpectedStatusCode)
		}

		return web.Document{}, fmt.Errorf("%s: %w", op, err)
	}

	domainDoc, err := web.NewDocument(web.DocumentProps{
		Title:     doc.Title,
		Byline:    doc.Byline,
		Markdown:  doc.Content,
		Length:    doc.Length,
		Excerpt:   doc.Excerpt,
		SiteName:  doc.SiteName,
		MainImage: doc.Image,
		AllImages: doc.Images,
		Favicon:   doc.Favicon,
		Language:  doc.Language,
	})
	if err != nil {
		return web.Document{}, fmt.Errorf("%s: invalid document from adapter: %w", op, err)
	}

	return domainDoc, nil
}
