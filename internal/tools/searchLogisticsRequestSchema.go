package tools

func searchLogisticsRequestsInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Búsqueda literal por nombre del proyecto, ubicación o tipo",
				"maxLength":   500,
			},
			"semanticQuery": map[string]any{
				"type":        "string",
				"description": "Descripción en lenguaje natural de las solicitudes buscadas",
				"maxLength":   500,
			},
			"statuses": map[string]any{
				"type":        "array",
				"description": "Estados permitidos para filtrar las solicitudes",
				"items": map[string]any{
					"type": "string",
					"enum": []string{
						"PENDING",
						"ASSIGNED",
						"COMPLETED",
						"CANCELLED",
					},
				},
				"uniqueItems": true,
			},
			"equipmentType": map[string]any{
				"type":        "string",
				"description": "Tipo exacto de maquinaria, por ejemplo EXCAVATOR o CRANE",
			},
			"near": map[string]any{
				"type":        "object",
				"description": "Coordenadas y radio para la búsqueda geográfica",
				"properties": map[string]any{
					"latitude": map[string]any{
						"type":        "number",
						"description": "Latitud resuelta de la ubicación",
						"minimum":     -90,
						"maximum":     90,
					},
					"longitude": map[string]any{
						"type":        "number",
						"description": "Longitud resuelta de la ubicación",
						"minimum":     -180,
						"maximum":     180,
					},
					"radiusKm": map[string]any{
						"type":        "number",
						"description": "Radio de búsqueda en kilómetros",
						"minimum":     0.1,
						"maximum":     500,
					},
				},
				"required": []string{
					"latitude",
					"longitude",
					"radiusKm",
				},
				"additionalProperties": false,
			},
			"minSemanticScore": map[string]any{
				"type":        "number",
				"description": "Similitud semántica mínima entre 0 y 1",
				"minimum":     0,
				"maximum":     1,
			},
			"page": map[string]any{
				"type":        "integer",
				"description": "Página solicitada",
				"minimum":     1,
			},
			"pageSize": map[string]any{
				"type":        "integer",
				"description": "Cantidad máxima de resultados",
				"minimum":     1,
				"maximum":     100,
			},
		},
		"additionalProperties": false,
	}
}
