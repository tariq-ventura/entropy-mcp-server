package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (c *Client) CreateRequest(ctx context.Context, input domain.CreateRequestInput) (*domain.Request, error) {
	var response domain.APIResponse[domain.Request]
	if err := c.do(ctx, http.MethodPost, "/api/v1/requests", input, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) GetRequest(ctx context.Context, id uuid.UUID) (*domain.Request, error) {
	var response domain.APIResponse[domain.Request]
	path := "/api/v1/requests/" + url.PathEscape(id.String())
	if err := c.do(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) SearchRequests(ctx context.Context, input domain.SearchRequestsInput) (*domain.SearchRequestsResult, error) {
	var response domain.APIResponse[domain.SearchRequestsResult]
	if err := c.do(ctx, http.MethodPost, "/api/v1/requests/search", input, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) ListMachinery(ctx context.Context, equipmentClass, status, search string) ([]domain.Machinery, error) {
	query := url.Values{"pageSize": {"100"}}
	if equipmentClass != "" {
		query.Set("equipmentClass", equipmentClass)
	}
	if status != "" {
		query.Set("status", status)
	}
	if search != "" {
		query.Set("search", search)
	}
	items := make([]domain.Machinery, 0)
	for page := int64(1); ; page++ {
		query.Set("page", fmt.Sprint(page))
		var response domain.ListResponse[domain.Machinery]
		if err := c.do(ctx, http.MethodGet, "/api/v1/equipments?"+query.Encode(), nil, &response); err != nil {
			return nil, err
		}
		items = append(items, response.Data...)
		if response.Pagination.TotalPages <= page {
			break
		}
	}
	return items, nil
}
func (c *Client) AssignMachinery(ctx context.Context, requestID, equipmentID uuid.UUID, reason string) (*domain.PrismaAssignment, error) {
	var response domain.APIResponse[domain.PrismaAssignment]
	path := fmt.Sprintf("/api/v1/requests/%s/assignment", url.PathEscape(requestID.String()))
	input := map[string]string{"equipmentId": equipmentID.String(), "reason": reason}
	if err := c.do(ctx, http.MethodPatch, path, input, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) ReleaseMachinery(ctx context.Context, requestID uuid.UUID, reason string) (*domain.Request, error) {
	var response domain.APIResponse[domain.Request]
	path := fmt.Sprintf("/api/v1/requests/%s/assignment?reason=%s", url.PathEscape(requestID.String()), url.QueryEscape(reason))
	if err := c.do(ctx, http.MethodDelete, path, nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
func (c *Client) UpdateMachineryStatus(ctx context.Context, equipmentID uuid.UUID, status string) (*domain.Machinery, error) {
	var response domain.APIResponse[domain.Machinery]
	path := fmt.Sprintf("/api/v1/equipments/%s/status", url.PathEscape(equipmentID.String()))
	if err := c.do(ctx, http.MethodPatch, path, map[string]string{"status": status}, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
