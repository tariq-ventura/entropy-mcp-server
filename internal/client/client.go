package client

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

type APIErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf(
		"logistic-service returned HTTP %d: %s: %s",
		e.StatusCode,
		e.Code,
		e.Message,
	)
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
