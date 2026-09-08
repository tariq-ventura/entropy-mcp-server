package domain

type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CreateRequestInput struct {
	EquipmentType string `json:"equipmentType" jsonschema:"Tipo de maquinaria, por ejemplo EXCAVATOR"`
	ProjectName   string `json:"projectName" jsonschema:"Nombre del proyecto"`

	Location Location `json:"location" jsonschema:"Ubicación donde se necesita la maquinaria"`

	Description  string `json:"description,omitempty" jsonschema:"Descripción del trabajo que se realizará"`
	Requirements string `json:"requirements,omitempty" jsonschema:"Requerimientos técnicos u operativos de la maquinaria"`

	StartDate string `json:"startDate" jsonschema:"Fecha inicial en formato RFC3339"`
	EndDate   string `json:"endDate" jsonschema:"Fecha final en formato RFC3339"`
}

type Request struct {
	ID            string `json:"id"`
	EquipmentType string `json:"equipmentType"`
	ProjectName   string `json:"projectName"`

	Location Location `json:"location"`

	// Compatibilidad con respuestas planas de logistic-service.
	LocationName string  `json:"locationName,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`

	Description  string `json:"description,omitempty"`
	Requirements string `json:"requirements,omitempty"`

	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
