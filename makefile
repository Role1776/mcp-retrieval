.PHONY: build cover cover-html test version

MODULE := app
BIN    := $(CURDIR)/bin/mcp-retrieval
COVER  := $(CURDIR)/bin/coverage.out
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)
LDFLAGS := -X github.com/Role1776/mcp-retrieval/app/internal/pkg/mcpserver.Version=$(VERSION)

build:
	@go -C $(MODULE) build -ldflags="$(LDFLAGS)" -o $(BIN) ./cmd/mcp-retrieval

version:
	@echo $(VERSION)

cover:
	@go -C $(MODULE) test -coverprofile=$(COVER) ./... > /dev/null
	@go -C $(MODULE) tool cover -func=$(COVER) | grep total:

cover-html:
	@go -C $(MODULE) test -coverprofile=$(COVER) ./... > /dev/null
	@go -C $(MODULE) tool cover -html=$(COVER)

test:
	@go -C $(MODULE) test -v ./...
