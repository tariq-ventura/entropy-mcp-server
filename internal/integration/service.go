package integration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
	"github.com/tariq-ventura/entropy-mcp-server/internal/projection"
	"golang.org/x/sync/errgroup"
)

type LogisticsSource interface {
	ListMachinery(context.Context, string, string, string) ([]domain.Machinery, error)
	SearchRequests(context.Context, domain.SearchRequestsInput) (*domain.SearchRequestsResult, error)
}

type FleetSource interface {
	ListVehicles(context.Context, string, string, string) ([]domain.Vehicle, error)
	ListTasks(context.Context, string) ([]domain.Task, error)
	ListMaintenance(context.Context, string, string) ([]domain.Maintenance, error)
}

type ProjectionStore interface {
	BeginSync(context.Context, time.Time) error
	FailSync(context.Context, error) error
	ApplySnapshot(context.Context, projection.Snapshot) error
	ListAllEquipment(context.Context) ([]projection.UnifiedEquipment, error)
	ListEquipment(context.Context, int, int, string, string, string, bool, bool) ([]projection.UnifiedEquipment, int64, error)
	GetEquipment(context.Context, string) (*projection.EquipmentDetail, error)
	ListAllMaintenance(context.Context) ([]projection.UnifiedMaintenance, error)
	ListTypes(context.Context) ([]projection.EquipmentType, error)
	GetRequest(context.Context, string) (*projection.UnifiedRequest, error)
	ListRequests(context.Context, int, int, string, string, string, string) ([]projection.UnifiedRequest, int64, error)
	ListAssignments(context.Context, int, int, string, string) ([]projection.UnifiedAssignment, int64, error)
	GetAssignmentByRequest(context.Context, string) (*projection.UnifiedAssignment, error)
	ListConflicts(context.Context, int, int) ([]projection.SyncConflict, int64, error)
	GetSyncState(context.Context) (*projection.SyncState, error)
	Dashboard(context.Context) (*projection.Dashboard, error)
}

type Service struct {
	logistics LogisticsSource
	fleet     FleetSource
	repo      ProjectionStore
	now       func() time.Time
	mu        sync.Mutex
}

func New(logistics LogisticsSource, fleet FleetSource, repo ProjectionStore) *Service {
	return &Service{logistics: logistics, fleet: fleet, repo: repo, now: time.Now}
}

func (s *Service) Sync(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	startedAt := s.now().UTC()
	if err := s.repo.BeginSync(ctx, startedAt); err != nil {
		return fmt.Errorf("marking projection sync as running: %w", err)
	}

	var machinery []domain.Machinery
	var vehicles []domain.Vehicle
	var requests []domain.Request
	var tasks []domain.Task
	var maintenance []domain.Maintenance
	group, groupContext := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		machinery, err = s.logistics.ListMachinery(groupContext, "", "", "")
		if err != nil {
			return fmt.Errorf("loading Prisma machinery: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		var err error
		vehicles, err = s.fleet.ListVehicles(groupContext, "", "", "")
		if err != nil {
			return fmt.Errorf("loading Startrack vehicles: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		var err error
		requests, err = s.listAllRequests(groupContext)
		if err != nil {
			return fmt.Errorf("loading Prisma requests: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		var err error
		tasks, err = s.fleet.ListTasks(groupContext, "")
		if err != nil {
			return fmt.Errorf("loading Startrack tasks: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		var err error
		maintenance, err = s.fleet.ListMaintenance(groupContext, "", "")
		if err != nil {
			return fmt.Errorf("loading Startrack maintenance: %w", err)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		_ = s.repo.FailSync(ctx, err)
		return err
	}

	snapshot := buildSnapshot(machinery, vehicles, requests, tasks, maintenance, uuid.NewString(), s.now().UTC())
	if err := s.repo.ApplySnapshot(ctx, snapshot); err != nil {
		wrapped := fmt.Errorf("applying unified projection: %w", err)
		_ = s.repo.FailSync(ctx, wrapped)
		return wrapped
	}
	return nil
}

func (s *Service) listAllRequests(ctx context.Context) ([]domain.Request, error) {
	result := make([]domain.Request, 0)
	for page := 1; ; page++ {
		response, err := s.logistics.SearchRequests(ctx, domain.SearchRequestsInput{Page: page, PageSize: 100})
		if err != nil {
			return nil, err
		}
		result = append(result, response.Requests...)
		if len(response.Requests) == 0 || len(result) >= int(response.Count) || len(response.Requests) < 100 {
			return result, nil
		}
	}
}

func (s *Service) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Minute
	}
	if err := s.Sync(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("initial unified projection sync failed", "error", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Sync(ctx); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("unified projection sync failed", "error", err)
			}
		}
	}
}

func buildSnapshot(machinery []domain.Machinery, vehicles []domain.Vehicle, requests []domain.Request, tasks []domain.Task, maintenance []domain.Maintenance, token string, syncedAt time.Time) projection.Snapshot {
	unified := UnifyEquipment(machinery, vehicles)
	snapshot := projection.Snapshot{Token: token, SyncedAt: syncedAt}
	typeNames := make(map[string]string)
	aliases := make(map[string]projection.EquipmentTypeAlias)
	aliasOwners := make(map[string]string)
	equipmentByKey := make(map[string]domain.UnifiedEquipment)
	equipmentIndex := make(map[string]int)

	addAlias := func(source, value, canonicalCode, entityKey string) {
		value = strings.TrimSpace(value)
		if value == "" || canonicalCode == "" {
			return
		}
		key := strings.ToLower(source) + ":" + Normalize(value)
		if previous, exists := aliasOwners[key]; exists && previous != canonicalCode {
			snapshot.Conflicts = append(snapshot.Conflicts, newConflict("equipment_type", entityKey, "alias", previous, canonicalCode, "the same source type alias resolves to two canonical types", token, syncedAt))
			return
		}
		aliasOwners[key] = canonicalCode
		aliases[key] = projection.EquipmentTypeAlias{Key: key, Source: source, SourceValue: value, Normalized: Normalize(value), TypeCode: canonicalCode, SyncToken: token, SyncedAt: syncedAt}
	}

	for _, item := range unified {
		code := Normalize(item.Type)
		if code == "" {
			code = "UNKNOWN"
		}
		if _, exists := typeNames[code]; !exists {
			typeNames[code] = fallback(item.Type, "Unknown")
		}
		addAlias("prisma", item.Type, code, item.EquipmentKey)
		addAlias("startrack", item.StartrackType, code, item.EquipmentKey)
		if index, exists := equipmentIndex[item.EquipmentKey]; exists {
			message := "multiple source records resolve to the same canonical equipment key"
			snapshot.Equipment[index].Conflicts = append(snapshot.Equipment[index].Conflicts, message)
			snapshot.Conflicts = append(snapshot.Conflicts, newConflict("equipment", item.EquipmentKey, "correlation", snapshot.Equipment[index].PrismaID, item.PrismaID, message, token, syncedAt))
			continue
		}
		equipmentByKey[item.EquipmentKey] = item
		equipmentIndex[item.EquipmentKey] = len(snapshot.Equipment)
		snapshot.Equipment = append(snapshot.Equipment, projection.UnifiedEquipment{
			EquipmentKey: item.EquipmentKey, EquipmentTypeCode: code, Linked: item.Linked, Available: item.Available,
			Name: item.Name, AssetNumber: item.AssetNumber, Type: item.Type, StartrackType: item.StartrackType,
			PrismaStatus: item.PrismaStatus, TrackingStatus: item.TrackingStatus, Company: item.Company,
			Year: item.Year, Color: item.Color, Brand: item.Brand, Model: item.Model, GroupName: item.Group,
			Tags: item.Tags, Driver: item.Driver, RemoteID: item.RemoteID, StartrackDescription: item.StartrackDescription,
			PrismaID: item.PrismaID, StartrackID: item.StartrackID, Conflicts: item.Conflicts, SyncToken: token, SyncedAt: syncedAt,
		})
		for _, message := range item.Conflicts {
			field := "correlation"
			if strings.Contains(strings.ToLower(message), "type mismatch") {
				field = "type"
			}
			snapshot.Conflicts = append(snapshot.Conflicts, newConflict("equipment", item.EquipmentKey, field, item.Type, item.StartrackType, message, token, syncedAt))
		}
	}

	for code, name := range typeNames {
		snapshot.Types = append(snapshot.Types, projection.EquipmentType{Code: code, Name: name, Active: true, SyncToken: token, SyncedAt: syncedAt})
	}
	for _, alias := range aliases {
		snapshot.Aliases = append(snapshot.Aliases, alias)
	}

	requestByID := make(map[string]domain.Request)
	for _, item := range requests {
		requestByID[item.ID] = item
		var equipmentKey *string
		if item.Machinery != nil && strings.TrimSpace(*item.Machinery) != "" {
			value := CorrelationKey(*item.Machinery)
			equipmentKey = &value
		}
		snapshot.Requests = append(snapshot.Requests, projection.UnifiedRequest{
			ID: item.ID, Project: item.Project, Type: item.Type, EquipmentTypeCode: Normalize(item.Type), Requester: item.Requester,
			StartDate: item.StartDate, EndDate: item.EndDate, Status: item.Status, Machinery: item.Machinery,
			EquipmentKey: equipmentKey, SourceCreatedAt: item.CreatedAt, SourceUpdatedAt: item.UpdatedAt, SyncToken: token, SyncedAt: syncedAt,
		})
	}

	for _, item := range maintenance {
		snapshot.Maintenance = append(snapshot.Maintenance, projection.UnifiedMaintenance{
			ID: item.ID, EquipmentKey: CorrelationKey(item.Vehicle), Vehicle: item.Vehicle, Reference: item.Reference,
			ServiceDate: item.ServiceDate, Odometer: item.Odometer, ServiceTime: item.ServiceTime, HourMeter: item.HourMeter,
			RepairReason: item.RepairReason, Provider: item.Provider, Mechanic: item.Mechanic, ServiceType: item.ServiceType,
			SourceCreatedAt: item.CreatedAt, SourceUpdatedAt: item.UpdatedAt, SyncToken: token, SyncedAt: syncedAt,
		})
	}

	for _, task := range tasks {
		requestID := markerValue(task.Description, "Entropy request")
		if requestID == "" {
			continue
		}
		assetNumber := markerValue(task.Description, "Prisma machinery")
		equipmentKey := CorrelationKey(assetNumber)
		request, found := requestByID[requestID]
		status, inconsistent := assignmentStatus(task.Status, request, found, equipmentKey)
		assignment := projection.UnifiedAssignment{
			ID: task.ID, RequestID: requestID, EquipmentKey: equipmentKey, TaskExternalID: task.TaskID, Status: status,
			Title: task.Title, Description: task.Description, TaskType: task.Type, ScheduledDate: task.ScheduledDate, Origin: task.Origin,
			Destination: task.Destination, Latitude: task.Latitude, Longitude: task.Longitude, Assignee: task.Assignee,
			PrismaStatus: request.Status, StartrackStatus: task.Status, Inconsistent: inconsistent, SyncToken: token, SyncedAt: syncedAt,
		}
		snapshot.Assignments = append(snapshot.Assignments, assignment)
		if inconsistent {
			message := "Prisma request and Startrack task do not represent the same assignment state"
			if !found {
				message = "Startrack task references a Prisma request that does not exist"
			}
			if equipmentKey == "" {
				message = "Startrack task does not contain a Prisma machinery marker"
			}
			snapshot.Conflicts = append(snapshot.Conflicts, newConflict("assignment", requestID, "status", request.Status, task.Status, message, token, syncedAt))
		}
		if equipmentKey != "" {
			if _, exists := equipmentByKey[equipmentKey]; !exists {
				snapshot.Conflicts = append(snapshot.Conflicts, newConflict("assignment", requestID, "equipment", assetNumber, task.TaskID, "assignment references equipment absent from the unified projection", token, syncedAt))
			}
		}
	}
	return snapshot
}

func assignmentStatus(taskStatus string, request domain.Request, requestFound bool, equipmentKey string) (string, bool) {
	if !requestFound || equipmentKey == "" {
		return "INCONSISTENT", true
	}
	switch strings.ToLower(strings.TrimSpace(taskStatus)) {
	case "pendiente", "pending":
		if (strings.EqualFold(request.Status, "Aprobada") || strings.EqualFold(request.Status, "APPROVED")) && request.Machinery != nil {
			return "ACTIVE", false
		}
	case "completada", "completed":
		if strings.EqualFold(request.Status, "Aprobada") || strings.EqualFold(request.Status, "APPROVED") {
			return "COMPLETED", false
		}
	case "cancelada", "cancelled", "canceled":
		if (strings.EqualFold(request.Status, "Pendiente") || strings.EqualFold(request.Status, "PENDING")) && request.Machinery == nil {
			return "CANCELLED", false
		}
	}
	return "INCONSISTENT", true
}

func markerValue(description, marker string) string {
	for _, line := range strings.Split(strings.ReplaceAll(description, "\r\n", "\n"), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), marker) {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

func newConflict(entityType, entityKey, field, prismaValue, startrackValue, message, token string, detectedAt time.Time) projection.SyncConflict {
	return projection.SyncConflict{ID: uuid.NewString(), EntityType: entityType, EntityKey: entityKey, Field: field, PrismaValue: prismaValue, StartrackValue: startrackValue, Message: message, SyncToken: token, DetectedAt: detectedAt}
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return strings.TrimSpace(value)
}

func (s *Service) Refresh(ctx context.Context) error { return s.Sync(ctx) }

func (s *Service) ListAllEquipment(ctx context.Context) ([]domain.UnifiedEquipment, error) {
	items, err := s.repo.ListAllEquipment(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.UnifiedEquipment, 0, len(items))
	for _, item := range items {
		result = append(result, equipmentToDomain(item))
	}
	return result, nil
}

func (s *Service) ListAllMaintenance(ctx context.Context) ([]domain.Maintenance, error) {
	items, err := s.repo.ListAllMaintenance(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Maintenance, 0, len(items))
	for _, item := range items {
		result = append(result, domain.Maintenance{ID: item.ID, Vehicle: item.Vehicle, Reference: item.Reference, ServiceDate: item.ServiceDate, Odometer: item.Odometer, ServiceTime: item.ServiceTime, HourMeter: item.HourMeter, RepairReason: item.RepairReason, Provider: item.Provider, Mechanic: item.Mechanic, ServiceType: item.ServiceType, CreatedAt: item.SourceCreatedAt, UpdatedAt: item.SourceUpdatedAt})
	}
	return result, nil
}

func (s *Service) GetRequest(ctx context.Context, id string) (*domain.Request, error) {
	item, err := s.repo.GetRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	return requestToDomain(*item), nil
}

func (s *Service) SearchRequests(ctx context.Context, input domain.SearchRequestsInput) (*domain.SearchRequestsResult, error) {
	status := ""
	if len(input.Statuses) == 1 {
		status = input.Statuses[0]
	}
	items, count, err := s.repo.ListRequests(ctx, input.Page, input.PageSize, input.Query, status, input.Type, input.Requester)
	if err != nil {
		return nil, err
	}
	result := &domain.SearchRequestsResult{Count: count, Requests: make([]domain.Request, 0, len(items))}
	for _, item := range items {
		if len(input.Statuses) > 1 && !containsFold(input.Statuses, item.Status) {
			continue
		}
		result.Requests = append(result.Requests, *requestToDomain(item))
	}
	if len(input.Statuses) > 1 {
		result.Count = int64(len(result.Requests))
	}
	return result, nil
}

func (s *Service) GetAssignment(ctx context.Context, requestID string) (*domain.UnifiedAssignmentStatus, error) {
	assignment, err := s.repo.GetAssignmentByRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	request, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	equipmentDetail, err := s.repo.GetEquipment(ctx, assignment.EquipmentKey)
	if err != nil {
		return nil, err
	}
	return &domain.UnifiedAssignmentStatus{
		Request:   *requestToDomain(*request),
		Equipment: equipmentToDomain(equipmentDetail.Equipment),
		Task: domain.Task{
			ID: assignment.ID, TaskID: assignment.TaskExternalID, Title: assignment.Title, Description: assignment.Description,
			Type: assignment.TaskType, Status: assignment.StartrackStatus, ScheduledDate: assignment.ScheduledDate,
			Origin: assignment.Origin, Destination: assignment.Destination, Latitude: assignment.Latitude,
			Longitude: assignment.Longitude, Assignee: assignment.Assignee,
		},
	}, nil
}

func containsFold(items []string, value string) bool {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}

func equipmentToDomain(item projection.UnifiedEquipment) domain.UnifiedEquipment {
	return domain.UnifiedEquipment{EquipmentKey: item.EquipmentKey, Linked: item.Linked, Available: item.Available, Name: item.Name, AssetNumber: item.AssetNumber, Type: item.Type, StartrackType: item.StartrackType, PrismaStatus: item.PrismaStatus, TrackingStatus: item.TrackingStatus, Company: item.Company, Year: item.Year, Color: item.Color, Brand: item.Brand, Model: item.Model, Group: item.GroupName, Tags: item.Tags, Driver: item.Driver, RemoteID: item.RemoteID, StartrackDescription: item.StartrackDescription, PrismaID: item.PrismaID, StartrackID: item.StartrackID, Conflicts: item.Conflicts}
}

func requestToDomain(item projection.UnifiedRequest) *domain.Request {
	return &domain.Request{ID: item.ID, Project: item.Project, Type: item.Type, Requester: item.Requester, StartDate: item.StartDate, EndDate: item.EndDate, Status: item.Status, Machinery: item.Machinery, CreatedAt: item.SourceCreatedAt, UpdatedAt: item.SourceUpdatedAt}
}

func (s *Service) ListEquipment(ctx context.Context, page, pageSize int, equipmentType, status, search string, onlyAvailable, onlyLinked bool) ([]projection.UnifiedEquipment, int64, error) {
	return s.repo.ListEquipment(ctx, page, pageSize, equipmentType, status, search, onlyAvailable, onlyLinked)
}
func (s *Service) GetEquipment(ctx context.Context, key string) (*projection.EquipmentDetail, error) {
	return s.repo.GetEquipment(ctx, key)
}
func (s *Service) ListTypes(ctx context.Context) ([]projection.EquipmentType, error) {
	return s.repo.ListTypes(ctx)
}
func (s *Service) ListRequests(ctx context.Context, page, pageSize int, query, status, equipmentType, requester string) ([]projection.UnifiedRequest, int64, error) {
	return s.repo.ListRequests(ctx, page, pageSize, query, status, equipmentType, requester)
}
func (s *Service) ListAssignments(ctx context.Context, page, pageSize int, status, search string) ([]projection.UnifiedAssignment, int64, error) {
	return s.repo.ListAssignments(ctx, page, pageSize, status, search)
}
func (s *Service) ListConflicts(ctx context.Context, page, pageSize int) ([]projection.SyncConflict, int64, error) {
	return s.repo.ListConflicts(ctx, page, pageSize)
}
func (s *Service) GetSyncState(ctx context.Context) (*projection.SyncState, error) {
	return s.repo.GetSyncState(ctx)
}
func (s *Service) Dashboard(ctx context.Context) (*projection.Dashboard, error) {
	return s.repo.Dashboard(ctx)
}
