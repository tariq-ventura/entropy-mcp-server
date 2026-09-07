package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) GetRecommendations(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input RequestIDInput,
) (*mcp.CallToolResult, domain.Recommendations, error) {
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.Recommendations](err)
	}

	recommendations, err :=
		t.logistics.GetRecommendations(ctx, requestID)

	if err != nil {
		return toolFailure[domain.Recommendations](err)
	}

	return nil, *recommendations, nil
}
