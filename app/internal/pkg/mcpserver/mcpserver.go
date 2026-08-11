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

type Config struct {
	Transport string `env:"MCP_TRANSPORT" envDefault:"stdio" validate:"oneof=stdio http"`
	Name      string `env:"MCP_NAME" envDefault:"mcp-retrieval"`
	Version   string `env:"MCP_VERSION" envDefault:"0.1.0"`
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
			Version: cfg.Version,
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
