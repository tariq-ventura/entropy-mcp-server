package tools

import (
	"testing"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func TestUnifyEquipmentUsesPrismaAvailability(t *testing.T) {
	machinery := []domain.Machinery{{ID: "prisma-1", Company: "The Hub", AssetNumber: "CF-03 - Cargador frontal 03", Name: "Cargador frontal 03", EquipmentClass: "Cargador frontal", Status: "Disponible"}}
	vehicles := []domain.Vehicle{{ID: "startrack-1", Description: "CF-03", Status: "Normal", Type: "Cargador frontal", Brand: "Caterpillar", Model: "950H", Driver: "MOT-014 - Adriana Steiner", RemoteID: "78093"}}
	result := UnifyEquipment(machinery, vehicles)
	if len(result) != 1 {
		t.Fatalf("expected one unified record, got %d", len(result))
	}
	item := result[0]
	if item.EquipmentKey != "CF03" {
		t.Fatalf("unexpected correlation key: %s", item.EquipmentKey)
	}
	if !item.Linked {
		t.Fatal("expected the Prisma and Startrack records to be linked")
	}
	if !item.Available {
		t.Fatal("Prisma Disponible must define availability")
	}
	if item.TrackingStatus != "Normal" {
		t.Fatalf("unexpected tracking status: %s", item.TrackingStatus)
	}
}

func TestUnifyEquipmentReportsTypeConflict(t *testing.T) {
	machinery := []domain.Machinery{{ID: "prisma-1", AssetNumber: "EX-01 - Excavadora", Name: "Excavadora", EquipmentClass: "Excavadora", Status: "Ocupada"}}
	vehicles := []domain.Vehicle{{ID: "startrack-1", Description: "EX-01", Status: "Normal", Type: "Retroexcavadora"}}
	result := UnifyEquipment(machinery, vehicles)
	if len(result) != 1 {
		t.Fatalf("expected one unified record, got %d", len(result))
	}
	if result[0].Available {
		t.Fatal("Ocupada must not be exposed as available")
	}
	if len(result[0].Conflicts) == 0 {
		t.Fatal("expected a type conflict")
	}
}

func TestCoordinateConversion(t *testing.T) {
	value := 13.3376152
	remote := toRemoteCoordinate(value)
	if remote != 133376152 {
		t.Fatalf("unexpected Startrack coordinate: %d", remote)
	}
	if got := fromRemoteCoordinate(remote); got != value {
		t.Fatalf("unexpected decimal coordinate: %.7f", got)
	}
}
