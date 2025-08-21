package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/brutski/go-code-graph/internal/analyzer"
)

func TestNewJSONExporter(t *testing.T) {
	exporter := NewJSONExporter()
	if exporter == nil {
		t.Fatal("Expected non-nil exporter")
	}
}

func TestExport(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-graph.json")

	// Create test graph
	graph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{
			{
				ID:       "package:main",
				Label:    "main",
				Type:     analyzer.NodeTypePackage,
				Package:  "main",
				FullName: "main",
				Size:     100,
				Position: analyzer.Position{
					Filename: "main.go",
					Line:     1,
				},
				Visibility: analyzer.VisibilityPublic,
				CreatedAt:  time.Now(),
			},
			{
				ID:         "function:main.main",
				Label:      "main",
				Type:       analyzer.NodeTypeFunction,
				Package:    "main",
				FullName:   "main.main",
				Size:       10,
				Complexity: 2,
				Position: analyzer.Position{
					Filename: "main.go",
					Line:     5,
				},
				Signature:  "func main()",
				Visibility: analyzer.VisibilityPublic,
				Metadata: map[string]interface{}{
					"test": "value",
				},
				CreatedAt: time.Now(),
			},
		},
		Edges: []analyzer.EnhancedEdge{
			{
				ID:     "edge1",
				Source: "package:main",
				Target: "function:main.main",
				Type:   analyzer.EdgeTypeHasMethod,
				Weight: 1,
				Position: analyzer.Position{
					Filename: "main.go",
					Line:     5,
				},
			},
		},
		Stats: analyzer.GraphStats{
			NodesByType: map[string]int{
				analyzer.NodeTypePackage:  1,
				analyzer.NodeTypeFunction: 1,
			},
			EdgesByType: map[string]int{
				analyzer.EdgeTypeHasMethod: 1,
			},
		},
		Metadata: analyzer.GraphMetadata{
			SourcePath: "/test/project",
			GoVersion:  "1.19",
			ModulePath: "test/module",
		},
	}

	exporter := NewJSONExporter()
	err := exporter.Export(graph, outputFile)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read and verify the exported data
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var exportedData ExportData
	err = json.Unmarshal(data, &exportedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported data: %v", err)
	}

	// Verify nodes
	if len(exportedData.Graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(exportedData.Graph.Nodes))
	}

	// Verify edges
	if len(exportedData.Graph.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(exportedData.Graph.Edges))
	}

	// Verify metadata
	if exportedData.Graph.Metadata.SourcePath != "/test/project" {
		t.Errorf("Expected source path '/test/project', got %s", exportedData.Graph.Metadata.SourcePath)
	}
}

func TestExportNilGraph(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "nil-graph.json")

	exporter := NewJSONExporter()
	err := exporter.Export(nil, outputFile)
	if err == nil {
		t.Error("Expected error when exporting nil graph")
	}
}

func TestExportEmptyGraph(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "empty-graph.json")

	emptyGraph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{},
		Edges: []analyzer.EnhancedEdge{},
		Stats: analyzer.GraphStats{
			NodesByType: map[string]int{},
			EdgesByType: map[string]int{},
		},
	}

	exporter := NewJSONExporter()
	err := exporter.Export(emptyGraph, outputFile)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var exportedData ExportData
	err = json.Unmarshal(data, &exportedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported data: %v", err)
	}

	if len(exportedData.Graph.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(exportedData.Graph.Nodes))
	}
	if len(exportedData.Graph.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(exportedData.Graph.Edges))
	}
}

func TestExportInvalidPath(t *testing.T) {
	graph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{},
		Edges: []analyzer.EnhancedEdge{},
	}

	exporter := NewJSONExporter()
	err := exporter.Export(graph, "/invalid/path/that/does/not/exist/output.json")
	if err == nil {
		t.Error("Expected error when exporting to invalid path")
	}
}

func TestExportSummary(t *testing.T) {
	tempDir := t.TempDir()
	summaryFile := filepath.Join(tempDir, "test-summary.json")

	graph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{
			{ID: "package:main", Type: analyzer.NodeTypePackage, Package: "main", Label: "main"},
			{ID: "package:utils", Type: analyzer.NodeTypePackage, Package: "utils", Label: "utils"},
			{ID: "1", Type: analyzer.NodeTypeFunction, Package: "main"},
			{ID: "2", Type: analyzer.NodeTypeStruct, Package: "main"},
			{ID: "3", Type: analyzer.NodeTypeInterface, Package: "utils"},
			{ID: "4", Type: analyzer.NodeTypeMethod, Package: "utils"},
		},
		Edges: []analyzer.EnhancedEdge{
			{Type: analyzer.EdgeTypeCalls},
			{Type: analyzer.EdgeTypeCalls},
			{Type: analyzer.EdgeTypeImplements},
		},
		Stats: analyzer.GraphStats{
			NodesByType: map[string]int{
				analyzer.NodeTypePackage:   2,
				analyzer.NodeTypeFunction:  1,
				analyzer.NodeTypeStruct:    1,
				analyzer.NodeTypeInterface: 1,
				analyzer.NodeTypeMethod:    1,
			},
			EdgesByType: map[string]int{
				analyzer.EdgeTypeCalls:      2,
				analyzer.EdgeTypeImplements: 1,
			},
		},
	}

	exporter := NewJSONExporter()
	err := exporter.ExportSummary(graph, summaryFile)
	if err != nil {
		t.Fatalf("ExportSummary failed: %v", err)
	}

	// Read and verify summary (note: ExportSummary appends ".summary" to the filename)
	data, err := os.ReadFile(summaryFile + ".summary")
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	var summary map[string]interface{}
	err = json.Unmarshal(data, &summary)
	if err != nil {
		t.Fatalf("Failed to unmarshal summary: %v", err)
	}

	// Verify summary contents
	if summary["total_nodes"].(float64) != 6 {
		t.Errorf("Expected total_nodes 6, got %v", summary["total_nodes"])
	}
	if summary["total_edges"].(float64) != 3 {
		t.Errorf("Expected total_edges 3, got %v", summary["total_edges"])
	}

	packages, ok := summary["packages"].([]interface{})
	if !ok || len(packages) != 2 {
		t.Error("Expected 2 packages in summary")
	}

	nodeBreakdown, ok := summary["node_breakdown"].(map[string]interface{})
	if !ok {
		t.Error("Expected node_breakdown in summary")
	} else {
		if nodeBreakdown["function"].(float64) != 1 {
			t.Errorf("Expected 1 function, got %v", nodeBreakdown["function"])
		}
	}
}

func TestExportWithEmbeddings(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "embeddings-graph.json")

	graph := &analyzer.Graph{
		Nodes: []analyzer.EnhancedNode{
			{
				ID:              "function:test",
				Label:           "test",
				Type:            analyzer.NodeTypeFunction,
				Package:         "main",
				FullName:        "main.test",
				Embedding:       []float32{0.1, 0.2, 0.3, 0.4, 0.5},
				EmbeddingModel:  "test-model",
				SemanticSummary: "A test function",
				Metadata: map[string]interface{}{
					"has_embedding": true,
					"semantic_hash": "abc123",
				},
			},
		},
		Edges: []analyzer.EnhancedEdge{},
	}

	exporter := NewJSONExporter()
	err := exporter.Export(graph, outputFile)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify embeddings are preserved
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var exportedData ExportData
	err = json.Unmarshal(data, &exportedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported data: %v", err)
	}

	node := exportedData.Graph.Nodes[0]
	if len(node.Embedding) != 5 {
		t.Errorf("Expected 5 embedding dimensions, got %d", len(node.Embedding))
	}
	if node.EmbeddingModel != "test-model" {
		t.Errorf("Expected embedding model 'test-model', got %s", node.EmbeddingModel)
	}
	if node.SemanticSummary != "A test function" {
		t.Errorf("Expected semantic summary 'A test function', got %s", node.SemanticSummary)
	}
}
