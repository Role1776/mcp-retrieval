package web

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnippet(t *testing.T) {
	cases := []struct {
		name  string
		props SnippetProps

		want    Snippet
		wantErr bool
	}{
		{
			name: "full",
			props: SnippetProps{
				Link:    "https://go.dev",
				Title:   "Go",
				Rank:    1,
				Source:  "go.dev",
				Snippet: "the language",
				Favicon: "https://go.dev/favicon.ico",
			},
			want: Snippet{
				Link:    "https://go.dev",
				Title:   "Go",
				Rank:    1,
				Source:  "go.dev",
				Snippet: "the language",
				Favicon: "https://go.dev/favicon.ico",
			},
		},
		{
			name:  "favicon is optional",
			props: SnippetProps{Link: "https://go.dev", Title: "Go", Rank: 1, Source: "go.dev", Snippet: "the language"},
			want:  Snippet{Link: "https://go.dev", Title: "Go", Rank: 1, Source: "go.dev", Snippet: "the language"},
		},
		{
			name:    "missing link",
			props:   SnippetProps{Title: "Go", Rank: 1, Source: "go.dev", Snippet: "the language"},
			wantErr: true,
		},
		{
			name:    "missing title",
			props:   SnippetProps{Link: "https://go.dev", Rank: 1, Source: "go.dev", Snippet: "the language"},
			wantErr: true,
		},
		{
			name:    "missing source",
			props:   SnippetProps{Link: "https://go.dev", Title: "Go", Rank: 1, Snippet: "the language"},
			wantErr: true,
		},
		{
			name:    "missing snippet",
			props:   SnippetProps{Link: "https://go.dev", Title: "Go", Rank: 1, Source: "go.dev"},
			wantErr: true,
		},
		{
			name:    "zero rank",
			props:   SnippetProps{Link: "https://go.dev", Title: "Go", Rank: 0, Source: "go.dev", Snippet: "the language"},
			wantErr: true,
		},
		{
			name:    "negative rank",
			props:   SnippetProps{Link: "https://go.dev", Title: "Go", Rank: -1, Source: "go.dev", Snippet: "the language"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			snippet, err := NewSnippet(tc.props)

			// assert
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, Snippet{}, snippet)

				return
			}

			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, snippet.ID)

			snippet.ID = uuid.Nil
			assert.Equal(t, tc.want, snippet)
		})
	}
}

func TestSnippetHost(t *testing.T) {
	cases := []struct {
		name string
		link string

		want string
	}{
		{name: "plain host", link: "https://example.com/page", want: "example.com"},
		{name: "www is stripped", link: "https://www.example.com/page", want: "example.com"},
		{name: "www only as a prefix", link: "https://wwwx.example.com", want: "wwwx.example.com"},
		{name: "port is kept", link: "https://example.com:8443/page", want: "example.com:8443"},
		{name: "no scheme means no host", link: "example.com/page", want: ""},
		{name: "empty link", link: "", want: ""},
		{name: "unparsable link", link: "https://[::1", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			host := Snippet{Link: tc.link}.Host()

			// assert
			assert.Equal(t, tc.want, host)
		})
	}
}

func TestSnippetReranked(t *testing.T) {
	cases := []struct {
		name    string
		snippet Snippet
		rank    int

		want int
	}{
		{name: "raises the rank", snippet: Snippet{Rank: 1}, rank: 7, want: 7},
		{name: "lowers the rank", snippet: Snippet{Rank: 9}, rank: 2, want: 2},
		{name: "same rank", snippet: Snippet{Rank: 3}, rank: 3, want: 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			original := tc.snippet

			// act
			reranked := tc.snippet.Reranked(tc.rank)

			// assert
			assert.Equal(t, tc.want, reranked.Rank)
			assert.Equal(t, original, tc.snippet, "the receiver must not be mutated")
		})
	}
}

func TestSnippetMarkdown(t *testing.T) {
	cases := []struct {
		name    string
		snippet Snippet

		want string
	}{
		{
			name:    "with text",
			snippet: Snippet{Rank: 1, Title: "Go", Link: "https://go.dev", Snippet: "the language"},
			want:    "1. Go\nhttps://go.dev\nthe language",
		},
		{
			name:    "without text",
			snippet: Snippet{Rank: 2, Title: "Docs", Link: "https://go.dev/doc"},
			want:    "2. Docs\nhttps://go.dev/doc",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			markdown := tc.snippet.Markdown()

			// assert
			assert.Equal(t, tc.want, markdown)
		})
	}
}

func TestSnippetsLen(t *testing.T) {
	cases := []struct {
		name     string
		snippets Snippets

		want int
	}{
		{name: "nil", snippets: nil, want: 0},
		{name: "empty", snippets: Snippets{}, want: 0},
		{name: "two", snippets: Snippets{{Rank: 1}, {Rank: 2}}, want: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.snippets.Len())
		})
	}
}

func TestSnippetsIsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		snippets Snippets

		want bool
	}{
		{name: "nil", snippets: nil, want: true},
		{name: "empty", snippets: Snippets{}, want: true},
		{name: "one", snippets: Snippets{{Rank: 1}}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.snippets.IsEmpty())
		})
	}
}

func TestSnippetsLimit(t *testing.T) {
	three := Snippets{{Title: "a"}, {Title: "b"}, {Title: "c"}}

	cases := []struct {
		name     string
		snippets Snippets
		n        int

		want Snippets
	}{
		{name: "keeps the first n", snippets: three, n: 2, want: Snippets{{Title: "a"}, {Title: "b"}}},
		{name: "n above the length keeps everything", snippets: three, n: 10, want: three},
		{name: "n equal to the length keeps everything", snippets: three, n: 3, want: three},
		{name: "zero n", snippets: three, n: 0, want: Snippets{}},
		{name: "negative n", snippets: three, n: -1, want: Snippets{}},
		{name: "empty receiver", snippets: Snippets{}, n: 2, want: Snippets{}},
		{name: "nil receiver", snippets: nil, n: 2, want: Snippets{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			limited := tc.snippets.Limit(tc.n)

			// assert
			assert.Equal(t, tc.want, limited)

			if len(limited) > 0 {
				limited[0].Title = "mutated"
				assert.NotEqual(t, "mutated", tc.snippets[0].Title, "the result must not alias the receiver")
			}
		})
	}
}

func TestSnippetsDedupe(t *testing.T) {
	cases := []struct {
		name     string
		snippets Snippets

		want Snippets
	}{
		{
			name:     "keeps the first occurrence of a link",
			snippets: Snippets{{Link: "a", Title: "first"}, {Link: "b"}, {Link: "a", Title: "second"}},
			want:     Snippets{{Link: "a", Title: "first"}, {Link: "b"}},
		},
		{
			name:     "different titles under the same link collapse",
			snippets: Snippets{{Link: "a", Title: "x"}, {Link: "a", Title: "y"}},
			want:     Snippets{{Link: "a", Title: "x"}},
		},
		{
			name:     "nothing to dedupe",
			snippets: Snippets{{Link: "a"}, {Link: "b"}},
			want:     Snippets{{Link: "a"}, {Link: "b"}},
		},
		{name: "empty receiver", snippets: Snippets{}, want: Snippets{}},
		{name: "nil receiver", snippets: nil, want: Snippets{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			deduped := tc.snippets.Dedupe()

			// assert
			assert.Equal(t, tc.want, deduped)
		})
	}
}

func TestSnippetsRerank(t *testing.T) {
	cases := []struct {
		name     string
		snippets Snippets

		want []int
	}{
		{name: "ranks start at one", snippets: Snippets{{Rank: 0}, {Rank: 0}, {Rank: 0}}, want: []int{1, 2, 3}},
		{name: "existing ranks are overwritten in place order", snippets: Snippets{{Rank: 9}, {Rank: 4}}, want: []int{1, 2}},
		{name: "empty receiver", snippets: Snippets{}, want: []int{}},
		{name: "nil receiver", snippets: nil, want: []int{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			original := append(Snippets{}, tc.snippets...)

			// act
			reranked := tc.snippets.Rerank()

			// assert
			ranks := make([]int, 0, len(reranked))
			for _, snippet := range reranked {
				ranks = append(ranks, snippet.Rank)
			}

			assert.Equal(t, tc.want, ranks)
			assert.Equal(t, original, append(Snippets{}, tc.snippets...), "the receiver must not be mutated")
		})
	}
}

func TestSnippetsSortedByRank(t *testing.T) {
	cases := []struct {
		name     string
		snippets Snippets

		want Snippets
	}{
		{
			name:     "sorts ascending",
			snippets: Snippets{{Rank: 3, Title: "c"}, {Rank: 1, Title: "a"}, {Rank: 2, Title: "b"}},
			want:     Snippets{{Rank: 1, Title: "a"}, {Rank: 2, Title: "b"}, {Rank: 3, Title: "c"}},
		},
		{
			name:     "equal ranks keep their order",
			snippets: Snippets{{Rank: 1, Title: "first"}, {Rank: 1, Title: "second"}},
			want:     Snippets{{Rank: 1, Title: "first"}, {Rank: 1, Title: "second"}},
		},
		{name: "empty receiver", snippets: Snippets{}, want: Snippets{}},
		{name: "nil receiver", snippets: nil, want: Snippets{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			original := append(Snippets{}, tc.snippets...)

			// act
			sorted := tc.snippets.SortedByRank()

			// assert
			assert.Equal(t, tc.want, sorted)
			assert.Equal(t, original, append(Snippets{}, tc.snippets...), "the receiver must not be mutated")
		})
	}
}

func TestSnippetsMarkdown(t *testing.T) {
	cases := []struct {
		name     string
		snippets Snippets

		want string
	}{
		{
			name: "entries are separated by a blank line",
			snippets: Snippets{
				{Rank: 1, Title: "Go", Link: "https://go.dev", Snippet: "the language"},
				{Rank: 2, Title: "Docs", Link: "https://go.dev/doc"},
			},
			want: "1. Go\nhttps://go.dev\nthe language\n\n2. Docs\nhttps://go.dev/doc",
		},
		{
			name:     "single entry has no separator",
			snippets: Snippets{{Rank: 1, Title: "Go", Link: "https://go.dev"}},
			want:     "1. Go\nhttps://go.dev",
		},
		{name: "empty receiver", snippets: Snippets{}, want: ""},
		{name: "nil receiver", snippets: nil, want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.snippets.Markdown())
		})
	}
}
