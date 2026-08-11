package web

import (
	"context"

	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
	"github.com/Role1776/mcp-retrieval/app/internal/transport/mcp/utils/mcperrors"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) searchImages(ctx context.Context, _ *mcpsdk.CallToolRequest, req dto.ImagesSearchRequest) (*mcpsdk.CallToolResult, dto.ImagesSearchResponse, error) {
	res, err := h.usecase.SearchImages(ctx, req)
	if err != nil {
		return mcperrors.Result(err), dto.ImagesSearchResponse{}, nil
	}

	return nil, res, nil
}
