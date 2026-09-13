package integration

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

func CorrelationKey(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.Index(value, " - "); index > 0 {
		value = value[:index]
	}
	return Normalize(value)
}

func Normalize(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}
		return -1
	}, strings.TrimSpace(value))
}

func SameText(left, right string) bool { return Normalize(left) == Normalize(right) }

func UnifyEquipment(machinery []domain.Machinery, vehicles []domain.Vehicle) []domain.UnifiedEquipment {
	vehicleByKey := make(map[string][]domain.Vehicle)
	for _, vehicle := range vehicles {
		key := CorrelationKey(vehicle.Description)
		if key == "" {
			key = "STARTRACK" + Normalize(vehicle.ID)
		}
		vehicleByKey[key] = append(vehicleByKey[key], vehicle)
	}
	used := make(map[string]bool)
	result := make([]domain.UnifiedEquipment, 0, len(machinery)+len(vehicles))
	for _, item := range machinery {
		key := CorrelationKey(item.AssetNumber)
		if key == "" {
			key = "PRISMA" + Normalize(item.ID)
		}
		matches := vehicleByKey[key]
		unified := domain.UnifiedEquipment{EquipmentKey: key, Name: item.Name, AssetNumber: item.AssetNumber, Type: item.EquipmentClass, PrismaStatus: item.Status, Company: item.Company, PrismaID: item.ID, Available: strings.EqualFold(strings.TrimSpace(item.Status), "Disponible")}
		if len(matches) > 0 {
			vehicle := matches[0]
			for _, match := range matches {
				used[match.ID] = true
			}
			unified.Linked = true
			unified.StartrackID = vehicle.ID
			unified.StartrackType = vehicle.Type
			unified.TrackingStatus = vehicle.Status
			unified.Year = vehicle.Year
			unified.Color = vehicle.Color
			unified.Brand = vehicle.Brand
			unified.Model = vehicle.Model
			unified.Group = vehicle.Group
			unified.Tags = vehicle.Tags
			unified.Driver = vehicle.Driver
			unified.RemoteID = vehicle.RemoteID
			unified.StartrackDescription = vehicle.Description
			if !SameText(item.EquipmentClass, vehicle.Type) {
				unified.Conflicts = append(unified.Conflicts, fmt.Sprintf("type mismatch: Prisma=%s, Startrack=%s", item.EquipmentClass, vehicle.Type))
			}
			if len(matches) > 1 {
				unified.Conflicts = append(unified.Conflicts, "multiple Startrack vehicles share the correlation key")
			}
		} else {
			unified.Conflicts = append(unified.Conflicts, "missing Startrack vehicle")
		}
		result = append(result, unified)
	}
	for _, vehicle := range vehicles {
		if used[vehicle.ID] {
			continue
		}
		key := CorrelationKey(vehicle.Description)
		if key == "" {
			key = "STARTRACK" + Normalize(vehicle.ID)
		}
		result = append(result, domain.UnifiedEquipment{EquipmentKey: key, Name: vehicle.Description, Type: vehicle.Type, StartrackType: vehicle.Type, TrackingStatus: vehicle.Status, Year: vehicle.Year, Color: vehicle.Color, Brand: vehicle.Brand, Model: vehicle.Model, Group: vehicle.Group, Tags: vehicle.Tags, Driver: vehicle.Driver, RemoteID: vehicle.RemoteID, StartrackDescription: vehicle.Description, StartrackID: vehicle.ID, Conflicts: []string{"missing Prisma machinery"}})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].EquipmentKey < result[j].EquipmentKey })
	return result
}

func FindEquipment(items []domain.UnifiedEquipment, value string) (domain.UnifiedEquipment, bool) {
	key := CorrelationKey(value)
	for _, item := range items {
		if item.EquipmentKey == key || strings.EqualFold(item.PrismaID, strings.TrimSpace(value)) || strings.EqualFold(item.StartrackID, strings.TrimSpace(value)) || SameText(item.AssetNumber, value) {
			return item, true
		}
	}
	return domain.UnifiedEquipment{}, false
}
