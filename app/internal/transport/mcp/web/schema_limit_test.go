package web

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// schemaLimitMaxItems is deliberately small so that the cap must come from
// the value passed to New, not from any hard-coded default.
const schemaLimitMaxItems = 3

type schemaLimitCase struct {
	tool   string
	argKey string
	item   func(i int) string
	// expect registers a usecase expectation that counts invocations.
	expect func(uc *Mockusecase, calls *atomic.Int32)
}

func schemaLimitCases() []schemaLimitCase {
	return []schemaLimitCase{
		{
			tool:   "web_search",
			argKey: "queries",
			item:   func(i int) string { return fmt.Sprintf("query %d", i) },
			expect: func(uc *Mockusecase, calls *atomic.Int32) {
				uc.EXPECT().SearchSnippets(gomock.Any(), gomock.Any()).
					DoAndReturn(func(context.Context, dto.SearchRequest) (dto.SearchResponse, error) {
						calls.Add(1)
						return dto.SearchResponse{}, nil
					}).AnyTimes()
			},
		},
		{
			tool:   "web_search_images",
			argKey: "queries",
			item:   func(i int) string { return fmt.Sprintf("image query %d", i) },
			expect: func(uc *Mockusecase, calls *atomic.Int32) {
				uc.EXPECT().SearchImages(gomock.Any(), gomock.Any()).
					DoAndReturn(func(context.Context, dto.ImagesSearchRequest) (dto.ImagesSearchResponse, error) {
						calls.Add(1)
						return dto.ImagesSearchResponse{}, nil
					}).AnyTimes()
			},
		},
		{
			tool:   "web_scrape",
			argKey: "urls",
			item:   func(i int) string { return fmt.Sprintf("https://example.com/page-%d", i) },
			expect: func(uc *Mockusecase, calls *atomic.Int32) {
				uc.EXPECT().ScrapePages(gomock.Any(), gomock.Any()).
					DoAndReturn(func(context.Context, dto.ScrapeRequest) (dto.ScrapeResponse, error) {
						calls.Add(1)
						return dto.ScrapeResponse{}, nil
					}).AnyTimes()
			},
		},
	}
}

// connectSchemaLimitClient registers the handler's tools on a real SDK server
// and returns a client session connected to it over in-memory transports.
func connectSchemaLimitClient(t *testing.T, uc *Mockusecase, maxItems int) *mcpsdk.ClientSession {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test-server", Version: "v0.0.0"}, nil)
	New(uc, maxItems).RegisterTools(server)

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}

func schemaLimitItems(n int, gen func(int) string) []string {
	items := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		items = append(items, gen(i))
	}

	return items
}

func TestRegisterToolsCapsListInSchema(t *testing.T) {
	// arrange
	uc := NewMockusecase(gomock.NewController(t))
	session := connectSchemaLimitClient(t, uc, schemaLimitMaxItems)

	// act
	list, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	tools := make(map[string]*mcpsdk.Tool, len(list.Tools))
	for _, tool := range list.Tools {
		tools[tool.Name] = tool
	}

	// assert
	type schemaShape struct {
		Properties map[string]struct {
			MaxItems *int `json:"maxItems"`
		} `json:"properties"`
	}

	for _, tc := range schemaLimitCases() {
		t.Run(tc.tool, func(t *testing.T) {
			tool, ok := tools[tc.tool]
			require.True(t, ok, "tool %q is not registered", tc.tool)

			raw, err := json.Marshal(tool.InputSchema)
			require.NoError(t, err)

			var schema schemaShape
			require.NoError(t, json.Unmarshal(raw, &schema))

			prop, ok := schema.Properties[tc.argKey]
			require.True(t, ok, "property %q missing from input schema of %q", tc.argKey, tc.tool)
			require.NotNil(t, prop.MaxItems, "property %q of %q has no maxItems", tc.argKey, tc.tool)
			assert.Equal(t, schemaLimitMaxItems, *prop.MaxItems)
		})
	}
}

func TestToolRejectsMoreThanMaxItems(t *testing.T) {
	for _, tc := range schemaLimitCases() {
		t.Run(tc.tool, func(t *testing.T) {
			// arrange
			uc := NewMockusecase(gomock.NewController(t))

			var calls atomic.Int32
			tc.expect(uc, &calls)

			session := connectSchemaLimitClient(t, uc, schemaLimitMaxItems)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// act
			res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
				Name: tc.tool,
				Arguments: map[string]any{
					tc.argKey: schemaLimitItems(schemaLimitMaxItems+1, tc.item),
				},
			})

			// assert
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.True(t, res.IsError)
			assert.Contains(t, resultText(t, res), "maxItems")
			assert.Zero(t, calls.Load(), "usecase must not be called when the list exceeds maxItems")
		})
	}
}

func TestToolAcceptsExactlyMaxItems(t *testing.T) {
	for _, tc := range schemaLimitCases() {
		t.Run(tc.tool, func(t *testing.T) {
			// arrange
			uc := NewMockusecase(gomock.NewController(t))

			var calls atomic.Int32
			tc.expect(uc, &calls)

			session := connectSchemaLimitClient(t, uc, schemaLimitMaxItems)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// act
			res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
				Name: tc.tool,
				Arguments: map[string]any{
					tc.argKey: schemaLimitItems(schemaLimitMaxItems, tc.item),
				},
			})

			// assert
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.False(t, res.IsError, "exactly maxItems items must not be rejected")
			assert.Equal(t, int32(1), calls.Load(), "usecase must be called exactly once")
		})
	}
}
