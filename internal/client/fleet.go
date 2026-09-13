package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) ListVehicles(ctx context.Context, vehicleType, status, search string) ([]domain.Vehicle, error) {
	query := url.Values{"pageSize": {"100"}}
	if vehicleType != "" {
		query.Set("type", vehicleType)
	}
	if status != "" {
		query.Set("status", status)
	}
	if search != "" {
		query.Set("search", search)
	}
	items := make([]domain.Vehicle, 0)
	for page := int64(1); ; page++ {
		query.Set("page", fmt.Sprint(page))
		var response domain.ListResponse[domain.Vehicle]
		if err := c.do(ctx, http.MethodGet, "/api/v1/vehicles?"+query.Encode(), nil, &response); err != nil {
			return nil, err
		}
		items = append(items, response.Data...)
		if response.Pagination.TotalPages <= page {
			break
		}
	}
	return items, nil
}
func (c *Client) GetVehicle(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	var response domain.APIResponse[domain.Vehicle]
	if err := c.do(ctx, http.MethodGet, "/api/v1/vehicles/"+url.PathEscape(id.String()), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) CreateTask(ctx context.Context, input domain.CreateTaskRequest) (*domain.Task, error) {
	var response domain.APIResponse[domain.Task]
	if err := c.do(ctx, http.MethodPost, "/api/v1/tasks", input, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) ListTasks(ctx context.Context, search string) ([]domain.Task, error) {
	query := url.Values{"pageSize": {"100"}}
	if search != "" {
		query.Set("search", search)
	}
	items := make([]domain.Task, 0)
	for page := int64(1); ; page++ {
		query.Set("page", fmt.Sprint(page))
		var response domain.ListResponse[domain.Task]
		if err := c.do(ctx, http.MethodGet, "/api/v1/tasks?"+query.Encode(), nil, &response); err != nil {
			return nil, err
		}
		items = append(items, response.Data...)
		if response.Pagination.TotalPages <= page {
			break
		}
	}
	return items, nil
}
func (c *Client) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	var response domain.APIResponse[domain.Task]
	if err := c.do(ctx, http.MethodGet, "/api/v1/tasks/"+url.PathEscape(id.String()), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Task, error) {
	var response domain.APIResponse[domain.Task]
	path := fmt.Sprintf("/api/v1/tasks/%s", url.PathEscape(id.String()))
	if err := c.do(ctx, http.MethodPatch, path, map[string]string{"status": status}, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/tasks/"+url.PathEscape(id.String()), nil, nil)
}
func (c *Client) CreateGeofence(ctx context.Context, input domain.CreateGeofenceRequest) (*domain.Geofence, error) {
	var response domain.APIResponse[domain.Geofence]
	if err := c.do(ctx, http.MethodPost, "/api/v1/geofences", input, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) ListGeofences(ctx context.Context, search string) ([]domain.Geofence, error) {
	query := url.Values{"pageSize": {"100"}}
	if search != "" {
		query.Set("search", search)
	}
	items := make([]domain.Geofence, 0)
	for page := int64(1); ; page++ {
		query.Set("page", fmt.Sprint(page))
		var response domain.ListResponse[domain.Geofence]
		if err := c.do(ctx, http.MethodGet, "/api/v1/geofences?"+query.Encode(), nil, &response); err != nil {
			return nil, err
		}
		items = append(items, response.Data...)
		if response.Pagination.TotalPages <= page {
			break
		}
	}
	return items, nil
}

func (c *Client) ListMaintenance(ctx context.Context, vehicle, search string) ([]domain.Maintenance, error) {
	query := url.Values{"pageSize": {"100"}}
	if vehicle != "" {
		query.Set("vehicle", vehicle)
	}
	if search != "" {
		query.Set("search", search)
	}
	items := make([]domain.Maintenance, 0)
	for page := int64(1); ; page++ {
		query.Set("page", fmt.Sprint(page))
		var response domain.ListResponse[domain.Maintenance]
		if err := c.do(ctx, http.MethodGet, "/api/v1/maintenance?"+query.Encode(), nil, &response); err != nil {
			return nil, err
		}
		items = append(items, response.Data...)
		if response.Pagination.TotalPages <= page {
			break
		}
	}
	return items, nil
}
