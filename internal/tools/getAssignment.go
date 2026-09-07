package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) GetAssignment(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input RequestIDInput,
) (*mcp.CallToolResult, domain.Assignment, error) {
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	assignment, err := t.logistics.GetAssignment(ctx, requestID)
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	return nil, *assignment, nil
}
