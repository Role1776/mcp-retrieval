package web

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
)

const (
	inflightWaitLimit = 5 * time.Second
	// inflightSettle is the short window in which a cap violation must show up
	// as an extra retriever entry; it bounds negative checks only.
	inflightSettle = 100 * time.Millisecond
)

// inflightGauge tracks one retriever method: calls running now, the peak of
// that number, and the total number of calls that reached the retriever.
type inflightGauge struct {
	current atomic.Int64
	peak    atomic.Int64
	entered atomic.Int64
}

func (g *inflightGauge) enter() func() {
	now := g.current.Add(1)
	for {
		p := g.peak.Load()
		if now <= p || g.peak.CompareAndSwap(p, now) {
			break
		}
	}
	g.entered.Add(1)

	return func() { g.current.Add(-1) }
}

// inflightRetriever blocks every call until the test releases it: one token
// on gate releases one call, open() releases all of them. Each method keeps
// its own gauge so the search, images and scrape limits can be checked
// separately.
type inflightRetriever struct {
	search inflightGauge
	images inflightGauge
	scrape inflightGauge

	gate     chan struct{}
	openCh   chan struct{}
	openOnce sync.Once
}

func newInflightRetriever(t *testing.T) *inflightRetriever {
	t.Helper()

	f := &inflightRetriever{gate: make(chan struct{}), openCh: make(chan struct{})}
	t.Cleanup(f.open)

	return f
}

func (f *inflightRetriever) open() {
	f.openOnce.Do(func() { close(f.openCh) })
}

func (f *inflightRetriever) pass(t *testing.T) {
	t.Helper()

	select {
	case f.gate <- struct{}{}:
	case <-time.After(inflightWaitLimit):
		t.Fatal("no blocked retriever call to release")
	}
}

func (f *inflightRetriever) block(ctx context.Context) error {
	select {
	case <-f.gate:
		return nil
	case <-f.openCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *inflightRetriever) Search(ctx context.Context, _ web.Query, _ string) (web.Snippets, error) {
	defer f.search.enter()()

	if err := f.block(ctx); err != nil {
		return nil, err
	}

	s, err := web.NewSnippet(web.SnippetProps{
		Link:    "https://example.com/1",
		Title:   "title",
		Rank:    1,
		Source:  "example.com",
		Snippet: "snippet",
	})
	if err != nil {
		return nil, err
	}

	return web.Snippets{s}, nil
}

func (f *inflightRetriever) Images(ctx context.Context, _ web.Query, _ string) (web.Images, error) {
	defer f.images.enter()()

	if err := f.block(ctx); err != nil {
		return nil, err
	}

	img, err := web.NewImage(web.ImageProps{
		URL:         "https://example.com/img/1.png",
		PageURL:     "https://example.com/page/1",
		Description: "image",
	})
	if err != nil {
		return nil, err
	}

	return web.Images{img}, nil
}

func (f *inflightRetriever) Scrape(ctx context.Context, _ web.Link, _ bool) (web.Document, error) {
	defer f.scrape.enter()()

	if err := f.block(ctx); err != nil {
		return web.Document{}, err
	}

	return web.NewDocument(web.DocumentProps{Title: "title", Markdown: "body", Length: 4})
}

func newInflightUseCase(f *inflightRetriever, searchCap, imagesCap int) *UseCase {
	cfg := &SearchConfig{
		MaxQueries:           10,
		DefaultResults:       5,
		MaxResults:           20,
		DefaultTimeoutMs:     5000,
		MaxTimeoutMs:         10000,
		MinTimeoutMs:         50,
		DefaultImages:        5,
		MaxImages:            10,
		DefaultDocumentChars: 20000,
		MaxDocumentChars:     20000,
		MaxSearchInFlight:    searchCap,
		MaxImagesInFlight:    imagesCap,
	}

	return New(f, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)
}

// inflightOutcome is what a call returned, copied out of the response so the
// helpers do not depend on the response type.
type inflightOutcome struct {
	err      error
	statuses []string
	counts   []int
}

func inflightGoSearch(uc *UseCase, req dto.SearchRequest) <-chan inflightOutcome {
	ch := make(chan inflightOutcome, 1)

	go func() {
		resp, err := uc.SearchSnippets(context.Background(), req)
		out := inflightOutcome{err: err}
		if err == nil {
			for _, r := range resp.Results {
				out.statuses = append(out.statuses, r.Status)
				out.counts = append(out.counts, r.Count)
			}
		}
		ch <- out
	}()

	return ch
}

func inflightGoImages(uc *UseCase, req dto.ImagesSearchRequest) <-chan inflightOutcome {
	ch := make(chan inflightOutcome, 1)

	go func() {
		resp, err := uc.SearchImages(context.Background(), req)
		out := inflightOutcome{err: err}
		if err == nil {
			for _, r := range resp.Results {
				out.statuses = append(out.statuses, r.Status)
				out.counts = append(out.counts, r.Count)
			}
		}
		ch <- out
	}()

	return ch
}

func inflightGoScrape(uc *UseCase, req dto.ScrapeRequest) <-chan inflightOutcome {
	ch := make(chan inflightOutcome, 1)

	go func() {
		resp, err := uc.ScrapePages(context.Background(), req)
		out := inflightOutcome{err: err}
		if err == nil {
			for _, r := range resp.Results {
				out.statuses = append(out.statuses, r.Status)
			}
		}
		ch <- out
	}()

	return ch
}

func inflightAwait(t *testing.T, ch <-chan inflightOutcome) inflightOutcome {
	t.Helper()

	select {
	case out := <-ch:
		return out
	case <-time.After(inflightWaitLimit):
		t.Fatal("call did not return")

		return inflightOutcome{}
	}
}

// inflightWaitUntil polls cond until it holds or the deadline passes.
func inflightWaitUntil(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(inflightWaitLimit)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// inflightStaysAtMost fails if counter goes above limit during the settle window.
func inflightStaysAtMost(t *testing.T, what string, counter *atomic.Int64, limit int64) {
	t.Helper()

	deadline := time.Now().Add(inflightSettle)
	for time.Now().Before(deadline) {
		if got := counter.Load(); got > limit {
			t.Fatalf("%s: %d calls reached the retriever, want at most %d", what, got, limit)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSearchInFlight_SearchCapOneRunsQueriesOneAtATime(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 1, 0)

	done := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"a", "b", "c"}, TimeoutMs: 10000})

	for n := int64(1); n <= 3; n++ {
		inflightWaitUntil(t, "next search call to reach the retriever", func() bool {
			return f.search.entered.Load() >= n
		})
		inflightStaysAtMost(t, "search cap 1", &f.search.entered, n)
		assert.Equal(t, int64(1), f.search.current.Load(), "running search calls before release %d", n)
		f.pass(t)
	}

	out := inflightAwait(t, done)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess, statusSuccess, statusSuccess}, out.statuses)
	assert.Equal(t, []int{1, 1, 1}, out.counts)
	assert.Equal(t, int64(1), f.search.peak.Load())
}

func TestSearchInFlight_SearchCapTwoIsSharedByConcurrentCalls(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 2, 0)

	first := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"a1", "a2", "a3"}, TimeoutMs: 10000})
	second := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"b1", "b2", "b3"}, TimeoutMs: 10000})

	inflightWaitUntil(t, "two search calls running", func() bool {
		return f.search.current.Load() == 2
	})
	inflightStaysAtMost(t, "search cap 2 across two calls", &f.search.entered, 2)

	f.open()

	for _, done := range []<-chan inflightOutcome{first, second} {
		out := inflightAwait(t, done)
		require.NoError(t, out.err)
		assert.Equal(t, []string{statusSuccess, statusSuccess, statusSuccess}, out.statuses)
	}

	assert.Equal(t, int64(2), f.search.peak.Load())
	assert.Equal(t, int64(6), f.search.entered.Load())
}

func TestSearchInFlight_SearchCapZeroIsUnlimited(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 0, 0)

	done := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"a", "b", "c"}, TimeoutMs: 10000})

	inflightWaitUntil(t, "all three search queries running at once", func() bool {
		return f.search.current.Load() == 3
	})

	f.open()

	out := inflightAwait(t, done)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess, statusSuccess, statusSuccess}, out.statuses)
	assert.Equal(t, int64(3), f.search.peak.Load())
}

func TestSearchInFlight_SearchCapZeroIsUnlimitedAcrossCalls(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 0, 0)

	first := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"a1", "a2", "a3"}, TimeoutMs: 10000})
	second := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"b1", "b2", "b3"}, TimeoutMs: 10000})

	inflightWaitUntil(t, "all six search queries running at once", func() bool {
		return f.search.current.Load() == 6
	})

	f.open()

	for _, done := range []<-chan inflightOutcome{first, second} {
		out := inflightAwait(t, done)
		require.NoError(t, out.err)
		assert.Equal(t, []string{statusSuccess, statusSuccess, statusSuccess}, out.statuses)
	}

	assert.Equal(t, int64(6), f.search.peak.Load())
}

func TestSearchInFlight_SearchSaturationDoesNotBlockImages(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 1, 0)

	holder := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"hold-1"}, TimeoutMs: 10000})
	inflightWaitUntil(t, "search slot taken by a blocked call", func() bool {
		return f.search.current.Load() == 1
	})

	queued := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"hold-2"}, TimeoutMs: 10000})
	inflightStaysAtMost(t, "search cap 1 while saturated", &f.search.entered, 1)

	images := inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"i1", "i2"}, TimeoutMs: 10000})
	inflightWaitUntil(t, "both image queries running while search is saturated", func() bool {
		return f.images.current.Load() == 2
	})
	assert.Equal(t, int64(1), f.search.current.Load(), "search slot must still be held")

	f.open()

	out := inflightAwait(t, holder)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess}, out.statuses)

	out = inflightAwait(t, queued)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess}, out.statuses)

	out = inflightAwait(t, images)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess, statusSuccess}, out.statuses)

	assert.Equal(t, int64(1), f.search.peak.Load())
	assert.Equal(t, int64(2), f.images.peak.Load())
}

func TestSearchInFlight_ImagesCapOneRunsQueriesOneAtATime(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 0, 1)

	first := inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"a1", "a2"}, TimeoutMs: 10000})
	second := inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"b1", "b2"}, TimeoutMs: 10000})

	for n := int64(1); n <= 4; n++ {
		inflightWaitUntil(t, "next image call to reach the retriever", func() bool {
			return f.images.entered.Load() >= n
		})
		inflightStaysAtMost(t, "images cap 1", &f.images.entered, n)
		assert.Equal(t, int64(1), f.images.current.Load(), "running image calls before release %d", n)
		f.pass(t)
	}

	for _, done := range []<-chan inflightOutcome{first, second} {
		out := inflightAwait(t, done)
		require.NoError(t, out.err)
		assert.Equal(t, []string{statusSuccess, statusSuccess}, out.statuses)
		assert.Equal(t, []int{1, 1}, out.counts)
	}

	assert.Equal(t, int64(1), f.images.peak.Load())
}

func TestSearchInFlight_ImagesCapZeroIsUnlimited(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 0, 0)

	done := inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"a", "b", "c"}, TimeoutMs: 10000})

	inflightWaitUntil(t, "all three image queries running at once", func() bool {
		return f.images.current.Load() == 3
	})

	f.open()

	out := inflightAwait(t, done)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess, statusSuccess, statusSuccess}, out.statuses)
	assert.Equal(t, int64(3), f.images.peak.Load())
}

func TestSearchInFlight_SearchTimesOutWhileWaitingForSlot(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 1, 0)

	holder := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"hold"}, TimeoutMs: 10000})
	inflightWaitUntil(t, "search slot taken by a blocked call", func() bool {
		return f.search.current.Load() == 1
	})

	waiter := inflightAwait(t, inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"x", "y"}, TimeoutMs: 200}))
	require.NoError(t, waiter.err, "a call whose items all timed out waiting for a slot is not an error")
	assert.Equal(t, []string{"timeout", "timeout"}, waiter.statuses)
	assert.Equal(t, int64(1), f.search.entered.Load(), "queries that never got a slot must not reach the retriever")

	f.open()

	out := inflightAwait(t, holder)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess}, out.statuses)
}

func TestSearchInFlight_ImagesTimeOutWhileWaitingForSlot(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 0, 1)

	holder := inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"hold"}, TimeoutMs: 10000})
	inflightWaitUntil(t, "images slot taken by a blocked call", func() bool {
		return f.images.current.Load() == 1
	})

	waiter := inflightAwait(t, inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"x", "y"}, TimeoutMs: 200}))
	require.NoError(t, waiter.err, "a call whose items all timed out waiting for a slot is not an error")
	assert.Equal(t, []string{"timeout", "timeout"}, waiter.statuses)
	assert.Equal(t, int64(1), f.images.entered.Load(), "queries that never got a slot must not reach the retriever")

	f.open()

	out := inflightAwait(t, holder)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess}, out.statuses)
}

func TestSearchInFlight_ScrapeIsNotLimitedBySearchCaps(t *testing.T) {
	t.Parallel()

	f := newInflightRetriever(t)
	uc := newInflightUseCase(f, 1, 1)

	searchHolder := inflightGoSearch(uc, dto.SearchRequest{Queries: []string{"hold"}, TimeoutMs: 10000})
	imagesHolder := inflightGoImages(uc, dto.ImagesSearchRequest{Queries: []string{"hold"}, TimeoutMs: 10000})
	inflightWaitUntil(t, "search and images slots both taken", func() bool {
		return f.search.current.Load() == 1 && f.images.current.Load() == 1
	})

	urls := []string{"https://example.com/1", "https://example.com/2", "https://example.com/3"}
	scrape := inflightGoScrape(uc, dto.ScrapeRequest{URLs: urls, TimeoutMs: 10000})

	inflightWaitUntil(t, "all three URLs being scraped at once", func() bool {
		return f.scrape.current.Load() == 3
	})

	f.open()

	out := inflightAwait(t, scrape)
	require.NoError(t, out.err)
	assert.Equal(t, []string{statusSuccess, statusSuccess, statusSuccess}, out.statuses)
	assert.Equal(t, int64(3), f.scrape.peak.Load())

	for _, done := range []<-chan inflightOutcome{searchHolder, imagesHolder} {
		out := inflightAwait(t, done)
		require.NoError(t, out.err)
		assert.Equal(t, []string{statusSuccess}, out.statuses)
	}
}
