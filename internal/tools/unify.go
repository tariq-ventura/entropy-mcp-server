package tools

import (
	"context"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
	"github.com/tariq-ventura/entropy-mcp-server/internal/integration"
)

func correlationKey(value string) string { return integration.CorrelationKey(value) }
func normalize(value string) string      { return integration.Normalize(value) }
func sameText(left, right string) bool   { return integration.SameText(left, right) }

func UnifyEquipment(machinery []domain.Machinery, vehicles []domain.Vehicle) []domain.UnifiedEquipment {
	return integration.UnifyEquipment(machinery, vehicles)
}

func (t *Tools) loadUnifiedEquipment(ctx context.Context) ([]domain.UnifiedEquipment, error) {
	if t.projection != nil {
		if err := t.projection.Refresh(ctx); err != nil {
			return nil, err
		}
		return t.projection.ListAllEquipment(ctx)
	}
	machinery, err := t.logistics.ListMachinery(ctx, "", "", "")
	if err != nil {
		return nil, err
	}
	vehicles, err := t.fleet.ListVehicles(ctx, "", "", "")
	if err != nil {
		return nil, err
	}
	return UnifyEquipment(machinery, vehicles), nil
}
func findUnified(items []domain.UnifiedEquipment, value string) (domain.UnifiedEquipment, bool) {
	return integration.FindEquipment(items, value)
}
