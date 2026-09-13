package tools

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/domain"
)

const recommendationAlgorithmVersion = "hybrid-v1"

type recommendationCandidate struct {
	equipment        domain.UnifiedEquipment
	summary          domain.MaintenanceSummary
	document         string
	reasons          []string
	warnings         []string
	maintenanceScore float64
	operationalScore float64
}

func (t *Tools) GetRecommendations(ctx context.Context, _ *mcp.CallToolRequest, input RequestIDInput) (*mcp.CallToolResult, domain.Recommendations, error) {
	requestID, err := parseUUID(input.RequestID, "requestId")
	if err != nil {
		return toolFailure[domain.Recommendations](err)
	}
	items, err := t.loadUnifiedEquipment(ctx)
	if err != nil {
		return toolFailure[domain.Recommendations](err)
	}
	var request *domain.Request
	if t.projection != nil {
		request, err = t.projection.GetRequest(ctx, requestID.String())
	} else {
		request, err = t.logistics.GetRequest(ctx, requestID)
	}
	if err != nil {
		return toolFailure[domain.Recommendations](err)
	}
	if !strings.EqualFold(strings.TrimSpace(request.Status), "Pendiente") && !strings.EqualFold(strings.TrimSpace(request.Status), "PENDING") {
		return toolFailure[domain.Recommendations](errors.New("recommendations are available only for pending Prisma requests"))
	}
	var maintenance []domain.Maintenance
	if t.projection != nil {
		maintenance, err = t.projection.ListAllMaintenance(ctx)
	} else {
		maintenance, err = t.fleet.ListMaintenance(ctx, "", "")
	}
	if err != nil {
		return toolFailure[domain.Recommendations](err)
	}
	maintenanceByKey := groupMaintenance(maintenance)
	candidates := make([]recommendationCandidate, 0)
	for _, item := range items {
		if !item.Linked || !item.Available || !sameText(item.Type, request.Type) {
			continue
		}
		history := maintenanceByKey[item.EquipmentKey]
		summary, maintenanceScore, maintenanceReasons, maintenanceWarnings := evaluateMaintenance(history, t.now().UTC())
		operationalScore, operationalReasons, operationalWarnings := evaluateOperationalReadiness(item)
		reasons := []string{"Disponible según Prisma", "Registro correlacionado con Startrack", "Clase compatible con la solicitud"}
		reasons = append(reasons, maintenanceReasons...)
		reasons = append(reasons, operationalReasons...)
		warnings := append([]string{}, maintenanceWarnings...)
		warnings = append(warnings, operationalWarnings...)
		candidates = append(candidates, recommendationCandidate{
			equipment: item, summary: summary,
			document: recommendationDocument(item, history), reasons: reasons, warnings: warnings,
			maintenanceScore: maintenanceScore, operationalScore: operationalScore,
		})
	}
	result := domain.Recommendations{
		RequestID: request.ID, AlgorithmVersion: recommendationAlgorithmVersion,
		EmbeddingProvider: t.embeddings.Provider(), EmbeddingModel: t.embeddings.Model(),
		Recommendations: make([]domain.Recommendation, 0, len(candidates)),
	}
	if len(candidates) == 0 {
		return nil, result, nil
	}
	queryVector, err := t.embeddings.EmbedQuery(ctx, recommendationQuery(*request))
	if err != nil {
		return toolFailure[domain.Recommendations](fmt.Errorf("embedding the Prisma request: %w", err))
	}
	documents := make([]string, len(candidates))
	for index := range candidates {
		documents[index] = candidates[index].document
	}
	documentVectors, err := t.embeddings.EmbedDocuments(ctx, documents)
	if err != nil {
		return toolFailure[domain.Recommendations](fmt.Errorf("embedding unified equipment: %w", err))
	}
	if len(documentVectors) != len(candidates) {
		return toolFailure[domain.Recommendations](errors.New("embedding provider returned an unexpected number of document vectors"))
	}
	for index, candidate := range candidates {
		semanticScore, similarityError := cosineScore(queryVector, documentVectors[index])
		if similarityError != nil {
			return toolFailure[domain.Recommendations](similarityError)
		}
		total := roundScore(semanticScore*0.50 + candidate.maintenanceScore*0.30 + candidate.operationalScore*0.20)
		reasons := append(candidate.reasons, fmt.Sprintf("Relevancia semántica: %.2f/100", semanticScore))
		result.Recommendations = append(result.Recommendations, domain.Recommendation{
			Score:     domain.RecommendationScore{Total: total, Semantic: roundScore(semanticScore), Maintenance: roundScore(candidate.maintenanceScore), Operational: roundScore(candidate.operationalScore)},
			Equipment: candidate.equipment, Maintenance: candidate.summary, Reasons: reasons, Warnings: candidate.warnings,
		})
	}
	sort.SliceStable(result.Recommendations, func(i, j int) bool {
		if result.Recommendations[i].Score.Total == result.Recommendations[j].Score.Total {
			return result.Recommendations[i].Equipment.EquipmentKey < result.Recommendations[j].Equipment.EquipmentKey
		}
		return result.Recommendations[i].Score.Total > result.Recommendations[j].Score.Total
	})
	for index := range result.Recommendations {
		result.Recommendations[index].Rank = index + 1
	}
	result.Count = len(result.Recommendations)
	return nil, result, nil
}

func recommendationQuery(request domain.Request) string {
	return fmt.Sprintf("Solicitud de maquinaria. Proyecto: %s. Clase requerida: %s. Solicitante: %s. Periodo: %s a %s.", request.Project, request.Type, request.Requester, request.StartDate, request.EndDate)
}

func recommendationDocument(item domain.UnifiedEquipment, history []domain.Maintenance) string {
	parts := []string{
		fmt.Sprintf("Maquinaria %s. Activo %s. Clase %s. Empresa %s.", item.Name, item.AssetNumber, item.Type, item.Company),
		fmt.Sprintf("Vehiculo Startrack %s. Marca %s. Modelo %s. Año %d. Grupo %s. Etiquetas %s. Conductor %s. Estado de rastreo %s.", item.StartrackDescription, item.Brand, item.Model, item.Year, item.Group, item.Tags, item.Driver, item.TrackingStatus),
	}
	for index, record := range sortedMaintenance(history) {
		if index >= 3 {
			break
		}
		parts = append(parts, fmt.Sprintf("Mantenimiento %s: tipo %s, motivo %s, referencia %s, odometro %.2f, horometro %.2f.", record.ServiceDate, record.ServiceType, record.RepairReason, record.Reference, record.Odometer, record.HourMeter))
	}
	return strings.Join(parts, " ")
}

func groupMaintenance(items []domain.Maintenance) map[string][]domain.Maintenance {
	result := make(map[string][]domain.Maintenance)
	for _, item := range items {
		key := correlationKey(item.Vehicle)
		result[key] = append(result[key], item)
	}
	return result
}

func evaluateMaintenance(history []domain.Maintenance, now time.Time) (domain.MaintenanceSummary, float64, []string, []string) {
	valid := make([]domain.Maintenance, 0, len(history))
	warnings := make([]string, 0)
	for _, record := range history {
		date, err := time.Parse(time.RFC3339, record.ServiceDate)
		if err != nil {
			warnings = append(warnings, "Se ignoró un mantenimiento con fecha inválida")
			continue
		}
		if date.After(now.Add(24 * time.Hour)) {
			warnings = append(warnings, "Se ignoró un mantenimiento con fecha futura")
			continue
		}
		valid = append(valid, record)
	}
	valid = sortedMaintenance(valid)
	if len(valid) == 0 {
		return domain.MaintenanceSummary{Signals: []string{"Sin historial válido"}}, 45, nil, append(warnings, "No existe historial de mantenimiento válido; se aplica un puntaje neutral-bajo")
	}
	latest := valid[0]
	latestDate, _ := time.Parse(time.RFC3339, latest.ServiceDate)
	ageDays := int(now.Sub(latestDate).Hours() / 24)
	score := 55.0
	signals := make([]string, 0)
	switch {
	case ageDays <= 90:
		score += 20
		signals = append(signals, "Servicio realizado en los últimos 90 días")
	case ageDays <= 180:
		score += 10
		signals = append(signals, "Servicio realizado en los últimos 180 días")
	case ageDays > 365:
		score -= 20
		signals = append(signals, "Último servicio con más de 365 días")
	default:
		signals = append(signals, "Servicio realizado entre 181 y 365 días")
	}
	serviceType := strings.ToLower(latest.ServiceType)
	if strings.Contains(serviceType, "prevent") {
		score += 15
		signals = append(signals, "Último servicio preventivo")
	} else if strings.Contains(serviceType, "correct") {
		score -= 5
		signals = append(signals, "Último servicio correctivo")
	}
	if strings.Contains(strings.ToLower(latest.RepairReason), "emerg") {
		score -= 10
		signals = append(signals, "Último motivo marcado como emergencia")
	}
	recentCorrective := 0
	for _, record := range valid {
		date, _ := time.Parse(time.RFC3339, record.ServiceDate)
		if now.Sub(date) <= 180*24*time.Hour && strings.Contains(strings.ToLower(record.ServiceType), "correct") {
			recentCorrective++
		}
	}
	if recentCorrective > 1 {
		penalty := math.Min(15, float64(recentCorrective-1)*5)
		score -= penalty
		signals = append(signals, fmt.Sprintf("%d servicios correctivos en los últimos 180 días", recentCorrective))
	}
	summary := domain.MaintenanceSummary{
		RecordCount: len(valid), LatestServiceDate: latest.ServiceDate, LatestServiceType: latest.ServiceType,
		LatestRepairReason: latest.RepairReason, LatestReference: latest.Reference,
		LatestOdometer: latest.Odometer, LatestHourMeter: latest.HourMeter,
		RecentCorrectiveCount: recentCorrective, Signals: signals,
	}
	reasons := []string{fmt.Sprintf("Mantenimiento: %.2f/100; último servicio %s", clampScore(score), latest.ServiceDate)}
	return summary, clampScore(score), reasons, warnings
}

func evaluateOperationalReadiness(item domain.UnifiedEquipment) (float64, []string, []string) {
	score := 50.0
	reasons := make([]string, 0)
	warnings := make([]string, 0)
	status := strings.ToLower(strings.TrimSpace(item.TrackingStatus))
	if status == "normal" || status == "activo" || status == "active" {
		score += 30
		reasons = append(reasons, "Estado de rastreo operativo en Startrack")
	} else {
		warnings = append(warnings, "El estado de rastreo Startrack no es Normal/Activo: "+item.TrackingStatus)
	}
	if strings.TrimSpace(item.Driver) != "" {
		score += 20
		reasons = append(reasons, "Conductor registrado en Startrack: "+item.Driver)
	} else {
		warnings = append(warnings, "El vehículo no tiene conductor registrado")
	}
	if len(item.Conflicts) > 0 {
		score -= math.Min(20, float64(len(item.Conflicts))*10)
		for _, conflict := range item.Conflicts {
			warnings = append(warnings, "Conflicto de datos unificados: "+conflict)
		}
	}
	return clampScore(score), reasons, warnings
}

func sortedMaintenance(items []domain.Maintenance) []domain.Maintenance {
	result := append([]domain.Maintenance(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		left, leftError := time.Parse(time.RFC3339, result[i].ServiceDate)
		right, rightError := time.Parse(time.RFC3339, result[j].ServiceDate)
		if leftError != nil || rightError != nil {
			return result[i].ServiceDate > result[j].ServiceDate
		}
		return left.After(right)
	})
	return result
}

func cosineScore(left, right []float32) (float64, error) {
	if len(left) == 0 || len(left) != len(right) {
		return 0, errors.New("embedding vectors have incompatible dimensions")
	}
	var dot, leftNorm, rightNorm float64
	for index := range left {
		l, r := float64(left[index]), float64(right[index])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0, errors.New("embedding provider returned a zero vector")
	}
	return clampScore(dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm)) * 100), nil
}

func clampScore(value float64) float64 { return math.Max(0, math.Min(100, value)) }
func roundScore(value float64) float64 { return math.Round(value*100) / 100 }
