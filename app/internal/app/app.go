package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	webAdapter "github.com/Role1776/mcp-retrieval/app/internal/adapter/web"
	"github.com/Role1776/mcp-retrieval/app/internal/config"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/mcpserver"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/server"
	mcpRouter "github.com/Role1776/mcp-retrieval/app/internal/transport/mcp/router"
	mcpWeb "github.com/Role1776/mcp-retrieval/app/internal/transport/mcp/web"
	webUseCase "github.com/Role1776/mcp-retrieval/app/internal/usecase/web"
)

const shutdownTimeout = 10 * time.Second

type stopper interface {
	Stop(ctx context.Context) error
}

type App struct {
	Config *config.Config
	Logger *slog.Logger
	srv    stopper
}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		Config: cfg,
		Logger: logger,
	}
}

func (a *App) Run() error {
	const op = "app.Run"
	retrievalClient, err := webAdapter.InitClient(a.Config.RetrieverAdapter)
	if err != nil {
		return fmt.Errorf("%s: failed to create retrieval client: %w", op, err)
	}

	webAdapter := webAdapter.New(retrievalClient)

	webUseCase := webUseCase.New(webAdapter, a.Logger, a.Config.SearchUseCase)

	mcpSrv := mcpserver.NewServer(a.Config.MCP)

	mcpRouter.New(mcpSrv.MCP(), mcpWeb.New(webUseCase)).RegisterAllTools()

	run, err := a.buildMCPServer(mcpSrv)
	if err != nil {
		return err
	}

	notifyCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		if err := run(notifyCtx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, io.EOF) {
			errCh <- fmt.Errorf("%s: failed to run mcp server: %w", op, err)
		}
		close(errCh)

	}()

	a.Logger.Info("MCP server has started",
		slog.String("op", op),
		slog.String("transport", a.Config.MCP.Transport),
	)

	select {
	case <-notifyCtx.Done():
		break
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	a.Logger.Info("shutting down", slog.String("op", op))

	if err := a.stop(ctxShutdown); err != nil {
		return err
	}

	return nil
}

func (a *App) buildMCPServer(mcpSrv *mcpserver.Server) (func(context.Context) error, error) {
	const op = "app.buildMCPServer"

	switch a.Config.MCP.Transport {
	case mcpserver.TransportStdio:
		a.srv = mcpSrv

		return mcpSrv.RunStdio, nil

	case mcpserver.TransportHTTP:
		if a.Config.Server.Port == "" {
			return nil, fmt.Errorf("%s: http transport: server port is not configured", op)
		}

		httpSrv := server.NewServer(a.Config.Server, mcpSrv.HTTPHandler())
		a.srv = httpSrv

		return httpSrv.Run, nil

	default:
		return nil, fmt.Errorf("%s: unknown mcp transport: %s", op, a.Config.MCP.Transport)
	}
}

func (a *App) stop(ctx context.Context) error {
	const op = "app.stop"

	if err := a.srv.Stop(ctx); err != nil {
		return fmt.Errorf("%s: failed to shutdown server: %w", op, err)
	}

	return nil
}
