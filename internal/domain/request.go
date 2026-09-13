package domain

type CreateRequestInput struct {
	Project   string `json:"project" jsonschema:"Proyecto de Prisma asociado a la solicitud"`
	Type      string `json:"type" jsonschema:"Clase de maquinaria requerida, por ejemplo Cargador frontal"`
	Requester string `json:"requester" jsonschema:"Nombre del gerente de proyecto que solicita la maquinaria"`
	StartDate string `json:"startDate" jsonschema:"Inicio del período en formato RFC3339"`
	EndDate   string `json:"endDate" jsonschema:"Fin del período en formato RFC3339"`
}
type Request struct {
	ID        string  `json:"id"`
	Project   string  `json:"project"`
	Type      string  `json:"type"`
	Requester string  `json:"requester"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Status    string  `json:"status"`
	Machinery *string `json:"machinery,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}
type SearchRequestsInput struct {
	Query     string   `json:"query,omitempty"`
	Statuses  []string `json:"statuses,omitempty"`
	Type      string   `json:"type,omitempty"`
	Requester string   `json:"requester,omitempty"`
	Page      int      `json:"page,omitempty"`
	PageSize  int      `json:"pageSize,omitempty"`
}
type SearchRequestsResult struct {
	Count    int64     `json:"count"`
	Requests []Request `json:"requests"`
}
