package client

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	serviceName string
	baseURL     string
	token       string
	http        *http.Client
}
type APIError struct {
	ServiceName string
	StatusCode  int
	Code        string
	Message     string
}
type APIErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s returned HTTP %d: %s: %s", e.ServiceName, e.StatusCode, e.Code, e.Message)
}
func NewClient(serviceName, baseURL, token string) *Client {
	return &Client{serviceName: serviceName, baseURL: strings.TrimRight(baseURL, "/"), token: strings.TrimSpace(token), http: &http.Client{Timeout: 15 * time.Second}}
}
