package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) GetLogisticsRequest(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input RequestIDInput,
) (*mcp.CallToolResult, domain.Request, error) {
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.Request](err)
	}

	request, err := t.logistics.GetRequest(ctx, requestID)
	if err != nil {
		return toolFailure[domain.Request](err)
	}

	return nil, *request, nil
}
