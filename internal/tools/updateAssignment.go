package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) updateAssignmentStatus(
	ctx context.Context,
	input UpdateAssignmentInput,
	status string,
) (*mcp.CallToolResult, domain.Assignment, error) {
	if !input.Confirmed {
		return toolFailure[domain.Assignment](
			fmt.Errorf(
				"%s requires explicit human confirmation",
				strings.ToLower(status),
			),
		)
	}

	assignmentID, err := parseUUID(
		input.AssignmentID,
		"assignmentId",
	)
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	if strings.TrimSpace(input.Reason) == "" {
		return toolFailure[domain.Assignment](
			errors.New("reason is required"),
		)
	}

	assignment, err := t.logistics.UpdateAssignmentStatus(
		ctx,
		assignmentID,
		domain.UpdateAssignmentStatusInput{
			Status: status,
			Reason: strings.TrimSpace(input.Reason),
		},
	)
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	return nil, *assignment, nil
}
