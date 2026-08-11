package web

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
)

func TestNewLink(t *testing.T) {
	// "https://example.com/" is exactly 20 bytes, so the padding below lands the
	// normalized link precisely on maxURLLength.
	const host = "https://example.com/"

	cases := []struct {
		name string
		raw  string

		want    string
		wantErr bool
	}{
		{name: "http", raw: "http://example.com/page", want: "http://example.com/page"},
		{name: "https", raw: "https://example.com/page", want: "https://example.com/page"},
		{name: "surrounding whitespace is trimmed", raw: "  https://example.com/page\n", want: "https://example.com/page"},
		{name: "uppercase scheme is lowercased", raw: "HTTPS://example.com/Page", want: "https://example.com/Page"},
		{name: "query and fragment are kept", raw: "https://example.com/p?q=1#top", want: "https://example.com/p?q=1#top"},
		{name: "non-ascii path is percent-encoded", raw: "https://example.com/é", want: "https://example.com/%C3%A9"},
		{name: "space in path is percent-encoded", raw: "https://example.com/a b", want: "https://example.com/a%20b"},
		{
			name: "non-ascii path within the limit once encoded",
			// each rune is 2 bytes raw and 6 bytes encoded.
			raw:  host + strings.Repeat("é", 100),
			want: host + strings.Repeat("%C3%A9", 100),
		},
		{
			name: "ascii path at the limit",
			raw:  host + strings.Repeat("a", maxURLLength-len(host)),
			want: host + strings.Repeat("a", maxURLLength-len(host)),
		},

		{
			name: "non-ascii path exceeding the limit only after encoding",
			// 1280 bytes raw, 3800 once encoded: the limit must be checked after parsing.
			raw:     host + strings.Repeat("é", 630),
			wantErr: true,
		},
		{name: "ascii path over the limit", raw: host + strings.Repeat("a", maxURLLength), wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "whitespace only", raw: "   \t\n", wantErr: true},
		{name: "no scheme", raw: "example.com/page", wantErr: true},
		{name: "scheme-relative", raw: "//example.com/page", wantErr: true},
		{name: "ftp scheme", raw: "ftp://example.com", wantErr: true},
		{name: "file scheme has no host", raw: "file:///etc/passwd", wantErr: true},
		{name: "javascript scheme", raw: "javascript:alert(1)", wantErr: true},
		{name: "data scheme", raw: "data:text/html,<h1>x</h1>", wantErr: true},
		{name: "https without host", raw: "https:///page", wantErr: true},
		{name: "space in host", raw: "http://exa mple.com", wantErr: true},
		{name: "unterminated ipv6 host", raw: "http://[::1", wantErr: true},
		{name: "control character", raw: "https://example.com/\x7f", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			link, err := NewLink(tc.raw)

			// assert
			if tc.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidURL)
				assert.Empty(t, link.String())

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, link.String())
			assert.LessOrEqual(t, len(link.String()), maxURLLength)
		})
	}
}

func TestNewLinks(t *testing.T) {
	cases := []struct {
		name string
		raw  []string

		want    []string
		wantErr bool
	}{
		{
			name: "all valid",
			raw:  []string{"https://a.example/1", "  http://b.example/2 "},
			want: []string{"https://a.example/1", "http://b.example/2"},
		},
		{name: "one invalid rejects the batch", raw: []string{"https://a.example/1", "not a url"}, wantErr: true},
		{name: "empty input", raw: nil, wantErr: true},
		{name: "empty slice", raw: []string{}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			links, err := NewLinks(tc.raw)

			// assert
			if tc.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidURL)
				assert.Nil(t, links)

				return
			}

			require.NoError(t, err)
			require.Len(t, links, len(tc.want))
			for i, want := range tc.want {
				assert.Equal(t, want, links[i].String())
			}
		})
	}
}

func TestLinkHost(t *testing.T) {
	cases := []struct {
		name string
		link Link

		want string
	}{
		{name: "plain host", link: Link{value: "https://example.com/page"}, want: "example.com"},
		{name: "host with port", link: Link{value: "https://example.com:8443/page?q=1"}, want: "example.com:8443"},
		{name: "userinfo is not part of the host", link: Link{value: "https://user@example.com/page"}, want: "example.com"},
		{name: "zero value", link: Link{}, want: ""},
		{name: "unparsable value", link: Link{value: "https://[::1"}, want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			host := tc.link.Host()

			// assert
			assert.Equal(t, tc.want, host)
		})
	}
}

func TestNewQuery(t *testing.T) {
	cases := []struct {
		name string
		raw  string

		want    string
		wantErr error
	}{
		{name: "plain", raw: "golang generics", want: "golang generics"},
		{name: "surrounding whitespace is trimmed", raw: "  golang generics\n", want: "golang generics"},
		{
			name: "non-ascii at the limit is counted in runes",
			// 512 runes of 2 bytes each: rejected if the limit were measured in bytes.
			raw:  strings.Repeat("é", maxQueryLength),
			want: strings.Repeat("é", maxQueryLength),
		},
		{name: "empty", raw: "", wantErr: domain.ErrEmptyQuery},
		{name: "whitespace only", raw: " \t\n", wantErr: domain.ErrEmptyQuery},
		{name: "ascii over the limit", raw: strings.Repeat("a", maxQueryLength+1), wantErr: domain.ErrQueryTooLong},
		{name: "non-ascii over the limit", raw: strings.Repeat("é", maxQueryLength+1), wantErr: domain.ErrQueryTooLong},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			query, err := NewQuery(tc.raw)

			// assert
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.True(t, query.IsZero())

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, query.String())
			assert.False(t, query.IsZero())
		})
	}
}

func TestNewQueries(t *testing.T) {
	cases := []struct {
		name string
		raw  []string

		want    []string
		wantErr error
	}{
		{name: "all valid", raw: []string{" first ", "second"}, want: []string{"first", "second"}},
		{name: "one invalid rejects the batch", raw: []string{"first", "  "}, wantErr: domain.ErrEmptyQuery},
		{
			name:    "one too long rejects the batch",
			raw:     []string{"first", strings.Repeat("a", maxQueryLength+1)},
			wantErr: domain.ErrQueryTooLong,
		},
		{name: "empty input", raw: nil, wantErr: domain.ErrEmptyQuery},
		{name: "empty slice", raw: []string{}, wantErr: domain.ErrEmptyQuery},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			queries, err := NewQueries(tc.raw)

			// assert
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, queries)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tc.want), queries.Len())
			assert.Equal(t, tc.want, queries.Strings())
		})
	}
}
