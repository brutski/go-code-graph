package analyzer

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/brutski/go-code-graph/internal/embeddings"
)

const (
	// DefaultParallelismLevel is the default number of concurrent embedding operations
	DefaultParallelismLevel = 5

	// ParallelismEnvVar is the environment variable to set parallelism level
	ParallelismEnvVar = "EMBEDDINGS_PARALLELISM_LEVEL"
)

// EmbeddingsGenerator handles embedding generation for nodes
type EmbeddingsGenerator struct {
	client           embeddings.Client
	graph            *Graph // Add graph reference for relationship-aware context
	parallelismLevel int
	hasher           *nodeHasher
	logger           *slog.Logger
}

// NewEmbeddingsGenerator creates a new embeddings generator
func NewEmbeddingsGenerator(client embeddings.Client, logger *slog.Logger) *EmbeddingsGenerator {
	return NewEmbeddingsGeneratorWithConfig(client, DefaultParallelismLevel, logger)
}

// NewEmbeddingsGeneratorWithConfig creates a new embeddings generator with explicit parallelism level
func NewEmbeddingsGeneratorWithConfig(client embeddings.Client, parallelismLevel int, logger *slog.Logger) *EmbeddingsGenerator {
	if parallelismLevel <= 0 {
		parallelismLevel = DefaultParallelismLevel
	}

	return &EmbeddingsGenerator{
		client:           client,
		graph:            nil, // Graph will be set later via SetGraph()
		parallelismLevel: parallelismLevel,
		hasher:           &nodeHasher{},
		logger:           logger,
	}
}

// SetGraph sets the graph for relationship-aware embeddings
func (e *EmbeddingsGenerator) SetGraph(graph *Graph) {
	if e != nil {
		e.graph = graph
	}
}

// IsEnabled returns true if embeddings generation is enabled
func (e *EmbeddingsGenerator) IsEnabled() bool {
	return e != nil && e.client != nil
}

// GetClient returns the underlying embeddings client
func (e *EmbeddingsGenerator) GetClient() embeddings.Client {
	return e.client
}

// GetModelName returns the embedding model name if available
func (e *EmbeddingsGenerator) GetModelName() string {
	if e.IsEnabled() {
		return e.client.GetModelName()
	}
	return ""
}

// GetDimensions returns the embedding dimensions
func (e *EmbeddingsGenerator) GetDimensions() int {
	if e.IsEnabled() {
		return e.client.GetDimensions()
	}
	return 0
}

// CreateEmbedding generates an embedding for the given text
func (e *EmbeddingsGenerator) CreateEmbedding(text string) ([]float32, error) {
	if !e.IsEnabled() {
		return nil, fmt.Errorf("embeddings generator is not enabled")
	}
	return e.client.CreateEmbedding(text)
}

// PrepareNodeForEmbedding prepares a node with semantic summary for potential embedding generation
func (e *EmbeddingsGenerator) PrepareNodeForEmbedding(node *Node) {
	if !e.IsEnabled() {
		return
	}

	// Create semantic summary and store it
	node.SemanticSummary = e.createSemanticSummary(node)
	node.EmbeddingModel = e.client.GetModelName()

	// Add hash metadata for incremental updates
	if e.hasher != nil {
		e.hasher.addHashMetadata(node)
	}
}

// NodeNeedsEmbedding checks if a node needs embedding generation or update
// based on comparing existing and new metadata
func (e *EmbeddingsGenerator) NodeNeedsEmbedding(node *Node, existingMetadata map[string]interface{}) bool {
	if !e.IsEnabled() || node.Metadata == nil {
		return false
	}

	// If no existing metadata, we need to generate embeddings
	if existingMetadata == nil {
		return true
	}

	// Check if node doesn't have embedding
	if hasEmbedding, ok := existingMetadata["has_embedding"].(bool); !ok || !hasEmbedding {
		return true
	}

	// Compare semantic hashes
	existingSemanticHash, hasExisting := existingMetadata["semantic_hash"].(string)
	newSemanticHash, hasNew := node.Metadata["semantic_hash"].(string)

	if !hasExisting || !hasNew {
		return true // Need to generate if we can't compare
	}

	// If semantic hash changed, we need new embedding
	return existingSemanticHash != newSemanticHash
}

// GetNodeMetadata returns the metadata needed for incremental updates
func (e *EmbeddingsGenerator) GetNodeMetadata(node *Node) map[string]interface{} {
	if node.Metadata == nil {
		return nil
	}

	// Return relevant metadata fields for storage
	metadata := make(map[string]interface{})
	if v, ok := node.Metadata["content_hash"]; ok {
		metadata["content_hash"] = v
	}
	if v, ok := node.Metadata["semantic_hash"]; ok {
		metadata["semantic_hash"] = v
	}
	if v, ok := node.Metadata["has_embedding"]; ok {
		metadata["has_embedding"] = v
	}
	if v, ok := node.Metadata["last_analyzed"]; ok {
		metadata["last_analyzed"] = v
	}
	if v, ok := node.Metadata["embedding_generated_at"]; ok {
		metadata["embedding_generated_at"] = v
	}

	return metadata
}

// PrepareNodeForStorage adds hash metadata to a node (V1 compatibility)
func PrepareNodeForStorage(node *Node) {
	hasher := &nodeHasher{}
	hasher.addHashMetadata(node)
}

// GenerateEmbeddingsForNodes generates embeddings for specific nodes concurrently
func (e *EmbeddingsGenerator) GenerateEmbeddingsForNodes(nodes []*Node) error {
	if !e.IsEnabled() || len(nodes) == 0 {
		return nil
	}

	totalNodes := len(nodes)
	e.logger.Debug("Generating embeddings for nodes",
		"nodeCount", totalNodes,
		"workers", e.parallelismLevel)

	// Use a semaphore to limit concurrent calls
	semaphore := make(chan struct{}, e.parallelismLevel)

	// Results channel to collect outcomes
	type result struct {
		nodeID    string
		embedding []float32
		err       error
	}
	resultChan := make(chan result, totalNodes)

	// Use WaitGroup to properly wait for all goroutines
	var wg sync.WaitGroup

	// Process all nodes concurrently
	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n *Node) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			embedding, err := e.client.CreateEmbedding(n.SemanticSummary)

			resultChan <- result{
				nodeID:    n.ID,
				embedding: embedding,
				err:       err,
			}
		}(i, node)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	var successful, failed int
	var errors []error
	nodeEmbeddings := make(map[string][]float32)

	for res := range resultChan {
		if res.err != nil {
			errors = append(errors, fmt.Errorf("node %s: %w", res.nodeID, res.err))
			failed++
		} else {
			nodeEmbeddings[res.nodeID] = res.embedding
			successful++
		}

		// Progress reporting for large datasets
		if totalNodes > 100 && (successful+failed)%100 == 0 {
			e.logger.Debug("Embedding generation progress",
				"processed", successful+failed,
				"total", totalNodes,
				"successful", successful,
				"failed", failed)
		}
	}

	// Report errors
	if len(errors) > 0 {
		e.logger.Warn("Embedding generation failures",
			"errorCount", len(errors),
			"totalNodes", totalNodes)
		for i, err := range errors {
			if i < 5 { // Show first 5 errors
				e.logger.Warn("Embedding error", "error", err)
			}
		}
		if len(errors) > 5 {
			e.logger.Warn("Additional errors omitted", "count", len(errors)-5)
		}
	}

	// Update nodes with successful embeddings
	for _, node := range nodes {
		if embedding, exists := nodeEmbeddings[node.ID]; exists {
			node.Embedding = embedding
			// Update metadata to reflect that embedding was generated
			if e.hasher != nil {
				e.hasher.addHashMetadata(node)
			}
		}
	}

	e.logger.Debug("Completed embedding generation",
		"successful", successful,
		"failed", failed)
	return nil
}

// NodeRelationships holds the relationships for a node
type NodeRelationships struct {
	Calls            []string // Functions this node calls
	CalledBy         []string // Functions that call this node
	Uses             []string // Types this node uses
	UsedBy           []string // Where this node is used
	Implements       []string // Interfaces implemented
	Returns          []string // Types returned
	Imports          []string // Packages imported
	HandlesError     bool     // Whether this node handles errors
	SpawnsGoroutines bool     // Whether this spawns goroutines
}

// getNodeRelationships gets all relationships for a node from the graph
func (e *EmbeddingsGenerator) getNodeRelationships(nodeID string) *NodeRelationships {
	rels := &NodeRelationships{
		Calls:      []string{},
		CalledBy:   []string{},
		Uses:       []string{},
		UsedBy:     []string{},
		Implements: []string{},
		Returns:    []string{},
		Imports:    []string{},
	}

	if e.graph == nil {
		return rels
	}

	// Scan all edges to find relationships
	for _, edge := range e.graph.Edges {
		switch edge.Type {
		case EdgeTypeCalls:
			if edge.Source == nodeID {
				if target := e.findNodeLabel(edge.Target); target != "" {
					rels.Calls = append(rels.Calls, target)
				}
			} else if edge.Target == nodeID {
				if source := e.findNodeLabel(edge.Source); source != "" {
					rels.CalledBy = append(rels.CalledBy, source)
				}
			}
		case EdgeTypeUses:
			if edge.Source == nodeID {
				if target := e.findNodeLabel(edge.Target); target != "" {
					rels.Uses = append(rels.Uses, target)
				}
			} else if edge.Target == nodeID {
				if source := e.findNodeLabel(edge.Source); source != "" {
					rels.UsedBy = append(rels.UsedBy, source)
				}
			}
		case EdgeTypeImplements:
			if edge.Source == nodeID {
				if target := e.findNodeLabel(edge.Target); target != "" {
					rels.Implements = append(rels.Implements, target)
				}
			}
		case EdgeTypeReturns:
			if edge.Source == nodeID {
				if target := e.findNodeLabel(edge.Target); target != "" {
					rels.Returns = append(rels.Returns, target)
				}
			}
		case EdgeTypeImports:
			if edge.Source == nodeID {
				if target := e.findNodeLabel(edge.Target); target != "" {
					rels.Imports = append(rels.Imports, target)
				}
			}
		case EdgeTypeHandlesError:
			if edge.Source == nodeID {
				rels.HandlesError = true
			}
		case EdgeTypeSpawnsGoroutine:
			if edge.Source == nodeID {
				rels.SpawnsGoroutines = true
			}
		}
	}

	return rels
}

// findNodeLabel finds a node's label by ID
func (e *EmbeddingsGenerator) findNodeLabel(nodeID string) string {
	if e.graph == nil {
		return ""
	}

	for i := range e.graph.Nodes {
		if e.graph.Nodes[i].ID == nodeID {
			return e.graph.Nodes[i].Label
		}
	}
	return ""
}

// createSemanticSummary creates a semantic summary of a node for embedding generation
func (e *EmbeddingsGenerator) createSemanticSummary(node *Node) string {
	// Get relationships from the graph
	rels := e.getNodeRelationships(node.ID)

	var parts []string

	// Basic info
	parts = append(parts, fmt.Sprintf("%s %s in %s", node.Type, node.Label, node.Package))

	// Add documentation if available
	if node.Documentation != "" {
		parts = append(parts, cleanDocumentation(node.Documentation))
	}

	// Add signature for functions/methods
	if (node.Type == NodeTypeFunction || node.Type == NodeTypeMethod) && node.Signature != "" {
		parts = append(parts, node.Signature)
	}

	// Add relationship-based context
	if len(rels.Calls) > 0 {
		parts = append(parts, fmt.Sprintf("calls: %s", summarizeList(rels.Calls, 5)))
	}

	if len(rels.Uses) > 0 {
		parts = append(parts, fmt.Sprintf("uses: %s", summarizeList(rels.Uses, 5)))
	}

	if len(rels.Returns) > 0 {
		parts = append(parts, fmt.Sprintf("returns: %s", strings.Join(rels.Returns, ", ")))
	}

	if len(rels.Implements) > 0 {
		parts = append(parts, fmt.Sprintf("implements: %s", strings.Join(rels.Implements, ", ")))
	}

	// Add behavioral info
	if rels.HandlesError {
		parts = append(parts, "handles errors")
	}

	if rels.SpawnsGoroutines {
		parts = append(parts, "spawns goroutines")
	}

	// Add complexity for functions if significant
	if (node.Type == NodeTypeFunction || node.Type == NodeTypeMethod) && node.Complexity > 10 {
		parts = append(parts, fmt.Sprintf("complexity %d", node.Complexity))
	}

	// Add structural info for types
	switch node.Type {
	case NodeTypeStruct:
		if node.TypeInfo.FieldCount > 0 {
			parts = append(parts, fmt.Sprintf("%d fields", node.TypeInfo.FieldCount))
		}
		if len(node.TypeInfo.MethodSet) > 0 {
			methods := []string{}
			for _, m := range node.TypeInfo.MethodSet {
				methods = append(methods, m.Name)
			}
			parts = append(parts, fmt.Sprintf("methods: %s", summarizeList(methods, 5)))
		}

	case NodeTypeInterface:
		if len(node.TypeInfo.MethodSet) > 0 {
			methods := []string{}
			for _, m := range node.TypeInfo.MethodSet {
				methods = append(methods, m.Name)
			}
			parts = append(parts, fmt.Sprintf("methods: %s", strings.Join(methods, ", ")))
		}
	}

	return strings.Join(parts, ". ")
}

// summarizeList truncates a list with a summary if too long
func summarizeList(items []string, max int) string {
	if len(items) <= max {
		return strings.Join(items, ", ")
	}
	return fmt.Sprintf("%s and %d more",
		strings.Join(items[:max], ", "), len(items)-max)
}

// cleanDocumentation removes excess whitespace and truncates very long documentation
func cleanDocumentation(doc string) string {
	// Remove extra whitespace
	doc = strings.TrimSpace(doc)
	doc = strings.ReplaceAll(doc, "\n\n", " ")
	doc = strings.ReplaceAll(doc, "\n", " ")
	doc = strings.ReplaceAll(doc, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(doc, "  ") {
		doc = strings.ReplaceAll(doc, "  ", " ")
	}

	// Truncate if too long
	const maxLen = 200
	if len(doc) > maxLen {
		// Try to cut at sentence boundary
		if idx := strings.LastIndex(doc[:maxLen], ". "); idx > 0 {
			doc = doc[:idx+1]
		} else {
			doc = doc[:maxLen-3] + "..."
		}
	}

	return doc
}
