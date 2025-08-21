package embeddings

import (
	"context"
	"testing"
)

func TestBatchEmbeddings(t *testing.T) {
	// Test with Mock client for deterministic behavior
	config := Config{
		Provider: "mock",
		Model:    "test-model",
		Options: map[string]interface{}{
			"dimensions": 384,
		},
	}

	client, err := NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test data
	texts := []string{
		"Function processOrder handles order processing logic",
		"Method validateInput checks user input for validity",
		"Struct UserData contains user information fields",
		"Interface PaymentProcessor defines payment operations",
		"Package main contains application entry point",
	}

	// Test batch processing
	embeddings, err := client.BatchCreateEmbeddings(texts)
	if err != nil {
		t.Fatalf("BatchCreateEmbeddings failed: %v", err)
	}

	// Verify results
	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}

	// Verify each embedding has correct dimensions
	expectedDims := client.GetDimensions()
	for i, embedding := range embeddings {
		if len(embedding) != expectedDims {
			t.Errorf("Embedding %d has %d dimensions, expected %d", i, len(embedding), expectedDims)
		}
	}

	// Test empty input
	emptyEmbeddings, err := client.BatchCreateEmbeddings([]string{})
	if err != nil {
		t.Errorf("BatchCreateEmbeddings failed on empty input: %v", err)
	}
	if len(emptyEmbeddings) != 0 {
		t.Errorf("Expected empty result for empty input, got %d embeddings", len(emptyEmbeddings))
	}

	// Test individual vs batch consistency (for Mock client)
	individual, err := client.CreateEmbedding(texts[0])
	if err != nil {
		t.Fatalf("CreateEmbedding failed: %v", err)
	}

	batch, err := client.BatchCreateEmbeddings(texts[:1])
	if err != nil {
		t.Fatalf("BatchCreateEmbeddings failed: %v", err)
	}

	// For Mock client, individual and batch should produce the same result
	if len(individual) != len(batch[0]) {
		t.Errorf("Individual and batch embeddings have different dimensions")
	}

	// Check if embeddings are the same (they should be for Mock client with deterministic behavior)
	for i := range individual {
		if individual[i] != batch[0][i] {
			t.Errorf("Individual and batch embeddings differ at index %d: %f vs %f", i, individual[i], batch[0][i])
		}
	}
}

func BenchmarkEmbeddingMethods(b *testing.B) {
	config := Config{
		Provider: "mock",
		Model:    "test-model",
		Options: map[string]interface{}{
			"dimensions": 384,
		},
	}

	client, err := NewClient(context.Background(), config)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	texts := []string{
		"Function processOrder handles order processing logic",
		"Method validateInput checks user input for validity",
		"Struct UserData contains user information fields",
		"Interface PaymentProcessor defines payment operations",
		"Package main contains application entry point",
		"Constructor NewService creates service instance",
		"Handler HandleRequest processes HTTP requests",
		"Validator CheckPermissions verifies user access",
		"Repository SaveData persists information",
		"Service ProcessTransaction handles payments",
	}

	b.Run("Individual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, text := range texts {
				_, err := client.CreateEmbedding(text)
				if err != nil {
					b.Fatalf("CreateEmbedding failed: %v", err)
				}
			}
		}
	})

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.BatchCreateEmbeddings(texts)
			if err != nil {
				b.Fatalf("BatchCreateEmbeddings failed: %v", err)
			}
		}
	})
}
