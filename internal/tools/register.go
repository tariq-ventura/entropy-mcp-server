package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (t *Tools) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "create_logistics_request", Description: "Crea en Prisma una solicitud de maquinaria Pendiente con proyecto, tipo, solicitante y período."}, t.CreateLogisticsRequest)
	mcp.AddTool(server, &mcp.Tool{Name: "get_logistics_request", Description: "Consulta por UUID una solicitud Prisma desde la proyección canónica reconciliada del MCP."}, t.GetLogisticsRequest)
	mcp.AddTool(server, &mcp.Tool{Name: "search_logistics_requests", Description: "Busca solicitudes reconciliadas por texto, estado, tipo y solicitante."}, t.SearchLogisticsRequests)
	mcp.AddTool(server, &mcp.Tool{Name: "list_unified_equipment", Description: "Consulta el inventario canónico del MCP. Prisma es la verdad de disponibilidad; Startrack aporta telemática, mantenimiento y conductor; los desacuerdos se publican como conflictos."}, t.ListUnifiedEquipment)
	mcp.AddTool(server, &mcp.Tool{Name: "get_unified_equipment", Description: "Obtiene un equipo unificado por clave correlacionada, número de activo o UUID de cualquiera de los sistemas."}, t.GetUnifiedEquipment)
	mcp.AddTool(server, &mcp.Tool{Name: "get_recommendations", Description: "Genera un ranking híbrido de maquinaria: aplica elegibilidad Prisma, une datos Startrack, evalúa historial de mantenimiento y calcula relevancia semántica con embeddings. Devuelve score, desglose, razones y advertencias; no realiza escrituras."}, t.GetRecommendations)
	mcp.AddTool(server, &mcp.Tool{Name: "create_assignment", Description: "Operación unificada: crea la tarea de traslado en Startrack y luego asigna la maquinaria y aprueba la solicitud en Prisma. Compensa la tarea si Prisma falla. Requiere confirmed=true."}, t.CreateAssignment)
	mcp.AddTool(server, &mcp.Tool{Name: "get_assignment", Description: "Consulta la asignación ya reconciliada por el MCP entre la solicitud Prisma, el equipo canónico y la tarea Startrack."}, t.GetAssignment)
	mcp.AddTool(server, &mcp.Tool{Name: "complete_assignment", Description: "Marca la tarea Startrack como Completada y deja la maquinaria Disponible en Prisma, conservando la solicitud Aprobada. Requiere confirmed=true."}, t.CompleteAssignment)
	mcp.AddTool(server, &mcp.Tool{Name: "cancel_assignment", Description: "Marca la tarea Startrack como Cancelada y revierte la solicitud Prisma a Pendiente liberando la maquinaria. Requiere confirmed=true."}, t.CancelAssignment)
	mcp.AddTool(server, &mcp.Tool{Name: "create_geofence", Description: "Crea una geocerca Startrack usando coordenadas decimales; el MCP las convierte al formato fijo de Startrack."}, t.CreateGeofence)
	mcp.AddTool(server, &mcp.Tool{Name: "list_geofences", Description: "Lista geocercas Startrack y devuelve también coordenadas decimales listas para n8n."}, t.ListGeofences)
}
