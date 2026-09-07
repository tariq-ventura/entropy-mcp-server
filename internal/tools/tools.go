package tools

import (
	"github.com/tariq-ventura/entropy-mcp-server/internal/client"
)

type Tools struct {
	logistics *client.Client
}

type RequestIDInput struct {
	RequestID string `json:"requestId" jsonschema:"UUID de la solicitud logística"`
}

type CreateAssignmentInput struct {
	RequestID   string `json:"requestId" jsonschema:"UUID de la solicitud logística"`
	EquipmentID string `json:"equipmentId" jsonschema:"UUID de la maquinaria seleccionada"`
	Reason      string `json:"reason" jsonschema:"Motivo de la asignación"`
	Confirmed   bool   `json:"confirmed" jsonschema:"Debe ser true únicamente después de confirmación humana explícita"`
}

type UpdateAssignmentInput struct {
	AssignmentID string `json:"assignmentId" jsonschema:"UUID de la asignación"`
	Reason       string `json:"reason" jsonschema:"Motivo de la operación"`
	Confirmed    bool   `json:"confirmed" jsonschema:"Debe ser true únicamente después de confirmación humana explícita"`
}

func New(logisticsClient *client.Client) *Tools {
	return &Tools{
		logistics: logisticsClient,
	}
}
