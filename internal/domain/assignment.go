package domain

type CreateAssignmentInput struct {
	EquipmentID string `json:"equipmentId"`
	Reason      string `json:"reason"`
}

type UpdateAssignmentStatusInput struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type Assignment struct {
	ID           string  `json:"id"`
	RequestID    string  `json:"requestId"`
	EquipmentID  string  `json:"equipmentId"`
	Status       string  `json:"status"`
	StatusReason string  `json:"statusReason"`
	AssignedAt   string  `json:"assignedAt"`
	CompletedAt  *string `json:"completedAt,omitempty"`
	CancelledAt  *string `json:"cancelledAt,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}
