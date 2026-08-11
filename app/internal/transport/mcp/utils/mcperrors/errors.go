package mcperrors

import (
	"errors"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func Result(err error) *mcpsdk.CallToolResult {
	var message string

	switch {
	case errors.Is(err, domain.ErrUnexpectedStatusCode):
		message = "upstream service unavailable"

	case errors.Is(err, domain.ErrInvalidRequest):
		message = "invalid request"

	case errors.Is(err, domain.ErrTooManyQueries):
		message = "too many queries"

	case errors.Is(err, domain.ErrTooManyURLs):
		message = "too many urls"

	case errors.Is(err, domain.ErrEmptyQuery):
		message = "query must not be empty"

	case errors.Is(err, domain.ErrQueryTooLong):
		message = "query is too long"

	case errors.Is(err, domain.ErrInvalidURL):
		message = "invalid url"

	case errors.Is(err, domain.ErrRobotsDenied):
		message = "robots.txt denied"

	case errors.Is(err, domain.ErrAllURLsFailed):
		message = "every url failed to be scraped; the pages may be unreachable or hold no extractable text"

	case errors.Is(err, domain.ErrAllQueriesFailed):
		message = "every query failed; the search upstream may be unreachable"

	default:
		message = "internal server error"
	}

	return &mcpsdk.CallToolResult{
		IsError: true,
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: message}},
	}
}
