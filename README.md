<h1 align="center">mcp-retrieval</h1>

<p align="center">
  <b>An MCP server that gives an LLM three web tools: search, image search, and page scraping — no API keys required.</b>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-8bc34a?style=for-the-badge" alt="License MIT"></a>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/MCP-Server-6E56CF?style=for-the-badge&logo=anthropic&logoColor=white" alt="MCP">
  <img src="https://img.shields.io/badge/Transport-stdio_%7C_http-26A5E4?style=for-the-badge" alt="Transport">
  <img src="https://img.shields.io/badge/DuckDuckGo-Search-DE5833?style=for-the-badge&logo=duckduckgo&logoColor=white" alt="DuckDuckGo">
  <img src="https://img.shields.io/badge/Bing-Images-008373?style=for-the-badge&logo=microsoftbing&logoColor=white" alt="Bing Images">
  <img src="https://img.shields.io/badge/uTLS-Fingerprint-1f6feb?style=for-the-badge" alt="uTLS">
</p>

<p align="center">
  <a href="#demo">Demo</a> ·
  <a href="#tools">Tools</a> ·
  <a href="#quick-start">Quick start</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#retrieval-engine">Retrieval engine</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="CONTRIBUTING.md">Contributing</a>
</p>

---

## Demo

<!-- TODO: replace the placeholder below with the real demo video.
     GitHub renders an <video> tag inline when the src points at an
     uploaded asset URL (https://github.com/user-attachments/assets/...). -->

<p align="center">
  <img src="https://img.shields.io/badge/%F0%9F%8E%AC_Demo_video-coming_soon-6E56CF?style=for-the-badge" alt="Demo video coming soon" height="48">
</p>

<p align="center"><i>An LLM using <code>web_search</code>, <code>web_search_images</code>, and <code>web_scrape</code> end to end — recording on the way.</i></p>

---

## What it is

`mcp-retrieval` is a [Model Context Protocol](https://modelcontextprotocol.io) server written in Go. It exposes web retrieval capabilities to any MCP-compatible client (Claude Desktop, IDE agents, custom LLM apps) as three read-only tools. Under the hood it uses the [`retrieval-go`](https://github.com/free-llms-foundation/retrieval-go) library to search the web and fetch pages, returning results as clean Markdown ready to hand to a model.

The library needs **no API keys**: web search goes through DuckDuckGo Lite, image search through Bing Images, and page fetching runs the HTML through a readability extractor before converting it to Markdown. To stay reliable against bot protection it impersonates real browsers at the TLS level and can rotate both browser fingerprints and proxies — see [Retrieval engine](#retrieval-engine).

Both transports the MCP SDK supports are available and expose the identical tool set:

- **stdio** — the client launches the binary and talks over stdin/stdout (the default, ideal for desktop clients).
- **http** — a long-running streamable HTTP server (useful for remote/shared deployments).

---

## Tools

| Tool | Description |
| :--- | :--- |
| **`web_search`** | Runs one or more queries in parallel and returns per-query deduplicated, reranked snippets with links. |
| **`web_search_images`** | Runs one or more image queries in parallel and returns per-query deduplicated image results. |
| **`web_scrape`** | Downloads one or more pages in parallel and returns the main article text as Markdown. |

All three are annotated as **read-only**. Each tool returns a structured JSON payload that matches its output schema; the SDK mirrors the same JSON into the text content block for clients that do not read `structuredContent`.

### `web_search`

| Parameter | Type | Default | Notes |
| :--- | :--- | :--- | :--- |
| `queries` | `[]string` | — | **Required.** Executed in parallel. |
| `max_results` | `int` | `5` | Snippets per query, capped at `max_results` config (`20`). |
| `timeout_ms` | `int64` | `5000` | Whole-call timeout; clamped to `[min, max]` from config. |
| `date` | `string` | — | Freshness filter: `d` (day), `w` (week), `m` (month), `y` (year). |

### `web_search_images`

| Parameter | Type | Default | Notes |
| :--- | :--- | :--- | :--- |
| `queries` | `[]string` | — | **Required.** Executed in parallel. |
| `max_images` | `int` | `5` | Images per query, capped at `max_images` config (`10`). |
| `timeout_ms` | `int64` | `5000` | Whole-call timeout; clamped to `[min, max]` from config. |
| `date` | `string` | — | Freshness filter: `d` / `w` / `m` / `y`. |

### `web_scrape`

| Parameter | Type | Default | Notes |
| :--- | :--- | :--- | :--- |
| `urls` | `[]string` | — | **Required.** Downloaded in parallel. |
| `robots_txt` | `bool` | `false` | Respect the page's `robots.txt`. |
| `timeout_ms` | `int64` | `5000` | Whole-call timeout; clamped to `[min, max]` from config. |
| `remove_links` | `bool` | `false` | Strip Markdown links from the text. |
| `max_chars` | `int` | `20000` | Truncate page text to N characters, capped at `max_document_chars` config (`20000`). |

> Both `queries`/`urls` lists are capped at `max_queries` (`10`) items per call. Queries must be ≤ 512 characters; URLs ≤ 2048 characters and `http`/`https` only.

### Results and counts

Every call fans out across the input list and returns one entry per query/URL, each with its own `status` — `success`, `failed`, or `timeout` — so a partial failure still returns the items that did work.

`count` is the number of items actually returned, and it can be **lower than the requested `max_results` / `max_images`**: duplicates within a single query's results are removed before the limit is applied, and the upstream may simply have fewer items to give. A smaller `count` is a normal outcome, not an error.

Deduplication is **per query, not across queries**. Each entry is deduplicated on its own, so a link found by two of the queries in the same call appears in both entries — dedupe the union yourself if you need it.

### Errors

Request-level failures are returned as a tool result with `isError: true` and a plain-text message, not as a JSON-RPC error — the model reads the message and can correct the call itself. Per-item failures never do this; they stay inside the payload as `status: "failed"` / `"timeout"`.

A call fails outright only when the input is rejected before any work starts, or when **every** item in it fails:

| Message | Meaning |
| :--- | :--- |
| `invalid request` | The arguments did not pass validation. |
| `too many queries` / `too many urls` | The list exceeds `MAX_QUERIES`. |
| `query must not be empty` | An empty query, or an empty `queries` list. |
| `query is too long` | A query exceeds 512 characters. |
| `invalid url` | A URL is malformed, over 2048 characters, or not `http`/`https`. |
| `robots.txt denied` | `robots_txt: true` and the page disallows fetching. |
| `upstream service unavailable` | The upstream answered with an unexpected status code. |
| `every url failed to be scraped; the pages may be unreachable or hold no extractable text` | All URLs failed. Individual causes are logged to `stderr`, not returned. |
| `every query failed; the search upstream may be unreachable` | All queries failed. |
| `internal server error` | Anything unclassified. |

The all-failed messages deliberately do not distinguish timeouts from other causes: a mixed batch can fail for several reasons at once, and the per-item `status` already carries that detail whenever at least one item survives.

### Known limitations

- **`web_scrape` handles HTML only.** Pages are run through a readability extractor, which needs article markup, so `text/plain` responses yield nothing and come back as `status: "failed"`. Raw-file hosts are the common case: `raw.githubusercontent.com`, `github.com/.../raw/...`, `cdn.jsdelivr.net`. Scrape the rendered page instead of the raw file.
- **`web_search_images` relevance is not guaranteed.** For some queries Bing Images serves a page that is not a result set, and it is parsed as though it were — the tool then returns unrelated images with `status: "success"`. Treat image results as best-effort and verify them before showing them to a user.
- **No JavaScript.** Pages are fetched as-is; content rendered client-side is invisible to the extractor.

---

## Quick start

### Install

Pick whichever fits — all three give the identical server.

**Container** (no Go toolchain needed):

```bash
docker pull ghcr.io/role1776/mcp-retrieval:latest
```

**Prebuilt binary** — grab the archive for your platform from the [latest release](https://github.com/Role1776/mcp-retrieval/releases/latest), unpack it, and put `mcp-retrieval` on your `PATH`.

**From source:**

```bash
go install github.com/Role1776/mcp-retrieval/app/cmd/mcp-retrieval@latest   # needs Go 1.25.5+
```

Or build the binary in place (the Go module lives in `app/`):

```bash
make build          # -> bin/mcp-retrieval
```

### Run

```bash
# defaults: stdio transport, no configuration needed
./bin/mcp-retrieval

# with an explicit env file
./bin/mcp-retrieval -env /absolute/path/to/.env
```

The one flag is optional:

| Flag | Meaning |
| :--- | :--- |
| `-env` | Path to a `.env` file. If omitted — or if the file does not exist — the server starts on defaults and whatever is already in the environment. There is no implicit lookup: under stdio the working directory is chosen by the MCP client, so a relative default would be unpredictable. |

### Connecting an MCP client (stdio)

Point your client at the built binary. Example Claude Desktop config:

```json
{
  "mcpServers": {
    "retrieval": {
      "command": "/absolute/path/to/app",
      "env": {
        "MAX_RESULTS": "20"
      }
    }
  }
}
```

The `env` block is optional — `"command"` alone is enough.

### Connecting an MCP client (container)

Run the image on stdio. Configuration still travels through the `env` block, but Docker needs each variable named on the command line with `-e` for it to reach the process:

```json
{
  "mcpServers": {
    "retrieval": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "MAX_RESULTS",
        "-e", "DEFAULT_TIMEOUT_MS",
        "ghcr.io/role1776/mcp-retrieval:latest"
      ],
      "env": {
        "MAX_RESULTS": "20",
        "DEFAULT_TIMEOUT_MS": "5000"
      }
    }
  }
}
```

`-i` is required — without it the container gets no stdin and the client sees the server die immediately. Clients that install from the [MCP Registry](https://registry.modelcontextprotocol.io) build this invocation themselves and prompt for the variables declared in [`server.json`](server.json).

### Running over HTTP

Set `MCP_TRANSPORT=http` and the server listens on `SERVER_PORT` at `MCP_PATH` (default `http://localhost:8080/mcp`).

---

## Configuration

Everything is configured through **environment variables**, and the result is validated before startup. Variables already present in the environment win over a `.env` file, so an MCP client's `env` block always takes effect. Every field has a sensible default, so the server runs with no configuration at all (stdio transport).

See [`.env.example`](.env.example) for the full list at its default values, ready to copy to `.env`.

### MCP server

| Env | Default | Notes |
| :--- | :--- | :--- |
| `MCP_TRANSPORT` | `stdio` | `stdio` or `http`. |
| `MCP_NAME` | `mcp-retrieval` | Server name advertised to clients. |
| `MCP_VERSION` | `0.1.0` | Server version advertised to clients. Identifies the build. |
| `MCP_PATH` | `/mcp` | HTTP route (http transport only). |

### HTTP server (http transport only)

| Env | Default |
| :--- | :--- |
| `SERVER_PORT` | `8080` |
| `SERVER_READ_TIMEOUT` | `60s` |
| `SERVER_WRITE_TIMEOUT` | `60s` |

### HTTP client and proxy

| Env | Default | Notes |
| :--- | :--- | :--- |
| `MAX_IDLE_CONNS_PER_HOST` | `100` | HTTP connection pooling. |
| `PROXY_HOST` | — | Optional. If set, requests are routed through a rotating-session proxy. |
| `PROXY_PORT` | — | Required when `PROXY_HOST` is set. |
| `PROXY_SCHEME` | — | Required when `PROXY_HOST` is set. |
| `PROXY_LOGIN` | — | Required when `PROXY_HOST` is set. |
| `PROXY_PASSWORD` | — | Required when `PROXY_HOST` is set. |

When a proxy is configured, each outbound request gets a unique session id appended to the login, so the upstream provider rotates the exit IP per request.

### Limits

| Env | Default |
| :--- | :--- |
| `MAX_QUERIES` | `10` |
| `DEFAULT_RESULTS` | `5` |
| `MAX_RESULTS` | `20` |
| `DEFAULT_TIMEOUT_MS` | `5000` |
| `MAX_TIMEOUT_MS` | `10000` |
| `MIN_TIMEOUT_MS` | `1000` |
| `DEFAULT_IMAGES` | `5` |
| `MAX_IMAGES` | `10` |
| `DEFAULT_DOCUMENT_CHARS` | `20000` |
| `MAX_DOCUMENT_CHARS` | `20000` |

### Logging

| Env | Default | Notes |
| :--- | :--- | :--- |
| `LOG_MODE` | `local` | `local` → text handler at debug level; `prod` → JSON handler at info level. Logs go to `stderr`. |

---

## Architecture

The project follows a clean, layered structure. Dependencies point inward toward the domain, and each layer talks to the next through interfaces.

```
app/                       the Go module: sources plus its build files
                           (Dockerfile, .dockerignore, .goreleaser.yaml)

cmd/mcp-retrieval/main.go  entry point: parse flags, load config, run app

internal/
  app/                     wiring + lifecycle (build server, run, graceful shutdown)
  config/                  config loading (.env → env vars → validate)
  domain/                  core types (Query, Link, Document, Snippet, Image) and errors
  dto/web/                 request/response shapes for the MCP tools
  transport/mcp/           MCP layer
    router/                registers every tool group on the MCP server
    web/                   tool handlers
    utils/                 schema helpers and error → tool-result mapping
  usecase/web/             business logic: validation, parallelism, timeouts, dedupe/limit/rerank
  adapter/web/             retrieval-go client wiring (search, images, scrape, proxy)
  pkg/                     reusable building blocks (mcpserver, server, logger, validator)
```

Request flow for a tool call:

```
MCP client → transport/mcp/web (handler) → usecase/web → adapter/web → retrieval-go → the web
                     ↑ maps errors               ↑ validates, fans out, limits results
```

Search and scrape both fan out across the input list concurrently and aggregate per-item results, each with its own status (`success`, `failed`, `timeout`). A call only fails outright when **every** item in it fails.

---

## Retrieval engine

All network work is delegated to [`retrieval-go`](https://github.com/free-llms-foundation/retrieval-go), configured in [`app/internal/adapter/web`](app/internal/adapter/web/retrieval.go). Worth knowing:

- **Sources.** Web search uses **DuckDuckGo Lite**; image search uses **Bing Images**; page fetching runs the raw HTML through a **readability** extractor and converts the main article to **Markdown** (tables included). No search-engine API keys are required.
- **Browser impersonation.** The adapter enables `WithBrowserRotation()`, so each request is sent from one of ~11 real browser profiles picked at random. Every profile pairs a genuine **TLS/JA3 fingerprint** (via [uTLS](https://github.com/refraction-networking/utls)) with a matching `User-Agent` and client-hint headers — Chrome 133/131/120 (Windows/macOS/Linux), Edge 131, Firefox 120 (Windows/macOS), Safari 18.4 (macOS), and iOS 18.4 Safari. This makes the traffic look like ordinary browsers rather than a Go HTTP client, which is what keeps the free sources reachable.
- **Proxy rotation.** When `PROXY_HOST` is configured, the adapter installs a proxy factory that appends a unique `session-<id>` to the proxy username on every request. With a session-based residential/rotating proxy provider, that yields a **fresh exit IP per request**, spreading load and avoiding rate limits. Without a proxy, requests go out directly.
- **Response handling.** Responses are transparently decompressed (`gzip`, `br`, `zstd`, `deflate`), and keep-alive is disabled (`WithDisableKeepAlive()`) so pooled connections don't pin a single fingerprint/IP across requests.

None of this needs configuration to work — the defaults above are applied automatically. Only proxy credentials are optional extras.

## Development

Everything Go lives in `app/`, so either use the makefile from the repository
root or pass `-C app` to the toolchain:

```bash
make build          # compile the binary
make test           # run tests

go -C app build ./...      # compile everything
go -C app test ./...       # run tests
go -C app vet ./...        # static checks
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for pull-request guidelines.

## License

Released under the [MIT License](LICENSE).
