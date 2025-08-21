package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestActualDuplicateQuery(t *testing.T) {
	driver, server := setupTestDatabase(t)
	defer driver.Close(context.Background())

	ctx := context.Background()
	records := executeAndLogQuery(t, server, ctx)

	validateResultParsing(t, records)
	checkMainFunctionDuplicates(t, records)
	testPatternDetection(t, server, ctx)

	t.Logf("TestActualDuplicateQuery completed")
}

// setupTestDatabase creates database connection and server instance
func setupTestDatabase(t *testing.T) (neo4j.DriverWithContext, *Server) {
	driver, err := neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "codeGraph123", ""))
	if err != nil {
		t.Skip("Cannot connect to Neo4j database")
	}

	server := &Server{
		neo4jDriver: driver,
	}

	return driver, server
}

// executeAndLogQuery executes the duplicate functions query and logs results
func executeAndLogQuery(t *testing.T, server *Server, ctx context.Context) []*neo4j.Record {
	t.Logf("Testing query: %s", duplicateFunctionsQuery)

	records, err := server.executeCypherQuery(ctx, duplicateFunctionsQuery, map[string]any{"workspace": "test-workspace"}, false)
	if err != nil {
		t.Fatalf("executeCypherQuery failed: %v", err)
	}

	t.Logf("Got %d records from database", len(records))
	logRecordDetails(t, records)

	return records
}

// logRecordDetails logs detailed information about Neo4j records
func logRecordDetails(t *testing.T, records []*neo4j.Record) {
	for i, record := range records {
		t.Logf("Record %d:", i)
		for j, key := range record.Keys {
			value := record.Values[j]
			t.Logf("  %s (%T): %v", key, value, value)

			if key == "packages" {
				if arr, ok := value.([]any); ok {
					t.Logf("    packages array length: %d", len(arr))
					for idx, pkg := range arr {
						t.Logf("    [%d]: %v", idx, pkg)
					}
				}
			}
		}
	}
}

// validateResultParsing tests the JSON parsing functionality
func validateResultParsing(t *testing.T, records []*neo4j.Record) {
	jsonData, err := parseNeo4jResults(records)
	if err != nil {
		t.Fatalf("parseNeo4jResults failed: %v", err)
	}
	t.Logf("Parsed JSON result:\n%s", string(jsonData))
}

// checkMainFunctionDuplicates specifically validates main function duplicates
func checkMainFunctionDuplicates(t *testing.T, records []*neo4j.Record) {
	for _, record := range records {
		for j, key := range record.Keys {
			if key == "functionName" && record.Values[j] == "main" {
				validateMainFunctionPackages(t, record)
				break
			}
		}
	}
}

// validateMainFunctionPackages checks for package duplicates in main function
func validateMainFunctionPackages(t *testing.T, record *neo4j.Record) {
	packagesIdx, countIdx := findPackageIndices(record)

	if packagesIdx == -1 || countIdx == -1 {
		return
	}

	packages := record.Values[packagesIdx]
	count := record.Values[countIdx]
	t.Logf("Main function - raw packages: %v (type: %T)", packages, packages)
	t.Logf("Main function - raw count: %v (type: %T)", count, count)

	if arr, ok := packages.([]any); ok {
		duplicates := findDuplicatePackages(arr)
		if len(duplicates) > 0 {
			t.Errorf("Found duplicates in packages array from Neo4j: %v", duplicates)
		}
	}
}

// findPackageIndices locates packages and duplicateCount field indices
func findPackageIndices(record *neo4j.Record) (packagesIdx, countIdx int) {
	packagesIdx = -1
	countIdx = -1

	for k, key := range record.Keys {
		if key == "packages" {
			packagesIdx = k
		}
		if key == "duplicateCount" {
			countIdx = k
		}
	}

	return packagesIdx, countIdx
}

// findDuplicatePackages identifies duplicate entries in packages array
func findDuplicatePackages(packages []any) []string {
	seen := make(map[string]bool)
	var duplicates []string

	for _, pkg := range packages {
		pkgStr := pkg.(string)
		if seen[pkgStr] {
			duplicates = append(duplicates, pkgStr)
		}
		seen[pkgStr] = true
	}

	return duplicates
}

// testPatternDetection validates pattern detection functionality
func testPatternDetection(t *testing.T, server *Server, ctx context.Context) {
	t.Log("Testing pattern detection")

	res, err := server.handlePatternDetection(ctx, nil, &mcp.CallToolParamsFor[PatternDetectionParams]{
		Arguments: PatternDetectionParams{
			PatternType: "duplicate",
		},
	})
	if err != nil {
		t.Logf("Pattern detection error: %v", err)
		return
	}

	if res != nil {
		for _, content := range res.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				t.Logf("Pattern detection result: %s", textContent.Text)
			}
		}
	}
}
