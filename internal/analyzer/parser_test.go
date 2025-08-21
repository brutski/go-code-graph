package analyzer

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/brutski/go-code-graph/internal/embeddings"
)

// createTestModule creates a temporary directory with a go.mod file
func createTestModule(t *testing.T) string {
	testDir := t.TempDir()

	// Create go.mod file
	err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(`module testmodule

go 1.21
`), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	return testDir
}

func TestParserAnalyzeCodebase(t *testing.T) {
	// Create test directory structure
	testDir := createTestModule(t)

	// Create test Go files
	mainFile := filepath.Join(testDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`
package main

import "fmt"

type Config struct {
	Name string
	Port int
}

func main() {
	fmt.Println("Hello, World!")
}

func NewConfig() *Config {
	return &Config{
		Name: "test",
		Port: 8080,
	}
}
`), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create parser with mock embeddings
	config := embeddings.DefaultMockConfig()
	client, err := embeddings.NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	logger := slog.Default()
	embeddingsGen := NewEmbeddingsGenerator(client, logger)
	parser := NewParser(embeddingsGen, logger)

	// Analyze the test codebase
	graph, err := parser.AnalyzeCodebase(testDir)
	if err != nil {
		t.Fatalf("AnalyzeCodebase failed: %v", err)
	}

	// Verify graph structure
	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Check nodes were created
	if len(graph.Nodes) == 0 {
		t.Error("Expected at least one node in the graph")
	}

	// Verify specific nodes exist
	expectedNodes := map[string]string{
		"package:testmodule":            NodeTypePackage,
		"struct:testmodule.Config":      NodeTypeStruct,
		"function:testmodule.main":      NodeTypeFunction,
		"function:testmodule.NewConfig": NodeTypeFunction,
		"field:testmodule.Config.Name":  NodeTypeField,
		"field:testmodule.Config.Port":  NodeTypeField,
	}

	nodeMap := make(map[string]*Node)
	for i := range graph.Nodes {
		nodeMap[graph.Nodes[i].ID] = &graph.Nodes[i]
	}

	for id, expectedType := range expectedNodes {
		node, exists := nodeMap[id]
		if !exists {
			t.Errorf("Expected node %s to exist", id)
			continue
		}
		if node.Type != expectedType {
			t.Errorf("Node %s: expected type %s, got %s", id, expectedType, node.Type)
		}
	}

	// Check edges
	if len(graph.Edges) == 0 {
		t.Error("Expected at least one edge in the graph")
	}

	// Verify constructor pattern detection
	hasConstructsEdge := false
	for _, edge := range graph.Edges {
		if edge.Type == "constructs" && edge.Source == "function:testmodule.NewConfig" && edge.Target == "struct:testmodule.Config" {
			hasConstructsEdge = true
			break
		}
	}
	if !hasConstructsEdge {
		t.Error("Expected constructs edge from NewConfig to Config")
	}
}

func TestParserWithInvalidPath(t *testing.T) {
	parser := NewParser(nil, nil)

	// Test with non-existent directory
	_, err := parser.AnalyzeCodebase("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestParserWithEmptyDirectory(t *testing.T) {
	testDir := t.TempDir()
	parser := NewParser(nil, nil)

	graph, err := parser.AnalyzeCodebase(testDir)
	if err != nil {
		t.Fatalf("AnalyzeCodebase failed: %v", err)
	}

	// Parser always creates 22 builtin type nodes (bool, string, int, error, etc.)
	// even for empty directories
	expectedBuiltinNodes := 22
	if len(graph.Nodes) != expectedBuiltinNodes {
		t.Errorf("Expected %d builtin nodes for empty directory, got %d", expectedBuiltinNodes, len(graph.Nodes))
	}
}

func TestParserWithSyntaxError(t *testing.T) {
	testDir := createTestModule(t)

	// Create invalid Go file
	invalidFile := filepath.Join(testDir, "invalid.go")
	err := os.WriteFile(invalidFile, []byte(`
package main

func main() {
	// Missing closing brace
`), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser(nil, nil)

	// Should handle syntax errors gracefully
	graph, err := parser.AnalyzeCodebase(testDir)
	if err != nil {
		// It's ok if it returns an error
		t.Logf("Got expected error: %v", err)
		return
	}

	// Or it might skip the invalid file
	if graph != nil && len(graph.Nodes) > 0 {
		t.Log("Parser handled syntax error by skipping invalid file")
	}
}

func TestParserInterfaceImplementation(t *testing.T) {
	testDir := createTestModule(t)

	// Create test files with interface and implementation
	interfaceFile := filepath.Join(testDir, "interface.go")
	err := os.WriteFile(interfaceFile, []byte(`
package main

type Writer interface {
	Write([]byte) (int, error)
}

type FileWriter struct {
	path string
}

func (fw *FileWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
`), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser(nil, nil)
	graph, err := parser.AnalyzeCodebase(testDir)
	if err != nil {
		t.Fatalf("AnalyzeCodebase failed: %v", err)
	}

	// Check for implements relationship
	hasImplementsEdge := false
	for _, edge := range graph.Edges {
		if edge.Type == "implements" && edge.Source == "struct:testmodule.FileWriter" && edge.Target == "interface:testmodule.Writer" {
			hasImplementsEdge = true
			break
		}
	}
	if !hasImplementsEdge {
		t.Error("Expected implements edge from FileWriter to Writer")
	}
}

func TestParserConcurrencyPatterns(t *testing.T) {
	testDir := createTestModule(t)

	// Create test file with goroutines and channels
	concurrentFile := filepath.Join(testDir, "concurrent.go")
	err := os.WriteFile(concurrentFile, []byte(`
package main

func worker(jobs <-chan int, results chan<- int) {
	for j := range jobs {
		results <- j * 2
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	go worker(jobs, results)
	
	jobs <- 1
	<-results
}
`), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser(nil, nil)
	graph, err := parser.AnalyzeCodebase(testDir)
	if err != nil {
		t.Fatalf("AnalyzeCodebase failed: %v", err)
	}

	// Check for goroutine spawn edge
	hasSpawnsEdge := false
	for _, edge := range graph.Edges {
		if edge.Type == EdgeTypeSpawnsGoroutine && edge.Source == "function:testmodule.main" && edge.Target == "function:testmodule.worker" {
			hasSpawnsEdge = true
			break
		}
	}
	if !hasSpawnsEdge {
		t.Error("Expected spawns edge from main to worker")
	}

	// Channel communication edges have been removed in simplification
}
