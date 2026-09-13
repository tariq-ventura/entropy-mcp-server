package tools

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func (t *Tools) CreateAssignment(ctx context.Context, _ *mcp.CallToolRequest, input CreateAssignmentInput) (*mcp.CallToolResult, domain.UnifiedAssignment, error) {
	if !input.Confirmed {
		return toolFailure[domain.UnifiedAssignment](errors.New("create_assignment requires explicit human confirmation"))
	}
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.UnifiedAssignment](err)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return toolFailure[domain.UnifiedAssignment](errors.New("reason is required"))
	}
	if input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180 {
		return toolFailure[domain.UnifiedAssignment](errors.New("latitude or longitude is outside the valid decimal range"))
	}
	scheduledDate, err := parseDate(input.ScheduledDate)
	if err != nil {
		return toolFailure[domain.UnifiedAssignment](errors.New("scheduledDate must be YYYY-MM-DD or RFC3339"))
	}
	items, err := t.loadUnifiedEquipment(ctx)
	if err != nil {
		return toolFailure[domain.UnifiedAssignment](err)
	}
	var request *domain.Request
	if t.projection != nil {
		request, err = t.projection.GetRequest(ctx, requestID.String())
	} else {
		request, err = t.logistics.GetRequest(ctx, requestID)
	}
	if err != nil {
		return toolFailure[domain.UnifiedAssignment](err)
	}
	if !strings.EqualFold(request.Status, "Pendiente") && !strings.EqualFold(request.Status, "PENDING") {
		return toolFailure[domain.UnifiedAssignment](errors.New("the Prisma request is not pending"))
	}
	equipment, found := findUnified(items, input.EquipmentKey)
	if !found {
		return toolFailure[domain.UnifiedAssignment](errors.New("unified equipment not found"))
	}
	if !equipment.Linked {
		return toolFailure[domain.UnifiedAssignment](errors.New("equipment must be linked between Prisma and Startrack"))
	}
	if !equipment.Available {
		return toolFailure[domain.UnifiedAssignment](errors.New("equipment is not Disponible in Prisma"))
	}
	if !sameText(equipment.Type, request.Type) {
		return toolFailure[domain.UnifiedAssignment](errors.New("equipment type does not match the Prisma request"))
	}
	assignee := strings.TrimSpace(input.Assignee)
	if assignee == "" {
		assignee = strings.TrimSpace(equipment.Driver)
	}
	if assignee == "" {
		return toolFailure[domain.UnifiedAssignment](errors.New("assignee is required because the Startrack vehicle has no driver"))
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return toolFailure[domain.UnifiedAssignment](errors.New("title is required"))
	}
	description := strings.TrimSpace(input.Description)
	if len(description) > 1700 {
		description = description[:1700]
	}
	description = strings.TrimSpace(description + "\nEntropy request: " + request.ID + "\nPrisma machinery: " + equipment.AssetNumber)
	task, err := t.fleet.CreateTask(ctx, domain.CreateTaskRequest{TaskID: strings.TrimSpace(input.TaskID), Title: title, Description: description, Type: strings.TrimSpace(input.TaskType), Status: "Pendiente", ScheduledDate: scheduledDate.Format("2006-01-02T15:04:05Z07:00"), Origin: strings.TrimSpace(input.Origin), Destination: strings.TrimSpace(input.Destination), Latitude: toRemoteCoordinate(input.Latitude), Longitude: toRemoteCoordinate(input.Longitude), Assignee: assignee})
	if err != nil {
		return toolFailure[domain.UnifiedAssignment](err)
	}
	prismaEquipmentID, err := parseUUID(equipment.PrismaID, "prismaId")
	if err != nil {
		taskID, taskIDError := parseUUID(task.ID, "taskId")
		if taskIDError == nil {
			_ = t.fleet.DeleteTask(ctx, taskID)
		}
		return toolFailure[domain.UnifiedAssignment](err)
	}
	assignment, err := t.logistics.AssignMachinery(ctx, requestID, prismaEquipmentID, strings.TrimSpace(input.Reason))
	if err != nil {
		taskID, parseError := parseUUID(task.ID, "taskId")
		if parseError != nil {
			return toolFailure[domain.UnifiedAssignment](fmt.Errorf("Prisma assignment failed and task compensation could not identify the task: %w", err))
		}
		if compensationError := t.fleet.DeleteTask(ctx, taskID); compensationError != nil {
			return toolFailure[domain.UnifiedAssignment](fmt.Errorf("Prisma assignment failed: %v; Startrack compensation failed: %w", err, compensationError))
		}
		return toolFailure[domain.UnifiedAssignment](err)
	}
	equipment.PrismaStatus = assignment.Equipment.Status
	equipment.Available = false
	t.refreshProjectionBestEffort(ctx)
	return nil, domain.UnifiedAssignment{Request: assignment.Request, Equipment: equipment, Task: *task, Compensated: false}, nil
}

func (t *Tools) GetAssignment(ctx context.Context, _ *mcp.CallToolRequest, input RequestIDInput) (*mcp.CallToolResult, domain.UnifiedAssignmentStatus, error) {
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	if t.projection != nil {
		if err := t.projection.Refresh(ctx); err != nil {
			return toolFailure[domain.UnifiedAssignmentStatus](err)
		}
		assignment, err := t.projection.GetAssignment(ctx, requestID.String())
		if err != nil {
			return toolFailure[domain.UnifiedAssignmentStatus](err)
		}
		return nil, *assignment, nil
	}
	items, err := t.loadUnifiedEquipment(ctx)
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	var request *domain.Request
	if t.projection != nil {
		request, err = t.projection.GetRequest(ctx, requestID.String())
	} else {
		request, err = t.logistics.GetRequest(ctx, requestID)
	}
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	if request.Machinery == nil {
		return toolFailure[domain.UnifiedAssignmentStatus](errors.New("the request has no assigned machinery"))
	}
	equipment, found := findUnified(items, *request.Machinery)
	if !found {
		return toolFailure[domain.UnifiedAssignmentStatus](errors.New("assigned machinery could not be unified"))
	}
	tasks, err := t.fleet.ListTasks(ctx, request.ID)
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	if len(tasks) == 0 {
		return toolFailure[domain.UnifiedAssignmentStatus](errors.New("no Startrack task is linked to the Prisma request"))
	}
	return nil, domain.UnifiedAssignmentStatus{Request: *request, Equipment: equipment, Task: tasks[0]}, nil
}

func (t *Tools) CompleteAssignment(ctx context.Context, _ *mcp.CallToolRequest, input UpdateAssignmentInput) (*mcp.CallToolResult, domain.UnifiedAssignmentStatus, error) {
	return t.updateAssignment(ctx, input, "Completada", false)
}
func (t *Tools) CancelAssignment(ctx context.Context, _ *mcp.CallToolRequest, input UpdateAssignmentInput) (*mcp.CallToolResult, domain.UnifiedAssignmentStatus, error) {
	return t.updateAssignment(ctx, input, "Cancelada", true)
}

func (t *Tools) updateAssignment(ctx context.Context, input UpdateAssignmentInput, taskStatus string, resetRequest bool) (*mcp.CallToolResult, domain.UnifiedAssignmentStatus, error) {
	if !input.Confirmed {
		return toolFailure[domain.UnifiedAssignmentStatus](errors.New(strings.ToLower(taskStatus) + " requires explicit human confirmation"))
	}
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	taskID, err := parseUUID(input.TaskID, "taskId")
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	items, err := t.loadUnifiedEquipment(ctx)
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	var request *domain.Request
	if t.projection != nil {
		request, err = t.projection.GetRequest(ctx, requestID.String())
	} else {
		request, err = t.logistics.GetRequest(ctx, requestID)
	}
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	if request.Machinery == nil {
		return toolFailure[domain.UnifiedAssignmentStatus](errors.New("the request has no assigned machinery"))
	}
	equipment, found := findUnified(items, *request.Machinery)
	if !found {
		return toolFailure[domain.UnifiedAssignmentStatus](errors.New("assigned machinery could not be unified"))
	}
	task, err := t.fleet.UpdateTaskStatus(ctx, taskID, taskStatus)
	if err != nil {
		return toolFailure[domain.UnifiedAssignmentStatus](err)
	}
	if resetRequest {
		request, err = t.logistics.ReleaseMachinery(ctx, requestID, strings.TrimSpace(input.Reason))
	} else {
		prismaID, parseError := parseUUID(equipment.PrismaID, "prismaId")
		if parseError != nil {
			err = parseError
		} else {
			_, err = t.logistics.UpdateMachineryStatus(ctx, prismaID, "Disponible")
		}
	}
	if err != nil {
		if _, rollbackError := t.fleet.UpdateTaskStatus(ctx, taskID, "Pendiente"); rollbackError != nil {
			return toolFailure[domain.UnifiedAssignmentStatus](fmt.Errorf("updating Prisma failed: %v; Startrack rollback failed: %w", err, rollbackError))
		}
		return toolFailure[domain.UnifiedAssignmentStatus](fmt.Errorf("updating Prisma failed after changing Startrack; Startrack was returned to Pendiente: %w", err))
	}
	t.refreshProjectionBestEffort(ctx)
	refreshed, loadErr := t.loadUnifiedEquipment(ctx)
	if loadErr == nil {
		if item, ok := findUnified(refreshed, equipment.EquipmentKey); ok {
			equipment = item
		}
	}
	return nil, domain.UnifiedAssignmentStatus{Request: *request, Equipment: equipment, Task: *task}, nil
}

func toRemoteCoordinate(value float64) int64 { return int64(math.Round(value * 10000000)) }
