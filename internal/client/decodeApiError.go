package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func decodeAPIError(response *http.Response) error {
	body, readError := io.ReadAll(
		io.LimitReader(response.Body, 64*1024),
	)
	if readError != nil {
		return &APIError{
			StatusCode: response.StatusCode,
			Code:       "upstream_error",
			Message:    "could not read logistic-service response",
		}
	}

	var apiError APIError

	if err := json.Unmarshal(body, &apiError); err != nil {
		apiError.Code = "upstream_error"
		apiError.Message = strings.TrimSpace(string(body))
	}

	if apiError.Code == "" {
		apiError.Code = "upstream_error"
	}

	if apiError.Message == "" {
		apiError.Message = http.StatusText(response.StatusCode)
	}

	return &APIError{
		StatusCode: response.StatusCode,
		Code:       apiError.Code,
		Message:    apiError.Message,
	}
}
