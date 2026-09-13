package embeddings_ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbedDocumentsUsesBatchEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/embed" || request.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		var input embeddingRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if input.Model != "test-model" || input.Dimensions != 2 || len(input.Input) != 2 {
			t.Fatalf("unexpected payload: %#v", input)
		}
		if input.Input[0] != "search_document: primer equipo" {
			t.Fatalf("missing document prefix: %q", input.Input[0])
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"embeddings": [][]float32{{1, 0}, {0, 1}}})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-model", 2)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	vectors, err := client.EmbedDocuments(context.Background(), []string{"primer equipo", "segundo equipo"})
	if err != nil {
		t.Fatalf("embedding documents: %v", err)
	}
	if len(vectors) != 2 || len(vectors[0]) != 2 {
		t.Fatalf("unexpected vectors: %#v", vectors)
	}
}
