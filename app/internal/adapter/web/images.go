package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"

	"github.com/free-llms-foundation/retrieval-go"
)

func (r *Retrieval) Images(ctx context.Context, query web.Query, date string) (web.Images, error) {
	const op = "retrieval.Images"

	images, err := r.client.SearchImagesWithQuery(ctx, query.String(), date)
	if err != nil {
		if errors.Is(err, retrieval.ErrUnexpectedStatusCode) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrUnexpectedStatusCode)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make(web.Images, 0, len(images))
	for _, image := range images {
		img, err := web.NewImage(web.ImageProps{
			URL:         image.URL,
			PageURL:     image.PageURL,
			Description: image.Desc,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: invalid image from adapter: %w", op, err)
		}
		result = append(result, img)
	}

	return result, nil
}
