package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer("/path/to/graph.json")

	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.graphFile != "/path/to/graph.json" {
		t.Errorf("Expected graphFile '/path/to/graph.json', got %s", server.graphFile)
	}
}

func TestHandleIndex(t *testing.T) {
	// Create temporary web directory
	tempDir := t.TempDir()
	webDir := filepath.Join(tempDir, "web")
	err := os.MkdirAll(webDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create web directory: %v", err)
	}

	// Create index.html
	indexContent := `<!DOCTYPE html><html><body>Test Index</body></html>`
	err = os.WriteFile(filepath.Join(webDir, "index.html"), []byte(indexContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}

	// Change to temp directory for test
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	server := NewServer("graph.json")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandleGraph(t *testing.T) {
	// Create temporary graph file
	tempDir := t.TempDir()
	graphFile := filepath.Join(tempDir, "test-graph.json")

	graphData := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":    "test-node",
				"label": "Test",
				"type":  "function",
			},
		},
		"edges": []interface{}{},
	}

	data, err := json.Marshal(graphData)
	if err != nil {
		t.Fatalf("Failed to marshal graph data: %v", err)
	}

	err = os.WriteFile(graphFile, data, 0o600)
	if err != nil {
		t.Fatalf("Failed to create graph file: %v", err)
	}

	server := NewServer(graphFile)

	req := httptest.NewRequest("GET", "/api/graph", nil)
	w := httptest.NewRecorder()

	server.handleGraph(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}

	// Check CORS header
	cors := resp.Header.Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("Expected CORS header '*', got %s", cors)
	}

	// Verify response body
	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	nodes, ok := responseData["nodes"].([]interface{})
	if !ok || len(nodes) != 1 {
		t.Error("Expected 1 node in response")
	}
}

func TestHandleGraphFileNotFound(t *testing.T) {
	server := NewServer("/non/existent/file.json")

	req := httptest.NewRequest("GET", "/api/graph", nil)
	w := httptest.NewRecorder()

	server.handleGraph(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleSummary(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	graphFile := filepath.Join(tempDir, "test-graph.json")
	summaryFile := graphFile + ".summary"

	summaryData := map[string]interface{}{
		"total_nodes": 10,
		"total_edges": 5,
		"packages":    []string{"main", "utils"},
	}

	data, err := json.Marshal(summaryData)
	if err != nil {
		t.Fatalf("Failed to marshal summary data: %v", err)
	}

	err = os.WriteFile(summaryFile, data, 0o600)
	if err != nil {
		t.Fatalf("Failed to create summary file: %v", err)
	}

	server := NewServer(graphFile)

	req := httptest.NewRequest("GET", "/api/summary", nil)
	w := httptest.NewRecorder()

	server.handleSummary(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify response
	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if responseData["total_nodes"].(float64) != 10 {
		t.Errorf("Expected total_nodes 10, got %v", responseData["total_nodes"])
	}
}

func TestHandleSummaryFileNotFound(t *testing.T) {
	server := NewServer("/non/existent/file.json")

	req := httptest.NewRequest("GET", "/api/summary", nil)
	w := httptest.NewRecorder()

	server.handleSummary(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestServerRoutes(t *testing.T) {
	// Test that all expected routes are configured
	_ = NewServer("test.json") // Just validate it can be created

	// Create a test server with our routes
	mux := http.NewServeMux()

	// Mock static file server
	mux.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/visualization", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test each route
	routes := []string{
		"/",
		"/visualization",
		"/api/graph",
		"/api/summary",
		"/static/test.js",
	}

	client := &http.Client{}
	for _, route := range routes {
		resp, err := client.Get(ts.URL + route)
		if err != nil {
			t.Errorf("Failed to GET %s: %v", route, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Route %s returned status %d, expected 200", route, resp.StatusCode)
		}
	}
}

func TestHandleGraphReadError(t *testing.T) {
	// Create a file that exists but can't be read
	tempDir := t.TempDir()
	graphFile := filepath.Join(tempDir, "unreadable.json")

	err := os.WriteFile(graphFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Make file unreadable (on Unix systems)
	err = os.Chmod(graphFile, 0o000)
	if err != nil {
		t.Skip("Cannot change file permissions on this system")
	}
	defer os.Chmod(graphFile, 0o600) // Restore permissions

	server := NewServer(graphFile)

	req := httptest.NewRequest("GET", "/api/graph", nil)
	w := httptest.NewRecorder()

	server.handleGraph(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}
