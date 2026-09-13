package tools

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) CreateLogisticsRequest(ctx context.Context, _ *mcp.CallToolRequest, input domain.CreateRequestInput) (*mcp.CallToolResult, domain.Request, error) {
	if strings.TrimSpace(input.Project) == "" || strings.TrimSpace(input.Type) == "" || strings.TrimSpace(input.Requester) == "" {
		return toolFailure[domain.Request](errors.New("project, type and requester are required"))
	}
	start, err := parseDate(input.StartDate)
	if err != nil {
		return toolFailure[domain.Request](errors.New("startDate must be YYYY-MM-DD or RFC3339"))
	}
	end, err := parseDate(input.EndDate)
	if err != nil {
		return toolFailure[domain.Request](errors.New("endDate must be YYYY-MM-DD or RFC3339"))
	}
	if !end.After(start) {
		return toolFailure[domain.Request](errors.New("endDate must be after startDate"))
	}
	input.Project = strings.TrimSpace(input.Project)
	input.Type = strings.TrimSpace(input.Type)
	input.Requester = strings.TrimSpace(input.Requester)
	input.StartDate = start.Format(time.RFC3339)
	input.EndDate = end.Format(time.RFC3339)
	request, err := t.logistics.CreateRequest(ctx, input)
	if err != nil {
		return toolFailure[domain.Request](err)
	}
	t.refreshProjectionBestEffort(ctx)
	return nil, *request, nil
}

func (t *Tools) GetLogisticsRequest(ctx context.Context, _ *mcp.CallToolRequest, input RequestIDInput) (*mcp.CallToolResult, domain.Request, error) {
	id, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.Request](err)
	}
	var request *domain.Request
	if t.projection != nil {
		if err := t.projection.Refresh(ctx); err != nil {
			return toolFailure[domain.Request](err)
		}
		request, err = t.projection.GetRequest(ctx, id.String())
	} else {
		request, err = t.logistics.GetRequest(ctx, id)
	}
	if err != nil {
		return toolFailure[domain.Request](err)
	}
	return nil, *request, nil
}

func (t *Tools) SearchLogisticsRequests(ctx context.Context, _ *mcp.CallToolRequest, input domain.SearchRequestsInput) (*mcp.CallToolResult, domain.SearchRequestsResult, error) {
	input.Query = strings.TrimSpace(input.Query)
	input.Type = strings.TrimSpace(input.Type)
	input.Requester = strings.TrimSpace(input.Requester)
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	var result *domain.SearchRequestsResult
	var err error
	if t.projection != nil {
		if err = t.projection.Refresh(ctx); err == nil {
			result, err = t.projection.SearchRequests(ctx, input)
		}
	} else {
		result, err = t.logistics.SearchRequests(ctx, input)
	}
	if err != nil {
		return toolFailure[domain.SearchRequestsResult](err)
	}
	return nil, *result, nil
}

func (t *Tools) refreshProjectionBestEffort(ctx context.Context) {
	if t.projection == nil {
		return
	}
	if err := t.projection.Refresh(ctx); err != nil {
		slog.Warn("source write succeeded but unified projection refresh failed", "error", err)
	}
}

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	return time.Parse("2006-01-02", value)
}
