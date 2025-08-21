package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/brutski/go-code-graph/internal/analyzer"
	"github.com/brutski/go-code-graph/internal/exporter"
	"github.com/brutski/go-code-graph/internal/server"
)

func main() {
	var (
		repoPath        = flag.String("repo", ".", "Path to the Go repository to analyze")
		output          = flag.String("output", "graph.json", "Output file path for the graph data")
		format          = flag.String("format", "json", "Output format (json)")
		serve           = flag.Bool("serve", false, "Start HTTP server for visualization")
		port            = flag.Int("port", 8080, "Port for HTTP server")
		summary         = flag.Bool("summary", true, "Generate summary file")
		allowedPackages = flag.String("include-packages", "", "Comma-separated list of external packages to include (e.g., 'github.com/gin-gonic/gin/...')")
	)
	flag.Parse()

	logger := slog.Default()

	// Validate repository path
	absRepoPath, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatalf("Invalid repository path: %v", err)
	}

	if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
		log.Fatalf("Repository path does not exist: %s", absRepoPath)
	}

	logger.Debug(fmt.Sprintf("🔍 Analyzing Go codebase at: %s", absRepoPath))

	// Parse allowed packages list
	var allowedPkgs []string
	if *allowedPackages != "" {
		allowedPkgs = strings.Split(*allowedPackages, ",")
		for i := range allowedPkgs {
			allowedPkgs[i] = strings.TrimSpace(allowedPkgs[i])
		}
		logger.Info(fmt.Sprintf("📦 Including external packages: %v", allowedPkgs))
	}

	// Create parser and analyze codebase (without embeddings for this CLI tool)
	parser := analyzer.NewParser(nil, logger)
	parser.SetAllowedPackages(allowedPkgs)
	graph, err := parser.AnalyzeCodebase(absRepoPath)
	if err != nil {
		log.Fatalf("Failed to analyze codebase: %v", err)
	}

	printStats(logger, graph)

	// Export graph data
	if err := exportGraph(graph, *output, *format, *summary); err != nil {
		log.Fatalf("Failed to export graph: %v", err)
	}

	// Show completion message (always shown)
	logger.Info(fmt.Sprintf("✅ Analysis complete: %s", *output))

	// Start web server if requested
	if *serve {
		logger.Info(fmt.Sprintf("🌐 Starting visualization server at http://localhost:%d", *port))
		srv := server.NewServer(*output)
		if err := srv.Start(*port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

// printStats prints basic statistics about the analyzed graph
func printStats(logger *slog.Logger, graph *analyzer.Graph) {
	logger.Info("\n📊 Analysis Results:\n")
	logger.Info(fmt.Sprintf("   Nodes: %d", len(graph.Nodes)))
	logger.Info(fmt.Sprintf("   Edges: %d", len(graph.Edges)))
	logger.Info(fmt.Sprintf("   Packages: %d", graph.Stats.PackageCount))

	logger.Info("\n📦 Node Types:\n")
	for nodeType, count := range graph.Stats.NodesByType {
		if count > 0 {
			logger.Info(fmt.Sprintf("   %s: %d", nodeType, count))
		}
	}

	logger.Info("\n🔗 Edge Types:\n")
	for edgeType, count := range graph.Stats.EdgesByType {
		if count > 0 {
			logger.Info(fmt.Sprintf("   %s: %d", edgeType, count))
		}
	}

	if len(graph.Stats.Dependencies) > 0 {
		logger.Info(fmt.Sprintf("\n📋 External Dependencies: %d", len(graph.Stats.Dependencies)))
		if len(graph.Stats.Dependencies) <= 10 {
			for _, dep := range graph.Stats.Dependencies {
				logger.Info(fmt.Sprintf("   - %s", dep))
			}
		} else {
			logger.Info("   (First 5 shown)")
			for i := 0; i < 5; i++ {
				logger.Info(fmt.Sprintf("   - %s", graph.Stats.Dependencies[i]))
			}
			logger.Info(fmt.Sprintf("   ... and %d more", len(graph.Stats.Dependencies)-5))
		}
	}
}

// exportGraph exports the graph in the specified format
func exportGraph(graph *analyzer.Graph, outputPath, format string, includeSummary bool) error {
	switch format {
	case "json":
		jsonExporter := exporter.NewJSONExporter()

		if err := jsonExporter.Export(graph, outputPath); err != nil {
			return fmt.Errorf("failed to export JSON: %w", err)
		}

		if includeSummary {
			if err := jsonExporter.ExportSummary(graph, outputPath); err != nil {
				return fmt.Errorf("failed to export summary: %w", err)
			}
		}

		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
