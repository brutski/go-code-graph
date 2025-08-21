package embeddings

import (
	"context"
	"os"
	"testing"
)

func TestConfigCreation(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name: "bedrock_config_from_env",
			envVars: map[string]string{
				"EMBEDDING_PROVIDER": "bedrock",
				"EMBEDDING_MODEL":    "amazon.titan-embed-text-v2:0",
				"AWS_REGION":         "us-east-1",
			},
			expected: Config{
				Provider: "bedrock",
				Model:    "amazon.titan-embed-text-v2:0",
				Options: map[string]interface{}{
					"region": "us-east-1",
				},
			},
		},
		{
			name: "openai_config_from_env",
			envVars: map[string]string{
				"EMBEDDING_PROVIDER": "openai",
				"EMBEDDING_MODEL":    "text-embedding-ada-002",
				"OPENAI_API_KEY":     "sk-test-key",
			},
			expected: Config{
				Provider: "openai",
				Model:    "text-embedding-ada-002",
				Options: map[string]interface{}{
					"apiKey": "sk-test-key",
				},
			},
		},
		{
			name: "mock_config_from_env",
			envVars: map[string]string{
				"EMBEDDING_PROVIDER": "mock",
				"EMBEDDING_MODEL":    "mock-model",
			},
			expected: Config{
				Provider: "mock",
				Model:    "mock-model",
				Options: map[string]interface{}{
					"dimensions": 384,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Create config based on environment
			config := createConfigFromEnv()

			// Verify provider and model
			if config.Provider != tt.expected.Provider {
				t.Errorf("Expected provider %s, got %s", tt.expected.Provider, config.Provider)
			}

			if config.Model != tt.expected.Model {
				t.Errorf("Expected model %s, got %s", tt.expected.Model, config.Model)
			}

			// Verify options
			for key, expectedValue := range tt.expected.Options {
				actualValue, exists := config.Options[key]
				if !exists {
					t.Errorf("Expected option %s to exist", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Expected option %s to be %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestDefaultConfigs(t *testing.T) {
	tests := []struct {
		name     string
		configFn func() Config
		expected Config
	}{
		{
			name:     "default_bedrock",
			configFn: DefaultBedrockConfig,
			expected: Config{
				Provider: "bedrock",
				Model:    "amazon.titan-embed-text-v2:0",
				Options: map[string]interface{}{
					"region": "us-east-1",
				},
			},
		},
		{
			name:     "default_openai",
			configFn: DefaultOpenAIConfig,
			expected: Config{
				Provider: "openai",
				Model:    "text-embedding-ada-002",
				Options:  map[string]interface{}{},
			},
		},
		{
			name:     "default_mock",
			configFn: DefaultMockConfig,
			expected: Config{
				Provider: "mock",
				Model:    "mock-model",
				Options: map[string]interface{}{
					"dimensions": 384,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.configFn()

			if config.Provider != tt.expected.Provider {
				t.Errorf("Expected provider %s, got %s", tt.expected.Provider, config.Provider)
			}

			if config.Model != tt.expected.Model {
				t.Errorf("Expected model %s, got %s", tt.expected.Model, config.Model)
			}

			// Verify options
			for key, expectedValue := range tt.expected.Options {
				actualValue, exists := config.Options[key]
				if !exists {
					t.Errorf("Expected option %s to exist", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Expected option %s to be %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestClientCreation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		config        Config
		expectError   bool
		expectedModel string
	}{
		{
			name: "valid_mock_client",
			config: Config{
				Provider: "mock",
				Model:    "mock-model",
				Options: map[string]interface{}{
					"dimensions": 384,
				},
			},
			expectError:   false,
			expectedModel: "mock-model",
		},
		{
			name: "invalid_provider",
			config: Config{
				Provider: "invalid",
				Model:    "test-model",
				Options:  map[string]interface{}{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ctx, tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("Expected client but got nil")
				return
			}

			if client.GetModelName() != tt.expectedModel {
				t.Errorf("Expected model name %s, got %s", tt.expectedModel, client.GetModelName())
			}
		})
	}
}

func TestEnvironmentVariableDetection(t *testing.T) {
	tests := []struct {
		name             string
		envVars          map[string]string
		expectedProvider string
	}{
		{
			name: "bedrock_auto_detection",
			envVars: map[string]string{
				"AWS_REGION": "us-west-2",
			},
			expectedProvider: "bedrock",
		},
		{
			name: "bedrock_with_profile",
			envVars: map[string]string{
				"AWS_PROFILE": "my-profile",
			},
			expectedProvider: "bedrock",
		},
		{
			name: "explicit_provider_overrides_detection",
			envVars: map[string]string{
				"EMBEDDING_PROVIDER": "openai",
				"AWS_REGION":         "us-east-1",
			},
			expectedProvider: "openai",
		},
		{
			name:             "no_provider_detected",
			envVars:          map[string]string{},
			expectedProvider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars first
			envVarsToClean := []string{
				"EMBEDDING_PROVIDER", "AWS_REGION", "AWS_PROFILE", "OPENAI_API_KEY",
			}
			for _, key := range envVarsToClean {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			provider := detectProviderFromEnv()
			if provider != tt.expectedProvider {
				t.Errorf("Expected provider %s, got %s", tt.expectedProvider, provider)
			}
		})
	}
}

// Helper functions to simulate the logic from main.go

func createConfigFromEnv() Config {
	provider := os.Getenv("EMBEDDING_PROVIDER")
	model := os.Getenv("EMBEDDING_MODEL")

	// Auto-detect provider if not specified
	if provider == "" {
		provider = detectProviderFromEnv()
	}

	if provider == "" {
		return Config{} // Empty config
	}

	// Create configuration
	config := Config{
		Provider: provider,
		Model:    model,
		Options:  make(map[string]interface{}),
	}

	// Set provider-specific defaults and options
	switch provider {
	case "bedrock":
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

	return config
}

func detectProviderFromEnv() string {
	// Check for explicit provider first
	if provider := os.Getenv("EMBEDDING_PROVIDER"); provider != "" {
		return provider
	}

	// Auto-detect based on available credentials
	if os.Getenv("AWS_REGION") != "" || os.Getenv("AWS_PROFILE") != "" {
		return "bedrock"
	}
	return ""
}
