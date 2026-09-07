package tools

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) CreateAssignment(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input CreateAssignmentInput,
) (*mcp.CallToolResult, domain.Assignment, error) {
	if !input.Confirmed {
		return toolFailure[domain.Assignment](
			errors.New(
				"create_assignment requires explicit human confirmation",
			),
		)
	}

	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	equipmentID, err := parseUUID(
		input.EquipmentID,
		"equipmentId",
	)
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	if strings.TrimSpace(input.Reason) == "" {
		return toolFailure[domain.Assignment](
			errors.New("reason is required"),
		)
	}

	assignment, err := t.logistics.CreateAssignment(
		ctx,
		requestID,
		domain.CreateAssignmentInput{
			EquipmentID: equipmentID.String(),
			Reason:      strings.TrimSpace(input.Reason),
		},
	)
	if err != nil {
		return toolFailure[domain.Assignment](err)
	}

	return nil, *assignment, nil
}
