package web

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
)

const (
	maxQueryLength = 512
	// maxURLLength bounds the percent-encoded URL, which is pure ASCII, so bytes and runes match.
	maxURLLength = 2048
)

type Query struct {
	value string
}

func NewQuery(raw string) (Query, error) {
	value := strings.TrimSpace(raw)

	if value == "" {
		return Query{}, domain.ErrEmptyQuery
	}

	if utf8.RuneCountInString(value) > maxQueryLength {
		return Query{}, domain.ErrQueryTooLong
	}

	return Query{value: value}, nil
}

func (q Query) String() string {
	return q.value
}

func (q Query) IsZero() bool {
	return q.value == ""
}

type Queries []Query

func NewQueries(raw []string) (Queries, error) {
	queries := make(Queries, 0, len(raw))
	for _, r := range raw {
		query, err := NewQuery(r)
		if err != nil {
			return nil, err
		}
		queries = append(queries, query)
	}

	if len(queries) == 0 {
		return nil, domain.ErrEmptyQuery
	}

	return queries, nil
}

func (q Queries) Len() int {
	return len(q)
}

func (q Queries) Strings() []string {
	result := make([]string, 0, len(q))
	for _, query := range q {
		result = append(result, query.String())
	}

	return result
}

type Link struct {
	value string
}

func NewLink(raw string) (Link, error) {
	value := strings.TrimSpace(raw)

	if value == "" {
		return Link{}, domain.ErrInvalidURL
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return Link{}, domain.ErrInvalidURL
	}

	if parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return Link{}, domain.ErrInvalidURL
	}

	normalized := parsed.String()
	if len(normalized) > maxURLLength {
		return Link{}, domain.ErrInvalidURL
	}

	return Link{value: normalized}, nil
}

func NewLinks(raw []string) ([]Link, error) {
	links := make([]Link, 0, len(raw))
	for _, r := range raw {
		link, err := NewLink(r)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if len(links) == 0 {
		return nil, domain.ErrInvalidURL
	}

	return links, nil
}

func (l Link) String() string {
	return l.value
}

func (l Link) Host() string {
	parsed, err := url.Parse(l.value)
	if err != nil {
		return ""
	}

	return parsed.Host
}
