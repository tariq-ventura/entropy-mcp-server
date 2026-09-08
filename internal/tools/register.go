package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (t *Tools) Register(server *mcp.Server) {
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "create_logistics_request",
			Description: "Crea una solicitud logística PENDING. " +
				"Utiliza esta herramienta después de extraer y confirmar " +
				"tipo de maquinaria, proyecto, ubicación, descripción, " +
				"requerimientos y fechas.",
		},
		t.CreateLogisticsRequest,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_logistics_request",
			Description: "Consulta una solicitud logística por su UUID.",
		},
		t.GetLogisticsRequest,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_recommendations",
			Description: "Obtiene el ranking determinista de maquinaria " +
				"para una solicitud PENDING. No reserva maquinaria.",
		},
		t.GetRecommendations,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "create_assignment",
			Description: "Asigna y reserva una maquinaria para una solicitud. " +
				"Es una operación con efectos reales y exige confirmed=true " +
				"después de recibir confirmación humana.",
		},
		t.CreateAssignment,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_assignment",
			Description: "Consulta la asignación asociada a una solicitud.",
		},
		t.GetAssignment,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "complete_assignment",
			Description: "Completa una asignación, completa la solicitud " +
				"y libera la maquinaria. Exige confirmación humana.",
		},
		t.CompleteAssignment,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "cancel_assignment",
			Description: "Cancela una asignación, cancela la solicitud " +
				"y libera la maquinaria. Exige confirmación humana.",
		},
		t.CancelAssignment,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "search_logistics_requests",
			Description: `
				Busca solicitudes logísticas usando texto literal, significado semántico,
				estado, tipo de maquinaria y proximidad geográfica.

				Usa semanticQuery cuando el usuario describa una necesidad o concepto.
				Usa query cuando proporcione el nombre exacto de un proyecto.
				Usa near cuando solicite resultados cerca de una ciudad o dirección.
				No inventes coordenadas: deben venir de Google Maps.
			`,
			InputSchema: searchLogisticsRequestsInputSchema(),
		},
		t.SearchLogisticsRequests,
	)
}
