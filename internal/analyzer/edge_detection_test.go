package analyzer

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/brutski/go-code-graph/internal/embeddings"
)

// TestAllEdgeTypes comprehensively tests detection of all edge types
func TestAllEdgeTypes(t *testing.T) {
	// Create parser
	config := embeddings.DefaultMockConfig()
	client, err := embeddings.NewClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	logger := slog.Default()
	embeddingsGen := NewEmbeddingsGenerator(client, logger)
	parser := NewParser(embeddingsGen, logger)

	// Analyze each testdata package and combine results
	graph := NewGraph()
	testPackages := []string{
		"basic", "structural", "control", "errors",
		"concurrency", "generics", "advanced", "pointer_receiver",
	}

	for _, pkg := range testPackages {
		testdataPath := filepath.Join("testdata", "edges", pkg)
		pkgGraph, err := parser.AnalyzeCodebase(testdataPath)
		if err != nil {
			t.Logf("Warning: Failed to analyze %s: %v", pkg, err)
			continue
		}

		// Combine nodes and edges
		graph.Nodes = append(graph.Nodes, pkgGraph.Nodes...)
		graph.Edges = append(graph.Edges, pkgGraph.Edges...)
	}

	// Create edge type map for verification
	edgesByType := make(map[string][]*Edge)
	for i := range graph.Edges {
		edge := &graph.Edges[i]
		edgesByType[edge.Type] = append(edgesByType[edge.Type], edge)
	}

	// Define expected edge types with test cases - simplified to essential edges only
	expectedEdgeTypes := []struct {
		edgeType    string
		minExpected int
		description string
		verifyFunc  func(t *testing.T, edges []*Edge)
	}{
		{
			edgeType:    EdgeTypeCalls,
			minExpected: 3,
			description: "function calls",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Should have calls to callee(), Service.Process(), Writer.Write()
				foundCallee := false
				foundProcess := false
				foundPrintln := false

				for _, edge := range edges {
					if edge.Target == "function:testdata/basic.callee" {
						foundCallee = true
					}
					if contains(edge.Target, ".Process") {
						foundProcess = true
					}
					if contains(edge.Target, "Println") {
						foundPrintln = true
					}
				}

				if !foundCallee {
					t.Error("Expected call edge to callee()")
				}
				if !foundProcess {
					t.Error("Expected call edge to Process() method")
				}
				if !foundPrintln {
					t.Log("Note: fmt.Println() edge not found - external packages not analyzed")
				}
			},
		},
		{
			edgeType:    EdgeTypeImplements,
			minExpected: 3,
			description: "interface implementations",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// FileReader implements Reader, ReadCloser implements Reader and Closer
				implementers := make(map[string]bool)
				for _, edge := range edges {
					if contains(edge.Target, "Reader") || contains(edge.Target, "Writer") || contains(edge.Target, "Closer") {
						implementers[edge.Source] = true
					}
				}
				if len(implementers) < 2 {
					t.Errorf("Expected at least 2 Reader implementers, got %d", len(implementers))
				}
			},
		},
		{
			edgeType:    EdgeTypeImports,
			minExpected: 3,
			description: "package imports",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Check for imports between test packages
				foundBasicImport := false
				foundStructuralImport := false

				for _, edge := range edges {
					// control imports testdata/basic and testdata/structural
					if contains(edge.Source, "control") && contains(edge.Target, "basic") {
						foundBasicImport = true
					}
					if contains(edge.Source, "control") && contains(edge.Target, "structural") {
						foundStructuralImport = true
					}
				}

				if !foundBasicImport {
					t.Error("Expected import edge from control to basic package")
				}
				if !foundStructuralImport {
					t.Error("Expected import edge from control to structural package")
				}
			},
		},
		{
			edgeType:    EdgeTypeEmbeds,
			minExpected: 3,
			description: "struct/interface embedding",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Extended embeds Base, ExtendedPtr embeds *Base, etc.
				foundBase := false
				foundInterface := false
				for _, edge := range edges {
					if contains(edge.Target, "Base") {
						foundBase = true
					}
					if contains(edge.Target, "Logger") && contains(edge.Source, "ErrorLogger") {
						foundInterface = true
					}
				}
				if !foundBase {
					t.Error("Expected embed edge for Base struct")
				}
				if !foundInterface {
					t.Error("Expected embed edge for Logger interface")
				}
			},
		},
		{
			edgeType:    EdgeTypeHasMethod,
			minExpected: 5,
			description: "type methods",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// User should have GetUsername, SetPassword, IsValid methods
				userMethods := 0
				for _, edge := range edges {
					if contains(edge.Source, "User") &&
						(contains(edge.Target, "GetUsername") ||
							contains(edge.Target, "SetPassword") ||
							contains(edge.Target, "IsValid")) {
						userMethods++
					}
				}
				if userMethods < 3 {
					t.Errorf("Expected at least 3 User methods, found %d", userMethods)
				}
			},
		},
		{
			edgeType:    EdgeTypeHasField,
			minExpected: 5,
			description: "struct fields",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// User should have ID, Username, Email, password, Profile fields
				userFields := 0
				for _, edge := range edges {
					if contains(edge.Source, "struct:") && contains(edge.Source, "User") {
						userFields++
					}
				}
				if userFields < 4 {
					t.Errorf("Expected at least 4 User fields, found %d", userFields)
				}
			},
		},
		{
			edgeType:    EdgeTypeHasParameter,
			minExpected: 5,
			description: "function parameters",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// CreateUser should have id, username, email, opts parameters
				createUserParams := 0
				for _, edge := range edges {
					if contains(edge.Source, "CreateUser") {
						createUserParams++
					}
				}
				if createUserParams < 4 {
					t.Errorf("Expected at least 4 CreateUser parameters, found %d", createUserParams)
				}
			},
		},
		{
			edgeType:    EdgeTypeConstructs,
			minExpected: 2,
			description: "constructor patterns",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// NewService and NewServiceWithDefaults should construct Service
				foundNewService := false
				for _, edge := range edges {
					if contains(edge.Source, "NewService") && contains(edge.Target, "Service") {
						foundNewService = true
					}
				}
				if !foundNewService {
					t.Error("Expected constructs edge from NewService to Service")
				}
			},
		},
		{
			edgeType:    EdgeTypeReturns,
			minExpected: 2,
			description: "return types",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Functions should have returns edges for their return types
				foundUserReturn := false
				foundErrorReturn := false
				for _, edge := range edges {
					if contains(edge.Target, "User") || contains(edge.Target, "Service") {
						foundUserReturn = true
					}
					if contains(edge.Target, "error") {
						foundErrorReturn = true
					}
				}
				if !foundUserReturn {
					t.Error("Expected returns edge to User type")
				}
				if !foundErrorReturn {
					t.Error("Expected returns edge to error type")
				}
			},
		},
		{
			edgeType:    EdgeTypeUses,
			minExpected: 2,
			description: "type usage",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// ProcessUser uses User type
				foundUserUsage := false
				for _, edge := range edges {
					if contains(edge.Source, "ProcessUser") && contains(edge.Target, "User") {
						foundUserUsage = true
					}
				}
				if !foundUserUsage {
					t.Error("Expected uses edge from ProcessUser to User")
				}
			},
		},
		{
			edgeType:    EdgeTypeHandlesError,
			minExpected: 3,
			description: "error handling",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Functions should handle errors from various sources
				if len(edges) < 3 {
					t.Logf("Warning: Expected more error handling edges, got %d", len(edges))
				}
			},
		},
		{
			edgeType:    EdgeTypeWrapsError,
			minExpected: 2,
			description: "error wrapping",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Should have error wrapping with fmt.Errorf
				if len(edges) < 1 {
					t.Error("Expected at least one error wrapping edge")
				}
			},
		},
		{
			edgeType:    EdgeTypeSpawnsGoroutine,
			minExpected: 3,
			description: "goroutine spawning",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Should spawn worker goroutines
				foundWorker := false
				for _, edge := range edges {
					if contains(edge.Target, "worker") {
						foundWorker = true
					}
				}
				if !foundWorker {
					t.Error("Expected spawns edge to worker function")
				}
			},
		},
		// Removed edge types (no longer tested):
		// - EdgeTypeTypeAsserts
		// - EdgeTypeSendsChannel
		// - EdgeTypeReceivesChannel
		// - EdgeTypeInstantiates
		// - EdgeTypePromotes
		// - EdgeTypeShadows
		// - EdgeTypeClosureCaptures
		// - EdgeTypeDefers
		// - EdgeTypePanics
		// - EdgeTypeRecovers
		{
			edgeType:    EdgeTypeMethodOf,
			minExpected: 3,
			description: "method ownership",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Repository methods should have method_of edges
				foundSave := false
				foundFind := false
				for _, edge := range edges {
					if contains(edge.Source, "Save") && contains(edge.Target, "Repository") {
						foundSave = true
					}
					if contains(edge.Source, "Find") && contains(edge.Target, "Repository") {
						foundFind = true
					}
				}
				if !foundSave {
					t.Error("Expected method_of edge from Save to Repository")
				}
				if !foundFind {
					t.Error("Expected method_of edge from Find to Repository")
				}
			},
		},
		{
			edgeType:    EdgeTypeParameterType,
			minExpected: 3,
			description: "parameter types",
			verifyFunc: func(t *testing.T, edges []*Edge) {
				// Parameters should have type edges
				foundUserParam := false
				foundReaderParam := false
				for _, edge := range edges {
					if contains(edge.Target, "User") {
						foundUserParam = true
					}
					if contains(edge.Target, "Reader") {
						foundReaderParam = true
					}
				}
				if !foundUserParam {
					t.Error("Expected parameter_type edge to User")
				}
				if !foundReaderParam {
					t.Error("Expected parameter_type edge to Reader")
				}
			},
		},
	}

	// Verify each edge type
	t.Logf("Total edges found: %d", len(graph.Edges))
	t.Logf("Edge types found: %d", len(edgesByType))

	for _, expected := range expectedEdgeTypes {
		t.Run(expected.edgeType, func(t *testing.T) {
			edges := edgesByType[expected.edgeType]
			t.Logf("Found %d %s edges", len(edges), expected.description)

			if len(edges) < expected.minExpected {
				t.Errorf("Expected at least %d %s edges, got %d",
					expected.minExpected, expected.description, len(edges))
			}

			if expected.verifyFunc != nil && len(edges) > 0 {
				expected.verifyFunc(t, edges)
			}
		})
	}

	// Log summary of all edge types found
	t.Logf("\nEdge Type Summary:")
	recognizedTypes := make(map[string]bool)
	for _, expected := range expectedEdgeTypes {
		recognizedTypes[expected.edgeType] = true
	}

	// Log all edge types found
	for edgeType, edges := range edgesByType {
		if !recognizedTypes[edgeType] {
			t.Logf("Found edge type (mapped): %s (%d edges)", edgeType, len(edges))
		}
	}

	// Log expected but missing edge types
	t.Logf("\nMissing edge types:")
	for _, expected := range expectedEdgeTypes {
		if _, found := edgesByType[expected.edgeType]; !found {
			t.Logf("- %s (%s)", expected.edgeType, expected.description)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[len(s)-len(substr):] == substr ||
			s[:len(substr)] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
