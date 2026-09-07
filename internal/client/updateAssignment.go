package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) UpdateAssignmentStatus(
	ctx context.Context,
	assignmentID uuid.UUID,
	input domain.UpdateAssignmentStatusInput,
) (*domain.Assignment, error) {
	var response domain.APIResponse[domain.Assignment]

	path := fmt.Sprintf(
		"/api/v1/assignments/%s/status",
		url.PathEscape(assignmentID.String()),
	)

	if err := c.do(
		ctx,
		http.MethodPatch,
		path,
		input,
		&response,
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
