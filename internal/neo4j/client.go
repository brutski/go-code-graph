package neo4j

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/brutski/go-code-graph/internal/analyzer"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Client wraps Neo4j database operations for code graphs
type Client struct {
	driver   neo4j.DriverWithContext
	database string
}

// Config holds Neo4j connection configuration
type Config struct {
	URI      string
	Username string
	Password string
	Database string
}

// NewClient creates a new Neo4j client
func NewClient(config Config) (*Client, error) {
	// Create driver
	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.Username, config.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	client := &Client{
		driver:   driver,
		database: config.Database,
	}

	// Test connection
	ctx := context.Background()
	if err := client.ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	log.Printf("Connected to Neo4j at %s", config.URI)
	return client, nil
}

// Close closes the Neo4j connection
func (c *Client) Close() error {
	return c.driver.Close(context.Background())
}

// ping tests the connection to Neo4j
func (c *Client) ping(ctx context.Context) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	_, err := session.Run(ctx, "RETURN 1", nil)
	return err
}

// ClearGraph removes all nodes and relationships from the graph
func (c *Client) ClearGraph(ctx context.Context) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	// Delete all relationships first, then nodes
	queries := []string{
		"MATCH ()-[r]-() DELETE r",
		"MATCH (n) DELETE n",
	}

	for _, query := range queries {
		_, err := session.Run(ctx, query, nil)
		if err != nil {
			return fmt.Errorf("failed to clear graph: %w", err)
		}
	}

	log.Println("Cleared existing graph data")
	return nil
}

// ImportGraph imports a complete graph into Neo4j
func (c *Client) ImportGraph(ctx context.Context, graph *analyzer.Graph) error {
	if graph == nil {
		return fmt.Errorf("graph cannot be nil")
	}

	log.Printf("Importing graph with %d nodes and %d edges", len(graph.Nodes), len(graph.Edges))

	// Clear existing graph
	if err := c.ClearGraph(ctx); err != nil {
		return err
	}

	// Create constraints and indexes first
	if err := c.createConstraintsAndIndexes(ctx); err != nil {
		return err
	}

	// Import nodes in batches
	if err := c.importNodes(ctx, graph.Nodes); err != nil {
		return err
	}

	// Import edges in batches
	if err := c.importEdges(ctx, graph.Edges); err != nil {
		return err
	}

	// Import metadata
	if err := c.importMetadata(ctx, graph); err != nil {
		return err
	}

	log.Println("Graph import completed successfully")
	return nil
}

// createConstraintsAndIndexes creates Neo4j constraints and indexes for optimal performance
func (c *Client) createConstraintsAndIndexes(ctx context.Context) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	constraints := []string{
		// Unique constraints for node IDs
		"CREATE CONSTRAINT node_id_unique IF NOT EXISTS FOR (n:CodeNode) REQUIRE n.id IS UNIQUE",

		// Indexes for common queries
		"CREATE INDEX node_type_idx IF NOT EXISTS FOR (n:CodeNode) ON (n.type)",
		"CREATE INDEX node_package_idx IF NOT EXISTS FOR (n:CodeNode) ON (n.package)",
		"CREATE INDEX node_label_idx IF NOT EXISTS FOR (n:CodeNode) ON (n.label)",
	}

	for _, constraint := range constraints {
		if _, err := session.Run(ctx, constraint, nil); err != nil {
			// Ignore errors for existing constraints/indexes
			log.Printf("Constraint/Index creation (may already exist): %s", constraint)
		}
	}

	return nil
}

// importNodes imports all nodes in batches
func (c *Client) importNodes(ctx context.Context, nodes []analyzer.EnhancedNode) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	batchSize := 1000
	for i := 0; i < len(nodes); i += batchSize {
		end := i + batchSize
		if end > len(nodes) {
			end = len(nodes)
		}

		batch := nodes[i:end]
		if err := c.importNodeBatch(ctx, session, batch); err != nil {
			return err
		}

		log.Printf("Imported %d/%d nodes", end, len(nodes))
	}

	return nil
}

// importNodeBatch imports a batch of nodes
func (c *Client) importNodeBatch(ctx context.Context, session neo4j.SessionWithContext, nodes []analyzer.EnhancedNode) error {
	params := make([]interface{}, len(nodes))
	for i, node := range nodes {
		params[i] = map[string]interface{}{
			"id":            node.ID,
			"label":         node.Label,
			"type":          node.Type,
			"package":       node.Package,
			"full_name":     node.FullName,
			"size":          node.Size,
			"level":         node.Level,
			"visibility":    node.Visibility,
			"signature":     node.Signature,
			"documentation": node.Documentation,
			"complexity":    node.Complexity,
			"filename":      node.Position.Filename,
			"line":          node.Position.Line,
			"column":        node.Position.Column,
			"created_at":    node.CreatedAt.Unix(),
			"tags":          node.Tags,

			// Type information
			"type_kind":       node.TypeInfo.Kind,
			"is_pointer":      node.TypeInfo.IsPointer,
			"is_slice":        node.TypeInfo.IsSlice,
			"is_array":        node.TypeInfo.IsArray,
			"is_channel":      node.TypeInfo.IsChannel,
			"channel_dir":     node.TypeInfo.ChannelDir,
			"is_generic":      node.TypeInfo.IsGeneric,
			"underlying_type": node.TypeInfo.UnderlyingType,
			"field_count":     node.TypeInfo.FieldCount,
			"is_exported":     node.TypeInfo.IsExported,
			"is_alias":        node.TypeInfo.IsAlias,
			"embedded_types":  node.TypeInfo.EmbeddedTypes,
		}
	}

	query := `
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
			signature: node.signature,
			documentation: node.documentation,
			complexity: node.complexity,
			filename: node.filename,
			line: node.line,
			column: node.column,
			created_at: node.created_at,
			tags: node.tags,
			type_kind: node.type_kind,
			is_pointer: node.is_pointer,
			is_slice: node.is_slice,
			is_array: node.is_array,
			is_channel: node.is_channel,
			channel_dir: node.channel_dir,
			is_generic: node.is_generic,
			underlying_type: node.underlying_type,
			field_count: node.field_count,
			is_exported: node.is_exported,
			is_alias: node.is_alias,
			embedded_types: node.embedded_types
		})
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"nodes": params,
	})

	return err
}

// importEdges imports all edges in batches
func (c *Client) importEdges(ctx context.Context, edges []analyzer.EnhancedEdge) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	batchSize := 1000
	for i := 0; i < len(edges); i += batchSize {
		end := i + batchSize
		if end > len(edges) {
			end = len(edges)
		}

		batch := edges[i:end]
		if err := c.importEdgeBatch(ctx, session, batch); err != nil {
			return err
		}

		log.Printf("Imported %d/%d edges", end, len(edges))
	}

	return nil
}

// importEdgeBatch imports a batch of edges
func (c *Client) importEdgeBatch(ctx context.Context, session neo4j.SessionWithContext, edges []analyzer.EnhancedEdge) error {
	// Group edges by type to create specific relationships
	edgesByType := make(map[string][]analyzer.EnhancedEdge)
	for _, edge := range edges {
		edgesByType[edge.Type] = append(edgesByType[edge.Type], edge)
	}

	for edgeType, edgeBatch := range edgesByType {
		params := make([]map[string]interface{}, len(edgeBatch))
		for i, edge := range edgeBatch {
			params[i] = map[string]interface{}{
				"source":      edge.Source,
				"target":      edge.Target,
				"weight":      edge.Weight,
				"context":     edge.Context,
				"conditional": edge.Conditional,
				"loop_depth":  edge.LoopDepth,
				"filename":    edge.Position.Filename,
				"line":        edge.Position.Line,
				"column":      edge.Position.Column,
				"created_at":  edge.CreatedAt.Unix(),
			}
		}

		// Sanitize edge type to be a valid Neo4j relationship type
		relationshipType := sanitizeLabel(edgeType)

		query := fmt.Sprintf(`
			UNWIND $edges AS edge
			MATCH (source:CodeNode {id: edge.source})
			MATCH (target:CodeNode {id: edge.target})
			CREATE (source)-[r:%s {
				weight: edge.weight,
				context: edge.context,
				conditional: edge.conditional,
				loop_depth: edge.loop_depth,
				filename: edge.filename,
				line: edge.line,
				column: edge.column,
				created_at: edge.created_at
			}]->(target)
		`, relationshipType)

		_, err := session.Run(ctx, query, map[string]interface{}{
			"edges": params,
		})
		if err != nil {
			return fmt.Errorf("failed to import edges of type %s: %w", edgeType, err)
		}
	}

	return nil
}

// importMetadata imports graph metadata
func (c *Client) importMetadata(ctx context.Context, graph *analyzer.Graph) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	metadataNode := map[string]interface{}{
		"type":                  "graph_metadata",
		"version":               graph.Version,
		"created_at":            graph.CreatedAt.Unix(),
		"total_nodes":           graph.Stats.TotalNodes,
		"total_edges":           graph.Stats.TotalEdges,
		"source_path":           graph.Metadata.SourcePath,
		"go_version":            graph.Metadata.GoVersion,
		"module_path":           graph.Metadata.ModulePath,
		"build_tags":            graph.Metadata.BuildTags,
		"dependencies":          graph.Stats.Dependencies,
		"package_count":         graph.Stats.PackageCount,
		"max_depth":             graph.Stats.MaxDepth,
		"cyclomatic_complexity": graph.Stats.CyclomaticComplexity,
	}

	query := `
		CREATE (meta:GraphMetadata {
			type: $metadata.type,
			version: $metadata.version,
			created_at: $metadata.created_at,
			total_nodes: $metadata.total_nodes,
			total_edges: $metadata.total_edges,
			source_path: $metadata.source_path,
			go_version: $metadata.go_version,
			module_path: $metadata.module_path,
			build_tags: $metadata.build_tags,
			dependencies: $metadata.dependencies,
			package_count: $metadata.package_count,
			max_depth: $metadata.max_depth,
			cyclomatic_complexity: $metadata.cyclomatic_complexity
		})
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"metadata": metadataNode,
	})

	return err
}

// GetGraphStats returns basic statistics about the stored graph
func (c *Client) GetGraphStats(ctx context.Context) (map[string]interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
	})
	defer session.Close(ctx)

	queries := map[string]string{
		"total_nodes": "MATCH (n:CodeNode) RETURN count(n) as count",
		"total_edges": "MATCH ()-[r]->() RETURN count(r) as count",
		"node_types":  "MATCH (n:CodeNode) RETURN n.type as type, count(n) as count ORDER BY count DESC",
		"edge_types":  "MATCH ()-[r]->() RETURN type(r) as type, count(r) as count ORDER BY count DESC",
		"packages":    "MATCH (n:CodeNode {type: 'package'}) RETURN n.label as package, n.package as path",
	}

	stats := make(map[string]interface{})

	for key, query := range queries {
		result, err := session.Run(ctx, query, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query %s: %w", key, err)
		}

		var records []map[string]interface{}
		for result.Next(ctx) {
			record := result.Record()
			recordMap := make(map[string]interface{})
			for _, key := range record.Keys {
				if value, ok := record.Get(key); ok {
					recordMap[key] = value
				}
			}
			records = append(records, recordMap)
		}

		if len(records) == 1 && key == "total_nodes" || key == "total_edges" {
			stats[key] = records[0]["count"]
		} else {
			stats[key] = records
		}
	}

	return stats, nil
}

// DefaultConfig returns a default Neo4j configuration
func DefaultConfig() Config {
	return Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: "codeGraph123",
		Database: "neo4j",
	}
}

// sanitizeLabel sanitizes a string to be a valid Neo4j label or relationship type
func sanitizeLabel(s string) string {
	if s == "" {
		return ""
	}

	// Remove leading digits
	for len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		s = s[1:]
	}

	if s == "" {
		return ""
	}

	// Convert to PascalCase
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	var result strings.Builder
	for _, part := range parts {
		if part != "" {
			// Capitalize first letter
			result.WriteString(strings.ToUpper(string(part[0])))
			// Lowercase the rest
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}
