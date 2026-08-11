package web

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/Role1776/mcp-retrieval/app/internal/pkg/validator"
)

type Snippet struct {
	ID      uuid.UUID `json:"id" validate:"required"`
	Link    string    `json:"link" validate:"required"`
	Title   string    `json:"title" validate:"required"`
	Rank    int       `json:"rank" validate:"required,gt=0"`
	Source  string    `json:"source" validate:"required"`
	Snippet string    `json:"snippet" validate:"required"`
	Favicon string    `json:"favicon" validate:"omitempty"`
}

type SnippetProps struct {
	Link    string
	Title   string
	Rank    int
	Source  string
	Snippet string
	Favicon string
}

func NewSnippet(props SnippetProps) (Snippet, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Snippet{}, err
	}

	snippet := Snippet{
		ID:      id,
		Title:   props.Title,
		Link:    props.Link,
		Source:  props.Source,
		Snippet: props.Snippet,
		Favicon: props.Favicon,
		Rank:    props.Rank,
	}

	if err := validator.Validate(snippet); err != nil {
		return Snippet{}, err
	}

	return snippet, nil
}

func (s Snippet) Host() string {
	parsed, err := url.Parse(s.Link)
	if err != nil {
		return ""
	}

	return strings.TrimPrefix(parsed.Host, "www.")
}

func (s Snippet) Reranked(rank int) Snippet {
	s.Rank = rank
	return s
}

func (s Snippet) Markdown() string {
	var b strings.Builder

	b.WriteString(strconv.Itoa(s.Rank))
	b.WriteString(". ")
	b.WriteString(s.Title)
	b.WriteString("\n")
	b.WriteString(s.Link)

	if s.Snippet != "" {
		b.WriteString("\n")
		b.WriteString(s.Snippet)
	}

	return b.String()
}

type Snippets []Snippet

func (s Snippets) Len() int {
	return len(s)
}

func (s Snippets) IsEmpty() bool {
	return len(s) == 0
}

func (s Snippets) Limit(n int) Snippets {
	if n <= 0 || len(s) == 0 {
		return Snippets{}
	}

	result := make(Snippets, min(len(s), n))
	copy(result, s)

	return result
}

func (s Snippets) Dedupe() Snippets {
	seen := make(map[string]struct{}, len(s))
	result := make(Snippets, 0, len(s))

	for _, snippet := range s {
		if _, ok := seen[snippet.Link]; ok {
			continue
		}
		seen[snippet.Link] = struct{}{}
		result = append(result, snippet)
	}

	return result
}

func (s Snippets) Rerank() Snippets {
	result := make(Snippets, 0, len(s))
	for i, snippet := range s {
		result = append(result, snippet.Reranked(i+1))
	}

	return result
}

func (s Snippets) SortedByRank() Snippets {
	result := make(Snippets, len(s))
	copy(result, s)

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Rank < result[j].Rank
	})

	return result
}

func (s Snippets) Markdown() string {
	parts := make([]string, 0, len(s))
	for _, snippet := range s {
		parts = append(parts, snippet.Markdown())
	}

	return strings.Join(parts, "\n\n")
}
