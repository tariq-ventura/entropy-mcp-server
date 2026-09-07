package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) CancelAssignment(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input UpdateAssignmentInput,
) (*mcp.CallToolResult, domain.Assignment, error) {
	return t.updateAssignmentStatus(
		ctx,
		input,
		"CANCELLED",
	)
}
