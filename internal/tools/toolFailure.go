package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func toolFailure[T any](
	err error,
) (*mcp.CallToolResult, T, error) {
	var zero T

	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: err.Error(),
			},
		},
	}, zero, nil
}
