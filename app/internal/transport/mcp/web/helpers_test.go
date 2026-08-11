package web

import (
	"testing"

	"github.com/stretchr/testify/require"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func resultText(t *testing.T, res *mcpsdk.CallToolResult) string {
	t.Helper()

	require.NotNil(t, res)
	require.Len(t, res.Content, 1)

	content, ok := res.Content[0].(*mcpsdk.TextContent)
	require.True(t, ok, "content is not a text block")

	return content.Text
}
