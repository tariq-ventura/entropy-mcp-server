package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) GetRecommendations(
	ctx context.Context,
	requestID uuid.UUID,
) (*domain.Recommendations, error) {
	var response domain.APIResponse[domain.Recommendations]

	path := fmt.Sprintf(
		"/api/v1/requests/%s/recommendations",
		url.PathEscape(requestID.String()),
	)

	if err := c.do(
		ctx,
		http.MethodGet,
		path,
		nil,
		&response,
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
