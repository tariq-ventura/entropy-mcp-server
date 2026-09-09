package client

import (
	"context"
	"net/http"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) SearchRequests(
	ctx context.Context,
	input domain.SearchRequestsInput,
) (*domain.SearchRequestsResult, error) {
	var response domain.APIResponse[domain.SearchRequestsResult]

	if err := c.do(
		ctx,
		http.MethodPost,
		"/api/v1/requests/search",
		input,
		&response,
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
