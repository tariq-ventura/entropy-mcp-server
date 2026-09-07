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
		cfg.LogisticsServiceURL,
		cfg.LogisticsServiceToken,
	)

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "entropy-platform",
			Version: "v0.1.0",
		},
		nil,
	)

	tools.New(logisticsClient).Register(mcpServer)

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

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
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
