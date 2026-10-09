package web

import (
	"context"

	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"

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

func boolPtr(v bool) *bool { return &v }

func readOnlyAnnotations() *mcpsdk.ToolAnnotations {
	return &mcpsdk.ToolAnnotations{
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: boolPtr(false),
		OpenWorldHint:   boolPtr(true),
	}
}

func (h *Handler) RegisterTools(s *mcpsdk.Server) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        toolSearch,
		Annotations: readOnlyAnnotations(),
		Title:       "Web search",
		Description: "Searches the web for one or more queries and returns snippets with links. " +
			"Queries run in parallel. IMPORTANT: Always write search queries in English for best results and relevance. " +
			"Use this to find sources; then call '" + toolScrape + "' to read a full page, " +
			"or '" + toolSearchImages + "' when you need pictures (both can be called in the same turn).\n\n" +
			"Freshness: set the 'date' field to restrict results by recency - 'd' (past day), " +
			"'w' (past week), 'm' (past month), 'y' (past year). Use it to prefer the most recent " +
			"pages when the user asks about news or anything time-sensitive; leave it empty for all time.\n\n" +
			"Query operators (put them inside the query string itself): " +
			"'site:example.com term' limits the search to one site; 'filetype:pdf term' restricts to a file type; " +
			"double quotes \"exact phrase\" force an exact match; a leading '-' excludes a word (term -foo); " +
			"'intitle:term' requires the word in the page title. Operators can be combined, e.g. " +
			"'site:arxiv.org filetype:pdf transformers'.\n\n" +
			"Results are reported per query with a 'status' field (success, failed or timeout). " +
			"A failed query does not fail the others; the call errors only if every query fails. " +
			"The upstream search engines may rate-limit or block requests, which shows up as a failed or timed-out query - retry later or rephrase.",
		OutputSchema: outputFor[dto.SearchResponse](),
	}, h.search)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        toolSearchImages,
		Annotations: readOnlyAnnotations(),
		Title:       "Web image search",
		Description: "Searches for images by one or more queries, executed in parallel. " +
			"IMPORTANT: Queries must be in English ONLY. " +
			"Use this only when you need pictures; use '" + toolSearch + "' for text answers and links, " +
			"and '" + toolScrape + "' to read a page.\n\n" +
			"Freshness: set the 'date' field to restrict results by recency - 'd' (past day), " +
			"'w' (past week), 'm' (past month), 'y' (past year); leave it empty for all time.\n\n" +
			"Query operators can be embedded in the query string: 'site:example.com term' limits to one site, " +
			"double quotes \"exact phrase\" force an exact match, and a leading '-' excludes a word.\n\n" +
			"Results are reported per query with a 'status' field (success, no relevant, failed or timeout). " +
			"A failed query does not fail the others; the call errors only if every query fails. " +
			"The upstream image search may rate-limit or block requests, which shows up as a failed or timed-out query.",
		OutputSchema: outputFor[dto.ImagesSearchResponse](),
	}, h.searchImages)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        toolScrape,
		Annotations: readOnlyAnnotations(),
		Title:       "Web scrape",
		Description: "Downloads pages by their links and returns the main text as markdown. " +
			"Use this to read a page in full after '" + toolSearch + "' found it; snippets alone are often not enough. " +
			"Links are fetched in parallel, up to 10 per call, so batch all links into one call.\n\n" +
			"Size: 'max_chars' is a per-page character limit (not tokens). A page cut short ends with '[truncated]' " +
			"and has 'truncated' set; the rest cannot be fetched afterwards.\n\n" +
			"Results are reported per URL with a 'status' field (success, failed or timeout). " +
			"A failed URL does not fail the others; the call errors only if every URL fails. " +
			"Target sites may block or rate-limit the request, which shows up as a failed URL.",
		OutputSchema: outputFor[dto.ScrapeResponse](),
	}, h.scrape)
}
