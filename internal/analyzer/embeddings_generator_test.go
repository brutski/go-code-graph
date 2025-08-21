package analyzer

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/brutski/go-code-graph/internal/embeddings"
)

func TestEmbeddingsGeneratorHashIntegration(t *testing.T) {
	// Create a mock embedding client
	config := embeddings.DefaultMockConfig()
	client, err := embeddings.NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Create embeddings generator
	generator := NewEmbeddingsGenerator(client, slog.Default())

	// Create a test node
	node := &Node{
		ID:         "test:function:main.TestFunc",
		Label:      "TestFunc",
		Type:       NodeTypeFunction,
		Package:    "main",
		FullName:   "main.TestFunc",
		Size:       10,
		Complexity: 3,
		Signature:  "func TestFunc() error",
		TypeInfo: TypeInfo{
			Kind: "function",
		},
		Visibility: VisibilityPublic,
		Position: Position{
			Filename: "test.go",
			Line:     10,
		},
		CreatedAt: time.Now(),
	}

	// Prepare node for embedding (should add semantic summary and hash metadata)
	generator.PrepareNodeForEmbedding(node)

	// Verify semantic summary was created
	if node.SemanticSummary == "" {
		t.Error("SemanticSummary should be generated")
	}

	// Verify hash metadata was added
	if node.Metadata == nil {
		t.Fatal("Metadata should be initialized")
	}

	if _, ok := node.Metadata["content_hash"].(string); !ok {
		t.Error("content_hash should be in metadata")
	}

	if _, ok := node.Metadata["semantic_hash"].(string); !ok {
		t.Error("semantic_hash should be in metadata")
	}

	// Save the semantic hash for comparison
	originalSemanticHash := node.Metadata["semantic_hash"].(string)
	originalSemanticSummary := node.SemanticSummary
	t.Logf("Original semantic summary: %q", originalSemanticSummary)
	t.Logf("Original semantic hash: %s", originalSemanticHash)

	// Test NodeNeedsEmbedding - should return true for new node
	if !generator.NodeNeedsEmbedding(node, nil) {
		t.Error("NodeNeedsEmbedding should return true for node without existing metadata")
	}

	// Simulate existing metadata with same hash
	existingMetadata := map[string]interface{}{
		"semantic_hash": originalSemanticHash,
		"has_embedding": true,
	}

	if generator.NodeNeedsEmbedding(node, existingMetadata) {
		t.Error("NodeNeedsEmbedding should return false when semantic hash hasn't changed")
	}

	// Simulate code change by modifying node properties
	node.Complexity = 15      // Change complexity to >10 to be included in semantic summary
	node.SemanticSummary = "" // Clear to force regeneration
	generator.PrepareNodeForEmbedding(node)

	newSemanticSummary := node.SemanticSummary
	newSemanticHash := node.Metadata["semantic_hash"].(string)
	t.Logf("New semantic summary: %q", newSemanticSummary)
	t.Logf("New semantic hash: %s", newSemanticHash)

	if newSemanticSummary == originalSemanticSummary {
		t.Error("Semantic summary should be regenerated when complexity changes")
	}

	if newSemanticHash == originalSemanticHash {
		t.Error("Semantic hash should change when semantic summary changes")
	}

	// Now it should need embedding update
	if !generator.NodeNeedsEmbedding(node, existingMetadata) {
		t.Error("NodeNeedsEmbedding should return true when semantic hash changed")
	}

	// Generate embeddings
	err = generator.GenerateEmbeddingsForNodes([]*Node{node})
	if err != nil {
		t.Fatalf("Failed to generate embeddings: %v", err)
	}

	// Verify embedding was generated
	if len(node.Embedding) == 0 {
		t.Error("Embedding should be generated")
	}

	// Verify has_embedding is now true
	if hasEmbedding, ok := node.Metadata["has_embedding"].(bool); !ok || !hasEmbedding {
		t.Error("has_embedding should be true after embedding generation")
	}
}

func TestEmbeddingsGeneratorGetNodeMetadata(t *testing.T) {
	generator := NewEmbeddingsGenerator(nil, slog.Default())

	// Test with node that has metadata
	node := &Node{
		Metadata: map[string]interface{}{
			"content_hash":           "abc123",
			"semantic_hash":          "def456",
			"has_embedding":          true,
			"last_analyzed":          int64(1234567890),
			"embedding_generated_at": int64(1234567890),
			"other_field":            "should not be included",
		},
	}

	metadata := generator.GetNodeMetadata(node)

	// Verify only relevant fields are returned
	if len(metadata) != 5 {
		t.Errorf("Expected 5 metadata fields, got %d", len(metadata))
	}

	if metadata["content_hash"] != "abc123" {
		t.Error("content_hash should be preserved")
	}

	if metadata["semantic_hash"] != "def456" {
		t.Error("semantic_hash should be preserved")
	}

	if _, ok := metadata["other_field"]; ok {
		t.Error("other_field should not be included in metadata")
	}
}
