package web

import (
	"context"

	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
	"github.com/Role1776/mcp-retrieval/app/internal/transport/mcp/utils/mcpschema"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolSearch       = "web_search"
	toolSearchImages = "web_search_images"
	toolScrape       = "web_scrape"
)

//go:generate mockgen -source=handler.go -destination=usecase_mock_test.go -package=web
type usecase interface {
	SearchSnippets(ctx context.Context, req dto.SearchRequest) (dto.SearchResponse, error)
	SearchImages(ctx context.Context, req dto.ImagesSearchRequest) (dto.ImagesSearchResponse, error)
	ScrapePages(ctx context.Context, req dto.ScrapeRequest) (dto.ScrapeResponse, error)
}

type Handler struct {
	usecase usecase
}

func New(usecase usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) RegisterTools(s *mcpsdk.Server) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: toolSearch,
		Annotations: &mcpsdk.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Title: "Web search",
		Description: "Searches the web for one or more queries and returns snippets with links. " +
			"Queries run in parallel. IMPORTANT: Always write search queries in English for best results and relevance. " +
			"For images use the '" + toolSearchImages + "' tool; " +
			"both tools can be called in the same turn when a query needs text and images.\n\n" +
			"Freshness: set the 'date' field to restrict results by recency — 'd' (past day), " +
			"'w' (past week), 'm' (past month), 'y' (past year). Use it to prefer the most recent " +
			"pages when the user asks about news or anything time-sensitive; leave it empty for all time.\n\n" +
			"Query operators (put them inside the query string itself): " +
			"'site:example.com term' limits the search to one site; 'filetype:pdf term' restricts to a file type; " +
			"double quotes \"exact phrase\" force an exact match; a leading '-' excludes a word (term -foo); " +
			"'intitle:term' requires the word in the page title. Operators can be combined, e.g. " +
			"'site:arxiv.org filetype:pdf transformers'.",
		OutputSchema: mcpschema.OutputFor[dto.SearchResponse](),
	}, h.search)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: toolSearchImages,
		Annotations: &mcpsdk.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Title: "Web image search",
		Description: "Searches for images by one or more queries, executed in parallel. " +
			"IMPORTANT: Queries must be in English ONLY.\n\n" +
			"Freshness: set the 'date' field to restrict results by recency — 'd' (past day), " +
			"'w' (past week), 'm' (past month), 'y' (past year); leave it empty for all time.\n\n" +
			"Query operators can be embedded in the query string: 'site:example.com term' limits to one site, " +
			"double quotes \"exact phrase\" force an exact match, and a leading '-' excludes a word.",
		OutputSchema: mcpschema.OutputFor[dto.ImagesSearchResponse](),
	}, h.searchImages)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: toolScrape,
		Annotations: &mcpsdk.ToolAnnotations{
			ReadOnlyHint: true,
		},
		Title: "Web scrape",
		Description: "Downloads pages by their links and returns the main text as markdown. " +
			"Links are fetched in parallel; a single call handles up to 10 of them, " +
			"so batch all links you need into one call instead of calling the tool per link.\n\n" +
			"Size: the 'max_chars' limit applies to each page separately, not to the whole " +
			"response — 5 links stay under the limit each and the response holds all 5 texts. " +
			"The limit counts characters, not tokens. A page cut short ends with '[truncated]' " +
			"and its 'truncated' field is set; the cut-off part cannot be fetched afterwards, " +
			"so treat what you got as all there is for that page.",
		OutputSchema: mcpschema.OutputFor[dto.ScrapeResponse](),
	}, h.scrape)
}
