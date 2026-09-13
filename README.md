# Entropy MCP Server

Capa de integración entre Prisma (`logistic-service`), Startrack (`fleet-service`), la UI y el flujo conversacional de n8n. Ninguno de los dos servicios fuente se comunica con el otro. El MCP reconcilia ambos en una proyección PostgreSQL canónica, expone esa misma verdad por herramientas MCP y por una API Gin de lectura, y coordina las escrituras distribuidas.

## Arquitectura

```mermaid
flowchart TB
    N["n8n / agente"] -->|MCP tools| M["Entropy integration service"]
    U["UI"] -->|Gin REST| M
    M --> C[("PostgreSQL canónico MCP")]
    M -->|REST| P["Logistic / Prisma"]
    M -->|REST| S["Fleet / Startrack"]
```

## Verdad unificada

La correlación usa el código inicial del número de activo Prisma y la descripción Startrack:

```text
Prisma:    CF-03 - Cargador frontal 03
Startrack: CF-03
Clave MCP: CF03
```

| Dato unificado | Sistema autoritativo |
| --- | --- |
| Disponibilidad y estado operativo | Prisma `status` |
| Nombre, empresa, número de activo y clase | Prisma |
| Estado de rastreo, marca, modelo, año, grupo, etiquetas, conductor e ID remoto | Startrack |
| Tareas, geocercas y mantenimiento telemático | Startrack |
| Identidad canónica, equivalencias de tipos, asignación unificada y conflictos | MCP |

El MCP no interpreta `Normal` de Startrack como `Disponible`. Si los tipos difieren, falta un registro o los estados de una asignación no concuerdan, publica el conflicto en `integration_sync_conflicts` y conserva el último snapshot completo. Nunca reemplaza un snapshot válido con datos parciales cuando uno de los servicios fuente falla.

La reconciliación se ejecuta al iniciar, cada `SYNC_INTERVAL`, bajo demanda con `POST /api/v1/sync` y antes de las lecturas críticas del agente. Equipos, tipos, solicitudes, asignaciones, mantenimiento y conflictos se actualizan en una sola transacción.

## Base de datos propia del MCP

El MCP requiere su propia base PostgreSQL. No usa ni comparte la base de Prisma ni la de Startrack: obtiene los datos por las APIs de `logistic-service` y `fleet-service`, los normaliza y persiste el snapshot canónico que consumen tanto la UI como las herramientas MCP.

Al iniciar, `internal/database/database.go` abre el pool y verifica PostgreSQL. Después, `internal/projection/repository.go` ejecuta `AutoMigrate` sobre los modelos de `internal/projection/models.go`. Si la conexión o la migración falla, el proceso no arranca.

| Tabla MCP | Contenido persistido |
| --- | --- |
| `integration_equipment_types` | Catálogo canónico de tipos |
| `integration_equipment_type_aliases` | Equivalencias de nombres Prisma y Startrack |
| `integration_unified_equipments` | Equipo normalizado y vínculo entre IDs fuente |
| `integration_unified_requests` | Solicitudes provenientes de Prisma |
| `integration_unified_maintenance` | Mantenimiento proveniente de Startrack, ligado a la clave canónica |
| `integration_unified_assignments` | Asignación reconciliada entre solicitud y tarea |
| `integration_sync_conflicts` | Valores sin equivalente o inconsistentes |
| `integration_sync_state` | Estado, error, fecha y conteos del último snapshot |

La actualización de un snapshot se hace dentro de una transacción: primero se insertan o actualizan todos los registros con un nuevo `sync_token`, luego se eliminan los registros ausentes y finalmente se marca la sincronización como `READY`. Así la UI y el agente no leen una combinación parcial de dos sincronizaciones.

## API Gin para la UI

La UI consulta esta API para mostrar exactamente los datos disponibles al agente. Las mutaciones conversacionales siguen pasando por n8n y las herramientas MCP, donde se aplican confirmación humana y compensación.

| Método | Ruta | Resultado |
| --- | --- | --- |
| GET | `/api/v1/dashboard` | Totales y fecha de última sincronización |
| GET | `/api/v1/equipment-types` | Tipos canónicos reconciliados |
| GET | `/api/v1/equipments` | Equipos unificados y paginados |
| GET | `/api/v1/equipments/:key` | Equipo con historial de mantenimiento |
| GET | `/api/v1/requests` | Solicitudes Prisma ordenadas |
| GET | `/api/v1/assignments` | Estado unificado Prisma + Startrack |
| GET | `/api/v1/conflicts` | Inconsistencias detectadas |
| GET | `/api/v1/sync-status` | Estado y conteos del reconciliador |
| POST | `/api/v1/sync` | Fuerza una reconciliación completa |

Los listados aceptan `page` y `pageSize`. Equipos también aceptan `type`, `status`, `search`, `onlyAvailable` y `onlyLinked`; solicitudes aceptan `search`, `status`, `type` y `requester`; asignaciones aceptan `status` y `search`. Todas las rutas `/api/v1` requieren `Authorization: Bearer $INTEGRATION_API_KEY` (si no se configura, usa `MCP_API_KEY`).

## Herramientas MCP

| Herramienta | Uso |
| --- | --- |
| `create_logistics_request` | Crear una solicitud Pendiente en Prisma |
| `get_logistics_request` | Consultar una solicitud Prisma |
| `search_logistics_requests` | Buscar por texto, estado, tipo o solicitante |
| `list_unified_equipment` | Listar el inventario combinado Prisma + Startrack |
| `get_unified_equipment` | Consultar un equipo por clave o ID |
| `get_recommendations` | Rankear equipos elegibles con embeddings, mantenimiento y preparación operativa |
| `create_assignment` | Crear tarea Startrack y asignar/aprobar en Prisma |
| `get_assignment` | Consultar la asignación reconciliada en la proyección canónica |
| `complete_assignment` | Completar tarea y liberar disponibilidad en Prisma |
| `cancel_assignment` | Cancelar tarea y revertir la solicitud Prisma |
| `create_geofence` | Crear geocerca Startrack con coordenadas decimales |
| `list_geofences` | Consultar geocercas con coordenadas decimales |

`create_assignment`, `complete_assignment` y `cancel_assignment` exigen `confirmed: true` después de confirmación humana.

## Algoritmo híbrido de recomendaciones

`get_recommendations` conserva reglas estrictas antes de calcular cualquier puntaje:

1. La solicitud debe estar `Pendiente`.
2. La maquinaria debe existir en Prisma y estar correlacionada con un vehículo Startrack.
3. Prisma debe indicar `Disponible`.
4. La clase Prisma debe coincidir con el tipo solicitado.

Después, el MCP consulta todo el historial de mantenimiento de Fleet, genera un documento unificado por candidato y calcula:

```text
score total = relevancia semántica * 0.50
            + mantenimiento       * 0.30
            + preparación         * 0.20
```

| Componente | Fuente | Criterios principales |
| --- | --- | --- |
| Relevancia semántica | Embeddings | Similitud coseno entre la solicitud Prisma y el documento Prisma + Startrack + últimos mantenimientos |
| Mantenimiento | Fleet | Antigüedad del último servicio, preventivo/correctivo, emergencias y correctivos recientes |
| Preparación operativa | Fleet | Estado de rastreo `Normal`/`Activo` y conductor registrado |

El puntaje de mantenimiento es una heurística explicable, no una predicción de falla. Sin lectura actual del horómetro ni un intervalo de servicio configurado, el MCP no afirma cuántas horas faltan para el próximo mantenimiento. Cada resultado incluye `score`, `maintenance`, `reasons` y `warnings`.

No se utiliza distancia porque el diccionario no proporciona la posición GPS actual del vehículo. Las coordenadas de tareas y geocercas representan destinos o zonas, no una ubicación vehicular autoritativa.

Ejemplo reducido de respuesta:

```json
{
  "requestId": "UUID-DE-SOLICITUD",
  "algorithmVersion": "hybrid-v1",
  "embeddingProvider": "vertex",
  "embeddingModel": "text-multilingual-embedding-002",
  "count": 1,
  "recommendations": [
    {
      "rank": 1,
      "score": {
        "total": 92,
        "semantic": 93,
        "maintenance": 85,
        "operational": 100
      },
      "equipment": {
        "equipmentKey": "CF03",
        "available": true
      },
      "maintenance": {
        "recordCount": 2,
        "latestServiceDate": "2026-08-20T00:00:00Z",
        "latestServiceType": "Preventivo"
      },
      "reasons": ["Disponible según Prisma"]
    }
  ]
}
```

## Flujo recomendado para n8n

1. Recoger proyecto, tipo, solicitante y período.
2. Llamar `create_logistics_request`.
3. Llamar `get_recommendations` con el `requestId`.
4. Mostrar opciones al usuario y pedir confirmación.
5. Resolver origen/destino con Google Maps si hace falta.
6. Llamar `create_assignment` con `confirmed: true`.

Ejemplo de creación de solicitud:

```json
{
  "project": "PROY-014 - The Hub - Proyecto Xi - La Unión",
  "type": "Cargador frontal",
  "requester": "María José López Ramírez",
  "startDate": "2026-09-11",
  "endDate": "2026-09-14"
}
```

Ejemplo de asignación unificada:

```json
{
  "requestId": "UUID-DE-SOLICITUD",
  "equipmentKey": "CF03",
  "taskId": "TAR-014",
  "title": "Traslado de cargador frontal CF-03",
  "description": "Traslado de cargador frontal",
  "taskType": "Traslado",
  "scheduledDate": "2026-09-12",
  "origin": "PLANTA ORIGEN",
  "destination": "PROY-014 - The Hub - Proyecto Xi - La Unión",
  "latitude": 13.3376152,
  "longitude": -87.8486967,
  "assignee": "MOT-014 - Adriana Steiner",
  "reason": "Asignación confirmada por el usuario",
  "confirmed": true
}
```

El MCP crea primero la tarea Startrack. Si la asignación transaccional de Prisma falla, elimina la tarea como compensación. La descripción de la tarea guarda el UUID de la solicitud y el activo Prisma; el reconciliador usa ambos marcadores para materializar una única asignación y detectar divergencias posteriores.

## Configuración

```env
PORT=3002
MCP_API_KEY=replace-me
INTEGRATION_API_KEY=replace-me
MCP_DB_USER=entropy
MCP_DB_PASSWORD=entropy
MCP_DB_NAME=entropy_mcp
MCP_DB_PORT=5436
MCP_DB_STRING=postgres://entropy:entropy@localhost:5436/entropy_mcp?sslmode=disable
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=30m
SYNC_INTERVAL=2m
UI_ALLOWED_ORIGIN=http://localhost:3000
LOGISTIC_SERVICE_URL=http://localhost:3001
LOGISTIC_SERVICE_TOKEN=
FLEET_SERVICE_URL=http://localhost:3000
FLEET_SERVICE_TOKEN=
EMBEDDING_PROVIDER=ollama
EMBEDDING_DIMENSIONS=768
OLLAMA_URL=http://localhost:11434
OLLAMA_EMBEDDING_MODEL=nomic-embed-text
```

Para Cloud Run se usa Vertex AI:

```env
EMBEDDING_PROVIDER=vertex
EMBEDDING_DIMENSIONS=768
GCP_PROJECT_ID=minecraft-server-501415
VERTEX_LOCATION=us-east4
EMBEDDING_MODEL=text-multilingual-embedding-002
```

La cuenta de servicio del MCP necesita permiso para ejecutar predicciones de Vertex AI. En desarrollo local con Ollama, descargue primero el modelo configurado:

```bash
ollama pull nomic-embed-text
```

Levantar PostgreSQL y el MCP juntos:

```bash
cp .env.example .env
docker compose up --build
```

El volumen `mcp_postgres_data` conserva la verdad unificada aunque se reinicien los contenedores. Fleet y Logistic deben estar disponibles en los puertos indicados en `.env`; dentro de Compose se accede a ellos mediante `host.docker.internal`.

Para ejecutar Go directamente y levantar únicamente la base:

```bash
docker compose up -d mcp-postgres
go run ./cmd/services
```

Comprobar la base y las tablas creadas:

```bash
curl http://localhost:3002/ready
docker compose exec mcp-postgres psql -U entropy -d entropy_mcp -c '\\dt integration_*'
```

`/health` confirma que el proceso HTTP está vivo. `/ready` devuelve `503` si PostgreSQL deja de responder y, cuando está disponible, informa también `syncStatus` (`NOT_SYNCED`, `RUNNING`, `READY` o `FAILED`). Un fallo temporal de Fleet o Logistic no elimina el último snapshot válido almacenado.

Endpoints del proceso:

| Método | Ruta | Autenticación |
| --- | --- | --- |
| GET | `/health` | Sin autenticación |
| GET | `/ready` | Sin autenticación; valida PostgreSQL |
| POST | `/mcp` | `Authorization: Bearer $MCP_API_KEY` |
| GET/POST | `/api/v1/*` | `Authorization: Bearer $INTEGRATION_API_KEY` |

Los manifiestos de Cloud Run incluyen las URL de ambos servicios, configuran el pool y leen `MCP_DB_STRING` desde Secret Manager. La base debe ser PostgreSQL y accesible mediante la instancia Cloud SQL configurada. La sonda de arranque consulta `/ready`, por lo que una revisión no recibe tráfico antes de completar la conexión y las migraciones. Si se habilita autenticación de servicio a servicio, inyecte los tokens en `LOGISTIC_SERVICE_TOKEN` y `FLEET_SERVICE_TOKEN`.
