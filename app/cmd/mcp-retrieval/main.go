package main

import (
	"flag"
	"log"

	"github.com/Role1776/mcp-retrieval/app/internal/app"
	"github.com/Role1776/mcp-retrieval/app/internal/config"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/logger"
)

func main() {
	var envPath string
	flag.StringVar(&envPath, "env", "", "path to .env file (optional)")
	flag.Parse()

	cfg, err := config.Load(envPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logger.SetupLogger(cfg.Logger)
	if err := app.New(cfg, logger).Run(); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
