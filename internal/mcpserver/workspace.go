package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/brutski/go-code-graph/internal/analyzer"
)

// convertFloat32ToFloat64 converts []float32 to []float64 for Neo4j compatibility
func convertFloat32ToFloat64(f32 []float32) []float64 {
	if f32 == nil {
		return nil
	}
	f64 := make([]float64, len(f32))
	for i, v := range f32 {
		f64[i] = float64(v)
	}
	return f64
}

// AnalyzeWorkspaceParams represents the parameters for workspace analysis
type AnalyzeWorkspaceParams struct {
	// WorkspacePath is the path to the workspace/repository to analyze
	WorkspacePath string `json:"workspacePath,omitempty"`
	// WorkspaceName is the name to use for this workspace
	WorkspaceName string `json:"workspaceName,omitempty"`
	// Incremental determines if this is an incremental analysis
	Incremental bool `json:"incremental,omitempty"`
	// AllowedPackages is a list of external packages to include in analysis
	AllowedPackages []string `json:"allowedPackages,omitempty"`
}

// handleAnalyzeWorkspace performs complete workspace analysis and import
func (s *Server) handleAnalyzeWorkspace(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[AnalyzeWorkspaceParams]) (*mcp.CallToolResultFor[any], error) {
	// 1. Validate and setup parameters
	workspacePath, workspaceName, err := s.validateAndSetupParams(params.Arguments)
	if err != nil {
		return s.createErrorResponse(err), nil
	}

	// 2. Prepare for incremental update if needed
	existingNodes, err := s.prepareIncrementalUpdate(ctx, workspaceName, params.Arguments.Incremental)
	if err != nil {
		s.logger.Debug("Failed to prepare incremental update", "error", err)
		// Continue anyway - worst case we regenerate everything
	}

	// 3. Analyze codebase and create graph
	start := time.Now()
	graph, parser, err := s.analyzeAndCreateGraph(workspacePath, params.Arguments.AllowedPackages)
	if err != nil {
		return s.createErrorResponse(err), nil
	}

	// 4. Add workspace context to all nodes and edges
	s.addWorkspaceContext(graph, workspaceName)

	// 5. Process embeddings
	err = s.processEmbeddings(parser, graph, params.Arguments.Incremental, existingNodes)
	if err != nil {
		s.logger.Warn("Failed to process embeddings", "error", err)
		// Continue - embeddings are not critical
	}

	analysisTime := time.Since(start)

	// 6. Import to Neo4j and handle cleanup
	importStart := time.Now()
	removedCount, err := s.importToNeo4jWithCleanup(ctx, graph, params.Arguments.Incremental, workspaceName)
	if err != nil {
		return s.createErrorResponse(err), nil
	}

	importTime := time.Since(importStart)
	totalTime := time.Since(start)

	// 7. Store workspace metadata
	if err := s.storeWorkspaceInfo(ctx, graph, workspaceName, workspacePath, analysisTime, importTime, totalTime, params.Arguments.Incremental); err != nil {
		s.logger.Warn("Failed to store workspace metadata", "error", err)
	}

	// 8. Set as current workspace
	s.setCurrentWorkspace(workspaceName)

	// 9. Format and return response
	return s.formatAnalysisResponse(graph, workspaceName, workspacePath, analysisTime, importTime, totalTime, params.Arguments.Incremental, removedCount), nil
}

// detectWorkspaceName auto-detects workspace name from go.mod or directory
func (s *Server) detectWorkspaceName(workspacePath string) (string, error) {
	// Try to read go.mod first
	goModPath := filepath.Join(workspacePath, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
				// Extract last part of module path for workspace name
				parts := strings.Split(moduleName, "/")
				workspaceName := parts[len(parts)-1]
				return s.sanitizeWorkspaceName(workspaceName), nil
			}
		}
	}

	// Fall back to directory name
	dirName := filepath.Base(workspacePath)
	return s.sanitizeWorkspaceName(dirName), nil
}

// sanitizeWorkspaceName ensures workspace name is valid for Neo4j
func (s *Server) sanitizeWorkspaceName(name string) string {
	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := reg.ReplaceAllString(name, "_")

	// Ensure it doesn't start with a number
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "ws_" + sanitized
	}

	if sanitized == "" {
		sanitized = "workspace"
	}

	return sanitized
}

// addWorkspaceContext adds workspace information to all nodes and edges
func (s *Server) addWorkspaceContext(graph *analyzer.Graph, workspaceName string) {
	// Add workspace to all nodes
	for i := range graph.Nodes {
		if graph.Nodes[i].Metadata == nil {
			graph.Nodes[i].Metadata = make(map[string]interface{})
		}
		graph.Nodes[i].Metadata["workspace"] = workspaceName

		// Prefix node ID with workspace
		originalID := graph.Nodes[i].ID
		graph.Nodes[i].ID = fmt.Sprintf("%s:%s", workspaceName, originalID)
	}

	// Update edge references to use new node IDs
	for i := range graph.Edges {
		graph.Edges[i].Source = fmt.Sprintf("%s:%s", workspaceName, graph.Edges[i].Source)
		graph.Edges[i].Target = fmt.Sprintf("%s:%s", workspaceName, graph.Edges[i].Target)

		if graph.Edges[i].Metadata == nil {
			graph.Edges[i].Metadata = make(map[string]interface{})
		}
		graph.Edges[i].Metadata["workspace"] = workspaceName
	}
}

// importGraphToNeo4j imports graph using direct Neo4j driver calls
func (s *Server) importGraphToNeo4j(ctx context.Context, graph *analyzer.Graph, incremental bool) (*ImportStats, error) {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Import nodes in batches
	batchSize := 500 // Increased for better performance
	s.logger.Debug("Starting Neo4j node import", "totalNodes", len(graph.Nodes), "totalEdges", len(graph.Edges), "batchSize", batchSize)

	nodeStartTime := time.Now()
	for i := 0; i < len(graph.Nodes); i += batchSize {
		end := i + batchSize
		if end > len(graph.Nodes) {
			end = len(graph.Nodes)
		}

		batch := graph.Nodes[i:end]
		params := make([]map[string]interface{}, len(batch))
		for j, node := range batch {
			params[j] = map[string]interface{}{
				"id":                     node.ID,
				"label":                  node.Label,
				"type":                   node.Type,
				"package":                node.Package,
				"full_name":              node.FullName,
				"size":                   node.Size,
				"level":                  node.Level,
				"visibility":             node.Visibility,
				"complexity":             node.Complexity,
				"workspace":              node.Metadata["workspace"],
				"filename":               node.Position.Filename,
				"line":                   node.Position.Line,
				"column":                 node.Position.Column,
				"signature":              node.Signature,
				"documentation":          node.Documentation,
				"semantic_summary":       node.SemanticSummary,
				"content_hash":           node.Metadata["content_hash"],
				"semantic_hash":          node.Metadata["semantic_hash"],
				"has_embedding":          node.Metadata["has_embedding"],
				"last_analyzed":          node.Metadata["last_analyzed"],
				"embedding_generated_at": node.Metadata["embedding_generated_at"],
				"embedding":              convertFloat32ToFloat64(node.Embedding),
				"embedding_model":        node.EmbeddingModel,
			}
		}

		// Choose query based on incremental flag
		var query string
		if incremental {
			// Use MERGE for incremental updates
			query = `
				UNWIND $nodes AS node
				MERGE (n:CodeNode {id: node.id})
				ON CREATE SET n.created_at = timestamp()
				ON MATCH SET n.updated_at = timestamp()
				SET n.label = node.label,
				    n.type = node.type,
				    n.package = node.package,
				    n.full_name = node.full_name,
				    n.size = node.size,
				    n.level = node.level,
				    n.visibility = node.visibility,
				    n.complexity = node.complexity,
				    n.workspace = node.workspace,
				    n.filename = node.filename,
				    n.line = node.line,
				    n.column = node.column,
				    n.signature = node.signature,
				    n.documentation = node.documentation,
				    n.semantic_summary = node.semantic_summary,
				    n.content_hash = node.content_hash,
				    n.semantic_hash = node.semantic_hash,
				    n.has_embedding = node.has_embedding,
				    n.last_analyzed = node.last_analyzed,
				    n.embedding_generated_at = node.embedding_generated_at,
				    n.embedding = node.embedding,
				    n.embedding_model = node.embedding_model,
				    n.stale = false,
				    n.last_seen = timestamp()
			`
		} else {
			// Use CREATE for full rebuilds
			query = `
				UNWIND $nodes AS node
				CREATE (n:CodeNode {
					id: node.id,
					label: node.label,
					type: node.type,
					package: node.package,
					full_name: node.full_name,
					size: node.size,
					level: node.level,
					visibility: node.visibility,
					complexity: node.complexity,
					workspace: node.workspace,
					filename: node.filename,
					line: node.line,
					column: node.column,
					signature: node.signature,
					documentation: node.documentation,
					semantic_summary: node.semantic_summary,
					content_hash: node.content_hash,
					semantic_hash: node.semantic_hash,
					has_embedding: node.has_embedding,
					last_analyzed: node.last_analyzed,
					embedding_generated_at: node.embedding_generated_at,
					embedding: node.embedding,
					embedding_model: node.embedding_model,
					created_at: timestamp()
				})
			`
		}

		_, err := session.Run(ctx, query, map[string]interface{}{
			"nodes": params,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to import node batch: %w", err)
		}

		// Log progress every 10 batches or on last batch
		if (i/batchSize+1)%10 == 0 || end == len(graph.Nodes) {
			progress := float64(end) / float64(len(graph.Nodes)) * 100
			elapsed := time.Since(nodeStartTime)
			s.logger.Debug("Neo4j node import progress",
				"processed", end,
				"total", len(graph.Nodes),
				"progress", fmt.Sprintf("%.1f%%", progress),
				"elapsed", elapsed.Round(time.Millisecond))
		}
	}

	s.logger.Debug("Neo4j node import complete", "duration", time.Since(nodeStartTime).Round(time.Millisecond))

	// Import edges in batches
	edgeStartTime := time.Now()
	s.logger.Debug("Starting Neo4j edge import", "totalEdges", len(graph.Edges))

	for i := 0; i < len(graph.Edges); i += batchSize {
		end := i + batchSize
		if end > len(graph.Edges) {
			end = len(graph.Edges)
		}

		batch := graph.Edges[i:end]
		params := make([]map[string]interface{}, len(batch))
		for j, edge := range batch {
			params[j] = map[string]interface{}{
				"source":    edge.Source,
				"target":    edge.Target,
				"type":      edge.Type,
				"weight":    edge.Weight,
				"workspace": edge.Metadata["workspace"],
			}
		}

		query := `
			UNWIND $edges AS edge
			MATCH (source:CodeNode {id: edge.source})
			MATCH (target:CodeNode {id: edge.target})
			CREATE (source)-[r:RELATES_TO {
				type: edge.type,
				weight: edge.weight,
				workspace: edge.workspace
			}]->(target)
		`

		_, err := session.Run(ctx, query, map[string]interface{}{
			"edges": params,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to import edge batch: %w", err)
		}

		// Log progress every 10 batches or on last batch
		if (i/batchSize+1)%10 == 0 || end == len(graph.Edges) {
			progress := float64(end) / float64(len(graph.Edges)) * 100
			elapsed := time.Since(edgeStartTime)
			s.logger.Debug("Neo4j edge import progress",
				"processed", end,
				"total", len(graph.Edges),
				"progress", fmt.Sprintf("%.1f%%", progress),
				"elapsed", elapsed.Round(time.Millisecond))
		}
	}

	s.logger.Debug("Neo4j edge import complete", "duration", time.Since(edgeStartTime).Round(time.Millisecond))

	// Return ImportStats - for now just return nil since we're not tracking stats yet
	return &ImportStats{}, nil
}

// clearWorkspaceData removes existing workspace data from Neo4j
func (s *Server) clearWorkspaceData(ctx context.Context, workspaceName string) error {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := "MATCH (n {workspace: $workspace}) DETACH DELETE n"
	_, err := session.Run(ctx, query, map[string]any{
		"workspace": workspaceName,
	})

	if err == nil {
		s.logger.Debug("Cleared existing data for workspace", "workspace", workspaceName)
	}

	return err
}

// countNodesByType counts nodes of a specific type
func (s *Server) countNodesByType(graph *analyzer.Graph, nodeType string) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Type == nodeType {
			count++
		}
	}
	return count
}

// countNodesWithEmbeddingsByType counts nodes of a specific type that have embeddings
func (s *Server) countNodesWithEmbeddingsByType(graph *analyzer.Graph, nodeType string) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Type == nodeType && len(node.Embedding) > 0 {
			count++
		}
	}
	return count
}

// storeWorkspaceMetadata stores workspace information in Neo4j
func (s *Server) storeWorkspaceMetadata(ctx context.Context, metadata map[string]any) error {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MERGE (w:Workspace {name: $workspace})
		SET w += $metadata
		RETURN w
	`

	_, err := session.Run(ctx, query, map[string]any{
		"workspace": metadata["workspace"],
		"metadata":  metadata,
	})

	return err
}

// setCurrentWorkspace sets the current active workspace (in-memory for now)
func (s *Server) setCurrentWorkspace(workspaceName string) {
	// For now, we'll add this to the server config
	// In a full implementation, this could be stored in Neo4j or a separate state store
	s.logger.Debug("Set current workspace", "workspace", workspaceName)
}

// ensureIndexesExist creates necessary Neo4j indexes for optimal performance
func (s *Server) ensureIndexesExist(ctx context.Context) error {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Get dynamic vector dimensions from embedding client
	vectorDimensions := 1024 // Default for Titan
	if s.embeddingsGenerator != nil {
		vectorDimensions = s.embeddingsGenerator.GetDimensions()
	}

	// Index creation queries from setup_neo4j_indexes.cypher
	indexQueries := []string{
		// Vector index for embeddings (if they exist)
		fmt.Sprintf(`CREATE VECTOR INDEX code_embeddings_index IF NOT EXISTS
		FOR (n:CodeNode) 
		ON n.embedding
		OPTIONS {
		  indexConfig: {
		    `+"`vector.dimensions`"+`: %d,
		    `+"`vector.similarity_function`"+`: 'cosine'
		  }
		}`, vectorDimensions),

		// Composite index for efficient filtering during hybrid search
		`CREATE INDEX code_type_package_index IF NOT EXISTS
		FOR (n:CodeNode) 
		ON (n.type, n.package)`,

		// Full-text search index
		`CREATE FULLTEXT INDEX code_fulltext_index IF NOT EXISTS
		FOR (n:CodeNode)
		ON EACH [n.label, n.full_name, n.semantic_summary]`,

		// Complexity-based filtering
		`CREATE INDEX code_complexity_index IF NOT EXISTS
		FOR (n:CodeNode)
		ON n.complexity`,

		// Relationship types for efficient traversal
		`CREATE INDEX relationship_type_index IF NOT EXISTS
		FOR ()-[r:RELATES_TO]-()
		ON r.type`,

		// Unique constraint for node IDs
		`CREATE CONSTRAINT unique_node_id IF NOT EXISTS
		FOR (n:CodeNode)
		REQUIRE n.id IS UNIQUE`,
	}

	// Execute each index creation query
	for _, query := range indexQueries {
		_, err := session.Run(ctx, query, nil)
		if err != nil {
			s.logger.Debug("Index creation query failed (may already exist)", "error", err)
			// Continue with other indexes even if one fails
		}
	}

	s.logger.Debug("Completed index creation process")
	return nil
}

// prepareNodesWithHashes adds hash metadata to all nodes for change detection
func (s *Server) prepareNodesWithHashes(graph *analyzer.Graph) {
	for i := range graph.Nodes {
		analyzer.PrepareNodeForStorage(&graph.Nodes[i])
	}
}

// ImportStats tracks statistics for the import process
type ImportStats struct {
	Created int
	Updated int
}

// ExistingNodeInfo stores minimal information about existing nodes for incremental updates
type ExistingNodeInfo struct {
	ID             string
	SemanticHash   string
	Embedding      []float64
	EmbeddingModel string
}

// markAllNodesAsStale marks all nodes in a workspace as potentially stale
func (s *Server) markAllNodesAsStale(ctx context.Context, workspaceName string) error {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (n:CodeNode {workspace: $workspace})
		SET n.stale = true, n.last_seen = timestamp()
		RETURN count(n) as marked
	`

	result, err := session.Run(ctx, query, map[string]any{
		"workspace": workspaceName,
	})
	if err != nil {
		return fmt.Errorf("failed to mark nodes as stale: %w", err)
	}

	if result.Next(ctx) {
		marked := result.Record().Values[0].(int64)
		s.logger.Debug("Marked nodes as stale", "workspace", workspaceName, "count", marked)
	}

	return nil
}

// removeStaleNodes removes nodes that are still marked as stale after incremental update
func (s *Server) removeStaleNodes(ctx context.Context, workspaceName string) (int, error) {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// First, get a preview of what will be deleted for logging
	previewQuery := `
		MATCH (n:CodeNode {workspace: $workspace, stale: true})
		RETURN n.type as type, count(n) as count
		ORDER BY type
	`

	previewResult, err := session.Run(ctx, previewQuery, map[string]any{
		"workspace": workspaceName,
	})
	if err == nil && previewResult != nil {
		s.logger.Debug("Stale nodes to be removed by type", "workspace", workspaceName)
		for previewResult.Next(ctx) {
			record := previewResult.Record()
			nodeType := record.Values[0].(string)
			count := record.Values[1].(int64)
			s.logger.Debug("  ", "type", nodeType, "count", count)
		}
	}

	// Now delete the stale nodes and their relationships
	deleteQuery := `
		MATCH (n:CodeNode {workspace: $workspace, stale: true})
		WITH n, n.id as nodeId, n.type as nodeType
		DETACH DELETE n
		RETURN count(n) as removed
	`

	result, err := session.Run(ctx, deleteQuery, map[string]any{
		"workspace": workspaceName,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to remove stale nodes: %w", err)
	}

	var removedCount int
	if result.Next(ctx) {
		removedCount = int(result.Record().Values[0].(int64))
		s.logger.Debug("Removed stale nodes", "workspace", workspaceName, "count", removedCount)
	}

	return removedCount, nil
}

// fetchExistingNodes fetches existing nodes with their semantic hashes and embeddings
func (s *Server) fetchExistingNodes(ctx context.Context, workspaceName string) (map[string]*ExistingNodeInfo, error) {
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (n:CodeNode {workspace: $workspace})
		WHERE n.semantic_hash IS NOT NULL 
		AND n.type IN ['function', 'method', 'struct', 'interface']
		RETURN n.id as id, 
		       n.semantic_hash as semantic_hash,
		       n.embedding as embedding,
		       n.embedding_model as embedding_model
	`

	result, err := session.Run(ctx, query, map[string]any{
		"workspace": workspaceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing nodes: %w", err)
	}

	existingNodes := make(map[string]*ExistingNodeInfo)
	for result.Next(ctx) {
		record := result.Record()
		nodeID := record.Values[0].(string)

		existingNode := &ExistingNodeInfo{
			ID: nodeID,
		}

		// Semantic hash
		if semanticHash, ok := record.Values[1].(string); ok {
			existingNode.SemanticHash = semanticHash
		}

		// Embedding (may be nil)
		if embedding, ok := record.Values[2].([]float64); ok {
			existingNode.Embedding = embedding
		}

		// Embedding model
		if embeddingModel, ok := record.Values[3].(string); ok {
			existingNode.EmbeddingModel = embeddingModel
		}

		existingNodes[nodeID] = existingNode
	}

	s.logger.Debug("Fetched existing nodes for incremental update",
		"workspace", workspaceName,
		"count", len(existingNodes))

	return existingNodes, nil
}

// generateEmbeddingsForDelta generates embeddings only for new or changed nodes
func (s *Server) generateEmbeddingsForDelta(parser *analyzer.Parser, graph *analyzer.Graph, existingNodes map[string]*ExistingNodeInfo) error {
	if s.embeddingsGenerator == nil {
		s.logger.Debug("No embedding client configured, skipping embedding generation")
		return nil
	}

	nodesToEmbed := make([]*analyzer.EnhancedNode, 0)
	preservedCount := 0

	for i := range graph.Nodes {
		node := &graph.Nodes[i]

		// Skip nodes that don't need embeddings
		if node.Type != analyzer.NodeTypeFunction && node.Type != analyzer.NodeTypeMethod &&
			node.Type != analyzer.NodeTypeStruct && node.Type != analyzer.NodeTypeInterface {
			continue
		}

		// Get current semantic hash
		currentSemanticHash := ""
		if hash, ok := node.Metadata["semantic_hash"].(string); ok {
			currentSemanticHash = hash
		}

		// Check if this node exists in the database with unchanged semantic hash
		if existingNode, exists := existingNodes[node.ID]; exists {
			if currentSemanticHash == existingNode.SemanticHash && len(existingNode.Embedding) > 0 {
				// Preserve existing embedding
				embedding := make([]float32, len(existingNode.Embedding))
				for j, v := range existingNode.Embedding {
					embedding[j] = float32(v)
				}
				node.Embedding = embedding
				node.EmbeddingModel = existingNode.EmbeddingModel
				preservedCount++
				continue
			}
		}

		// Node is new or changed - needs embedding generation
		nodesToEmbed = append(nodesToEmbed, node)
	}

	s.logger.Debug("Delta embedding analysis",
		"totalNodes", len(graph.Nodes),
		"preserved", preservedCount,
		"toGenerate", len(nodesToEmbed))

	if len(nodesToEmbed) > 0 {
		return parser.GenerateEmbeddingsForNodes(nodesToEmbed)
	}

	return nil
}

// generateEmbeddingsForAllNodes generates embeddings for all applicable nodes
func (s *Server) generateEmbeddingsForAllNodes(parser *analyzer.Parser, graph *analyzer.Graph) error {
	if s.embeddingsGenerator == nil {
		s.logger.Debug("No embedding client configured, skipping embedding generation")
		return nil
	}

	nodesToEmbed := make([]*analyzer.EnhancedNode, 0)

	for i := range graph.Nodes {
		node := &graph.Nodes[i]

		// Only generate embeddings for specific node types
		if node.Type == analyzer.NodeTypeFunction || node.Type == analyzer.NodeTypeMethod ||
			node.Type == analyzer.NodeTypeStruct || node.Type == analyzer.NodeTypeInterface {
			nodesToEmbed = append(nodesToEmbed, node)
		}
	}

	s.logger.Debug("Full embedding generation",
		"totalNodes", len(graph.Nodes),
		"toGenerate", len(nodesToEmbed))

	if len(nodesToEmbed) > 0 {
		return parser.GenerateEmbeddingsForNodes(nodesToEmbed)
	}

	return nil
}

// validateAndSetupParams validates and sets up analysis parameters
func (s *Server) validateAndSetupParams(params AnalyzeWorkspaceParams) (workspacePath, workspaceName string, err error) {
	// Get workspace path with default to current directory
	workspacePath = params.WorkspacePath
	if workspacePath == "" {
		workspacePath, err = os.Getwd()
		if err != nil {
			return "", "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Auto-detect workspace name if not provided
	workspaceName = params.WorkspaceName
	if workspaceName == "" {
		workspaceName, err = s.detectWorkspaceName(workspacePath)
		if err != nil {
			return "", "", fmt.Errorf("failed to detect workspace name: %w", err)
		}
	}

	return workspacePath, workspaceName, nil
}

// prepareIncrementalUpdate prepares the database for incremental or full update
func (s *Server) prepareIncrementalUpdate(ctx context.Context, workspaceName string, incremental bool) (map[string]*ExistingNodeInfo, error) {
	if !incremental {
		// Full rebuild - clear existing workspace data
		if err := s.clearWorkspaceData(ctx, workspaceName); err != nil {
			s.logger.Debug("Failed to clear existing workspace data", "workspace", workspaceName, "error", err)
		}
		return nil, nil
	}

	// Incremental update - mark all nodes as potentially stale
	if err := s.markAllNodesAsStale(ctx, workspaceName); err != nil {
		s.logger.Debug("Failed to mark nodes as stale", "workspace", workspaceName, "error", err)
		// Continue anyway - worst case we won't delete orphaned nodes
	}

	// Fetch existing nodes to check semantic hashes
	existingNodes, err := s.fetchExistingNodes(ctx, workspaceName)
	if err != nil {
		s.logger.Warn("Failed to fetch existing nodes for hash comparison", "error", err)
		// Continue without hash comparison - will regenerate all embeddings
		return nil, nil
	}

	return existingNodes, nil
}

// analyzeAndCreateGraph analyzes the codebase and creates the graph
func (s *Server) analyzeAndCreateGraph(workspacePath string, allowedPackages []string) (*analyzer.Graph, *analyzer.Parser, error) {
	parser := analyzer.NewParser(s.embeddingsGenerator, s.logger)
	
	if len(allowedPackages) > 0 {
		parser.SetAllowedPackages(allowedPackages)
		s.logger.Info("Including external packages", "packages", allowedPackages)
	}
	
	graph, err := parser.AnalyzeCodebase(workspacePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze codebase: %w", err)
	}

	return graph, parser, nil
}

// processEmbeddings handles all embedding-related processing
func (s *Server) processEmbeddings(parser *analyzer.Parser, graph *analyzer.Graph, incremental bool, existingNodes map[string]*ExistingNodeInfo) error {
	// Prepare nodes with hash metadata for change detection
	// IMPORTANT: This must be called BEFORE checking semantic hashes
	// but AFTER semantic summaries are generated
	s.prepareNodesWithHashes(graph)

	// Generate embeddings based on incremental or full rebuild
	if s.embeddingsGenerator != nil {
		s.logger.Info("Starting embedding generation...")
	}

	if incremental && existingNodes != nil {
		// For incremental updates, only generate embeddings for new or changed nodes
		if err := s.generateEmbeddingsForDelta(parser, graph, existingNodes); err != nil {
			return err
		}
	} else {
		// For full rebuilds, generate embeddings for all applicable nodes
		if err := s.generateEmbeddingsForAllNodes(parser, graph); err != nil {
			return err
		}
	}

	// Update has_embedding metadata
	for i := range graph.Nodes {
		if graph.Nodes[i].Metadata == nil {
			graph.Nodes[i].Metadata = make(map[string]interface{})
		}
		graph.Nodes[i].Metadata["has_embedding"] = len(graph.Nodes[i].Embedding) > 0
	}

	return nil
}

// importToNeo4jWithCleanup imports to Neo4j and handles cleanup for incremental updates
func (s *Server) importToNeo4jWithCleanup(ctx context.Context, graph *analyzer.Graph, incremental bool, workspaceName string) (int, error) {
	// IMPORTANT: Create indexes BEFORE importing data for better performance
	s.logger.Info("Creating Neo4j indexes for optimal import performance...")
	if err := s.ensureIndexesExist(ctx); err != nil {
		s.logger.Warn("Failed to create indexes", "error", err)
		// Continue - we can still import without indexes, just slower
	}

	// Import to Neo4j using direct driver calls
	s.logger.Info("Starting Neo4j import...")
	_, err := s.importGraphToNeo4j(ctx, graph, incremental)
	if err != nil {
		return 0, fmt.Errorf("failed to import to Neo4j: %w", err)
	}

	// For incremental updates, remove nodes that are still marked as stale
	var removedCount int
	if incremental {
		s.logger.Debug("Removing stale nodes...")
		removedCount, err = s.removeStaleNodes(ctx, workspaceName)
		if err != nil {
			s.logger.Warn("Failed to remove stale nodes", "workspace", workspaceName, "error", err)
			// Continue - this is not critical
		}
	}

	return removedCount, nil
}

// storeWorkspaceInfo creates and stores workspace metadata
func (s *Server) storeWorkspaceInfo(ctx context.Context, graph *analyzer.Graph, workspaceName, workspacePath string, analysisTime, importTime, totalTime time.Duration, incremental bool) error {
	workspaceInfo := map[string]any{
		"workspace":    workspaceName,
		"path":         workspacePath,
		"nodes":        len(graph.Nodes),
		"edges":        len(graph.Edges),
		"packages":     s.countNodesByType(graph, analyzer.NodeTypePackage),
		"functions":    s.countNodesByType(graph, analyzer.NodeTypeFunction),
		"methods":      s.countNodesByType(graph, analyzer.NodeTypeMethod),
		"structs":      s.countNodesByType(graph, analyzer.NodeTypeStruct),
		"interfaces":   s.countNodesByType(graph, analyzer.NodeTypeInterface),
		"analysisTime": analysisTime.String(),
		"importTime":   importTime.String(),
		"totalTime":    totalTime.String(),
		"analyzedAt":   time.Now().Format(time.RFC3339),
		"incremental":  incremental,
	}

	return s.storeWorkspaceMetadata(ctx, workspaceInfo)
}

// formatAnalysisResponse formats the analysis response with statistics
func (s *Server) formatAnalysisResponse(graph *analyzer.Graph, workspaceName, workspacePath string, analysisTime, importTime, totalTime time.Duration, incremental bool, removedCount int) *mcp.CallToolResultFor[any] {
	// Format response
	summary := fmt.Sprintf("✅ Workspace '%s' analyzed successfully", workspaceName)

	// Calculate embedding statistics
	funcCount := s.countNodesByType(graph, analyzer.NodeTypeFunction)
	funcWithEmbeddings := s.countNodesWithEmbeddingsByType(graph, analyzer.NodeTypeFunction)
	methodCount := s.countNodesByType(graph, analyzer.NodeTypeMethod)
	methodWithEmbeddings := s.countNodesWithEmbeddingsByType(graph, analyzer.NodeTypeMethod)
	structCount := s.countNodesByType(graph, analyzer.NodeTypeStruct)
	structWithEmbeddings := s.countNodesWithEmbeddingsByType(graph, analyzer.NodeTypeStruct)
	interfaceCount := s.countNodesByType(graph, analyzer.NodeTypeInterface)
	interfaceWithEmbeddings := s.countNodesWithEmbeddingsByType(graph, analyzer.NodeTypeInterface)

	// Debug log to verify function is being called
	s.logger.Debug("Embedding statistics",
		"functions", funcCount, "funcEmbeddings", funcWithEmbeddings,
		"methods", methodCount, "methodEmbeddings", methodWithEmbeddings)

	// Count total nodes with embeddings for debug
	totalWithEmbeddings := funcWithEmbeddings + methodWithEmbeddings + structWithEmbeddings + interfaceWithEmbeddings

	// Build details string with incremental update info
	incrementalInfo := ""
	if incremental {
		incrementalInfo = fmt.Sprintf("\n\n📈 Incremental Update:\n   Nodes removed (stale): %d", removedCount)
	}

	details := fmt.Sprintf(`
📊 Analysis Results:
   Workspace: %s
   Path: %s
   Nodes: %d
   Edges: %d
   
📦 Node Types:
   Packages: %d
   Functions: %d (with embeddings: %d)
   Methods: %d (with embeddings: %d)
   Structs: %d (with embeddings: %d)
   Interfaces: %d (with embeddings: %d)

🧠 Embedding Debug:
   Total nodes with embeddings: %d
   Functions checked: %d, with embeddings: %d
   Methods checked: %d, with embeddings: %d

⏱️ Performance:
   Analysis: %s
   Import: %s
   Total: %s%s
   
Ready for intelligent code queries!`,
		workspaceName,
		workspacePath,
		len(graph.Nodes),
		len(graph.Edges),
		s.countNodesByType(graph, analyzer.NodeTypePackage),
		funcCount, funcWithEmbeddings,
		methodCount, methodWithEmbeddings,
		structCount, structWithEmbeddings,
		interfaceCount, interfaceWithEmbeddings,
		totalWithEmbeddings,
		funcCount, funcWithEmbeddings,
		methodCount, methodWithEmbeddings,
		analysisTime.String(),
		importTime.String(),
		totalTime.String(),
		incrementalInfo,
	)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: details},
		},
	}
}

// createErrorResponse creates an MCP error response
func (s *Server) createErrorResponse(err error) *mcp.CallToolResultFor[any] {
	return &mcp.CallToolResultFor[any]{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}
}
