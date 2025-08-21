package analyzer

// ParserInterface defines the contract for all parser implementations
type ParserInterface interface {
	// AnalyzeCodebase analyzes a Go codebase and builds a graph
	AnalyzeCodebase(rootPath string) (*Graph, error)

	// GenerateEmbeddingsForNodes generates embeddings for specific nodes
	GenerateEmbeddingsForNodes(nodes []*EnhancedNode) error

	// GetGraph returns the current graph
	GetGraph() *Graph
}
