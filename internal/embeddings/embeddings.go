package embeddings

import (
	"context"
	"errors"
	"strings"

	embeddings_ollama "github.com/tariq-ventura/entropy-mcp-server/internal/embeddings/ollama"
	embeddings_vertex "github.com/tariq-ventura/entropy-mcp-server/internal/embeddings/vertex"
)

type Client interface {
	EmbedQuery(context.Context, string) ([]float32, error)
	EmbedDocuments(context.Context, []string) ([][]float32, error)
	Provider() string
	Model() string
	Dimensions() int
}

type Config struct {
	Provider       string
	Dimensions     int
	OllamaURL      string
	OllamaModel    string
	GCPProjectID   string
	VertexLocation string
	VertexModel    string
}

func New(ctx context.Context, cfg Config) (Client, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "ollama":
		return embeddings_ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel, cfg.Dimensions)
	case "vertex":
		return embeddings_vertex.NewClient(ctx, cfg.GCPProjectID, cfg.VertexLocation, cfg.VertexModel, cfg.Dimensions)
	default:
		return nil, errors.New("EMBEDDING_PROVIDER must be vertex or ollama")
	}
}
