package web

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
)

func TestCanceledUsecaseErrorMapsToRequestCanceled(t *testing.T) {
	// wrapped on purpose: the handler must match with errors.Is.
	canceledErr := fmt.Errorf("search: %w", domain.ErrCanceled)

	t.Run("search", func(t *testing.T) {
		req := dto.SearchRequest{Queries: []string{"go"}}

		ctrl := gomock.NewController(t)
		uc := NewMockusecase(ctrl)
		uc.EXPECT().SearchSnippets(gomock.Any(), req).Return(dto.SearchResponse{}, canceledErr)

		result, _, err := New(uc, 10).search(context.Background(), nil, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Equal(t, "request canceled", resultText(t, result))
	})

	t.Run("search images", func(t *testing.T) {
		req := dto.ImagesSearchRequest{Queries: []string{"cats"}}

		ctrl := gomock.NewController(t)
		uc := NewMockusecase(ctrl)
		uc.EXPECT().SearchImages(gomock.Any(), req).Return(dto.ImagesSearchResponse{}, canceledErr)

		result, _, err := New(uc, 10).searchImages(context.Background(), nil, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Equal(t, "request canceled", resultText(t, result))
	})

	t.Run("scrape", func(t *testing.T) {
		req := dto.ScrapeRequest{URLs: []string{"https://example.com/a"}}

		ctrl := gomock.NewController(t)
		uc := NewMockusecase(ctrl)
		uc.EXPECT().ScrapePages(gomock.Any(), req).Return(dto.ScrapeResponse{}, canceledErr)

		result, _, err := New(uc, 10).scrape(context.Background(), nil, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Equal(t, "request canceled", resultText(t, result))
	})
}
