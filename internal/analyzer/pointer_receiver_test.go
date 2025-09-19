package analyzer

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/brutski/go-code-graph/internal/embeddings"
)

// TestPointerReceiverMethodCallEdges tests the fix for GitHub issue #3
// This test verifies that call edges are properly constructed when a struct value
// calls a method with a pointer receiver
func TestPointerReceiverMethodCallEdges(t *testing.T) {
	// Create parser
	config := embeddings.DefaultMockConfig()
	client, err := embeddings.NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	logger := slog.Default()
	embeddingsGen := NewEmbeddingsGenerator(client, logger)
	parser := NewParser(embeddingsGen, logger)

	// Analyze the pointer_receiver test package
	testdataPath := filepath.Join("testdata", "edges", "pointer_receiver")
	graph, err := parser.AnalyzeCodebase(testdataPath)
	if err != nil {
		t.Fatalf("Failed to analyze testdata: %v", err)
	}

	// Look for the specific edge that should be created:
	// TestPointerReceiverCall -> Service.PubackEncode -> *BasicFileWriter.Write

	// First, find all call edges
	callEdges := make([]*Edge, 0)
	for i := range graph.Edges {
		edge := &graph.Edges[i]
		if edge.Type == EdgeTypeCalls {
			callEdges = append(callEdges, edge)
		}
	}

	t.Logf("Found %d call edges in pointer_receiver test", len(callEdges))

	// Look for specific edges that should exist with our fix
	foundPubackEncodeCall := false
	foundPointerWriteCall := false
	foundValueWriteCall := false

	for _, edge := range callEdges {
		t.Logf("Call edge: %s -> %s", edge.Source, edge.Target)

		// Look for the call from TestPointerReceiverCall to Service.PubackEncode
		if contains(edge.Source, "TestPointerReceiverCall") && contains(edge.Target, "Service.PubackEncode") {
			foundPubackEncodeCall = true
		}

		// Look for the call to *BasicFileWriter.Write (the key fix)
		// The target should contain the full path with the * prefix
		if contains(edge.Target, "*testdata/pointer_receiver.BasicFileWriter.Write") {
			foundPointerWriteCall = true
		}

		// Look for the call from TestValueCallsPointerMethod to *BasicFileWriter.Write
		if contains(edge.Source, "TestValueCallsPointerMethod") && contains(edge.Target, "*testdata/pointer_receiver.BasicFileWriter.Write") {
			foundValueWriteCall = true
		}
	}

	// Verify that the key edges are found
	if !foundPubackEncodeCall {
		t.Error("Expected call edge from TestPointerReceiverCall to Service.PubackEncode")
	}

	if !foundPointerWriteCall {
		t.Error("Expected call edge to *testdata/pointer_receiver.BasicFileWriter.Write (this verifies the fix for issue #3)")
	}

	if !foundValueWriteCall {
		t.Error("Expected call edge from TestValueCallsPointerMethod to *testdata/pointer_receiver.BasicFileWriter.Write")
	}

	// Log nodes for debugging
	t.Logf("Found %d nodes and %d edges", len(graph.Nodes), len(graph.Edges))

	// Log all nodes to understand what's being created
	for _, node := range graph.Nodes {
		t.Logf("Node: Type=%s, Label=%s, ID=%s", node.Type, node.Label, node.ID)
	}

	// Log all edges to understand what relationships exist
	for _, edge := range graph.Edges {
		t.Logf("Edge: %s -> %s (type: %s)", edge.Source, edge.Target, edge.Type)
	}
}

// Note: contains and findSubstring helper functions are already defined in edge_detection_test.go