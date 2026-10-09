package web

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
)

func TestScrapePages_Validation(t *testing.T) {
	eleven := make([]string, 11)
	for i := range eleven {
		eleven[i] = "https://example.com/" + string(rune('a'+i))
	}

	cases := []struct {
		name string

		urls []string

		wantErr error
	}{
		{name: "more than 10 urls", urls: eleven, wantErr: domain.ErrTooManyURLs},
		{name: "not a url", urls: []string{"https://example.com/ok", "not a url"}, wantErr: domain.ErrInvalidURL},
		{name: "relative path", urls: []string{"https://example.com/ok", "/relative/path"}, wantErr: domain.ErrInvalidURL},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			uc := newTestUseCase(fake)

			_, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: tc.urls})

			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, fake.recorded())
		})
	}
}

func TestScrapePages_NilURLsIsError(t *testing.T) {
	fake := newFake(nil)
	uc := newTestUseCase(fake)

	_, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{})

	require.Error(t, err)
	assert.Empty(t, fake.recorded())
}

func TestScrapePages_ResultsKeepInputOrderAndTitle(t *testing.T) {
	urls := []string{"https://example.com/one", "https://example.com/two", "https://example.com/three"}
	fake := newFake(map[string]behavior{
		urls[0]: {doc: makeDoc(t, "One", "body one")},
		urls[1]: {doc: makeDoc(t, "Two", "body two")},
		urls[2]: {doc: makeDoc(t, "Three", "body three")},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: urls})

	require.NoError(t, err)
	require.Len(t, resp.Results, 3)
	for i, title := range []string{"One", "Two", "Three"} {
		assert.Equal(t, urls[i], resp.Results[i].URL)
		assert.Equal(t, statusSuccess, resp.Results[i].Status)
		assert.Equal(t, title, resp.Results[i].ScrapedData.Title)
	}
}

func TestScrapePages_RobotsTxtIsPassedUnchanged(t *testing.T) {
	cases := []struct {
		name string

		robotsTxt bool
	}{
		{name: "true", robotsTxt: true},
		{name: "false", robotsTxt: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{doc: makeDoc(t, "T", "body")}
			uc := newTestUseCase(fake)

			_, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: []string{"https://example.com/a"}, RobotsTxt: tc.robotsTxt})

			require.NoError(t, err)
			calls := fake.recorded()
			require.Len(t, calls, 1)
			assert.Equal(t, tc.robotsTxt, calls[0].robotsTxt)
		})
	}
}

func TestScrapePages_RobotsDeniedMarksOnlyThatURLFailed(t *testing.T) {
	fake := newFake(map[string]behavior{
		"https://example.com/denied": {err: domain.ErrRobotsDenied},
		"https://example.com/ok":     {doc: makeDoc(t, "OK", "body")},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{
		URLs:      []string{"https://example.com/denied", "https://example.com/ok"},
		RobotsTxt: true,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.Equal(t, statusFailed, resp.Results[0].Status)
	assert.Equal(t, statusSuccess, resp.Results[1].Status)
}

func TestScrapePages_AllFailedReturnsError(t *testing.T) {
	fake := newFake(nil)
	fake.fallback = behavior{err: errors.New("fetch failed")}
	uc := newTestUseCase(fake)

	_, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: []string{"https://example.com/a", "https://example.com/b"}})

	assert.ErrorIs(t, err, domain.ErrAllURLsFailed)
}

func TestScrapePages_RemoveLinks(t *testing.T) {
	const md = "see [docs](https://example.com/docs) here"

	cases := []struct {
		name string

		removeLinks bool

		wantMarkdown string
	}{
		{name: "links are replaced by their text", removeLinks: true, wantMarkdown: "see docs here"},
		{name: "markdown is unchanged when false", removeLinks: false, wantMarkdown: md},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{doc: makeDoc(t, "T", md)}
			uc := newTestUseCase(fake)

			resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: []string{"https://example.com/a"}, RemoveLinks: tc.removeLinks})

			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			assert.Equal(t, tc.wantMarkdown, resp.Results[0].ScrapedData.Markdown)
		})
	}
}

func TestScrapePages_MaxChars(t *testing.T) {
	cases := []struct {
		name string

		pageLen  int
		maxChars int

		wantTruncated bool
	}{
		{name: "zero uses default 20000, longer page is truncated", pageLen: 25000, maxChars: 0, wantTruncated: true},
		{name: "above 20000 is clamped to 20000", pageLen: 25000, maxChars: 50000, wantTruncated: true},
		{name: "explicit limit truncates a longer page", pageLen: 500, maxChars: 100, wantTruncated: true},
		{name: "page shorter than limit is unchanged", pageLen: 50, maxChars: 100, wantTruncated: false},
		{name: "page shorter than default is unchanged", pageLen: 1000, maxChars: 0, wantTruncated: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page := strings.Repeat("word ", tc.pageLen/5)
			fake := newFake(nil)
			fake.fallback = behavior{doc: makeDoc(t, "T", page)}
			uc := newTestUseCase(fake)

			resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: []string{"https://example.com/a"}, MaxChars: tc.maxChars})

			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			got := resp.Results[0].ScrapedData
			assert.Equal(t, tc.wantTruncated, got.Truncated)
			if tc.wantTruncated {
				assert.True(t, strings.HasSuffix(got.Markdown, "[truncated]"))
			} else {
				assert.Equal(t, page, got.Markdown)
			}
		})
	}
}

func TestScrapePages_MaxCharsIsPerPage(t *testing.T) {
	long := strings.Repeat("word ", 100)
	short := "tiny"
	fake := newFake(map[string]behavior{
		"https://example.com/long":  {doc: makeDoc(t, "Long", long)},
		"https://example.com/short": {doc: makeDoc(t, "Short", short)},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{
		URLs:     []string{"https://example.com/long", "https://example.com/short"},
		MaxChars: 100,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.True(t, resp.Results[0].ScrapedData.Truncated)
	assert.False(t, resp.Results[1].ScrapedData.Truncated)
	assert.Equal(t, short, resp.Results[1].ScrapedData.Markdown)
}

func TestScrapePages_TimeoutClamping(t *testing.T) {
	cases := []struct {
		name string

		timeoutMs int64

		min time.Duration
		max time.Duration
	}{
		{name: "default when zero", timeoutMs: 0, min: 4000 * time.Millisecond, max: 5000 * time.Millisecond},
		{name: "below minimum is clamped to 1000", timeoutMs: 10, min: 500 * time.Millisecond, max: 1000 * time.Millisecond},
		{name: "above maximum is clamped to 10000", timeoutMs: 60000, min: 9000 * time.Millisecond, max: 10000 * time.Millisecond},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{doc: makeDoc(t, "T", "body")}
			uc := newTestUseCase(fake)

			_, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: []string{"https://example.com/a"}, TimeoutMs: tc.timeoutMs})

			require.NoError(t, err)
			calls := fake.recorded()
			require.Len(t, calls, 1)
			require.True(t, calls[0].hasDL)
			remaining := time.Until(calls[0].deadline)
			assert.GreaterOrEqual(t, remaining, tc.min)
			assert.LessOrEqual(t, remaining, tc.max)
		})
	}
}

func TestScrapePages_UnfinishedURLGetsTimeoutStatus(t *testing.T) {
	t.Parallel()

	fake := newFake(map[string]behavior{
		"https://example.com/fast": {doc: makeDoc(t, "Fast", "body")},
		"https://example.com/slow": {block: true},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{
		URLs:      []string{"https://example.com/slow", "https://example.com/fast"},
		TimeoutMs: 1000,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.Equal(t, statusTimeout, resp.Results[0].Status)
	assert.Equal(t, statusSuccess, resp.Results[1].Status)
}
