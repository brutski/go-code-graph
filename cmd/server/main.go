package main

import (
	"flag"
	"log"
	"os"

	"github.com/brutski/go-code-graph/internal/server"
)

func main() {
	var (
		graphFile = flag.String("graph", "graph.json", "Path to the graph JSON file")
		port      = flag.Int("port", 8080, "Port to serve the visualization on")
	)
	flag.Parse()

	// Check if graph file exists
	if _, err := os.Stat(*graphFile); os.IsNotExist(err) {
		log.Fatalf("Graph file %s not found. Please run the analyzer first.", *graphFile)
	}

	// Create and start server
	srv := server.NewServer(*graphFile)
	log.Printf("🚀 Starting Code Graph Visualization Server")
	log.Printf("📊 Graph data: %s", *graphFile)
	log.Printf("🌐 Server: http://localhost:%d", *port)
	log.Printf("📈 Visualization: http://localhost:%d/visualization", *port)

	if err := srv.Start(*port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
