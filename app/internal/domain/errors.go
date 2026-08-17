package domain

import "errors"

var (
	ErrRobotsDenied         = errors.New("robots denied")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrTooManyQueries       = errors.New("too many queries")
	ErrTooManyURLs          = errors.New("too many urls")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrEmptyQuery           = errors.New("empty query")
	ErrQueryTooLong         = errors.New("query too long")
	ErrInvalidURL           = errors.New("invalid url")
	ErrNoChunks             = errors.New("no chunks")
	ErrAllURLsFailed        = errors.New("all urls failed")
	ErrAllQueriesFailed     = errors.New("all queries failed")
	ErrNoRelevantImages     = errors.New("no relevant images")
)
