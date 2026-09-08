package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) SearchLogisticsRequests(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input domain.SearchRequestsInput,
) (*mcp.CallToolResult, domain.SearchRequestsResult, error) {
	input.Query = strings.TrimSpace(input.Query)
	input.EquipmentType = strings.ToUpper(
		strings.TrimSpace(input.EquipmentType),
	)

	if input.Page == 0 {
		input.Page = 1
	}

	if input.PageSize == 0 {
		input.PageSize = 20
	}

	result, err := t.logistics.SearchRequests(ctx, input)

	if err != nil {
		return toolFailure[domain.SearchRequestsResult](err)
	}

	return nil, *result, nil
}
