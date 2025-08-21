package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/brutski/go-code-graph/internal/analyzer"
	"github.com/brutski/go-code-graph/internal/neo4j"
)

func main() {
	var (
		inputFile = flag.String("input", "", "JSON file containing the analyzed graph")
		uri       = flag.String("uri", "bolt://localhost:7687", "Neo4j URI")
		username  = flag.String("username", "neo4j", "Neo4j username")
		password  = flag.String("password", "codeGraph123", "Neo4j password")
		database  = flag.String("database", "neo4j", "Neo4j database name")
	)
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -input=<graph.json>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("🔄 Importing graph from %s to Neo4j...\n", *inputFile)

	// Read the graph JSON file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// Parse JSON - assuming the format from our exporter
	var exportData struct {
		Graph *analyzer.Graph `json:"graph"`
	}

	if err := json.Unmarshal(data, &exportData); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	if exportData.Graph == nil {
		log.Fatalf("No graph data found in JSON file")
	}

	// Connect to Neo4j
	config := neo4j.Config{
		URI:      *uri,
		Username: *username,
		Password: *password,
		Database: *database,
	}

	client, err := neo4j.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer client.Close()

	// Import the graph
	ctx := context.Background()
	if err := client.ImportGraph(ctx, exportData.Graph); err != nil {
		log.Fatalf("Failed to import graph: %v", err)
	}

	// Get and display statistics
	stats, err := client.GetGraphStats(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get stats: %v", err)
	} else {
		fmt.Printf("\n📊 Import completed successfully!\n")
		fmt.Printf("   Nodes imported: %v\n", stats["total_nodes"])
		fmt.Printf("   Edges imported: %v\n", stats["total_edges"])

		if nodeTypes, ok := stats["node_types"].([]map[string]interface{}); ok {
			fmt.Printf("\n📦 Node types:\n")
			for _, nt := range nodeTypes {
				fmt.Printf("   %s: %v\n", nt["type"], nt["count"])
			}
		}

		if edgeTypes, ok := stats["edge_types"].([]map[string]interface{}); ok {
			fmt.Printf("\n🔗 Edge types:\n")
			for _, et := range edgeTypes {
				fmt.Printf("   %s: %v\n", et["type"], et["count"])
			}
		}
	}

	fmt.Printf("\n🌐 Neo4j Browser: http://localhost:7474\n")
	fmt.Printf("   Username: %s\n", *username)
	fmt.Printf("   Password: %s\n", *password)
	fmt.Printf("\n💡 Try these queries in Neo4j Browser:\n")
	fmt.Printf("   MATCH (n) RETURN n LIMIT 25\n")
	fmt.Printf("   MATCH (n:CodeNode {type: 'package'}) RETURN n\n")
	fmt.Printf("   MATCH (n:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(m:CodeNode) RETURN n, r, m LIMIT 10\n")
}
