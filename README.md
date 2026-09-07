# Entropy MCP Server

Servidor Model Context Protocol (MCP) de Entropy Platform. Expone como herramientas las operaciones principales del `logistic-service`, permitiendo que asistentes de IA y orquestadores como n8n consulten solicitudes, obtengan recomendaciones y administren asignaciones sin acceder directamente a la base de datos.

## 1. Propósito

`entropy-mcp-server` funciona como una capa controlada entre un cliente MCP y el API REST de logística.

Sus responsabilidades son:

- publicar herramientas con contratos JSON entendibles por un agente de IA;
- validar UUID y campos operativos antes de invocar el backend;
- exigir confirmación humana para operaciones que modifican asignaciones;
- autenticarse ante `logistic-service` cuando se configura un token;
- convertir errores HTTP del backend en resultados MCP con `isError: true`;
- propagar la cancelación de las solicitudes mediante `context.Context`.

Este servicio no:

- accede directamente a PostgreSQL;
- calcula el ranking de maquinaria;
- administra flotas o equipos;
- reemplaza las reglas de negocio de `logistic-service`;
- utiliza un modelo de lenguaje por sí mismo.

## 2. Arquitectura

```mermaid
flowchart LR
    U["Usuario"] --> A["Agente IA / n8n"]
    A -->|"MCP + Bearer"| M["Entropy MCP Server"]
    M -->|"REST + token opcional"| L["Logistic Service"]
    L --> F["Fleet Service"]
    L --> D[("PostgreSQL")]
```

El agente decide qué herramienta solicitar según la conversación. El MCP valida y traduce esa llamada a HTTP. Las decisiones operativas permanecen en los microservicios.

### Flujo de una asignación

```mermaid
sequenceDiagram
    participant U as Usuario
    participant A as Agente / n8n
    participant M as MCP Server
    participant L as Logistic Service
    participant F as Fleet Service

    U->>A: Solicita maquinaria
    A->>M: create_logistics_request
    M->>L: POST /requests
    L-->>M: Solicitud PENDING
    A->>M: get_recommendations
    M->>L: GET /recommendations
    L->>F: Consultar equipos disponibles
    F-->>L: Equipos elegibles
    L-->>M: Ranking determinista
    A-->>U: Presenta opciones y solicita confirmación
    U->>A: Confirma una maquinaria
    A->>M: create_assignment (confirmed=true)
    M->>L: POST /assignment
    L-->>M: Asignación creada
    M-->>A: Resultado estructurado
```

## 3. Tecnologías

| Componente | Tecnología |
| --- | --- |
| Lenguaje | Go 1.27 |
| Protocolo | Model Context Protocol sobre Streamable HTTP |
| SDK MCP | `github.com/modelcontextprotocol/go-sdk` v1.7.0 |
| Identificadores | `github.com/google/uuid` v1.6.0 |
| Comunicación interna | HTTP/JSON |
| Imagen de ejecución | Distroless Debian 12 |
| Puerto predeterminado | `3002` |

## 4. Estructura del proyecto

```text
entropy-mcp-server/
├── cmd/services/main.go
├── internal/
│   ├── client/
│   │   ├── client.go
│   │   ├── do.go
│   │   ├── decodeApiError.go
│   │   └── operaciones de Logistics
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── request.go
│   │   ├── recomendation.go
│   │   ├── assignment.go
│   │   └── response.go
│   └── tools/
│       ├── register.go
│       ├── tools.go
│       ├── validaciones
│       └── implementaciones MCP
├── Dockerfile
├── go.mod
└── go.sum
```

### Responsabilidad de cada paquete

| Paquete | Responsabilidad |
| --- | --- |
| `cmd/services` | Inicialización, rutas HTTP, autenticación, servidor y apagado controlado. |
| `internal/config` | Lectura y validación de variables de entorno. |
| `internal/client` | Cliente HTTP reutilizable para `logistic-service`. |
| `internal/domain` | Contratos de entrada y salida usados por Tools y Client. |
| `internal/tools` | Registro, validación y ejecución de las herramientas MCP. |

## 5. Endpoints HTTP

| Método | Ruta | Autenticación | Uso |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Verifica que el proceso HTTP esté activo. |
| `POST` | `/mcp` | Bearer token | Procesa mensajes JSON-RPC de MCP. |

Respuesta de salud:

```json
{
  "status": "ok",
  "service": "entropy-mcp-server"
}
```

> `/health` solo confirma que el servidor está ejecutándose. Actualmente no comprueba la disponibilidad de `logistic-service`.

## 6. Herramientas MCP

### Resumen

| Herramienta | Tipo | Efecto operacional | Confirmación explícita |
| --- | --- | --- | --- |
| `create_logistics_request` | Escritura | Crea una solicitud `PENDING`. | Gestionada por la conversación; no posee campo `confirmed`. |
| `get_logistics_request` | Lectura | Consulta una solicitud. | No |
| `get_recommendations` | Lectura | Obtiene un ranking; no reserva equipos. | No |
| `create_assignment` | Escritura | Asigna y reserva una maquinaria. | Sí, `confirmed=true` |
| `get_assignment` | Lectura | Consulta la asignación de una solicitud. | No |
| `complete_assignment` | Escritura | Completa la asignación y libera la maquinaria. | Sí, `confirmed=true` |
| `cancel_assignment` | Escritura | Cancela la asignación y libera la maquinaria. | Sí, `confirmed=true` |

### `create_logistics_request`

Crea una solicitud logística. La ubicación debe llegar resuelta a nombre, latitud y longitud; el MCP no geocodifica direcciones.

Entrada:

```json
{
  "equipmentType": "EXCAVATOR",
  "projectName": "Proyecto Escalón Galerías",
  "location": {
    "name": "P.º Gral. Escalón 3700, San Salvador, El Salvador",
    "latitude": 13.7022056,
    "longitude": -89.2299316
  },
  "startDate": "2030-09-10T08:00:00Z",
  "endDate": "2030-09-15T17:00:00Z"
}
```

Consideraciones:

- las fechas se documentan como RFC3339;
- el tipo debe coincidir con los valores aceptados por Logistics;
- esta herramienta no acepta la propiedad `confirmed`;
- la validación completa de fechas y estados corresponde a `logistic-service`.

### `get_logistics_request`

Entrada:

```json
{
  "requestId": "06ab32ff-9b61-4b8d-aaa9-4fd72fa99658"
}
```

Devuelve la solicitud encontrada o un resultado MCP de error cuando el UUID es inválido o Logistics rechaza la consulta.

### `get_recommendations`

Entrada:

```json
{
  "requestId": "06ab32ff-9b61-4b8d-aaa9-4fd72fa99658"
}
```

Salida conceptual:

```json
{
  "requestId": "06ab32ff-9b61-4b8d-aaa9-4fd72fa99658",
  "count": 1,
  "recommendations": [
    {
      "equipmentId": "e1b4798b-256f-4f4d-abf3-fd5213da6fcd",
      "code": "EQ-EXC-001",
      "type": "EXCAVATOR",
      "brand": "Caterpillar",
      "model": "320",
      "serialNumber": "CAT320-DEMO-001",
      "year": 2025,
      "capacityTons": 23,
      "location": {
        "name": "San Salvador",
        "latitude": 13.6929,
        "longitude": -89.2182
      },
      "distanceKm": 1.7,
      "engineHours": 1000,
      "nextMaintenanceHours": 1500,
      "maintenanceHoursRemaining": 500,
      "fuelPercent": 80,
      "score": 96.49,
      "reasons": [
        "Disponible",
        "Cerca del proyecto",
        "Margen de mantenimiento suficiente"
      ]
    }
  ]
}
```

El MCP no genera el score. Devuelve el ranking producido por `logistic-service`.

### `create_assignment`

Entrada:

```json
{
  "requestId": "06ab32ff-9b61-4b8d-aaa9-4fd72fa99658",
  "equipmentId": "e1b4798b-256f-4f4d-abf3-fd5213da6fcd",
  "reason": "Seleccionada por ser la mejor recomendación",
  "confirmed": true
}
```

Reglas aplicadas por el MCP:

- `confirmed` debe ser `true`;
- `requestId` y `equipmentId` deben ser UUID válidos;
- `reason` no puede estar vacío.

La operación real se ejecuta en:

```http
POST /api/v1/requests/{requestId}/assignment
```

### `get_assignment`

Consulta la asignación vinculada a una solicitud.

```json
{
  "requestId": "06ab32ff-9b61-4b8d-aaa9-4fd72fa99658"
}
```

### `complete_assignment`

```json
{
  "assignmentId": "770501e7-7c49-4081-9151-0f11668837e3",
  "reason": "Trabajo finalizado",
  "confirmed": true
}
```

Solicita a Logistics el cambio de estado a `COMPLETED`. El comportamiento esperado del flujo completo es completar la solicitud y liberar la maquinaria.

### `cancel_assignment`

```json
{
  "assignmentId": "770501e7-7c49-4081-9151-0f11668837e3",
  "reason": "El proyecto fue cancelado",
  "confirmed": true
}
```

Solicita el cambio de estado a `CANCELLED`. También requiere confirmación explícita.

## 7. Relación entre herramientas y Logistics API

| Herramienta MCP | Método REST | Ruta en `logistic-service` |
| --- | --- | --- |
| `create_logistics_request` | `POST` | `/api/v1/requests` |
| `get_logistics_request` | `GET` | `/api/v1/requests/{requestId}` |
| `get_recommendations` | `GET` | `/api/v1/requests/{requestId}/recommendations` |
| `create_assignment` | `POST` | `/api/v1/requests/{requestId}/assignment` |
| `get_assignment` | `GET` | `/api/v1/requests/{requestId}/assignment` |
| `complete_assignment` | `PATCH` | `/api/v1/assignments/{assignmentId}/status` |
| `cancel_assignment` | `PATCH` | `/api/v1/assignments/{assignmentId}/status` |

## 8. Configuración

### Variables de entorno

| Variable | Requerida | Predeterminado | Descripción |
| --- | --- | --- | --- |
| `PORT` | No | `3002` | Puerto HTTP del MCP Server. Debe indicarse sin `:`. |
| `LOGISTIC_SERVICE_URL` | Sí | — | URL base de Logistics, sin necesidad de `/` final. |
| `LOGISTIC_SERVICE_TOKEN` | No | Vacío | Bearer token enviado por el MCP a Logistics. |
| `MCP_API_KEY` | Sí | — | Token Bearer requerido para acceder a `/mcp`. |

Ejemplo local:

```bash
export PORT=3002
export LOGISTIC_SERVICE_URL=http://localhost:3001
export MCP_API_KEY=entropy-local-secret
export LOGISTIC_SERVICE_TOKEN=
```

En Docker Compose se debe usar el nombre DNS del servicio, no `localhost`:

```env
LOGISTIC_SERVICE_URL=http://logistic-service:3001
```

## 9. Ejecución local

Requisitos:

- Go 1.27 o una versión compatible con el `go.mod`;
- `logistic-service` disponible;
- `fleet-service` disponible cuando Logistics calcule recomendaciones o modifique equipos.

```bash
go mod download
go run ./cmd/services
```

Verificar salud:

```bash
curl -s http://localhost:3002/health | jq
```

### Probar MCP manualmente

Descubrir las herramientas disponibles:

```bash
curl -s -X POST http://localhost:3002/mcp \
  -H "Authorization: Bearer entropy-local-secret" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2026-07-28" \
  -d '{
    "jsonrpc": "2.0",
    "id": "tools-list-1",
    "method": "tools/list",
    "params": {
      "_meta": {
        "io.modelcontextprotocol/protocolVersion": "2026-07-28",
        "io.modelcontextprotocol/clientInfo": {
          "name": "manual-test",
          "version": "1.0.0"
        },
        "io.modelcontextprotocol/clientCapabilities": {}
      }
    }
  }' | jq
```

Invocar una herramienta de consulta:

```bash
REQUEST_ID="06ab32ff-9b61-4b8d-aaa9-4fd72fa99658"

jq -n \
  --arg request_id "$REQUEST_ID" \
  '{
    jsonrpc: "2.0",
    id: "get-request-1",
    method: "tools/call",
    params: {
      name: "get_logistics_request",
      arguments: {requestId: $request_id},
      _meta: {
        "io.modelcontextprotocol/protocolVersion": "2026-07-28",
        "io.modelcontextprotocol/clientInfo": {
          name: "manual-test",
          version: "1.0.0"
        },
        "io.modelcontextprotocol/clientCapabilities": {}
      }
    }
  }' |
curl -s -X POST http://localhost:3002/mcp \
  -H "Authorization: Bearer entropy-local-secret" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2026-07-28" \
  --data-binary @- |
jq
```

## 10. Ejecución con Docker

Construir:

```bash
docker build -t entropy-mcp-server:local .
```

Ejecutar contra un Logistics accesible desde el host:

```bash
docker run --rm \
  --name entropy-mcp-server \
  -p 3002:3002 \
  -e PORT=3002 \
  -e LOGISTIC_SERVICE_URL=http://host.docker.internal:3001 \
  -e MCP_API_KEY=entropy-local-secret \
  entropy-mcp-server:local
```

En Linux puede ser necesario agregar:

```bash
--add-host=host.docker.internal:host-gateway
```

### Docker Compose

```yaml
services:
  entropy-mcp-server:
    build:
      context: ./entropy-mcp-server
    container_name: entropy-mcp-server
    environment:
      PORT: "3002"
      LOGISTIC_SERVICE_URL: http://logistic-service:3001
      LOGISTIC_SERVICE_TOKEN: ${LOGISTIC_SERVICE_TOKEN:-}
      MCP_API_KEY: ${MCP_API_KEY}
    ports:
      - "3002:3002"
    depends_on:
      - logistic-service
    restart: unless-stopped
```

`depends_on` controla el orden de inicio, pero no garantiza que Logistics esté listo. Para un entorno más estable se recomienda health check y reintentos con backoff.

## 11. Configuración en n8n

En el nodo **MCP Client Tool**:

| Campo | Valor |
| --- | --- |
| Endpoint | `http://entropy-mcp-server:3002/mcp` |
| Server Transport | `HTTP Streamable` |
| Authentication | Bearer Auth |
| Token | Mismo valor configurado en `MCP_API_KEY` |
| Tools to Include | Las necesarias para el agente |

Separación recomendada:

- **Consultas:** `get_logistics_request`, `get_recommendations`, `get_assignment`.
- **Creación:** `create_logistics_request`.
- **Operaciones controladas:** `create_assignment`, `complete_assignment`, `cancel_assignment`.

Las herramientas controladas deben pasar por revisión humana. El agente solo debe enviar `confirmed=true` después de recibir una confirmación explícita del usuario.

Si se utiliza Chat Trigger junto con nodos de respuesta, configure **Response Mode** como **Using Response Nodes**.

## 12. Manejo de errores

Los errores se dividen en tres grupos:

1. **Autenticación MCP:** `/mcp` responde `401 Unauthorized` cuando falta el Bearer token o no coincide con `MCP_API_KEY`.
2. **Validación de herramienta:** UUID inválido, razón vacía o falta de confirmación producen un resultado MCP con `isError: true`.
3. **Error del backend:** respuestas HTTP no exitosas de Logistics se convierten en `APIError` y luego en un resultado MCP fallido.

Ejemplo conceptual de error MCP:

```json
{
  "content": [
    {
      "type": "text",
      "text": "create_assignment requires explicit human confirmation"
    }
  ],
  "isError": true
}
```

El uso de `toolFailure` devuelve el fallo como resultado válido del protocolo. Esto permite que el agente explique el problema sin provocar necesariamente un error de transporte JSON-RPC.

## 13. Seguridad

La implementación actual incluye:

- Bearer token obligatorio en `/mcp`;
- comparación del token en tiempo constante;
- límite de cuerpo de solicitud de 1 MiB;
- timeouts del servidor HTTP;
- timeout de 10 segundos para Logistics;
- validación de UUID;
- confirmación humana para asignar, completar y cancelar;
- propagación de cancelación a llamadas HTTP internas;
- imagen final distroless y binario estático.

Para despliegues fuera de la demo:

- almacenar tokens en Secret Manager;
- usar TLS y evitar publicar `/mcp` directamente sin un control perimetral;
- reemplazar el token compartido por autenticación de servicio e identidad de usuario;
- aplicar autorización por herramienta y rol;
- rotar secretos y separar credenciales DEV/PROD;
- evitar registrar payloads que contengan tokens o datos sensibles;
- agregar rate limiting y auditoría de operaciones destructivas.

## 14. Pruebas y validación

Validación básica del proyecto:

```bash
go test ./...
go vet ./...
go build ./cmd/services
```

Pruebas mínimas de integración:

1. `/health` responde `200`.
2. `/mcp` sin token responde `401`.
3. el cliente descubre las siete herramientas.
4. se crea y consulta una solicitud.
5. se obtienen recomendaciones sin modificar equipos.
6. `create_assignment` con `confirmed=false` es rechazado.
7. `create_assignment` con confirmación reserva el equipo.
8. completar o cancelar libera el equipo.
9. UUID inexistentes y errores de Logistics aparecen como `isError: true`.

## 15. Comportamiento operativo esperado

| Acción | Request | Assignment | Equipment |
| --- | --- | --- | --- |
| Crear solicitud | `PENDING` | — | Sin cambio |
| Consultar recomendaciones | `PENDING` | — | Sin cambio |
| Crear asignación | `ASSIGNED` | Estado inicial definido por Logistics | `RESERVED` |
| Completar asignación | `COMPLETED` | `COMPLETED` | `AVAILABLE` |
| Cancelar asignación | Estado definido por Logistics | `CANCELLED` | `AVAILABLE` |

La consistencia de estas transiciones es responsabilidad de `logistic-service` y `fleet-service`; el MCP solamente invoca la operación correspondiente.

## 16. Limitaciones actuales del MVP

- autenticación mediante un único token compartido;
- no existe autorización por usuario, rol o herramienta;
- `/health` no comprueba dependencias;
- no hay reintentos, circuit breaker ni idempotency key;
- no incluye métricas ni trazas distribuidas propias;
- no hay pruebas unitarias dentro del proyecto adjunto;
- el MCP depende del contrato REST actual de Logistics;
- la geocodificación se realiza fuera del MCP, por ejemplo en un subflujo de n8n;
- la disponibilidad por rango de fechas depende de la implementación de Logistics; el MCP no debe inferir por qué un ranking viene vacío;
- no existe una herramienta para listar solicitudes o asignaciones.

## 17. Mejoras recomendadas

Prioridad inmediata:

1. agregar pruebas unitarias para validaciones, confirmación y cliente HTTP;
2. agregar readiness que compruebe el acceso a Logistics;
3. instrumentar logs estructurados, métricas y OpenTelemetry;
4. implementar reintentos únicamente para operaciones seguras e idempotentes;
5. agregar un identificador de correlación propagado entre n8n, MCP y Logistics;
6. incorporar autenticación por identidad y RBAC;
7. agregar idempotencia en creación de solicitudes y asignaciones;
8. publicar contratos y versionar cambios incompatibles.

### Observación sobre errores HTTP

El proyecto contiene `APIErrorResponse`, pero `decodeAPIError` deserializa actualmente el cuerpo directamente sobre `APIError`. Para conservar el código enviado por Logistics (`{"error":"...","message":"..."}`), conviene deserializar primero `APIErrorResponse` y mapear `Error` hacia `APIError.Code`.

## 18. Principio de diseño

El modelo de lenguaje interpreta la intención; el MCP limita las capacidades expuestas; los microservicios aplican reglas deterministas y mantienen el estado operacional.

Esto evita que el agente:

- consulte directamente la base de datos;
- invente reglas de asignación;
- reserve maquinaria sin confirmación;
- modifique estados fuera de los endpoints autorizados.
