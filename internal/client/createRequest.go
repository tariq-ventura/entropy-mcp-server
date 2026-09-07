package client

import (
	"context"
	"net/http"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) CreateRequest(
	ctx context.Context,
	input domain.CreateRequestInput,
) (*domain.Request, error) {
	var response domain.APIResponse[domain.Request]

	if err := c.do(
		ctx,
		http.MethodPost,
		"/api/v1/requests",
		input,
		&response,
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
