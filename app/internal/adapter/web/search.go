package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"

	"github.com/free-llms-foundation/retrieval-go"
)

func (r *Retrieval) Search(ctx context.Context, query web.Query, date string) (web.Snippets, error) {
	const op = "retrieval.Search"

	pages, err := r.client.SearchWithQuery(ctx, query.String(), date)
	if err != nil {
		if errors.Is(err, retrieval.ErrUnexpectedStatusCode) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrUnexpectedStatusCode)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make(web.Snippets, 0, len(pages))
	for i, page := range pages {
		snippet, err := web.NewSnippet(web.SnippetProps{
			Title:   page.Title,
			Rank:    i + 1,
			Source:  page.Source,
			Link:    page.Link,
			Snippet: page.Snippet,
			Favicon: page.Favicon,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: invalid snippet from adapter: %w", op, err)
		}
		result = append(result, snippet)
	}

	return result, nil
}
