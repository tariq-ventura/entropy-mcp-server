package integration

import (
	"testing"
	"time"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func TestBuildSnapshotUnifiesSourcesAndDerivesActiveAssignment(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	machineryID := "11111111-1111-1111-1111-111111111111"
	requestID := "22222222-2222-2222-2222-222222222222"
	asset := "EQ-001"
	snapshot := buildSnapshot(
		[]domain.Machinery{{ID: machineryID, AssetNumber: asset, Name: "Cargador 1", EquipmentClass: "Cargador frontal", Status: "Asignado"}},
		[]domain.Vehicle{{ID: "33333333-3333-3333-3333-333333333333", Description: "EQ-001 - Cargador 1", Type: "Cargador Frontal", Status: "Normal", Driver: "Ana"}},
		[]domain.Request{{ID: requestID, Project: "Proyecto Norte", Type: "Cargador frontal", Status: "Aprobada", Machinery: &asset}},
		[]domain.Task{{ID: "44444444-4444-4444-4444-444444444444", TaskID: "TAR-001", Status: "Pendiente", Description: "Traslado\nEntropy request: " + requestID + "\nPrisma machinery: " + asset}},
		[]domain.Maintenance{{ID: "55555555-5555-5555-5555-555555555555", Vehicle: asset, ServiceType: "Preventivo"}},
		"sync-token",
		now,
	)

	if len(snapshot.Equipment) != 1 || !snapshot.Equipment[0].Linked {
		t.Fatalf("expected one linked equipment, got %#v", snapshot.Equipment)
	}
	if snapshot.Equipment[0].EquipmentTypeCode != "CARGADORFRONTAL" {
		t.Fatalf("unexpected canonical type code: %s", snapshot.Equipment[0].EquipmentTypeCode)
	}
	if len(snapshot.Types) != 1 || len(snapshot.Aliases) != 2 {
		t.Fatalf("expected one type and two source aliases, got types=%d aliases=%d", len(snapshot.Types), len(snapshot.Aliases))
	}
	if len(snapshot.Requests) != 1 || snapshot.Requests[0].EquipmentKey == nil || *snapshot.Requests[0].EquipmentKey != "EQ001" {
		t.Fatalf("request was not linked to canonical equipment: %#v", snapshot.Requests)
	}
	if len(snapshot.Assignments) != 1 || snapshot.Assignments[0].Status != "ACTIVE" || snapshot.Assignments[0].Inconsistent {
		t.Fatalf("expected an active consistent assignment, got %#v", snapshot.Assignments)
	}
	if len(snapshot.Maintenance) != 1 || snapshot.Maintenance[0].EquipmentKey != "EQ001" {
		t.Fatalf("maintenance was not linked to canonical equipment: %#v", snapshot.Maintenance)
	}
}

func TestBuildSnapshotPublishesAssignmentConflict(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	snapshot := buildSnapshot(nil, nil, nil, []domain.Task{{ID: "task", Status: "Pendiente", Description: "Entropy request: missing-request"}}, nil, "sync-token", now)
	if len(snapshot.Assignments) != 1 || snapshot.Assignments[0].Status != "INCONSISTENT" {
		t.Fatalf("expected an inconsistent assignment, got %#v", snapshot.Assignments)
	}
	if len(snapshot.Conflicts) == 0 || snapshot.Conflicts[0].EntityType != "assignment" {
		t.Fatalf("expected an assignment conflict, got %#v", snapshot.Conflicts)
	}
}
