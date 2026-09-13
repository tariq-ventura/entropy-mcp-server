package tools

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

type LogisticsService interface {
	CreateRequest(context.Context, domain.CreateRequestInput) (*domain.Request, error)
	GetRequest(context.Context, uuid.UUID) (*domain.Request, error)
	SearchRequests(context.Context, domain.SearchRequestsInput) (*domain.SearchRequestsResult, error)
	ListMachinery(context.Context, string, string, string) ([]domain.Machinery, error)
	AssignMachinery(context.Context, uuid.UUID, uuid.UUID, string) (*domain.PrismaAssignment, error)
	ReleaseMachinery(context.Context, uuid.UUID, string) (*domain.Request, error)
	UpdateMachineryStatus(context.Context, uuid.UUID, string) (*domain.Machinery, error)
}

type FleetService interface {
	ListVehicles(context.Context, string, string, string) ([]domain.Vehicle, error)
	CreateTask(context.Context, domain.CreateTaskRequest) (*domain.Task, error)
	ListTasks(context.Context, string) ([]domain.Task, error)
	GetTask(context.Context, uuid.UUID) (*domain.Task, error)
	UpdateTaskStatus(context.Context, uuid.UUID, string) (*domain.Task, error)
	DeleteTask(context.Context, uuid.UUID) error
	CreateGeofence(context.Context, domain.CreateGeofenceRequest) (*domain.Geofence, error)
	ListGeofences(context.Context, string) ([]domain.Geofence, error)
	ListMaintenance(context.Context, string, string) ([]domain.Maintenance, error)
}

type EmbeddingService interface {
	EmbedQuery(context.Context, string) ([]float32, error)
	EmbedDocuments(context.Context, []string) ([][]float32, error)
	Provider() string
	Model() string
	Dimensions() int
}

// UnifiedProjection is the MCP-owned read model shared by tools and the UI API.
// Writes still go to the authoritative source service and are reconciled here.
type UnifiedProjection interface {
	Refresh(context.Context) error
	ListAllEquipment(context.Context) ([]domain.UnifiedEquipment, error)
	ListAllMaintenance(context.Context) ([]domain.Maintenance, error)
	GetRequest(context.Context, string) (*domain.Request, error)
	SearchRequests(context.Context, domain.SearchRequestsInput) (*domain.SearchRequestsResult, error)
	GetAssignment(context.Context, string) (*domain.UnifiedAssignmentStatus, error)
}

type Tools struct {
	logistics  LogisticsService
	fleet      FleetService
	embeddings EmbeddingService
	projection UnifiedProjection
	now        func() time.Time
}

func New(logistics LogisticsService, fleet FleetService, embeddingClient EmbeddingService, projections ...UnifiedProjection) *Tools {
	var unifiedProjection UnifiedProjection
	if len(projections) > 0 {
		unifiedProjection = projections[0]
	}
	return &Tools{logistics: logistics, fleet: fleet, embeddings: embeddingClient, projection: unifiedProjection, now: time.Now}
}

type RequestIDInput struct {
	RequestID string `json:"requestId" jsonschema:"UUID de la solicitud en Prisma"`
}
type UnifiedEquipmentInput struct {
	EquipmentKey string `json:"equipmentKey" jsonschema:"Clave correlacionada, número de activo o UUID de Prisma/Startrack"`
}
type ListUnifiedEquipmentInput struct {
	Type          string `json:"type,omitempty" jsonschema:"Clase o tipo de equipo"`
	Search        string `json:"search,omitempty" jsonschema:"Texto para buscar por activo, nombre, marca o modelo"`
	OnlyAvailable bool   `json:"onlyAvailable,omitempty" jsonschema:"Devuelve solo maquinaria Disponible en Prisma"`
	OnlyLinked    bool   `json:"onlyLinked,omitempty" jsonschema:"Devuelve solo registros correlacionados entre Prisma y Startrack"`
}
type CreateAssignmentInput struct {
	RequestID     string  `json:"requestId" jsonschema:"UUID de la solicitud Prisma"`
	EquipmentKey  string  `json:"equipmentKey" jsonschema:"Clave del equipo devuelta por get_recommendations o list_unified_equipment"`
	TaskID        string  `json:"taskId" jsonschema:"Identificador externo de la tarea Startrack, por ejemplo TAR-014"`
	Title         string  `json:"title" jsonschema:"Título de la tarea de traslado"`
	Description   string  `json:"description,omitempty" jsonschema:"Descripción adicional de la tarea"`
	TaskType      string  `json:"taskType" jsonschema:"Tipo de tarea Startrack, por ejemplo Traslado"`
	ScheduledDate string  `json:"scheduledDate" jsonschema:"Fecha programada YYYY-MM-DD o RFC3339"`
	Origin        string  `json:"origin" jsonschema:"Origen del traslado"`
	Destination   string  `json:"destination" jsonschema:"Destino del traslado"`
	Latitude      float64 `json:"latitude" jsonschema:"Latitud decimal del destino"`
	Longitude     float64 `json:"longitude" jsonschema:"Longitud decimal del destino"`
	Assignee      string  `json:"assignee,omitempty" jsonschema:"Motorista; si se omite se usa el conductor de Startrack"`
	Reason        string  `json:"reason" jsonschema:"Motivo de la asignación"`
	Confirmed     bool    `json:"confirmed" jsonschema:"Debe ser true después de confirmación humana explícita"`
}
type UpdateAssignmentInput struct {
	RequestID string `json:"requestId" jsonschema:"UUID de la solicitud Prisma"`
	TaskID    string `json:"taskId" jsonschema:"UUID interno de la tarea Startrack"`
	Reason    string `json:"reason" jsonschema:"Motivo de la operación"`
	Confirmed bool   `json:"confirmed" jsonschema:"Debe ser true después de confirmación humana explícita"`
}
type CreateGeofenceInput struct {
	GeofenceID       string  `json:"geofenceId" jsonschema:"Identificador externo de Startrack"`
	Name             string  `json:"name" jsonschema:"Nombre de la geocerca, normalmente el proyecto Prisma"`
	Group            string  `json:"group,omitempty" jsonschema:"Grupo de geocercas"`
	AdditionalMargin float64 `json:"additionalMargin" jsonschema:"Radio o margen adicional en metros"`
	Latitude         float64 `json:"latitude" jsonschema:"Latitud decimal"`
	Longitude        float64 `json:"longitude" jsonschema:"Longitud decimal"`
}
type SearchInput struct {
	Search string `json:"search,omitempty" jsonschema:"Texto por nombre o identificador"`
}
