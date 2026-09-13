package domain

type Task struct {
	ID            string `json:"id"`
	TaskID        string `json:"taskId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	ScheduledDate string `json:"scheduledDate"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	Latitude      int64  `json:"latitude"`
	Longitude     int64  `json:"longitude"`
	Assignee      string `json:"assignee"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}
type CreateTaskRequest struct {
	TaskID        string `json:"taskId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	ScheduledDate string `json:"scheduledDate"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	Latitude      int64  `json:"latitude"`
	Longitude     int64  `json:"longitude"`
	Assignee      string `json:"assignee"`
}
