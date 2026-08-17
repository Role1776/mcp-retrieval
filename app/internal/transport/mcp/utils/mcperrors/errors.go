package mcperrors

import (
	"errors"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func Result(err error) *mcpsdk.CallToolResult {
	var message string
	var isFatal bool

	switch {
	case errors.Is(err, domain.ErrUnexpectedStatusCode):
		message = "upstream service unavailable"
		isFatal = true
	case errors.Is(err, domain.ErrInvalidRequest):
		message = "invalid request"
		isFatal = true

	case errors.Is(err, domain.ErrTooManyQueries):
		message = "too many queries"
		isFatal = true

	case errors.Is(err, domain.ErrTooManyURLs):
		message = "too many urls"
		isFatal = true

	case errors.Is(err, domain.ErrEmptyQuery):
		message = "query must not be empty"
		isFatal = true

	case errors.Is(err, domain.ErrQueryTooLong):
		message = "query is too long"
		isFatal = true

	case errors.Is(err, domain.ErrInvalidURL):
		message = "invalid url"
		isFatal = true

	case errors.Is(err, domain.ErrRobotsDenied):
		message = "robots.txt denied"

	case errors.Is(err, domain.ErrAllURLsFailed):
		message = "every url failed to be scraped; the pages may be unreachable or hold no extractable text"
		isFatal = true

	case errors.Is(err, domain.ErrAllQueriesFailed):
		message = "every query failed; the search upstream may be unreachable"
		isFatal = true

	default:
		message = "internal server error"
		isFatal = true
	}

	return &mcpsdk.CallToolResult{
		IsError: isFatal,
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: message}},
	}
}
