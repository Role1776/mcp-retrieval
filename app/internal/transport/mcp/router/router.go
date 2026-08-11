package router

import (
	mcpWeb "github.com/Role1776/mcp-retrieval/app/internal/transport/mcp/web"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Router struct {
	mcp *mcpsdk.Server
	web *mcpWeb.Handler
}

func New(mcpServer *mcpsdk.Server, web *mcpWeb.Handler) *Router {
	return &Router{
		mcp: mcpServer,
		web: web,
	}
}

func (r *Router) RegisterAllTools() {
	r.web.RegisterTools(r.mcp)
}
