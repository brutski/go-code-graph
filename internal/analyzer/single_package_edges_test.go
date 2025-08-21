package analyzer

import (
	"log/slog"
	"path/filepath"
	"testing"
)

func TestSinglePackageEdgeCounts(t *testing.T) {
	// Analyze only the single_package_test directory
	testDir := filepath.Join("..", "..", "testdata", "single_package_test")

	// Create parser with nil embeddings generator and default logger
	logger := slog.Default()
	parser := NewParser(nil, logger)

	graph, err := parser.AnalyzeCodebase(testDir)
	if err != nil {
		t.Fatalf("Failed to analyze single package: %v", err)
	}

	// Shadow and recover edge types have been removed in the simplification
	// These were rarely used edge types that added complexity without significant value

	// Log summary
	t.Logf("Single package analysis complete:")
	t.Logf("  - Total nodes: %d", len(graph.Nodes))
	t.Logf("  - Total edges: %d", len(graph.Edges))

	// Verify we have some basic edges
	if len(graph.Edges) == 0 {
		t.Error("Expected at least some edges in the graph")
	}

	// Log edge types found
	edgeTypes := make(map[string]int)
	for _, edge := range graph.Edges {
		edgeTypes[edge.Type]++
	}

	t.Logf("Edge types found:")
	for edgeType, count := range edgeTypes {
		t.Logf("  - %s: %d", edgeType, count)
	}
}
