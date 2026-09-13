package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func decodeAPIError(serviceName string, response *http.Response) error {
	body, readError := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if readError != nil {
		return &APIError{ServiceName: serviceName, StatusCode: response.StatusCode, Code: "upstream_error", Message: "could not read " + serviceName + " response"}
	}
	var payload APIErrorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		payload.Error = "upstream_error"
		payload.Message = strings.TrimSpace(string(body))
	}
	if payload.Error == "" {
		payload.Error = "upstream_error"
	}
	if payload.Message == "" {
		payload.Message = http.StatusText(response.StatusCode)
	}
	return &APIError{ServiceName: serviceName, StatusCode: response.StatusCode, Code: payload.Error, Message: payload.Message}
}
