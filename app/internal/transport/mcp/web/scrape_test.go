package web

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
)

func TestScrape(t *testing.T) {
	req := dto.ScrapeRequest{URLs: []string{"https://go.dev"}, MaxChars: 100}

	okResponse := dto.ScrapeResponse{
		Results: []dto.ScrapeResult{
			{
				URL:         "https://go.dev",
				Status:      "ok",
				ScrapedData: dto.Document{Title: "Go", Markdown: "body"},
			},
		},
		Metadata: dto.ScrapeMetadata{TotalRequestTimeMs: 12},
	}

	cases := []struct {
		name string
		res  dto.ScrapeResponse
		err  error

		wantIsError bool
		wantText    string
		wantRes     dto.ScrapeResponse
	}{
		{
			name:    "success",
			res:     okResponse,
			wantRes: okResponse,
		},
		{
			name:    "success with empty results",
			res:     dto.ScrapeResponse{},
			wantRes: dto.ScrapeResponse{},
		},
		{
			name:        "domain error is mapped",
			err:         domain.ErrRobotsDenied,
			wantIsError: false,
			wantText:    "robots.txt denied",
			wantRes:     dto.ScrapeResponse{},
		},
		{
			name:        "all urls failed is mapped",
			err:         domain.ErrAllURLsFailed,
			wantIsError: true,
			wantText:    "every url failed to be scraped; the pages may be unreachable or hold no extractable text",
			wantRes:     dto.ScrapeResponse{},
		},
		{
			name:        "unknown error falls back to internal",
			err:         errors.New("boom"),
			wantIsError: true,
			wantText:    "internal server error",
			wantRes:     dto.ScrapeResponse{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			uc.EXPECT().ScrapePages(gomock.Any(), req).Return(tc.res, tc.err)

			h := New(uc)

			// act
			result, res, err := h.scrape(context.Background(), nil, req)

			// assert
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)

			if tc.err == nil {
				assert.Nil(t, result)

				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tc.wantIsError, result.IsError)
			assert.Equal(t, tc.wantText, resultText(t, result))
		})
	}
}
