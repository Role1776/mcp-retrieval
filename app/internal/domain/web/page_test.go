package web

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
)

func TestNewDocument(t *testing.T) {
	cases := []struct {
		name  string
		props DocumentProps

		want    Document
		wantErr bool
	}{
		{
			name: "full",
			props: DocumentProps{
				Title:     "Go",
				Byline:    "The Go Authors",
				Markdown:  "the language",
				Length:    12,
				Excerpt:   "an excerpt",
				SiteName:  "go.dev",
				MainImage: "https://go.dev/a.png",
				AllImages: []string{"https://go.dev/a.png"},
				Favicon:   "https://go.dev/favicon.ico",
				Language:  "en",
			},
			want: Document{
				title:     "Go",
				byline:    "The Go Authors",
				markdown:  "the language",
				length:    12,
				excerpt:   "an excerpt",
				siteName:  "go.dev",
				mainImage: "https://go.dev/a.png",
				allImages: []string{"https://go.dev/a.png"},
				favicon:   "https://go.dev/favicon.ico",
				language:  "en",
			},
		},
		{
			name:  "only the required fields",
			props: DocumentProps{Title: "Go", Markdown: "the language", Length: 12},
			want:  Document{title: "Go", markdown: "the language", length: 12},
		},
		{
			name: "nil all images",
			props: DocumentProps{
				Title:     "Go",
				Byline:    "The Go Authors",
				Markdown:  "the language",
				Length:    12,
				Excerpt:   "an excerpt",
				SiteName:  "go.dev",
				MainImage: "https://go.dev/a.png",
				AllImages: nil,
				Favicon:   "https://go.dev/favicon.ico",
				Language:  "en",
			},
			want: Document{
				title:     "Go",
				byline:    "The Go Authors",
				markdown:  "the language",
				length:    12,
				excerpt:   "an excerpt",
				siteName:  "go.dev",
				mainImage: "https://go.dev/a.png",
				allImages: nil,
				favicon:   "https://go.dev/favicon.ico",
				language:  "en",
			},
		},
		{name: "missing title", props: DocumentProps{Markdown: "the language", Length: 12}, wantErr: true},
		{name: "missing markdown", props: DocumentProps{Title: "Go", Length: 12}, wantErr: true},
		{
			name:    "negative length",
			props:   DocumentProps{Title: "Go", Markdown: "the language", Length: -1},
			wantErr: true,
		},
		{
			name:    "invalid main image url",
			props:   DocumentProps{Title: "Go", Markdown: "the language", Length: 12, MainImage: "invalid-url"},
			wantErr: true,
		},
		{
			name:    "invalid all images url",
			props:   DocumentProps{Title: "Go", Markdown: "the language", Length: 12, AllImages: []string{"invalid-url"}},
			wantErr: true,
		},
		{
			name:    "invalid favicon url",
			props:   DocumentProps{Title: "Go", Markdown: "the language", Length: 12, Favicon: "invalid-url"},
			wantErr: true,
		},
		{name: "empty props", props: DocumentProps{}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			doc, err := NewDocument(tc.props)

			// assert
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, Document{}, doc)

				return
			}

			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, doc.id)

			doc.id = uuid.Nil
			assert.Equal(t, tc.want, doc)
		})
	}
}

func TestDocumentIsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		markdown string

		want bool
	}{
		{name: "empty", markdown: "", want: true},
		{name: "whitespace only", markdown: "  \n\t\n ", want: true},
		{name: "text", markdown: "the language", want: false},
		{name: "text surrounded by whitespace", markdown: "\n text \n", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			doc := Document{markdown: tc.markdown}

			// act & assert
			assert.Equal(t, tc.want, doc.IsEmpty())
		})
	}
}

func TestDocumentRemoveAllLinks(t *testing.T) {
	cases := []struct {
		name     string
		markdown string
		length   int

		want       string
		wantLength int
	}{
		{
			name:       "link is replaced by its text",
			markdown:   "see [Go](https://go.dev) now",
			length:     28,
			want:       "see Go now",
			wantLength: 10,
		},
		{
			name:       "several links in one line",
			markdown:   "[a](https://a.example) and [b](https://b.example)",
			length:     49,
			want:       "a and b",
			wantLength: 7,
		},
		{
			name: "an image keeps its leading bang",
			// the regex does not know about images, so "![alt](url)" collapses to "!alt".
			markdown:   "![alt](https://img.example/a.png)",
			length:     33,
			want:       "!alt",
			wantLength: 4,
		},
		{
			name:       "empty link text is left alone",
			markdown:   "[](https://x.example)",
			length:     21,
			want:       "[](https://x.example)",
			wantLength: 21,
		},
		{
			name: "a closing paren inside the url cuts the match short",
			// documents the known limitation of the regex.
			markdown:   "[a](https://x.example/(paren))",
			length:     30,
			want:       "a)",
			wantLength: 2,
		},
		{
			name: "reference-style links are not touched and the length is left as is",
			// without "](" the method returns before recomputing Length.
			markdown:   "[a][b] plain",
			length:     999,
			want:       "[a][b] plain",
			wantLength: 999,
		},
		{
			name: "length is recounted in runes",
			// "héllo" is 5 runes but 6 bytes.
			markdown:   "see [héllo](https://x.example)",
			length:     0,
			want:       "see héllo",
			wantLength: 9,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			doc := Document{markdown: tc.markdown, length: tc.length}

			// act
			doc.RemoveAllLinks()

			// assert
			assert.Equal(t, tc.want, doc.markdown)
			assert.Equal(t, tc.wantLength, doc.length)
		})
	}
}

func TestDocumentTruncate(t *testing.T) {
	cases := []struct {
		name     string
		markdown string
		limit    int

		wantTruncated bool
		wantErr       error
	}{
		{name: "zero limit", markdown: strings.Repeat("a", 100), limit: 0},
		{name: "negative limit", markdown: strings.Repeat("a", 100), limit: -1},
		{name: "shorter than the limit", markdown: strings.Repeat("a", 10), limit: 50},
		{name: "exactly at the limit", markdown: strings.Repeat("a", 50), limit: 50},
		{name: "empty markdown", markdown: "", limit: 50},
		{
			name: "non-ascii under the limit is counted in runes",
			// 50 runes, 100 bytes: truncated if the limit were measured in bytes.
			markdown: strings.Repeat("é", 50),
			limit:    50,
		},
		{name: "one rune over the limit", markdown: strings.Repeat("a", 51), limit: 50, wantTruncated: true},
		{name: "cut on a word boundary", markdown: strings.Repeat("word ", 100), limit: 50, wantTruncated: true},
		{name: "cut on a paragraph boundary", markdown: strings.Repeat("para one.\n\npara two.\n\n", 20), limit: 40, wantTruncated: true},
		{name: "no separator to cut on", markdown: strings.Repeat("a", 300), limit: 50, wantTruncated: true},
		{name: "non-ascii over the limit", markdown: strings.Repeat("é ", 100), limit: 50, wantTruncated: true},
		{
			name: "whitespace only leaves the splitter with nothing to keep",
			// every chunk is dropped as blank, so there is no first chunk to cut down to.
			markdown: strings.Repeat(" ", 100),
			limit:    1,
			wantErr:  domain.ErrNoChunks,
		},
		{
			name:     "newlines only",
			markdown: "\n\n\n\n",
			limit:    1,
			wantErr:  domain.ErrNoChunks,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			doc := Document{markdown: tc.markdown}

			// act
			err := doc.Truncate(tc.limit)

			// assert
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, tc.markdown, doc.markdown, "a failed truncation must leave the document alone")
				assert.False(t, doc.truncated)
				assert.Zero(t, doc.length)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantTruncated, doc.truncated)

			if !tc.wantTruncated {
				assert.Equal(t, tc.markdown, doc.markdown)
				assert.Zero(t, doc.length, "Length is only recomputed on truncation")

				return
			}

			require.True(t, strings.HasSuffix(doc.markdown, truncationSuffix))

			body := strings.TrimSuffix(doc.markdown, truncationSuffix)
			assert.LessOrEqual(t, utf8.RuneCountInString(body), tc.limit)
			assert.True(t, strings.HasPrefix(tc.markdown, body), "truncation must keep the beginning of the text")
			assert.Equal(t, utf8.RuneCountInString(doc.markdown), doc.length, "Length counts the suffix too")
		})
	}
}

func TestDocumentText(t *testing.T) {
	cases := []struct {
		name string
		doc  Document

		want string
	}{
		{
			name: "title and body only",
			doc:  Document{title: "Go", markdown: "the language"},
			want: "# Go\n\nthe language",
		},
		{
			name: "with the site name",
			doc:  Document{title: "Go", siteName: "go.dev", markdown: "the language"},
			want: "# Go\nsource: go.dev\n\nthe language",
		},
		{
			name: "with the byline",
			doc:  Document{title: "Go", byline: "The Go Authors", markdown: "the language"},
			want: "# Go\nauthor: The Go Authors\n\nthe language",
		},
		{
			name: "the site name comes before the byline",
			doc:  Document{title: "Go", siteName: "go.dev", byline: "The Go Authors", markdown: "the language"},
			want: "# Go\nsource: go.dev\nauthor: The Go Authors\n\nthe language",
		},
		{
			name: "empty document",
			doc:  Document{},
			want: "# \n\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.doc.Text())
		})
	}
}
