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

func TestSearchImages(t *testing.T) {
	req := dto.ImagesSearchRequest{Queries: []string{"cats"}, MaxImages: 2}

	okResponse := dto.ImagesSearchResponse{
		Results: []dto.ImagesResult{
			{
				Query:  "cats",
				Status: "ok",
				Count:  1,
				Images: []dto.Image{{URL: "https://img/1.png", PageURL: "https://page/1", Description: "cat"}},
			},
		},
		Metadata: dto.ImagesMetadata{TotalRequestTimeMs: 12},
	}

	cases := []struct {
		name string
		res  dto.ImagesSearchResponse
		err  error

		wantIsError bool
		wantText    string
		wantRes     dto.ImagesSearchResponse
	}{
		{
			name:    "success",
			res:     okResponse,
			wantRes: okResponse,
		},
		{
			name:    "success with empty results",
			res:     dto.ImagesSearchResponse{},
			wantRes: dto.ImagesSearchResponse{},
		},
		{
			name:        "domain error is mapped",
			err:         domain.ErrQueryTooLong,
			wantIsError: true,
			wantText:    "query is too long",
			wantRes:     dto.ImagesSearchResponse{},
		},
		{
			name:        "all queries failed is mapped",
			err:         domain.ErrAllQueriesFailed,
			wantIsError: true,
			wantText:    "every query failed; the search upstream may be unreachable",
			wantRes:     dto.ImagesSearchResponse{},
		},
		{
			name:        "unknown error falls back to internal",
			err:         errors.New("boom"),
			wantIsError: true,
			wantText:    "internal server error",
			wantRes:     dto.ImagesSearchResponse{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			uc.EXPECT().SearchImages(gomock.Any(), req).Return(tc.res, tc.err)

			h := New(uc)

			// act
			result, res, err := h.searchImages(context.Background(), nil, req)

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
