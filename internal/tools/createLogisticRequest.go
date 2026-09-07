package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) CreateLogisticsRequest(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input domain.CreateRequestInput,
) (*mcp.CallToolResult, domain.Request, error) {
	request, err := t.logistics.CreateRequest(ctx, input)
	if err != nil {
		return toolFailure[domain.Request](err)
	}

	return nil, *request, nil
}
