package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Server provides HTTP server for graph visualization
type Server struct {
	graphFile string
}

// NewServer creates a new visualization server
func NewServer(graphFile string) *Server {
	return &Server{
		graphFile: graphFile,
	}
}

// Start starts the HTTP server on the specified port
func (s *Server) Start(port int) error {
	// Setup routes
	mux := http.NewServeMux()

	// Serve static files (CSS, JS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Serve the main landing page
	mux.HandleFunc("/", s.handleIndex)

	// Serve the visualization page
	mux.HandleFunc("/visualization", s.handleVisualization)

	// API endpoints to serve graph data
	mux.HandleFunc("/api/graph", s.handleGraph)
	mux.HandleFunc("/api/summary", s.handleSummary)

	// Start server with proper timeouts
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server at http://localhost%s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * 1000000000, // 15 seconds
		WriteTimeout: 15 * 1000000000, // 15 seconds
		IdleTimeout:  60 * 1000000000, // 60 seconds
	}

	return server.ListenAndServe()
}

// handleIndex serves the main landing page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

// handleVisualization serves the interactive graph visualization page
func (s *Server) handleVisualization(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/visualization.html")
}

// handleGraph serves the graph data as JSON
func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	// Check if graph file exists
	if _, err := os.Stat(s.graphFile); os.IsNotExist(err) {
		http.Error(w, "Graph file not found", http.StatusNotFound)
		return
	}

	// Read and serve the graph file
	data, err := os.ReadFile(s.graphFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading graph file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for development
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing graph response: %v", err)
	}
}

// handleSummary serves the summary data as JSON
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	summaryFile := s.graphFile + ".summary"

	// Check if summary file exists
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		http.Error(w, "Summary file not found", http.StatusNotFound)
		return
	}

	// Read and serve the summary file
	data, err := os.ReadFile(summaryFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading summary file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for development
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing summary response: %v", err)
	}
}
