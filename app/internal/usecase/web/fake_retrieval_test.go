package web

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
)

// behavior describes what the fake returns for one input string.
type behavior struct {
	snippets web.Snippets
	images   web.Images
	doc      web.Document
	err      error
	// block makes the call wait for ctx.Done() and return the context error.
	block bool
}

type call struct {
	input     string
	date      string
	robotsTxt bool
	deadline  time.Time
	hasDL     bool
}

type fakeRetriever struct {
	mu        sync.Mutex
	behaviors map[string]behavior
	fallback  behavior
	calls     []call
}

func newFake(behaviors map[string]behavior) *fakeRetriever {
	return &fakeRetriever{behaviors: behaviors}
}

func (f *fakeRetriever) handle(ctx context.Context, input, date string, robotsTxt bool) behavior {
	dl, ok := ctx.Deadline()

	f.mu.Lock()
	f.calls = append(f.calls, call{input: input, date: date, robotsTxt: robotsTxt, deadline: dl, hasDL: ok})
	b, found := f.behaviors[input]
	if !found {
		b = f.fallback
	}
	f.mu.Unlock()

	return b
}

func (f *fakeRetriever) Search(ctx context.Context, query web.Query, date string) (web.Snippets, error) {
	b := f.handle(ctx, query.String(), date, false)
	if b.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return b.snippets, b.err
}

func (f *fakeRetriever) Images(ctx context.Context, query web.Query, date string) (web.Images, error) {
	b := f.handle(ctx, query.String(), date, false)
	if b.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return b.images, b.err
}

func (f *fakeRetriever) Scrape(ctx context.Context, link web.Link, robotsTxt bool) (web.Document, error) {
	b := f.handle(ctx, link.String(), "", robotsTxt)
	if b.block {
		<-ctx.Done()
		return web.Document{}, ctx.Err()
	}

	return b.doc, b.err
}

func (f *fakeRetriever) recorded() []call {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]call(nil), f.calls...)
}

func newTestUseCase(f *fakeRetriever) *UseCase {
	cfg := &SearchConfig{
		MaxQueries:           10,
		DefaultResults:       5,
		MaxResults:           20,
		DefaultTimeoutMs:     5000,
		MaxTimeoutMs:         10000,
		MinTimeoutMs:         1000,
		DefaultImages:        5,
		MaxImages:            10,
		DefaultDocumentChars: 20000,
		MaxDocumentChars:     20000,
	}

	return New(f, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)
}

func makeSnippets(t *testing.T, n int) web.Snippets {
	t.Helper()

	out := make(web.Snippets, 0, n)
	for i := 1; i <= n; i++ {
		s, err := web.NewSnippet(web.SnippetProps{
			Link:    fmt.Sprintf("https://example.com/%d", i),
			Title:   fmt.Sprintf("title %d", i),
			Rank:    i,
			Source:  "example.com",
			Snippet: fmt.Sprintf("snippet %d", i),
		})
		require.NoError(t, err)
		out = append(out, s)
	}

	return out
}

func makeImages(t *testing.T, n int) web.Images {
	t.Helper()

	out := make(web.Images, 0, n)
	for i := 1; i <= n; i++ {
		img, err := web.NewImage(web.ImageProps{
			URL:         fmt.Sprintf("https://example.com/img/%d.png", i),
			PageURL:     fmt.Sprintf("https://example.com/page/%d", i),
			Description: fmt.Sprintf("image %d", i),
		})
		require.NoError(t, err)
		out = append(out, img)
	}

	return out
}

func makeDoc(t *testing.T, title, markdown string) web.Document {
	t.Helper()

	doc, err := web.NewDocument(web.DocumentProps{
		Title:    title,
		Markdown: markdown,
		Length:   len([]rune(markdown)),
	})
	require.NoError(t, err)

	return doc
}
