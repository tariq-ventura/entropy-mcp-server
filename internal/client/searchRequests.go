package client

import (
	"context"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

const httpMethodQuery = "QUERY"

func (c *Client) SearchRequests(
	ctx context.Context,
	input domain.SearchRequestsInput,
) (*domain.SearchRequestsResult, error) {
	var response domain.APIResponse[domain.SearchRequestsResult]

	if err := c.do(
		ctx,
		httpMethodQuery,
		"/api/v1/requests",
		input,
		&response,
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
