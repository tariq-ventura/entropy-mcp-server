package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) CreateAssignment(
	ctx context.Context,
	requestID uuid.UUID,
	input domain.CreateAssignmentInput,
) (*domain.Assignment, error) {
	var response domain.APIResponse[domain.Assignment]

	path := fmt.Sprintf(
		"/api/v1/requests/%s/assignment",
		url.PathEscape(requestID.String()),
	)

	if err := c.do(
		ctx,
		http.MethodPost,
		path,
		input,
		&response,
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
