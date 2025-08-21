package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockClient implements EmbeddingClient for AWS Bedrock Titan embeddings
type BedrockClient struct {
	client *bedrockruntime.Client
	model  string
}

// TitanEmbeddingRequest represents the request format for Titan embeddings
type TitanEmbeddingRequest struct {
	InputText string `json:"inputText"`
}

// TitanEmbeddingResponse represents the response format from Titan embeddings
type TitanEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewBedrockClient creates a new Bedrock embedding client with the specified model
func NewBedrockClient(ctx context.Context, model string) (*BedrockClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	// Use default model if none specified
	if model == "" {
		model = "amazon.titan-embed-text-v2:0"
	}

	return &BedrockClient{
		client: client,
		model:  model,
	}, nil
}

// CreateEmbedding generates an embedding for the given text using Titan V2
func (b *BedrockClient) CreateEmbedding(text string) ([]float32, error) {
	// Prepare the request payload
	request := TitanEmbeddingRequest{
		InputText: text,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke Bedrock model
	result, err := b.client.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.model),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	// Parse the response
	var response TitanEmbeddingResponse
	if err := json.NewDecoder(bytes.NewReader(result.Body)).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding received from Bedrock")
	}

	return response.Embedding, nil
}

// BatchCreateEmbeddings generates embeddings for multiple texts efficiently
func (b *BedrockClient) BatchCreateEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	// Process in batches to avoid overwhelming the API
	batchSize := 10 // Reasonable batch size for Bedrock

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		// Process batch
		for j := i; j < end; j++ {
			embedding, err := b.CreateEmbedding(texts[j])
			if err != nil {
				return nil, fmt.Errorf("failed to create embedding for text %d: %w", j, err)
			}
			embeddings[j] = embedding
		}
	}

	return embeddings, nil
}

// GetModelName returns the model name/identifier being used
func (b *BedrockClient) GetModelName() string {
	return b.model
}

// GetDimensions returns the embedding dimensions for this model
func (b *BedrockClient) GetDimensions() int {
	return 1024 // Titan V2 produces 1024-dimensional embeddings
}

// GetModelInfo returns information about the embedding model
func (b *BedrockClient) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"model_name":       b.model,
		"model_provider":   "amazon",
		"embedding_size":   1024, // Titan V2 produces 1024-dimensional embeddings
		"max_input_tokens": 8192, // Titan V2 supports up to 8192 tokens
	}
}
