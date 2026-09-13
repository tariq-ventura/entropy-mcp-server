package embeddings_vertex

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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const cloudScope = "https://www.googleapis.com/auth/cloud-platform"

type Client struct {
	projectID   string
	location    string
	model       string
	dimensions  int
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
}

type embeddingRequest struct {
	Instances  []embeddingInstance `json:"instances"`
	Parameters embeddingParameters `json:"parameters"`
}

type embeddingInstance struct {
	Content  string `json:"content"`
	TaskType string `json:"task_type"`
}

type embeddingParameters struct {
	AutoTruncate         bool `json:"autoTruncate"`
	OutputDimensionality int  `json:"outputDimensionality"`
}

type embeddingResponse struct {
	Predictions []struct {
		Embeddings struct {
			Values []float32 `json:"values"`
		} `json:"embeddings"`
	} `json:"predictions"`
}

func NewClient(ctx context.Context, projectID, location, model string, dimensions int) (*Client, error) {
	projectID = strings.TrimSpace(projectID)
	location = strings.TrimSpace(location)
	model = strings.TrimSpace(model)
	if projectID == "" {
		return nil, errors.New("GCP_PROJECT_ID is required when EMBEDDING_PROVIDER=vertex")
	}
	if location == "" {
		return nil, errors.New("VERTEX_LOCATION is required when EMBEDDING_PROVIDER=vertex")
	}
	if model == "" {
		return nil, errors.New("EMBEDDING_MODEL is required when EMBEDDING_PROVIDER=vertex")
	}
	if dimensions <= 0 {
		return nil, errors.New("EMBEDDING_DIMENSIONS must be greater than zero")
	}
	tokenSource, err := google.DefaultTokenSource(ctx, cloudScope)
	if err != nil {
		return nil, fmt.Errorf("creating Google token source: %w", err)
	}
	return &Client{projectID: projectID, location: location, model: model, dimensions: dimensions, tokenSource: oauth2.ReuseTokenSource(nil, tokenSource), httpClient: &http.Client{Timeout: 45 * time.Second}}, nil
}

func (c *Client) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	vectors, err := c.embed(ctx, []string{text}, "RETRIEVAL_QUERY")
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (c *Client) EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error) {
	return c.embed(ctx, documents, "RETRIEVAL_DOCUMENT")
}

func (c *Client) embed(ctx context.Context, contents []string, taskType string) ([][]float32, error) {
	if len(contents) == 0 {
		return [][]float32{}, nil
	}
	instances := make([]embeddingInstance, len(contents))
	for index, content := range contents {
		content = strings.TrimSpace(content)
		if content == "" {
			return nil, errors.New("embedding content is empty")
		}
		instances[index] = embeddingInstance{Content: content, TaskType: taskType}
	}
	payload, err := json.Marshal(embeddingRequest{Instances: instances, Parameters: embeddingParameters{AutoTruncate: true, OutputDimensionality: c.dimensions}})
	if err != nil {
		return nil, fmt.Errorf("encoding Vertex AI request: %w", err)
	}
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict", c.location, c.projectID, c.location, c.model)
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("getting Google access token: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating Vertex AI request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("calling Vertex AI: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 8*1024))
		return nil, fmt.Errorf("Vertex AI returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	var result embeddingResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding Vertex AI response: %w", err)
	}
	if len(result.Predictions) != len(contents) {
		return nil, fmt.Errorf("Vertex AI returned %d embeddings for %d inputs", len(result.Predictions), len(contents))
	}
	vectors := make([][]float32, len(result.Predictions))
	for index, prediction := range result.Predictions {
		if len(prediction.Embeddings.Values) != c.dimensions {
			return nil, fmt.Errorf("invalid Vertex AI embedding dimensions: expected %d, received %d", c.dimensions, len(prediction.Embeddings.Values))
		}
		vectors[index] = prediction.Embeddings.Values
	}
	return vectors, nil
}

func (c *Client) Provider() string { return "vertex" }
func (c *Client) Model() string    { return c.model }
func (c *Client) Dimensions() int  { return c.dimensions }
