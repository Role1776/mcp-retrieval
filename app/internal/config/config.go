package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"

	webAdapter "github.com/Role1776/mcp-retrieval/app/internal/adapter/web"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/logger"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/mcpserver"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/server"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/validator"
	webUseCase "github.com/Role1776/mcp-retrieval/app/internal/usecase/web"
)

type Config struct {
	MCP              *mcpserver.Config
	Server           *server.Config
	Logger           *logger.Config
	RetrieverAdapter *webAdapter.RetrieverConfig
	SearchUseCase    *webUseCase.SearchConfig
}

func Load(envPath string) (*Config, error) {
	cfg := &Config{
		MCP:              &mcpserver.Config{},
		Server:           &server.Config{},
		Logger:           &logger.Config{},
		RetrieverAdapter: &webAdapter.RetrieverConfig{},
		SearchUseCase:    &webUseCase.SearchConfig{},
	}

	if err := loadEnvFile(envPath); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func loadEnvFile(envPath string) error {
	var err error
	if envPath != "" {
		err = godotenv.Load(envPath)
	}

	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		fmt.Fprintf(os.Stderr, "%s no env file found at %q, using the ambient environment\n",
			time.Now().Format(time.RFC3339), envPath)

		return nil
	default:
		return fmt.Errorf("load env file %s: %w", envPath, err)
	}
}
