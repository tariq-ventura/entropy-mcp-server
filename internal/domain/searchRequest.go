package domain

type SearchRequestsInput struct {
	Query         string        `json:"query,omitempty" jsonschema:"Texto relacionado con el proyecto, ubicación o tipo de maquinaria"`
	Statuses      []string      `json:"statuses,omitempty" jsonschema:"Estados que deben incluirse"`
	EquipmentType string        `json:"equipmentType,omitempty" jsonschema:"Tipo de maquinaria"`
	Near          *NearLocation `json:"near,omitempty" jsonschema:"Ubicación y radio para la búsqueda geográfica"`
	Page          int           `json:"page,omitempty" jsonschema:"Página solicitada"`
	PageSize      int           `json:"pageSize,omitempty" jsonschema:"Resultados por página, máximo 100"`
}

type NearLocation struct {
	Latitude  float64 `json:"latitude" jsonschema:"Latitud resuelta de la ubicación"`
	Longitude float64 `json:"longitude" jsonschema:"Longitud resuelta de la ubicación"`
	RadiusKM  float64 `json:"radiusKm" jsonschema:"Radio de búsqueda en kilómetros"`
}

type SearchRequestItem struct {
	ID            string   `json:"id"`
	EquipmentType string   `json:"equipmentType"`
	ProjectName   string   `json:"projectName"`
	LocationName  string   `json:"locationName"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	StartDate     string   `json:"startDate"`
	EndDate       string   `json:"endDate"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
	DistanceKM    *float64 `json:"distanceKm,omitempty"`
}

type SearchRequestsResult struct {
	Count    int64               `json:"count"`
	Requests []SearchRequestItem `json:"requests"`
}
