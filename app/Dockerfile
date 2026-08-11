# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mcp-retrieval ./cmd/mcp-retrieval

FROM gcr.io/distroless/static-debian12:nonroot

# Proves to the MCP Registry that this image belongs to the published server.
# The value must match "name" in server.json exactly.
LABEL io.modelcontextprotocol.server.name="io.github.Role1776/mcp-retrieval"
LABEL org.opencontainers.image.source="https://github.com/Role1776/mcp-retrieval"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=build /out/mcp-retrieval /mcp-retrieval

# stdio transport by default: the client spawns the container and talks over
# stdin/stdout, so no port is exposed. Set MCP_TRANSPORT=http to serve over HTTP.
ENTRYPOINT ["/mcp-retrieval"]
