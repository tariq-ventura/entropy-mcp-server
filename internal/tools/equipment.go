package tools

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) ListUnifiedEquipment(ctx context.Context, _ *mcp.CallToolRequest, input ListUnifiedEquipmentInput) (*mcp.CallToolResult, domain.UnifiedEquipmentList, error) {
	items, err := t.loadUnifiedEquipment(ctx)
	if err != nil {
		return toolFailure[domain.UnifiedEquipmentList](err)
	}
	filtered := make([]domain.UnifiedEquipment, 0, len(items))
	for _, item := range items {
		if input.OnlyAvailable && !item.Available {
			continue
		}
		if input.OnlyLinked && !item.Linked {
			continue
		}
		if input.Type != "" && !sameText(input.Type, item.Type) {
			continue
		}
		if input.Search != "" {
			haystack := strings.Join([]string{item.EquipmentKey, item.Name, item.AssetNumber, item.Type, item.Brand, item.Model, item.RemoteID}, " ")
			if !strings.Contains(strings.ToLower(haystack), strings.ToLower(strings.TrimSpace(input.Search))) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return nil, domain.UnifiedEquipmentList{Count: len(filtered), Equipment: filtered}, nil
}

func (t *Tools) GetUnifiedEquipment(ctx context.Context, _ *mcp.CallToolRequest, input UnifiedEquipmentInput) (*mcp.CallToolResult, domain.UnifiedEquipment, error) {
	if strings.TrimSpace(input.EquipmentKey) == "" {
		return toolFailure[domain.UnifiedEquipment](errors.New("equipmentKey is required"))
	}
	items, err := t.loadUnifiedEquipment(ctx)
	if err != nil {
		return toolFailure[domain.UnifiedEquipment](err)
	}
	item, found := findUnified(items, input.EquipmentKey)
	if !found {
		return toolFailure[domain.UnifiedEquipment](errors.New("unified equipment not found"))
	}
	return nil, item, nil
}
