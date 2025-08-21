package embeddings

import (
	"context"
	"fmt"
)

// Client interface for generating embeddings
type Client interface {
	CreateEmbedding(text string) ([]float32, error)
	GetModelName() string // Returns the model name/identifier being used
	GetDimensions() int   // Returns the embedding dimensions for this model
}

// NewClient creates a new embedding client based on the provided configuration
func NewClient(ctx context.Context, config Config) (Client, error) {
	switch config.Provider {
	case "bedrock":
		return NewBedrockClientFromConfig(ctx, config)
	case "openai":
		return NewOpenAIClientFromConfig(ctx, config)
	case "mock":
		return NewMockClientFromConfig(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", config.Provider)
	}
}

// NewBedrockClientFromConfig creates a new Bedrock client with the given configuration
func NewBedrockClientFromConfig(ctx context.Context, config Config) (Client, error) {
	return NewBedrockClient(ctx, config.Model)
}

// NewOpenAIClientFromConfig creates a new OpenAI client with the given configuration
func NewOpenAIClientFromConfig(ctx context.Context, config Config) (Client, error) {
	return NewOpenAIClient(ctx, config)
}

// NewMockClientFromConfig creates a new Mock client with the given configuration
func NewMockClientFromConfig(ctx context.Context, config Config) (Client, error) {
	return NewMockClient(config), nil
}
