package web

import (
	"context"

	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
	"github.com/Role1776/mcp-retrieval/app/internal/transport/mcp/utils/mcperrors"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) search(ctx context.Context, _ *mcpsdk.CallToolRequest, req dto.SearchRequest) (*mcpsdk.CallToolResult, dto.SearchResponse, error) {
	res, err := h.usecase.SearchSnippets(ctx, req)
	if err != nil {
		return mcperrors.Result(err), dto.SearchResponse{}, nil
	}

	return nil, res, nil
}
