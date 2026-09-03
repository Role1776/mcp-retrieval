package web

import (
	"context"

	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) scrape(ctx context.Context, _ *mcpsdk.CallToolRequest, req dto.ScrapeRequest) (*mcpsdk.CallToolResult, dto.ScrapeResponse, error) {
	res, err := h.usecase.ScrapePages(ctx, req)
	if err != nil {
		return result(err), dto.ScrapeResponse{}, nil
	}

	return nil, res, nil
}
