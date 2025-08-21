package analyzer

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/brutski/go-code-graph/internal/embeddings"
)

// TestFunctionVariableCalls tests that function variables don't cause panics
func TestFunctionVariableCalls(t *testing.T) {
	// Create parser
	config := embeddings.DefaultMockConfig()
	client, err := embeddings.NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	logger := slog.Default()
	embeddingsGen := NewEmbeddingsGenerator(client, logger)
	parser := NewParser(embeddingsGen, logger)

	// Analyze the function_vars test file
	testdataPath := filepath.Join("testdata", "edges", "advanced")
	graph, err := parser.AnalyzeCodebase(testdataPath)
	if err != nil {
		t.Fatalf("Failed to analyze testdata: %v", err)
	}

	// Verify we found the expected functions
	expectedFunctions := []string{
		"regularFunc",
		"TestFunctionVariables",
		"TestMethodValues",
	}

	foundFunctions := make(map[string]bool)
	for _, node := range graph.Nodes {
		if node.Type == NodeTypeFunction {
			foundFunctions[node.Label] = true
		}
	}

	for _, expected := range expectedFunctions {
		if !foundFunctions[expected] {
			t.Errorf("Expected to find function %s, but it was not found", expected)
		}
	}

	// Verify we found the Calculator type and its method
	foundCalculator := false
	foundAddMethod := false
	for _, node := range graph.Nodes {
		if node.Type == NodeTypeStruct && node.Label == "Calculator" {
			foundCalculator = true
		}
		if node.Type == NodeTypeMethod && node.Label == "Add" {
			foundAddMethod = true
		}
	}

	if !foundCalculator {
		t.Error("Expected to find Calculator struct")
	}
	if !foundAddMethod {
		t.Error("Expected to find Add method")
	}

	// The key test is that we didn't panic - if we got here, the fix worked
	t.Log("Successfully analyzed code with function variables without panicking")
}

// TestFunctionVariableEdges tests edge detection with function variables
func TestFunctionVariableEdges(t *testing.T) {
	// Create parser
	config := embeddings.DefaultMockConfig()
	client, err := embeddings.NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	logger := slog.Default()
	embeddingsGen := NewEmbeddingsGenerator(client, logger)
	parser := NewParser(embeddingsGen, logger)

	// Analyze the function_vars test file
	testdataPath := filepath.Join("testdata", "edges", "advanced")
	graph, err := parser.AnalyzeCodebase(testdataPath)
	if err != nil {
		t.Fatalf("Failed to analyze testdata: %v", err)
	}

	// Count call edges - we should still detect calls to regularFunc
	callCount := 0
	for _, edge := range graph.Edges {
		if edge.Type == "calls" {
			callCount++
			t.Logf("Found call edge: %s -> %s", edge.Source, edge.Target)
		}
	}

	// We should have at least some call edges (calls to println, etc)
	if callCount == 0 {
		t.Error("Expected to find at least some call edges")
	}

	t.Logf("Found %d call edges", callCount)
}
