package exporter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/brutski/go-code-graph/internal/analyzer"
)

// JSONExporter exports graph data to JSON format
type JSONExporter struct{}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{}
}

// Export exports the graph to a JSON file
func (e *JSONExporter) Export(graph *analyzer.Graph, filename string) error {
	if graph == nil {
		return fmt.Errorf("cannot export nil graph")
	}

	// Create the export data structure optimized for visualization
	exportData := &ExportData{
		Graph: graph,
		Config: &VisualizationConfig{
			NodeColors: map[string]string{
				analyzer.NodeTypePackage:   "#4CAF50",
				analyzer.NodeTypeStruct:    "#2196F3",
				analyzer.NodeTypeInterface: "#FF9800",
				analyzer.NodeTypeFunction:  "#9C27B0",
				analyzer.NodeTypeMethod:    "#E91E63",
				analyzer.NodeTypeField:     "#607D8B",
				analyzer.NodeTypeConstant:  "#795548",
				analyzer.NodeTypeVariable:  "#009688",
			},
			EdgeColors: map[string]string{
				analyzer.EdgeTypeCalls:        "#666666",
				analyzer.EdgeTypeImplements:   "#4CAF50",
				analyzer.EdgeTypeImports:      "#2196F3",
				analyzer.EdgeTypeEmbeds:       "#FF5722",
				analyzer.EdgeTypeHasMethod:    "#9C27B0",
				analyzer.EdgeTypeHasField:     "#607D8B",
				analyzer.EdgeTypeUses:         "#795548",
				analyzer.EdgeTypeReturns:      "#009688",
				analyzer.EdgeTypeHasParameter: "#FFC107",
			},
			Layout: LayoutConfig{
				Name:          "cose",
				NodeRepulsion: 8000,
				EdgeLength:    100,
				Gravity:       0.1,
				NumIter:       1000,
				InitialTemp:   200,
				CoolingFactor: 0.95,
				MinTemp:       1.0,
			},
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("Graph exported to %s (%d nodes, %d edges)\n",
		filename, len(graph.Nodes), len(graph.Edges))

	return nil
}

// ExportData contains the complete data structure for visualization
type ExportData struct {
	Graph  *analyzer.Graph      `json:"graph"`
	Config *VisualizationConfig `json:"config"`
}

// VisualizationConfig contains configuration for the web visualization
type VisualizationConfig struct {
	NodeColors map[string]string `json:"node_colors"`
	EdgeColors map[string]string `json:"edge_colors"`
	Layout     LayoutConfig      `json:"layout"`
}

// LayoutConfig contains layout algorithm configuration
type LayoutConfig struct {
	Name          string  `json:"name"`
	NodeRepulsion int     `json:"node_repulsion"`
	EdgeLength    int     `json:"edge_length"`
	Gravity       float64 `json:"gravity"`
	NumIter       int     `json:"num_iter"`
	InitialTemp   float64 `json:"initial_temp"`
	CoolingFactor float64 `json:"cooling_factor"`
	MinTemp       float64 `json:"min_temp"`
}

// ExportSummary creates a summary report of the exported graph
func (e *JSONExporter) ExportSummary(graph *analyzer.Graph, filename string) error {
	if graph == nil {
		return fmt.Errorf("cannot export summary for nil graph")
	}

	summary := map[string]interface{}{
		"total_nodes":    len(graph.Nodes),
		"total_edges":    len(graph.Edges),
		"statistics":     graph.Stats,
		"node_breakdown": graph.Stats.NodesByType,
		"edge_breakdown": graph.Stats.EdgesByType,
		"packages":       make([]PackageSummary, 0),
	}

	// Create package summaries
	packageSummaries := make(map[string]*PackageSummary)

	for _, node := range graph.Nodes {
		if node.Type == analyzer.NodeTypePackage {
			filesCount := ""
			if f, ok := node.Metadata["files"]; ok {
				if val, ok := f.(int); ok {
					filesCount = fmt.Sprintf("%d", val)
				}
			}
			importsCount := ""
			if i, ok := node.Metadata["imports"]; ok {
				if val, ok := i.(int); ok {
					importsCount = fmt.Sprintf("%d", val)
				}
			}

			packageSummaries[node.Package] = &PackageSummary{
				Name:      node.Label,
				Path:      node.Package,
				Files:     filesCount,
				Imports:   importsCount,
				Types:     0,
				Functions: 0,
				Methods:   0,
			}
		}
	}

	// Count types, functions, and methods per package
	for _, node := range graph.Nodes {
		if summary, exists := packageSummaries[node.Package]; exists {
			switch node.Type {
			case analyzer.NodeTypeStruct, analyzer.NodeTypeInterface:
				summary.Types++
			case analyzer.NodeTypeFunction:
				summary.Functions++
			case analyzer.NodeTypeMethod:
				summary.Methods++
			}
		}
	}

	// Convert map to slice
	packages := make([]PackageSummary, 0, len(packageSummaries))
	for _, pkg := range packageSummaries {
		packages = append(packages, *pkg)
	}
	summary["packages"] = packages

	// Marshal and write
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	summaryFilename := filename + ".summary"
	if err := os.WriteFile(summaryFilename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	fmt.Printf("Summary exported to %s\n", summaryFilename)
	return nil
}

// PackageSummary contains summary information about a package
type PackageSummary struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Files     string `json:"files"`
	Imports   string `json:"imports"`
	Types     int    `json:"types"`
	Functions int    `json:"functions"`
	Methods   int    `json:"methods"`
}
