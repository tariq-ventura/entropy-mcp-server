package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

type fakeLogistics struct {
	request      domain.Request
	machinery    []domain.Machinery
	assignError  error
	assignCalled bool
}

func (f *fakeLogistics) CreateRequest(context.Context, domain.CreateRequestInput) (*domain.Request, error) {
	return &f.request, nil
}
func (f *fakeLogistics) GetRequest(context.Context, uuid.UUID) (*domain.Request, error) {
	return &f.request, nil
}
func (f *fakeLogistics) SearchRequests(context.Context, domain.SearchRequestsInput) (*domain.SearchRequestsResult, error) {
	return &domain.SearchRequestsResult{}, nil
}
func (f *fakeLogistics) ListMachinery(context.Context, string, string, string) ([]domain.Machinery, error) {
	return f.machinery, nil
}
func (f *fakeLogistics) AssignMachinery(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) (*domain.PrismaAssignment, error) {
	f.assignCalled = true
	if f.assignError != nil {
		return nil, f.assignError
	}
	request := f.request
	request.Status = "Aprobada"
	asset := f.machinery[0].AssetNumber
	request.Machinery = &asset
	equipment := f.machinery[0]
	equipment.Status = "Ocupada"
	return &domain.PrismaAssignment{Request: request, Equipment: equipment}, nil
}
func (f *fakeLogistics) ReleaseMachinery(context.Context, uuid.UUID, string) (*domain.Request, error) {
	return &f.request, nil
}
func (f *fakeLogistics) UpdateMachineryStatus(context.Context, uuid.UUID, string) (*domain.Machinery, error) {
	return &f.machinery[0], nil
}

type fakeFleet struct {
	vehicles    []domain.Vehicle
	maintenance []domain.Maintenance
	createdTask domain.Task
	deleted     bool
}

func (f *fakeFleet) ListVehicles(context.Context, string, string, string) ([]domain.Vehicle, error) {
	return f.vehicles, nil
}
func (f *fakeFleet) CreateTask(context.Context, domain.CreateTaskRequest) (*domain.Task, error) {
	return &f.createdTask, nil
}
func (f *fakeFleet) ListTasks(context.Context, string) ([]domain.Task, error) {
	return []domain.Task{f.createdTask}, nil
}
func (f *fakeFleet) GetTask(context.Context, uuid.UUID) (*domain.Task, error) {
	return &f.createdTask, nil
}
func (f *fakeFleet) UpdateTaskStatus(context.Context, uuid.UUID, string) (*domain.Task, error) {
	return &f.createdTask, nil
}
func (f *fakeFleet) DeleteTask(context.Context, uuid.UUID) error {
	f.deleted = true
	return nil
}
func (f *fakeFleet) CreateGeofence(context.Context, domain.CreateGeofenceRequest) (*domain.Geofence, error) {
	return &domain.Geofence{}, nil
}
func (f *fakeFleet) ListGeofences(context.Context, string) ([]domain.Geofence, error) {
	return nil, nil
}
func (f *fakeFleet) ListMaintenance(context.Context, string, string) ([]domain.Maintenance, error) {
	return f.maintenance, nil
}

type fakeEmbeddings struct {
	query     []float32
	documents [][]float32
}

func (f *fakeEmbeddings) EmbedQuery(context.Context, string) ([]float32, error) {
	if f.query == nil {
		return []float32{1, 0}, nil
	}
	return f.query, nil
}
func (f *fakeEmbeddings) EmbedDocuments(_ context.Context, documents []string) ([][]float32, error) {
	if f.documents != nil {
		return f.documents, nil
	}
	result := make([][]float32, len(documents))
	for index := range documents {
		result[index] = []float32{1, 0}
	}
	return result, nil
}
func (f *fakeEmbeddings) Provider() string { return "fake" }
func (f *fakeEmbeddings) Model() string    { return "fake-model" }
func (f *fakeEmbeddings) Dimensions() int  { return 2 }

func assignmentFixture() (*fakeLogistics, *fakeFleet, CreateAssignmentInput) {
	requestID := uuid.New().String()
	machineryID := uuid.New().String()
	logistics := &fakeLogistics{
		request: domain.Request{ID: requestID, Type: "Cargador frontal", Status: "Pendiente"},
		machinery: []domain.Machinery{{
			ID: machineryID, AssetNumber: "CF-03 - Cargador frontal 03", Name: "Cargador frontal 03",
			EquipmentClass: "Cargador frontal", Status: "Disponible",
		}},
	}
	fleet := &fakeFleet{
		vehicles:    []domain.Vehicle{{ID: uuid.New().String(), Description: "CF-03", Type: "Cargador frontal", Driver: "MOT-014"}},
		createdTask: domain.Task{ID: uuid.New().String(), TaskID: "TAR-014", Status: "Pendiente"},
	}
	input := CreateAssignmentInput{
		RequestID: requestID, EquipmentKey: "CF03", TaskID: "TAR-014", Title: "Traslado CF-03",
		TaskType: "Traslado", ScheduledDate: "2026-09-12", Origin: "PLANTA", Destination: "PROY-014",
		Latitude: 13.3376152, Longitude: -87.8486967, Reason: "Confirmado", Confirmed: true,
	}
	return logistics, fleet, input
}

func TestCreateAssignmentWritesBothSystems(t *testing.T) {
	logistics, fleet, input := assignmentFixture()
	result, output, err := New(logistics, fleet, &fakeEmbeddings{}).CreateAssignment(context.Background(), nil, input)
	if err != nil || result != nil {
		t.Fatalf("unexpected MCP failure: result=%v err=%v", result, err)
	}
	if !logistics.assignCalled {
		t.Fatal("expected Prisma assignment to be called")
	}
	if fleet.deleted {
		t.Fatal("successful operation must not compensate the Startrack task")
	}
	if output.Request.Status != "Aprobada" || output.Equipment.PrismaStatus != "Ocupada" {
		t.Fatalf("unexpected unified output: %#v", output)
	}
}

func TestCreateAssignmentCompensatesStartrackWhenPrismaFails(t *testing.T) {
	logistics, fleet, input := assignmentFixture()
	logistics.assignError = errors.New("Prisma unavailable")
	result, _, err := New(logistics, fleet, &fakeEmbeddings{}).CreateAssignment(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("MCP failures must be returned as tool results: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected an MCP error result")
	}
	if !fleet.deleted {
		t.Fatal("expected Startrack task compensation")
	}
}
