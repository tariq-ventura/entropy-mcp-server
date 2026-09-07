package domain

type APIResponse[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
}

type APIErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
