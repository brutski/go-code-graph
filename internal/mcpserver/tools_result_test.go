package mcpserver

import (
	"encoding/json"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestParseNeo4jResults(t *testing.T) {
	// Create mock Neo4j records that simulate the duplicate function query results
	mockRecords := []*neo4j.Record{
		{
			Keys:   []string{"functionName", "packages", "duplicateCount"},
			Values: []any{"main", []any{"pkg1", "pkg2", "pkg3"}, int64(3)},
		},
		{
			Keys:   []string{"functionName", "packages", "duplicateCount"},
			Values: []any{"NewClient", []any{"pkg1", "pkg2"}, int64(2)},
		},
	}

	// Test parseNeo4jResults function
	jsonData, err := parseNeo4jResults(mockRecords)
	if err != nil {
		t.Fatalf("parseNeo4jResults failed: %v", err)
	}

	// Parse the JSON to verify structure
	var results []map[string]any
	if err := json.Unmarshal(jsonData, &results); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify we have 2 results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify first result
	if results[0]["functionName"] != "main" {
		t.Errorf("Expected functionName 'main', got %v", results[0]["functionName"])
	}

	packages := results[0]["packages"].([]any)
	if len(packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(packages))
	}

	// Check for duplicates in packages array
	packageMap := make(map[string]bool)
	for _, pkg := range packages {
		pkgStr := pkg.(string)
		if packageMap[pkgStr] {
			t.Errorf("Found duplicate package: %s", pkgStr)
		}
		packageMap[pkgStr] = true
	}

	t.Logf("Parsed results: %s", string(jsonData))
}

func TestConvertNeo4jValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "string value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "array of strings",
			input:    []any{"a", "b", "c"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "nested array",
			input:    []any{[]any{"inner1", "inner2"}, "outer"},
			expected: []any{[]any{"inner1", "inner2"}, "outer"},
		},
		{
			name: "map",
			input: map[string]any{
				"key1": "value1",
				"key2": []any{"a", "b"},
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": []any{"a", "b"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertNeo4jValue(tt.input)

			// Convert both to JSON for comparison
			expectedJSON, _ := json.Marshal(tt.expected)
			resultJSON, _ := json.Marshal(result)

			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("convertNeo4jValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
