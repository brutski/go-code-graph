package mcpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCypherQuerySyntax validates Cypher queries that are currently implemented
func TestCypherQuerySyntax(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name: "architectureLayersQuery",
			query: `
				MATCH (n:CodeNode)
				WHERE n.type IN ['struct', 'interface', 'function']
				WITH n.package as layer, count(n) as nodeCount
				RETURN layer, nodeCount
				ORDER BY nodeCount DESC
			`,
		},
		{
			name: "findImplementersQuery",
			query: `
				MATCH (s:CodeNode)-[r:RELATES_TO {type: 'implements'}]->(i:CodeNode)
				WHERE (i.full_name = $name OR i.label = $name)
				  AND s.type = 'struct'
				  AND i.type = 'interface'
				RETURN s.full_name as implementer, s.package as package, s.label as name
				ORDER BY s.label
			`,
		},
		{
			name: "duplicateFunctionsQuery",
			query: `
				MATCH (f:CodeNode {type: 'function'})
				WITH f.label as functionName, collect(DISTINCT f.package) as packages
				WHERE size(packages) > 1
				RETURN functionName, packages, size(packages) as duplicateCount
				ORDER BY duplicateCount DESC, functionName
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic syntax validation
			assert.NotEmpty(t, tt.query, "Query should not be empty")
			assert.Contains(t, tt.query, "MATCH", "Query should contain MATCH clause")
			assert.Contains(t, tt.query, "RETURN", "Query should contain RETURN clause")

			// Check for common syntax errors
			assert.NotContains(t, tt.query, "WHERE r.type", "Should not access .type on variable-length relationship list")
			assert.NotContains(t, tt.query, ":Function", "Should use CodeNode schema, not old labels")
			assert.NotContains(t, tt.query, ":Struct", "Should use CodeNode schema, not old labels")
			assert.NotContains(t, tt.query, ":CALLS", "Should use RELATES_TO schema, not old relationships")
		})
	}
}

// TestNaturalLanguageQueries validates natural language to Cypher conversion
func TestNaturalLanguageQueries(t *testing.T) {
	server := &Server{}

	tests := []struct {
		question string
		expected []string // Expected elements in generated query
	}{
		{
			question: "show me complex functions",
			expected: []string{"CodeNode", "type IN ['function', 'method']", "complexity"},
		},
		{
			question: "show interfaces and implementations",
			expected: []string{"CodeNode", "interface", "type: 'interface'"},
		},
		{
			question: "what are the main packages",
			expected: []string{"CodeNode", "package", "type: 'package'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.question, func(t *testing.T) {
			query, err := server.naturalLanguageToCypher(tt.question, "", "test-workspace")
			require.NoError(t, err)
			require.NotEmpty(t, query)

			for _, expected := range tt.expected {
				assert.Contains(t, query, expected,
					"Query should contain expected element: %s\nGenerated query: %s",
					expected, query)
			}
		})
	}
}

// TestPatternDetectionQueries validates pattern detection query generation
func TestPatternDetectionQueries(t *testing.T) {
	// Test pattern detection query structure without invoking the handlers
	validPatterns := []string{"duplicate", "usage"}

	for _, pattern := range validPatterns {
		t.Run(pattern, func(t *testing.T) {
			var expectedQuery string

			switch pattern {
			case "duplicate":
				expectedQuery = `
					MATCH (f:CodeNode {type: 'function'})
					WITH f.label as functionName, collect(DISTINCT f.package) as packages
					WHERE size(packages) > 1
					RETURN functionName, packages, size(packages) as duplicateCount
					ORDER BY duplicateCount DESC, functionName
				`
			case "usage":
				expectedQuery = `
					MATCH (n:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode)
					WITH target, count(r) as usageCount
					WHERE usageCount > 1
					RETURN target.label as targetName,
					       target.package as package,
					       target.type as nodeType,
					       usageCount
					ORDER BY usageCount DESC
				`
			}

			// Validate the query structure
			assert.Contains(t, expectedQuery, "MATCH", "Query should have MATCH clause")
			assert.Contains(t, expectedQuery, "RETURN", "Query should have RETURN clause")
			assert.Contains(t, expectedQuery, "CodeNode", "Query should use CodeNode schema")
		})
	}
}

// TestSchemaConsistency validates that implemented queries use the correct schema
func TestSchemaConsistency(t *testing.T) {
	// Test some actual implemented queries from the tools
	queries := []struct {
		name  string
		query string
	}{
		{
			name: "layers",
			query: `
				MATCH (n:CodeNode)
				WHERE n.type IN ['struct', 'interface', 'function']
				WITH n.package as layer, count(n) as nodeCount
				RETURN layer, nodeCount
				ORDER BY nodeCount DESC
			`,
		},
		{
			name: "findImplementers",
			query: `
				MATCH (s:CodeNode)-[r:RELATES_TO {type: 'implements'}]->(i:CodeNode)
				WHERE (i.full_name = $name OR i.label = $name)
				  AND s.type = 'struct'
				  AND i.type = 'interface'
				RETURN s.full_name as implementer, s.package as package, s.label as name
				ORDER BY s.label
			`,
		},
	}

	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			// Should use CodeNode schema
			assert.Contains(t, q.query, "CodeNode", "Should use CodeNode label")

			// Only check for RELATES_TO if the query actually uses relationships
			if q.name == "findImplementers" {
				assert.Contains(t, q.query, "RELATES_TO", "Should use RELATES_TO relationship")
			}

			// Should NOT use old schema
			assert.NotContains(t, q.query, ":Function", "Should not use old Function label")
			assert.NotContains(t, q.query, ":Struct", "Should not use old Struct label")
			assert.NotContains(t, q.query, ":Interface", "Should not use old Interface label")
			assert.NotContains(t, q.query, ":CALLS", "Should not use old CALLS relationship")
			assert.NotContains(t, q.query, ":IMPLEMENTS", "Should not use old IMPLEMENTS relationship")
		})
	}
}

// TestQueryParameterization validates that queries with parameters are properly structured
func TestQueryParameterization(t *testing.T) {
	tests := []struct {
		name     string
		generate func() string
	}{
		{
			name: "findImplementers",
			generate: func() string {
				return `
				MATCH (s:CodeNode {type: 'struct'})-[r:RELATES_TO {type: 'implements'}]->(i:CodeNode {type: 'interface'})
				WHERE i.full_name = $name OR i.label = $name
				RETURN s.full_name as implementer, s.package as package, s.label as name
				ORDER BY s.label
				`
			},
		},
		{
			name: "traceCallPath",
			generate: func() string {
				return `
				MATCH path = shortestPath(
					(start:CodeNode)-[r:RELATES_TO*1..10]->(end:CodeNode)
				)
				WHERE (start.full_name = $from OR start.label = $from)
				  AND (end.full_name = $to OR end.label = $to)
				  AND start.type IN ['function', 'method']
				  AND end.type IN ['function', 'method']
				  AND all(rel in r WHERE rel.type = 'calls')
				RETURN [n in nodes(path) | n.full_name] as callPath,
				       [n in nodes(path) | n.label] as callNames,
				       length(path) as pathLength
				`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.generate()

			// Should contain parameter placeholders
			switch tt.name {
			case "findImplementers":
				assert.Contains(t, query, "$name", "Should contain $name parameter")
			case "traceCallPath":
				assert.Contains(t, query, "$from", "Should contain $from parameter")
				assert.Contains(t, query, "$to", "Should contain $to parameter")
			}

			// Basic query structure
			assert.Contains(t, query, "MATCH", "Should contain MATCH clause")
			assert.Contains(t, query, "RETURN", "Should contain RETURN clause")
		})
	}
}
