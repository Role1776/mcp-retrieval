package web

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
)

// callerCancelAfter returns a context that the caller cancels after d.
func callerCancelAfter(t *testing.T, d time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	time.AfterFunc(d, cancel)

	return ctx
}

func TestSearchSnippets_CallerCanceledReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	_, err := uc.SearchSnippets(callerCancelAfter(t, 100*time.Millisecond), dto.SearchRequest{Queries: []string{"a", "b"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCanceled)
	assert.NotErrorIs(t, err, domain.ErrAllQueriesFailed)
}

func TestSearchSnippets_CallerCanceledWithOrdinaryFailureReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	fake := newFake(map[string]behavior{
		"bad":  {err: errors.New("upstream down")},
		"slow": {block: true},
	})
	uc := newTestUseCase(fake)

	_, err := uc.SearchSnippets(callerCancelAfter(t, 100*time.Millisecond), dto.SearchRequest{Queries: []string{"bad", "slow"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCanceled)
	assert.NotErrorIs(t, err, domain.ErrAllQueriesFailed)
}

func TestSearchSnippets_CallTimeoutIsNotCancellation(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	resp, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a", "b"}, TimeoutMs: 1000})

	require.NoError(t, err)
	assert.NotErrorIs(t, err, domain.ErrCanceled)
	require.Len(t, resp.Results, 2)
	for _, r := range resp.Results {
		assert.Equal(t, "timeout", r.Status)
	}
}

func TestSearchSnippets_OrdinaryFailuresWithoutCancelStillReturnAllFailed(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{err: errors.New("upstream down")}
	uc := newTestUseCase(fake)

	_, err := uc.SearchSnippets(context.Background(), dto.SearchRequest{Queries: []string{"a", "b"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAllQueriesFailed)
	assert.NotErrorIs(t, err, domain.ErrCanceled)
}

func TestSearchImages_CallerCanceledReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	_, err := uc.SearchImages(callerCancelAfter(t, 100*time.Millisecond), dto.ImagesSearchRequest{Queries: []string{"a", "b"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCanceled)
	assert.NotErrorIs(t, err, domain.ErrAllQueriesFailed)
}

func TestSearchImages_CallerCanceledWithOrdinaryFailureReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	fake := newFake(map[string]behavior{
		"bad":  {err: errors.New("upstream down")},
		"slow": {block: true},
	})
	uc := newTestUseCase(fake)

	_, err := uc.SearchImages(callerCancelAfter(t, 100*time.Millisecond), dto.ImagesSearchRequest{Queries: []string{"bad", "slow"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCanceled)
	assert.NotErrorIs(t, err, domain.ErrAllQueriesFailed)
}

func TestSearchImages_CallTimeoutIsNotCancellation(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	resp, err := uc.SearchImages(context.Background(), dto.ImagesSearchRequest{Queries: []string{"a", "b"}, TimeoutMs: 1000})

	require.NoError(t, err)
	assert.NotErrorIs(t, err, domain.ErrCanceled)
	require.Len(t, resp.Results, 2)
	for _, r := range resp.Results {
		assert.Equal(t, "timeout", r.Status)
	}
}

func TestSearchImages_OrdinaryFailuresWithoutCancelStillReturnAllFailed(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{err: errors.New("upstream down")}
	uc := newTestUseCase(fake)

	_, err := uc.SearchImages(context.Background(), dto.ImagesSearchRequest{Queries: []string{"a", "b"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAllQueriesFailed)
	assert.NotErrorIs(t, err, domain.ErrCanceled)
}

func TestScrapePages_CallerCanceledReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	_, err := uc.ScrapePages(callerCancelAfter(t, 100*time.Millisecond), dto.ScrapeRequest{
		URLs: []string{"https://example.com/a", "https://example.com/b"},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCanceled)
	assert.NotErrorIs(t, err, domain.ErrAllURLsFailed)
}

func TestScrapePages_CallerCanceledWithOrdinaryFailureReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	fake := newFake(map[string]behavior{
		"https://example.com/bad":  {err: errors.New("fetch failed")},
		"https://example.com/slow": {block: true},
	})
	uc := newTestUseCase(fake)

	_, err := uc.ScrapePages(callerCancelAfter(t, 100*time.Millisecond), dto.ScrapeRequest{
		URLs: []string{"https://example.com/bad", "https://example.com/slow"},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCanceled)
	assert.NotErrorIs(t, err, domain.ErrAllURLsFailed)
}

func TestScrapePages_CallTimeoutIsNotCancellation(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{block: true}
	uc := newTestUseCase(fake)

	urls := []string{"https://example.com/a", "https://example.com/b"}

	resp, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: urls, TimeoutMs: 1000})

	require.NoError(t, err)
	assert.NotErrorIs(t, err, domain.ErrCanceled)
	require.Len(t, resp.Results, 2)
	for _, r := range resp.Results {
		assert.Equal(t, "timeout", r.Status)
	}
}

func TestScrapePages_OrdinaryFailuresWithoutCancelStillReturnAllFailed(t *testing.T) {
	t.Parallel()

	fake := newFake(nil)
	fake.fallback = behavior{err: errors.New("fetch failed")}
	uc := newTestUseCase(fake)

	_, err := uc.ScrapePages(context.Background(), dto.ScrapeRequest{URLs: []string{"https://example.com/a", "https://example.com/b"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAllURLsFailed)
	assert.NotErrorIs(t, err, domain.ErrCanceled)
}
