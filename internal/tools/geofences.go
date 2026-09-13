package tools

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) CreateGeofence(ctx context.Context, _ *mcp.CallToolRequest, input CreateGeofenceInput) (*mcp.CallToolResult, domain.UnifiedGeofence, error) {
	if strings.TrimSpace(input.GeofenceID) == "" || strings.TrimSpace(input.Name) == "" {
		return toolFailure[domain.UnifiedGeofence](errors.New("geofenceId and name are required"))
	}
	if input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180 {
		return toolFailure[domain.UnifiedGeofence](errors.New("latitude or longitude is outside the valid decimal range"))
	}
	item, err := t.fleet.CreateGeofence(ctx, domain.CreateGeofenceRequest{GeofenceID: strings.TrimSpace(input.GeofenceID), Name: strings.TrimSpace(input.Name), Group: strings.TrimSpace(input.Group), AdditionalMargin: input.AdditionalMargin, Latitude: toRemoteCoordinate(input.Latitude), Longitude: toRemoteCoordinate(input.Longitude)})
	if err != nil {
		return toolFailure[domain.UnifiedGeofence](err)
	}
	return nil, domain.UnifiedGeofence{Geofence: *item, Latitude: fromRemoteCoordinate(item.Latitude), Longitude: fromRemoteCoordinate(item.Longitude)}, nil
}

func (t *Tools) ListGeofences(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, domain.UnifiedGeofenceList, error) {
	items, err := t.fleet.ListGeofences(ctx, strings.TrimSpace(input.Search))
	if err != nil {
		return toolFailure[domain.UnifiedGeofenceList](err)
	}
	result := make([]domain.UnifiedGeofence, 0, len(items))
	for _, item := range items {
		result = append(result, domain.UnifiedGeofence{Geofence: item, Latitude: fromRemoteCoordinate(item.Latitude), Longitude: fromRemoteCoordinate(item.Longitude)})
	}
	return nil, domain.UnifiedGeofenceList{Count: len(result), Geofences: result}, nil
}

func fromRemoteCoordinate(value int64) float64 { return float64(value) / 10000000 }
