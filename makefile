.PHONY: build cover cover-html test

MODULE := app
BIN    := $(CURDIR)/bin/mcp-retrieval
COVER  := $(CURDIR)/bin/coverage.out

build:
	@go -C $(MODULE) build -o $(BIN) ./cmd/mcp-retrieval

cover:
	@go -C $(MODULE) test -coverprofile=$(COVER) ./... > /dev/null
	@go -C $(MODULE) tool cover -func=$(COVER) | grep total:

cover-html:
	@go -C $(MODULE) test -coverprofile=$(COVER) ./... > /dev/null
	@go -C $(MODULE) tool cover -html=$(COVER)

test:
	@go -C $(MODULE) test -v ./...
