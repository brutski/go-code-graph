package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	// OpenAI model constants
	modelTextEmbeddingAda002 = "text-embedding-ada-002"
)

// OpenAIClient implements EmbeddingClient for OpenAI embeddings
type OpenAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// OpenAIEmbeddingRequest represents the request format for OpenAI embeddings
type OpenAIEmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// OpenAIEmbeddingResponse represents the response format from OpenAI embeddings
type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIClient creates a new OpenAI embedding client
func NewOpenAIClient(ctx context.Context, config Config) (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Try to get from config options
		if key, ok := config.Options["apiKey"].(string); ok {
			apiKey = key
		} else {
			return nil, fmt.Errorf("OpenAI API key not found in environment or config")
		}
	}

	model := config.Model
	if model == "" {
		model = modelTextEmbeddingAda002 // Default model
	}

	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CreateEmbedding generates an embedding for the given text using OpenAI
func (o *OpenAIClient) CreateEmbedding(text string) ([]float32, error) {
	request := OpenAIEmbeddingRequest{
		Input: []string{text},
		Model: o.model,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status code: %d", resp.StatusCode)
	}

	var response OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response from OpenAI")
	}

	return response.Data[0].Embedding, nil
}

// BatchCreateEmbeddings generates embeddings for multiple texts efficiently using OpenAI
func (o *OpenAIClient) BatchCreateEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// OpenAI supports up to 2048 inputs in a single request, but we'll use smaller batches
	// to avoid timeout issues and make progress visible for large datasets
	batchSize := 100
	embeddings := make([][]float32, len(texts))

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		batchEmbeddings, err := o.processBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to process batch starting at index %d: %w", i, err)
		}

		// Copy batch results to main array
		copy(embeddings[i:end], batchEmbeddings)
	}

	return embeddings, nil
}

// processBatch handles a single batch of texts for OpenAI API
func (o *OpenAIClient) processBatch(texts []string) ([][]float32, error) {
	request := OpenAIEmbeddingRequest{
		Input: texts,
		Model: o.model,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status code: %d", resp.StatusCode)
	}

	var response OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(response.Data))
	}

	// Extract embeddings in the correct order (OpenAI returns them with index)
	embeddings := make([][]float32, len(texts))
	for _, item := range response.Data {
		if item.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid index %d in response", item.Index)
		}
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// GetModelName returns the model name/identifier being used
func (o *OpenAIClient) GetModelName() string {
	return o.model
}

// GetDimensions returns the embedding dimensions for this model
func (o *OpenAIClient) GetDimensions() int {
	// Common OpenAI embedding dimensions
	switch o.model {
	case modelTextEmbeddingAda002:
		return 1536
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	default:
		return 1536 // Default assumption
	}
}
