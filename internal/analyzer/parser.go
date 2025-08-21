package analyzer

import (
	"fmt"
	"go/types"
	"log/slog"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Parser orchestrates the v2 code analysis
type Parser struct {
	graph                *Graph
	nodeBuilder          *NodeBuilder
	relationshipAnalyzer *RelationshipAnalyzer
	embeddingsGenerator  *EmbeddingsGenerator
	logger               *slog.Logger
	allowedPackages      []string // List of external packages to include
}

// Ensure Parser implements ParserInterface at compile time
var _ ParserInterface = (*Parser)(nil)

// Note: NewParser is now defined in parser_compat.go for V1 API compatibility
// This file contains the core V2 parser implementation

// SetAllowedPackages sets the list of external packages to include in analysis
func (p *Parser) SetAllowedPackages(packages []string) {
	p.allowedPackages = packages
}

// AnalyzeCodebase analyzes a Go codebase and builds a graph
func (p *Parser) AnalyzeCodebase(rootPath string) (*Graph, error) {
	p.logger.Info("Starting codebase analysis", "rootPath", rootPath)

	// Configure packages loading with all necessary information
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedModule,
		Dir: rootPath,
	}

	// Load all packages from the current module
	loadPatterns := []string{"./..."}

	// Add allowed external packages to load patterns
	if len(p.allowedPackages) > 0 {
		loadPatterns = append(loadPatterns, p.allowedPackages...)
	}

	pkgs, err := packages.Load(cfg, loadPatterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	p.logger.Debug("Loaded packages", "count", len(pkgs))
	for _, pkg := range pkgs {
		p.logger.Debug("Package loaded", "path", pkg.PkgPath, "name", pkg.Name, "errors", len(pkg.Errors))
	}

	// Filter valid packages (same logic as v1)
	validPackages := p.filterValidPackages(pkgs)
	p.logger.Debug("Valid packages after filtering", "count", len(validPackages))

	// Phase 1: Create all nodes in a single pass
	p.logger.Debug("Phase 1: Building nodes", "packageCount", len(validPackages))

	// First, ensure builtin types exist
	p.nodeBuilder.EnsureBuiltinTypes()
	p.logger.Debug("Created builtin type nodes")

	for _, pkg := range validPackages {
		visitor := NewASTVisitor(pkg, p.nodeBuilder, p.relationshipAnalyzer)
		visitor.BuildNodes()
	}

	// Build node index
	nodeIndex := make(map[string]bool)
	for _, node := range p.graph.Nodes {
		nodeIndex[node.ID] = true
	}

	// Phase 2: Analyze relationships across all packages
	p.logger.Debug("Phase 2: Analyzing relationships")
	p.relationshipAnalyzer.SetNodeIndex(nodeIndex)
	for _, pkg := range validPackages {
		p.relationshipAnalyzer.AddPackage(pkg)
	}
	p.relationshipAnalyzer.AnalyzeAllRelationships()

	// Phase 3: Pattern detection is now integrated into relationship analysis
	// Constructor patterns are detected in analyzeConstructorPatterns
	// Other patterns have been removed as rarely used

	// Phase 4: Prepare nodes for embedding (but don't generate yet)
	p.logger.Debug("Phase 4: Preparing semantic summaries")
	p.prepareSemanticSummaries()

	// Update statistics
	p.updateStats(validPackages)

	// Edge type normalization no longer needed

	// Phase 5: Detect cross-package interface implementations
	p.logger.Debug("Phase 5: Detecting cross-package interface implementations")
	p.detectCrossPackageInterfaceImplementations(validPackages)

	p.logger.Info("Analysis complete",
		"nodeCount", len(p.graph.Nodes),
		"edgeCount", len(p.graph.Edges))

	return p.graph, nil
}

// GenerateEmbeddingsForNodes generates embeddings for specific nodes
func (p *Parser) GenerateEmbeddingsForNodes(nodes []*EnhancedNode) error {
	if p.embeddingsGenerator == nil || !p.embeddingsGenerator.IsEnabled() || len(nodes) == 0 {
		return nil
	}

	// Since Node is a type alias for EnhancedNode, we can pass them directly
	return p.embeddingsGenerator.GenerateEmbeddingsForNodes(nodes)
}

// GetGraph returns the current graph
func (p *Parser) GetGraph() *Graph {
	return p.graph
}

// detectCrossPackageInterfaceImplementations detects interface implementations across packages
func (p *Parser) detectCrossPackageInterfaceImplementations(packages []*packages.Package) {
	// Collect all interfaces and concrete types
	var interfaces []*types.Named
	var interfaceNodes []string
	var concreteTypes []*types.Named

	// Build a node index for quick lookup
	nodeIndex := make(map[string]bool)
	for _, node := range p.graph.Nodes {
		nodeIndex[node.ID] = true
	}

	// Collect interfaces and concrete types from all packages
	for _, pkg := range packages {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if typeName, ok := obj.(*types.TypeName); ok {
				// Get the actual type, handling aliases
				typ := typeName.Type()
				if alias, ok := typ.(*types.Alias); ok {
					typ = alias.Underlying()
				}

				// Only process Named types
				named, isNamed := typ.(*types.Named)
				if !isNamed {
					continue
				}

				if _, ok := named.Underlying().(*types.Interface); ok {
					// It's an interface
					interfaces = append(interfaces, named)
					nodeID := GenerateNodeID(NodeTypeInterface, pkg.PkgPath, typeName.Name())
					interfaceNodes = append(interfaceNodes, nodeID)
				} else if _, ok := named.Underlying().(*types.Struct); ok {
					// It's a struct
					concreteTypes = append(concreteTypes, named)
				}
			}
		}
	}

	// Check each concrete type against each interface
	implementationsFound := 0
	for i, iface := range interfaces {
		ifaceNodeID := interfaceNodes[i]
		if !nodeIndex[ifaceNodeID] {
			continue
		}

		for _, concrete := range concreteTypes {
			concretePkg := ""
			if concrete.Obj().Pkg() != nil {
				concretePkg = concrete.Obj().Pkg().Path()
			}
			concreteNodeID := GenerateNodeID(NodeTypeStruct, concretePkg, concrete.Obj().Name())

			if !nodeIndex[concreteNodeID] {
				continue
			}

			// Skip if same package (already handled by per-package analysis)
			if concrete.Obj().Pkg() != nil && iface.Obj().Pkg() != nil &&
				concrete.Obj().Pkg().Path() == iface.Obj().Pkg().Path() {
				continue
			}

			// Check if the concrete type implements the interface
			ifaceType := iface.Underlying().(*types.Interface)
			if types.Implements(concrete, ifaceType) || types.Implements(types.NewPointer(concrete), ifaceType) {
				edge := Edge{
					Source: concreteNodeID,
					Target: ifaceNodeID,
					Type:   EdgeTypeImplements,
					Weight: 2,
				}
				p.graph.AddEdge(edge)
				implementationsFound++
			}
		}
	}

	p.logger.Info("Cross-package interface implementation detection complete",
		"implementationsFound", implementationsFound)
}

// filterValidPackages filters out packages with errors and external packages
func (p *Parser) filterValidPackages(pkgs []*packages.Package) []*packages.Package {
	// First, determine the module path
	var modulePath string
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			modulePath = pkg.Module.Path
			break
		}
	}
	p.logger.Debug("Determined module path", "modulePath", modulePath)

	var validPackages []*packages.Package
	for _, pkg := range pkgs {
		// Skip packages with errors
		if len(pkg.Errors) > 0 {
			p.logger.Debug("Skipping package with errors", "path", pkg.PkgPath, "errors", len(pkg.Errors))
			continue
		}

		// Skip external packages unless they're in the allowed list
		if isExternalPackageWithModule(pkg.PkgPath, modulePath) {
			// Check if this package is in the allowed list
			isAllowed := false
			for _, allowed := range p.allowedPackages {
				// Handle patterns like "github.com/gin-gonic/gin/..."
				allowedBase := strings.TrimSuffix(allowed, "/...")
				if pkg.PkgPath == allowedBase || strings.HasPrefix(pkg.PkgPath, allowedBase+"/") {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				p.logger.Debug("Skipping external package", "path", pkg.PkgPath, "modulePath", modulePath)
				continue
			} else {
				p.logger.Debug("Including allowed external package", "path", pkg.PkgPath)
			}
		}

		validPackages = append(validPackages, pkg)
	}

	return validPackages
}

// prepareSemanticSummaries prepares nodes for embedding generation
func (p *Parser) prepareSemanticSummaries() {
	if p.embeddingsGenerator == nil {
		return
	}

	for i := range p.graph.Nodes {
		node := &p.graph.Nodes[i]

		// Only prepare summaries for high-value nodes
		switch node.Type {
		case NodeTypeFunction, NodeTypeMethod,
			NodeTypeStruct, NodeTypeInterface:
			p.embeddingsGenerator.PrepareNodeForEmbedding(node)
		}
	}
}

// updateStats updates graph statistics
func (p *Parser) updateStats(packages []*packages.Package) {
	// Count dependencies
	deps := make(map[string]bool)
	for _, pkg := range packages {
		for path := range pkg.Imports {
			if isExternalPackageWithModule(path, "") {
				deps[path] = true
			}
		}
	}

	p.graph.Stats.Dependencies = make([]string, 0, len(deps))
	for dep := range deps {
		p.graph.Stats.Dependencies = append(p.graph.Stats.Dependencies, dep)
	}

	p.graph.Stats.PackageCount = p.graph.Stats.NodesByType[NodeTypePackage]
	p.graph.Stats.MaxDepth = 3 // packages -> types -> functions
}

// normalizeEdgeTypes removed - edge type migration no longer needed
// func (p *Parser) normalizeEdgeTypes() {
// 	// EdgeTypeCreates has been removed, use EdgeTypeConstructs directly
// }

// NewParser is the actual V2 parser constructor
// We rename the original NewParser to avoid conflicts
func NewParser(embeddingsGenerator *EmbeddingsGenerator, logger *slog.Logger) *Parser {
	// Use default logger if none provided
	if logger == nil {
		logger = slog.Default()
	}

	graph := NewGraph()
	nodeBuilder := NewNodeBuilder(graph)
	relationshipAnalyzer := NewRelationshipAnalyzer(graph, nodeBuilder)

	// Set the graph on the embeddings generator for relationship-aware context
	if embeddingsGenerator != nil {
		embeddingsGenerator.SetGraph(graph)
	}

	return &Parser{
		graph:                graph,
		nodeBuilder:          nodeBuilder,
		relationshipAnalyzer: relationshipAnalyzer,
		embeddingsGenerator:  embeddingsGenerator,
		logger:               logger,
	}
}
