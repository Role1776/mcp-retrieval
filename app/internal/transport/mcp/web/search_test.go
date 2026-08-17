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

func TestSearch(t *testing.T) {
	req := dto.SearchRequest{Queries: []string{"go"}, MaxResults: 2}

	okResponse := dto.SearchResponse{
		Results: []dto.Result{
			{
				Query:    "go",
				Status:   "ok",
				Count:    1,
				Snippets: []dto.Snippet{{Rank: 1, Title: "Go", Link: "https://go.dev", Snippet: "the language"}},
			},
		},
		Metadata: dto.SearchMetadata{TotalRequestTimeMs: 12},
	}

	cases := []struct {
		name string
		res  dto.SearchResponse
		err  error

		wantIsError bool
		wantText    string
		wantRes     dto.SearchResponse
	}{
		{
			name:    "success",
			res:     okResponse,
			wantRes: okResponse,
		},
		{
			name:    "success with empty results",
			res:     dto.SearchResponse{},
			wantRes: dto.SearchResponse{},
		},
		{
			name:        "domain error is mapped",
			err:         domain.ErrTooManyQueries,
			wantIsError: true,
			wantText:    "too many queries",
			wantRes:     dto.SearchResponse{},
		},
		{
			name:        "all queries failed is mapped",
			err:         domain.ErrAllQueriesFailed,
			wantIsError: true,
			wantText:    "every query failed; the search upstream may be unreachable",
			wantRes:     dto.SearchResponse{},
		},
		{
			name:        "unknown error falls back to internal",
			err:         errors.New("boom"),
			wantIsError: true,
			wantText:    "internal server error",
			wantRes:     dto.SearchResponse{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			uc.EXPECT().SearchSnippets(gomock.Any(), req).Return(tc.res, tc.err)

			h := New(uc)

			// act
			result, res, err := h.search(context.Background(), nil, req)

			// assert
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)

			if !tc.wantIsError {
				assert.Nil(t, result)

				return
			}

			require.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Equal(t, tc.wantText, resultText(t, result))
		})
	}
}
