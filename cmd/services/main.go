package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tariq-ventura/entropy-mcp-server/internal/client"
	"github.com/tariq-ventura/entropy-mcp-server/internal/config"
	"github.com/tariq-ventura/entropy-mcp-server/internal/database"
	"github.com/tariq-ventura/entropy-mcp-server/internal/embeddings"
	"github.com/tariq-ventura/entropy-mcp-server/internal/httpapi"
	"github.com/tariq-ventura/entropy-mcp-server/internal/integration"
	"github.com/tariq-ventura/entropy-mcp-server/internal/projection"
	"github.com/tariq-ventura/entropy-mcp-server/internal/tools"
)

func bearerAuthentication(
	expectedToken string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		authorization := strings.TrimSpace(
			request.Header.Get("Authorization"),
		)

		providedToken := strings.TrimSpace(
			strings.TrimPrefix(authorization, "Bearer "),
		)

		validLength := len(providedToken) == len(expectedToken)

		validToken := subtle.ConstantTimeCompare(
			[]byte(providedToken),
			[]byte(expectedToken),
		) == 1

		if !validLength || !validToken {
			response.Header().Set(
				"Content-Type",
				"application/json",
			)

			response.WriteHeader(http.StatusUnauthorized)

			_ = json.NewEncoder(response).Encode(
				map[string]string{
					"error":   "unauthorized",
					"message": "Bearer token inválido",
				},
			)

			return
		}

		next.ServeHTTP(response, request)
	})
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	logisticsClient := client.NewClient(
		"logistic-service",
		cfg.LogisticsServiceURL,
		cfg.LogisticsServiceToken,
	)
	fleetClient := client.NewClient(
		"fleet-service",
		cfg.FleetServiceURL,
		cfg.FleetServiceToken,
	)
	embeddingClient, err := embeddings.New(context.Background(), embeddings.Config{
		Provider:       cfg.EmbeddingProvider,
		Dimensions:     cfg.EmbeddingDimensions,
		OllamaURL:      cfg.OllamaURL,
		OllamaModel:    cfg.OllamaEmbeddingModel,
		GCPProjectID:   cfg.GCPProjectID,
		VertexLocation: cfg.VertexLocation,
		VertexModel:    cfg.EmbeddingModel,
	})
	if err != nil {
		slog.Error("embedding setup failed", "error", err)
		os.Exit(1)
	}
	databaseContext, cancelDatabase := context.WithTimeout(context.Background(), 30*time.Second)
	mcpDatabase, err := database.Open(databaseContext, database.Config{
		ConnectionString: cfg.MCPDatabaseString,
		MaxOpenConns:     cfg.DBMaxOpenConns,
		MaxIdleConns:     cfg.DBMaxIdleConns,
		ConnMaxLifetime:  cfg.DBConnMaxLifetime,
	})
	cancelDatabase()
	if err != nil {
		slog.Error("MCP PostgreSQL setup failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mcpDatabase.Close(); err != nil {
			slog.Error("MCP PostgreSQL shutdown failed", "error", err)
		}
	}()
	projectionRepository, err := projection.NewRepository(mcpDatabase.GORM())
	if err != nil {
		slog.Error("unified projection repository setup failed", "error", err)
		os.Exit(1)
	}
	migrationContext, cancelMigration := context.WithTimeout(context.Background(), 30*time.Second)
	if err := projectionRepository.Migrate(migrationContext); err != nil {
		cancelMigration()
		slog.Error("unified projection migration failed", "error", err)
		os.Exit(1)
	}
	cancelMigration()
	integrationService := integration.New(logisticsClient, fleetClient, projectionRepository)
	backgroundContext, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()
	go integrationService.Run(backgroundContext, cfg.SyncInterval)

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "entropy-platform",
			Version: "v0.1.0",
		},
		nil,
	)

	tools.New(logisticsClient, fleetClient, embeddingClient, integrationService).Register(mcpServer)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server {
			return mcpServer
		},
		&mcp.StreamableHTTPOptions{
			Stateless:                    true,
			JSONResponse:                 true,
			MaxRequestBodyBytes:          1 << 20,
			PropagateRequestCancellation: true,
			Logger:                       slog.Default(),
		},
	)

	mux := http.NewServeMux()

	mux.Handle(
		"/mcp",
		bearerAuthentication(
			cfg.MCPAPIKey,
			mcpHandler,
		),
	)
	mux.Handle("/api/v1/", httpapi.New(integrationService, cfg.IntegrationAPIKey, cfg.UIAllowedOrigin))

	mux.HandleFunc("/health", func(
		response http.ResponseWriter,
		_ *http.Request,
	) {
		response.Header().Set(
			"Content-Type",
			"application/json",
		)

		response.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(response).Encode(map[string]string{
			"status":  "ok",
			"service": "entropy-mcp-server",
		})
	})
	mux.HandleFunc("/ready", func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		response.Header().Set("Content-Type", "application/json")
		readinessContext, cancelReadiness := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancelReadiness()

		if err := mcpDatabase.Ping(readinessContext); err != nil {
			response.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(response).Encode(map[string]string{
				"status":   "not_ready",
				"service":  "entropy-mcp-server",
				"database": "unavailable",
			})
			return
		}
		state, err := projectionRepository.GetSyncState(readinessContext)
		if err != nil {
			response.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(response).Encode(map[string]string{
				"status":   "not_ready",
				"service":  "entropy-mcp-server",
				"database": "unavailable",
			})
			return
		}

		response.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(response).Encode(map[string]string{
			"status":     "ready",
			"service":    "entropy-mcp-server",
			"database":   "ready",
			"syncStatus": state.Status,
		})
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		slog.Info(
			"starting MCP server",
			"address",
			server.Addr,
			"mcp_endpoint",
			"/mcp",
			"database",
			"postgresql",
		)

		serverError <- server.ListenAndServe()
	}()

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-shutdownContext.Done():
		slog.Info("shutdown signal received")

	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}
	cancelBackground()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("MCP server stopped")
}
