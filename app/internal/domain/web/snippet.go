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
	id      uuid.UUID
	link    string
	title   string
	rank    int
	source  string
	snippet string
	favicon string
}

type SnippetProps struct {
	Link    string `validate:"required"`
	Title   string `validate:"required"`
	Rank    int    `validate:"required,gt=0"`
	Source  string `validate:"required"`
	Snippet string `validate:"required"`
	Favicon string `validate:"omitempty"`
}

func NewSnippet(props SnippetProps) (Snippet, error) {
	if err := validator.Validate(props); err != nil {
		return Snippet{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Snippet{}, err
	}

	return Snippet{
		id:      id,
		title:   props.Title,
		link:    props.Link,
		source:  props.Source,
		snippet: props.Snippet,
		favicon: props.Favicon,
		rank:    props.Rank,
	}, nil
}

type SnippetView struct {
	ID      uuid.UUID
	Link    string
	Title   string
	Rank    int
	Source  string
	Snippet string
	Favicon string
}

func (s Snippet) View() SnippetView {
	return SnippetView{
		ID:      s.id,
		Link:    s.link,
		Title:   s.title,
		Rank:    s.rank,
		Source:  s.source,
		Snippet: s.snippet,
		Favicon: s.favicon,
	}
}

func (s Snippet) Host() string {
	parsed, err := url.Parse(s.link)
	if err != nil {
		return ""
	}

	return strings.TrimPrefix(parsed.Host, "www.")
}

func (s Snippet) Reranked(rank int) Snippet {
	s.rank = rank
	return s
}

func (s Snippet) Markdown() string {
	var b strings.Builder

	b.WriteString(strconv.Itoa(s.rank))
	b.WriteString(". ")
	b.WriteString(s.title)
	b.WriteString("\n")
	b.WriteString(s.link)

	if s.snippet != "" {
		b.WriteString("\n")
		b.WriteString(s.snippet)
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

func (s Snippets) Views() []SnippetView {
	views := make([]SnippetView, 0, len(s))
	for _, snippet := range s {
		views = append(views, snippet.View())
	}

	return views
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
		if _, ok := seen[snippet.link]; ok {
			continue
		}
		seen[snippet.link] = struct{}{}
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
		return result[i].rank < result[j].rank
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
