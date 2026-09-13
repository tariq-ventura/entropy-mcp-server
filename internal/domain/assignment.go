package domain

type StatusTransition struct {
	ID         string `json:"id"`
	RequestID  string `json:"requestId"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	Reason     string `json:"reason"`
	ChangedAt  string `json:"changedAt"`
}
type PrismaAssignment struct {
	Request    Request          `json:"request"`
	Equipment  Machinery        `json:"equipment"`
	Transition StatusTransition `json:"transition"`
}
type UnifiedAssignment struct {
	Request     Request          `json:"request"`
	Equipment   UnifiedEquipment `json:"equipment"`
	Task        Task             `json:"task"`
	Compensated bool             `json:"compensated"`
}
type UnifiedAssignmentStatus struct {
	Request   Request          `json:"request"`
	Equipment UnifiedEquipment `json:"equipment"`
	Task      Task             `json:"task"`
}
type RecommendationScore struct {
	Total       float64 `json:"total"`
	Semantic    float64 `json:"semantic"`
	Maintenance float64 `json:"maintenance"`
	Operational float64 `json:"operational"`
}
type Recommendation struct {
	Rank        int                 `json:"rank"`
	Score       RecommendationScore `json:"score"`
	Equipment   UnifiedEquipment    `json:"equipment"`
	Maintenance MaintenanceSummary  `json:"maintenance"`
	Reasons     []string            `json:"reasons"`
	Warnings    []string            `json:"warnings,omitempty"`
}
type Recommendations struct {
	RequestID         string           `json:"requestId"`
	AlgorithmVersion  string           `json:"algorithmVersion"`
	EmbeddingProvider string           `json:"embeddingProvider"`
	EmbeddingModel    string           `json:"embeddingModel"`
	Count             int              `json:"count"`
	Recommendations   []Recommendation `json:"recommendations"`
}
