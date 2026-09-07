package tools

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func parseUUID(value string, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"%s must be a valid UUID",
			field,
		)
	}

	return id, nil
}
