package embeddings

// Config represents the configuration for embedding providers
type Config struct {
	Provider string                 `json:"provider"` // "bedrock", "openai", "mock"
	Model    string                 `json:"model"`    // Model identifier
	Options  map[string]interface{} `json:"options"`  // Provider-specific options
}

// DefaultBedrockConfig returns a default configuration for AWS Bedrock
func DefaultBedrockConfig() Config {
	return Config{
		Provider: "bedrock",
		Model:    "amazon.titan-embed-text-v2:0",
		Options: map[string]interface{}{
			"region": "us-east-1",
		},
	}
}

// DefaultOpenAIConfig returns a default configuration for OpenAI
func DefaultOpenAIConfig() Config {
	return Config{
		Provider: "openai",
		Model:    "text-embedding-ada-002",
		Options:  map[string]interface{}{},
	}
}

// DefaultMockConfig returns a default configuration for Mock (testing)
func DefaultMockConfig() Config {
	return Config{
		Provider: "mock",
		Model:    "mock-model",
		Options: map[string]interface{}{
			"dimensions": 384, // Smaller dimensions for testing
		},
	}
}
