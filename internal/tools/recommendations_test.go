package tools

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func TestRecommendationsUseSemanticMaintenanceAndOperationalSignals(t *testing.T) {
	requestID := uuid.New().String()
	logistics := &fakeLogistics{
		request: domain.Request{ID: requestID, Project: "The Hub La Unión", Type: "Cargador frontal", Requester: "María", Status: "Pendiente", StartDate: "2026-09-15T00:00:00Z", EndDate: "2026-09-20T00:00:00Z"},
		machinery: []domain.Machinery{
			{ID: uuid.New().String(), AssetNumber: "CF-03 - Cargador frontal 03", Name: "Cargador frontal 03", EquipmentClass: "Cargador frontal", Status: "Disponible"},
			{ID: uuid.New().String(), AssetNumber: "CF-04 - Cargador frontal 04", Name: "Cargador frontal 04", EquipmentClass: "Cargador frontal", Status: "Disponible"},
			{ID: uuid.New().String(), AssetNumber: "CF-05 - Cargador frontal 05", Name: "Cargador frontal 05", EquipmentClass: "Cargador frontal", Status: "Ocupada"},
			{ID: uuid.New().String(), AssetNumber: "EX-01 - Excavadora 01", Name: "Excavadora 01", EquipmentClass: "Excavadora", Status: "Disponible"},
		},
	}
	fleet := &fakeFleet{
		vehicles: []domain.Vehicle{
			{ID: uuid.New().String(), Description: "CF-03", Type: "Cargador frontal", Status: "Normal", Driver: "MOT-014", Tags: "The Hub"},
			{ID: uuid.New().String(), Description: "CF-04", Type: "Cargador frontal", Status: "Normal", Driver: "MOT-015"},
			{ID: uuid.New().String(), Description: "CF-05", Type: "Cargador frontal", Status: "Normal", Driver: "MOT-016"},
			{ID: uuid.New().String(), Description: "EX-01", Type: "Excavadora", Status: "Normal", Driver: "MOT-017"},
		},
		maintenance: []domain.Maintenance{
			{Vehicle: "CF-03", ServiceDate: "2026-08-20T00:00:00Z", ServiceType: "Preventivo", RepairReason: "Programado", Reference: "Servicio general"},
			{Vehicle: "CF-04", ServiceDate: "2025-01-10T00:00:00Z", ServiceType: "Correctivo", RepairReason: "Emergencia", Reference: "Falla hidráulica"},
		},
	}
	embeddingClient := &fakeEmbeddings{query: []float32{1, 0}, documents: [][]float32{{1, 0}, {0.8, 0.6}}}
	toolset := New(logistics, fleet, embeddingClient)
	toolset.now = func() time.Time { return time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC) }
	result, output, err := toolset.GetRecommendations(context.Background(), nil, RequestIDInput{RequestID: requestID})
	if err != nil || result != nil {
		t.Fatalf("unexpected MCP failure: result=%v err=%v", result, err)
	}
	if output.Count != 2 {
		t.Fatalf("expected two eligible recommendations, got %d", output.Count)
	}
	if output.Recommendations[0].Equipment.EquipmentKey != "CF03" {
		t.Fatalf("expected CF03 first, got %s", output.Recommendations[0].Equipment.EquipmentKey)
	}
	if output.Recommendations[0].Score.Maintenance <= output.Recommendations[1].Score.Maintenance {
		t.Fatal("recent preventive maintenance should score above old emergency corrective maintenance")
	}
	if output.AlgorithmVersion != recommendationAlgorithmVersion || output.EmbeddingProvider != "fake" {
		t.Fatalf("missing algorithm metadata: %#v", output)
	}
}

func TestCosineScoreRejectsIncompatibleVectors(t *testing.T) {
	if _, err := cosineScore([]float32{1}, []float32{1, 0}); err == nil {
		t.Fatal("expected incompatible vector error")
	}
}
