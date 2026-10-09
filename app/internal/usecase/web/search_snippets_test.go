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

func TestSearchSnippets_Validation(t *testing.T) {
	cases := []struct {
		name string

		queries []string

		wantErr error
	}{
		{name: "more than 10 queries", queries: strings.Split(strings.Repeat("q,", 11), ",")[:11], wantErr: domain.ErrTooManyQueries},
		{name: "empty query rejects the call", queries: []string{"ok", ""}, wantErr: domain.ErrEmptyQuery},
		{name: "query longer than 512 characters", queries: []string{"ok", strings.Repeat("a", 513)}, wantErr: domain.ErrQueryTooLong},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			uc := newTestUseCase(fake)

			_, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: tc.queries})

			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, fake.recorded())
		})
	}
}

func TestSearchSnippets_NilQueriesIsError(t *testing.T) {
	fake := newFake(nil)
	uc := newTestUseCase(fake)

	_, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{})

	require.Error(t, err)
	assert.Empty(t, fake.recorded())
}

func TestSearchSnippets_AcceptedQueryLengths(t *testing.T) {
	cases := []struct {
		name string

		query string
	}{
		{name: "exactly 512 ASCII characters", query: strings.Repeat("a", 512)},
		{name: "512 Cyrillic letters", query: strings.Repeat("ж", 512)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{snippets: makeSnippets(t, 1)}
			uc := newTestUseCase(fake)

			resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{tc.query}})

			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			assert.Equal(t, statusSuccess, resp.Results[0].Status)
		})
	}
}

func TestSearchSnippets_TenQueriesAreAccepted(t *testing.T) {
	fake := newFake(nil)
	fake.fallback = behavior{snippets: makeSnippets(t, 1)}
	uc := newTestUseCase(fake)

	queries := []string{"q0", "q1", "q2", "q3", "q4", "q5", "q6", "q7", "q8", "q9"}

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: queries})

	require.NoError(t, err)
	assert.Len(t, resp.Results, 10)
}

func TestSearchSnippets_ResultsKeepInputOrderAndQuery(t *testing.T) {
	fake := newFake(map[string]behavior{
		"first":  {snippets: makeSnippets(t, 1)},
		"second": {snippets: makeSnippets(t, 2)},
		"third":  {snippets: makeSnippets(t, 3)},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"first", "second", "third"}})

	require.NoError(t, err)
	require.Len(t, resp.Results, 3)
	for i, want := range []string{"first", "second", "third"} {
		assert.Equal(t, want, resp.Results[i].Query)
		assert.Equal(t, i+1, resp.Results[i].Count)
		assert.Equal(t, statusSuccess, resp.Results[i].Status)
	}
}

func TestSearchSnippets_DateIsPassedUnchanged(t *testing.T) {
	cases := []struct {
		name string

		date string
	}{
		{name: "week", date: "w"},
		{name: "empty", date: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{snippets: makeSnippets(t, 1)}
			uc := newTestUseCase(fake)

			_, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a", "b"}, Date: tc.date})

			require.NoError(t, err)
			calls := fake.recorded()
			require.Len(t, calls, 2)
			for _, c := range calls {
				assert.Equal(t, tc.date, c.date)
			}
		})
	}
}

func TestSearchSnippets_MaxResultsLimit(t *testing.T) {
	cases := []struct {
		name string

		maxResults int

		want int
	}{
		{name: "zero uses default 5", maxResults: 0, want: 5},
		{name: "explicit 3", maxResults: 3, want: 3},
		{name: "above 20 is clamped to 20", maxResults: 100, want: 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{snippets: makeSnippets(t, 30)}
			uc := newTestUseCase(fake)

			resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a"}, MaxResults: tc.maxResults})

			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			assert.Len(t, resp.Results[0].Snippets, tc.want)
			assert.Equal(t, tc.want, resp.Results[0].Count)
		})
	}
}

func TestSearchSnippets_MaxResultsIsPerQuery(t *testing.T) {
	fake := newFake(nil)
	fake.fallback = behavior{snippets: makeSnippets(t, 30)}
	uc := newTestUseCase(fake)

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a", "b"}, MaxResults: 3})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	for _, r := range resp.Results {
		assert.Len(t, r.Snippets, 3)
		assert.Equal(t, 3, r.Count)
	}
}

func TestSearchSnippets_FailedQueryDoesNotAffectOthers(t *testing.T) {
	fake := newFake(map[string]behavior{
		"good": {snippets: makeSnippets(t, 2)},
		"bad":  {err: errors.New("upstream down")},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"bad", "good"}})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.Equal(t, statusFailed, resp.Results[0].Status)
	assert.Equal(t, statusSuccess, resp.Results[1].Status)
	assert.Equal(t, 2, resp.Results[1].Count)
}

func TestSearchSnippets_AllFailedReturnsError(t *testing.T) {
	fake := newFake(nil)
	fake.fallback = behavior{err: errors.New("upstream down")}
	uc := newTestUseCase(fake)

	_, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a", "b"}})

	assert.ErrorIs(t, err, domain.ErrAllQueriesFailed)
}

func TestSearchSnippets_TimeoutClamping(t *testing.T) {
	cases := []struct {
		name string

		timeoutMs int64

		min time.Duration
		max time.Duration
	}{
		{name: "default when zero", timeoutMs: 0, min: 4000 * time.Millisecond, max: 5000 * time.Millisecond},
		{name: "below minimum is clamped to 1000", timeoutMs: 10, min: 500 * time.Millisecond, max: 1000 * time.Millisecond},
		{name: "within range is kept", timeoutMs: 3000, min: 2000 * time.Millisecond, max: 3000 * time.Millisecond},
		{name: "above maximum is clamped to 10000", timeoutMs: 60000, min: 9000 * time.Millisecond, max: 10000 * time.Millisecond},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFake(nil)
			fake.fallback = behavior{snippets: makeSnippets(t, 1)}
			uc := newTestUseCase(fake)

			_, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a"}, TimeoutMs: tc.timeoutMs})

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

func TestSearchSnippets_UnfinishedQueryGetsTimeoutStatus(t *testing.T) {
	t.Parallel()

	fake := newFake(map[string]behavior{
		"fast": {snippets: makeSnippets(t, 1)},
		"slow": {block: true},
	})
	uc := newTestUseCase(fake)

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"slow", "fast"}, TimeoutMs: 1000})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.Equal(t, statusTimeout, resp.Results[0].Status)
	assert.Equal(t, statusSuccess, resp.Results[1].Status)
}

func TestSearchSnippets_AllTimedOutIsNotAnError(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a", "b"}, TimeoutMs: 1000})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	for i, q := range []string{"a", "b"} {
		assert.Equal(t, q, resp.Results[i].Query)
		assert.Equal(t, statusTimeout, resp.Results[i].Status)
	}
}
