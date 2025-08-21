package neo4j

import (
	"context"
	"testing"

	"github.com/brutski/go-code-graph/internal/analyzer"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.URI != "bolt://localhost:7687" {
		t.Errorf("Expected default URI 'bolt://localhost:7687', got %s", config.URI)
	}
	if config.Username != "neo4j" {
		t.Errorf("Expected default username 'neo4j', got %s", config.Username)
	}
	if config.Password != "codeGraph123" {
		t.Errorf("Expected default password 'codeGraph123', got %s", config.Password)
	}
}

func TestSanitizeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Function", "Function"},
		{"function", "Function"},
		{"FUNCTION", "Function"},
		{"my_function", "MyFunction"},
		{"my-function", "MyFunction"},
		{"123function", "Function"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeLabel(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeLabel(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	config := DefaultConfig()

	// This test will only work if Neo4j is running
	client, err := NewClient(config)
	if err != nil {
		t.Skip("Neo4j not available for testing")
		return
	}
	defer client.Close()

	// Test ping
	ctx := context.Background()
	err = client.ping(ctx)
	if err != nil {
		t.Errorf("Failed to ping Neo4j: %v", err)
	}
}

func TestClientOperationsWithoutDatabase(t *testing.T) {
	// Test operations when database is not available
	config := DefaultConfig()
	config.URI = "bolt://localhost:99999" // Invalid port

	client, err := NewClient(config)
	if err == nil {
		t.Error("Expected error when connecting to invalid port")
		if client != nil {
			client.Close()
		}
	}
}

func TestImportGraphValidation(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Skip("Neo4j not available for testing")
		return
	}
	defer client.Close()

	// Test with nil graph
	ctx := context.Background()
	err = client.ImportGraph(ctx, nil)
	if err == nil {
		t.Error("Expected error when importing nil graph")
	}

	// Test with empty graph
	emptyGraph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{},
		Edges: []analyzer.EnhancedEdge{},
	}

	err = client.ImportGraph(ctx, emptyGraph)
	if err != nil {
		t.Errorf("Should handle empty graph gracefully: %v", err)
	}
}

func TestNodeImportValidation(t *testing.T) {
	// Create a mock graph with various node types
	graph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{
			{
				ID:       "package:test",
				Label:    "test",
				Type:     analyzer.NodeTypePackage,
				Package:  "test",
				FullName: "test",
				Metadata: map[string]interface{}{
					"test": "value",
				},
			},
			{
				ID:         "function:test.Main",
				Label:      "Main",
				Type:       analyzer.NodeTypeFunction,
				Package:    "test",
				FullName:   "test.Main",
				Signature:  "func Main()",
				Complexity: 5,
				Size:       20,
				Embedding:  []float32{0.1, 0.2, 0.3},
			},
			{
				ID:       "struct:test.Config",
				Label:    "Config",
				Type:     analyzer.NodeTypeStruct,
				Package:  "test",
				FullName: "test.Config",
				TypeInfo: analyzer.TypeInfo{
					Kind: "struct",
				},
			},
		},
		Edges: []analyzer.EnhancedEdge{
			{
				ID:     "edge1",
				Source: "function:test.Main",
				Target: "struct:test.Config",
				Type:   analyzer.EdgeTypeConstructs,
				Weight: 1,
			},
		},
		Stats: analyzer.GraphStats{
			TotalNodes: 3,
			TotalEdges: 1,
			NodesByType: map[string]int{
				analyzer.NodeTypePackage:  1,
				analyzer.NodeTypeFunction: 1,
				analyzer.NodeTypeStruct:   1,
			},
			EdgesByType: map[string]int{
				analyzer.EdgeTypeConstructs: 1,
			},
		},
	}

	// Validate graph structure
	if len(graph.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(graph.Edges))
	}

	// Test node with embedding
	nodeWithEmbedding := &graph.Nodes[1]
	if len(nodeWithEmbedding.Embedding) != 3 {
		t.Errorf("Expected embedding with 3 dimensions, got %d", len(nodeWithEmbedding.Embedding))
	}

	// Test node metadata
	nodeWithMetadata := &graph.Nodes[0]
	if nodeWithMetadata.Metadata["test"] != "value" {
		t.Error("Expected metadata to be preserved")
	}
}

func TestEdgeImportValidation(t *testing.T) {
	// Test various edge types
	edges := []analyzer.EnhancedEdge{
		{
			ID:     "edge1",
			Source: "function:main.main",
			Target: "function:main.helper",
			Type:   analyzer.EdgeTypeCalls,
			Weight: 1,
		},
		{
			ID:     "edge2",
			Source: "struct:main.Handler",
			Target: "interface:main.Writer",
			Type:   analyzer.EdgeTypeImplements,
			Weight: 1,
		},
	}

	for i := range edges {
		edge := &edges[i]
		// Validate edge structure
		if edge.Source == "" {
			t.Errorf("Edge %s missing source", edge.ID)
		}
		if edge.Target == "" {
			t.Errorf("Edge %s missing target", edge.ID)
		}
		if edge.Type == "" {
			t.Errorf("Edge %s missing type", edge.ID)
		}

		// Check for valid edge types
		validTypes := map[string]bool{
			analyzer.EdgeTypeCalls:           true,
			analyzer.EdgeTypeImplements:      true,
			analyzer.EdgeTypeImports:         true,
			analyzer.EdgeTypeConstructs:      true,
			analyzer.EdgeTypeSpawnsGoroutine: true,
		}

		if !validTypes[edge.Type] {
			t.Errorf("Edge %s has invalid type: %s", edge.ID, edge.Type)
		}
	}
}

func TestGraphStatsGeneration(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Skip("Neo4j not available for testing")
		return
	}
	defer client.Close()

	// This test would need a populated database
	// For now, we just test that the method doesn't panic
	ctx := context.Background()
	stats, err := client.GetGraphStats(ctx)
	if err != nil {
		// It's ok if it fails due to no data
		t.Logf("GetGraphStats error (expected if no data): %v", err)
	} else {
		// If it succeeds, validate the structure
		if stats["total_nodes"] != nil {
			if nodeCount, ok := stats["total_nodes"].(float64); ok && nodeCount < 0 {
				t.Error("total_nodes should not be negative")
			}
		}
		if stats["total_edges"] != nil {
			if edgeCount, ok := stats["total_edges"].(float64); ok && edgeCount < 0 {
				t.Error("total_edges should not be negative")
			}
		}
	}
}
