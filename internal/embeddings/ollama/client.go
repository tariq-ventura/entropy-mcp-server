package embeddings_ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	model      string
	dimensions int
	httpClient *http.Client
}

type embeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Truncate   bool     `json:"truncate"`
	Dimensions int      `json:"dimensions"`
	KeepAlive  string   `json:"keep_alive"`
}

type embeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func NewClient(baseURL, model string, dimensions int) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	model = strings.TrimSpace(model)
	if baseURL == "" {
		return nil, errors.New("OLLAMA_URL is required when EMBEDDING_PROVIDER=ollama")
	}
	if model == "" {
		return nil, errors.New("OLLAMA_EMBEDDING_MODEL is required when EMBEDDING_PROVIDER=ollama")
	}
	if dimensions <= 0 {
		return nil, errors.New("EMBEDDING_DIMENSIONS must be greater than zero")
	}
	return &Client{baseURL: baseURL, model: model, dimensions: dimensions, httpClient: &http.Client{Timeout: 45 * time.Second}}, nil
}

func (c *Client) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	vectors, err := c.embed(ctx, []string{"search_query: " + strings.TrimSpace(text)})
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (c *Client) EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error) {
	inputs := make([]string, len(documents))
	for index, document := range documents {
		inputs[index] = "search_document: " + strings.TrimSpace(document)
	}
	return c.embed(ctx, inputs)
}

func (c *Client) embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if len(inputs) == 0 {
		return [][]float32{}, nil
	}
	for _, input := range inputs {
		if strings.TrimSpace(input) == "" {
			return nil, errors.New("embedding input is empty")
		}
	}
	payload, err := json.Marshal(embeddingRequest{Model: c.model, Input: inputs, Truncate: true, Dimensions: c.dimensions, KeepAlive: "10m"})
	if err != nil {
		return nil, fmt.Errorf("encoding Ollama request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating Ollama request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("calling Ollama: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 8*1024))
		return nil, fmt.Errorf("Ollama returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	var result embeddingResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding Ollama response: %w", err)
	}
	if len(result.Embeddings) != len(inputs) {
		return nil, fmt.Errorf("Ollama returned %d embeddings for %d inputs", len(result.Embeddings), len(inputs))
	}
	for _, vector := range result.Embeddings {
		if len(vector) != c.dimensions {
			return nil, fmt.Errorf("invalid Ollama embedding dimensions: expected %d, received %d", c.dimensions, len(vector))
		}
	}
	return result.Embeddings, nil
}

func (c *Client) Provider() string { return "ollama" }
func (c *Client) Model() string    { return c.model }
func (c *Client) Dimensions() int  { return c.dimensions }
