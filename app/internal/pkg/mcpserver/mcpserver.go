package mcpserver

import (
	"context"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	TransportStdio = "stdio"
	TransportHTTP  = "http"
)

// Version is reported in the MCP initialize handshake. "dev" is the value for
// local builds only: the makefile, Dockerfile and GoReleaser all overwrite it
// from the git tag via -ldflags -X, which needs a var and not a const.
var Version = "dev"

type Config struct {
	Transport string `env:"MCP_TRANSPORT" envDefault:"stdio" validate:"oneof=stdio http"`
	Name      string `env:"MCP_NAME" envDefault:"mcp-retrieval"`
	Path      string `env:"MCP_PATH" envDefault:"/mcp"` // used only when transport is http
}

type Server struct {
	mcp *mcpsdk.Server
	cfg *Config
}

func NewServer(cfg *Config) *Server {
	return &Server{
		mcp: mcpsdk.NewServer(&mcpsdk.Implementation{
			Name:    cfg.Name,
			Version: Version,
		}, nil),
		cfg: cfg,
	}
}

func (s *Server) MCP() *mcpsdk.Server {
	return s.mcp
}

func (s *Server) RunStdio(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcpsdk.StdioTransport{})
}

func (s *Server) HTTPHandler() http.Handler {
	streamable := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return s.mcp
	}, nil)

	mux := http.NewServeMux()
	mux.Handle(s.cfg.Path, streamable)

	return mux
}

func (s *Server) Stop(_ context.Context) error {
	return nil
}
