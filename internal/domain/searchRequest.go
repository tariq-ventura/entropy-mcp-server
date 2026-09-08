package domain

type SearchRequestsInput struct {
	Query         string   `json:"query,omitempty" jsonschema:"Búsqueda literal por proyecto, ubicación o tipo"`
	SemanticQuery string   `json:"semanticQuery,omitempty" jsonschema:"Descripción en lenguaje natural de las solicitudes buscadas"`
	Statuses      []string `json:"statuses,omitempty" jsonschema:"Estados permitidos"`
	EquipmentType string   `json:"equipmentType,omitempty" jsonschema:"Tipo exacto de maquinaria"`

	Near *NearLocation `json:"near,omitempty"`

	MinSemanticScore *float64 `json:"minSemanticScore,omitempty" jsonschema:"Similitud mínima entre 0 y 1"`

	Page     int `json:"page,omitempty"`
	PageSize int `json:"pageSize,omitempty"`
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
