package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/brutski/go-code-graph/internal/analyzer"
	"github.com/brutski/go-code-graph/internal/embeddings"
	"github.com/brutski/go-code-graph/internal/mcpserver"
)

const (
	// Provider constants
	providerBedrock = "bedrock"

	// Default parallelism level for embeddings
	defaultParallelismLevel = 5
)

// createLogger creates a structured logger based on LOG_LEVEL environment variable
func createLogger() *slog.Logger {
	level := getEnvOrDefault("LOG_LEVEL", "info")

	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // Default to info
	}

	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	})
	return slog.New(textHandler)
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEmbeddingsParallelismLevel returns the parallelism level from environment variable
func getEmbeddingsParallelismLevel(logger *slog.Logger) int {
	if envVal := os.Getenv("EMBEDDINGS_PARALLELISM_LEVEL"); envVal != "" {
		if level, err := strconv.Atoi(envVal); err == nil && level > 0 {
			logger.Debug("Using embeddings parallelism level from environment", "level", level)
			return level
		} else {
			logger.Warn("Invalid EMBEDDINGS_PARALLELISM_LEVEL, using default",
				"value", envVal, "default", defaultParallelismLevel)
		}
	}
	return defaultParallelismLevel
}

// createEmbeddingClient creates an embedding client from environment variables
func createEmbeddingClient(ctx context.Context, logger *slog.Logger) embeddings.Client {
	provider := os.Getenv("EMBEDDING_PROVIDER")
	model := os.Getenv("EMBEDDING_MODEL")

	// Auto-detect provider if not specified
	if provider == "" {
		if os.Getenv("AWS_REGION") != "" || os.Getenv("AWS_PROFILE") != "" {
			provider = providerBedrock
		}
	}

	if provider == "" {
		logger.Info("No embedding provider configured")
		return nil
	}

	// Create configuration
	config := embeddings.Config{
		Provider: provider,
		Model:    model,
		Options:  make(map[string]interface{}),
	}

	// Set provider-specific defaults and options
	switch provider {
	case providerBedrock:
		if model == "" {
			config.Model = "amazon.titan-embed-text-v2:0"
		}
		if region := os.Getenv("AWS_REGION"); region != "" {
			config.Options["region"] = region
		}
	case "openai":
		if model == "" {
			config.Model = "text-embedding-ada-002"
		}
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			config.Options["apiKey"] = apiKey
		}
	case "mock":
		if model == "" {
			config.Model = "mock-model"
		}
		config.Options["dimensions"] = 384
	}

	// Create client
	client, err := embeddings.NewClient(ctx, config)
	if err != nil {
		logger.Warn("Failed to initialize embedding client, continuing without embeddings",
			"provider", provider, "error", err)
		return nil
	}

	logger.Info("Connected to embedding service successfully",
		"provider", provider, "model", client.GetModelName())

	return client
}

func main() {
	// Create structured logger
	logger := createLogger()

	// Parse configuration from environment variables only
	neo4jURI := getEnvOrDefault("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := getEnvOrDefault("NEO4J_USER", "neo4j")
	neo4jPassword := getEnvOrDefault("NEO4J_PASSWORD", "password")

	ctx := context.Background()

	// Create Neo4j client
	neo4jDriver, err := neo4j.NewDriverWithContext(
		neo4jURI,
		neo4j.BasicAuth(neo4jUser, neo4jPassword, ""),
	)
	if err != nil {
		logger.Error("Failed to create Neo4j driver", "error", err)
		os.Exit(1)
	}
	defer neo4jDriver.Close(ctx)

	// Test Neo4j connection
	if err := neo4jDriver.VerifyConnectivity(ctx); err != nil {
		logger.Error("Failed to connect to Neo4j", "error", err)
		os.Exit(1)
	}

	logger.Info("Connected to Neo4j successfully", "uri", neo4jURI)

	// Create embedding client
	embeddingClient := createEmbeddingClient(ctx, logger)

	// Get embeddings parallelism level
	embeddingsParallelism := getEmbeddingsParallelismLevel(logger)
	logger.Info("Embeddings configuration",
		"enabled", embeddingClient != nil,
		"parallelismLevel", embeddingsParallelism)

	embeddingsGenerator := analyzer.NewEmbeddingsGeneratorWithConfig(embeddingClient, embeddingsParallelism, logger)

	server, err := mcpserver.NewServer(neo4jDriver, embeddingsGenerator, logger)
	if err != nil {
		logger.Error("Failed to create MCP server", "error", err)
		os.Exit(1)
	}
	defer server.Close()

	logger.Info("Starting Code Graph MCP Server", "neo4j_uri", neo4jURI, "transport", "stdio")

	// Run the server over stdin/stdout until the client disconnects
	if err := server.Run(ctx, mcp.NewStdioTransport()); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
