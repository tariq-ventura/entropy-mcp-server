package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) do(
	ctx context.Context,
	method string,
	path string,
	input any,
	output any,
) error {
	var body io.Reader = http.NoBody

	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		body = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		method,
		c.baseURL+path,
		body,
	)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		request.Header.Set(
			"Authorization",
			"Bearer "+c.token,
		)
	}

	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("calling logistic-service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return decodeAPIError(response)
	}

	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf(
			"decoding logistic-service response: %w",
			err,
		)
	}

	return nil
}
