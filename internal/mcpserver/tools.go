package mcpserver

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

//go:embed queries/duplicate_functions.cypher
var duplicateFunctionsQuery string

//go:embed queries/usage_functions.cypher
var usageFunctionsQuery string

//go:embed queries/find_implementers.cypher
var findImplementersQuery string

//go:embed queries/trace_call_path.cypher
var traceCallPathQuery string

//go:embed queries/error_handlers.cypher
var errorHandlersQuery string

//go:embed queries/goroutine_spawners.cypher
var goroutineSpawnersQuery string

//go:embed queries/type_constructors.cypher
var typeConstructorsQuery string

//go:embed queries/embedded_types.cypher
var embeddedTypesQuery string

// CypherQueryParams represents the parameters for a Cypher query
type CypherQueryParams struct {
	// Query is the Cypher query to execute against the code graph
	Query string `json:"query"`
	// Parameters are optional query parameters for parameterized queries
	Parameters map[string]any `json:"parameters,omitempty"`
	// Explain returns the query execution plan instead of results
	Explain bool `json:"explain,omitempty"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// NaturalQueryParams represents the parameters for a natural language query
type NaturalQueryParams struct {
	// Question is the natural language question about the codebase
	Question string `json:"question"`
	// Context provides additional context for better query generation
	Context string `json:"context,omitempty"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// ImpactAnalysisParams represents the parameters for impact analysis
type ImpactAnalysisParams struct {
	// NodeID is the node identifier to analyze (e.g., function:pkg.FunctionName)
	NodeID string `json:"nodeId"`
	// ChangeType specifies the type of change: signature, delete, modify
	ChangeType string `json:"changeType,omitempty"`
	// MaxDepth is the maximum depth for impact analysis
	MaxDepth int `json:"maxDepth,omitempty"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// PatternDetectionParams represents the parameters for pattern detection
type PatternDetectionParams struct {
	// PatternType specifies the type of pattern: duplicate, similar, antipattern, usage
	PatternType string `json:"patternType"`
	// Filter provides additional filters for pattern matching
	Filter map[string]any `json:"filter,omitempty"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// FindImplementersParams represents the parameters for finding implementers of an interface
type FindImplementersParams struct {
	// InterfaceName is the full name of the interface (e.g., "io.Writer")
	InterfaceName string `json:"interfaceName"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// TraceCallPathParams represents the parameters for tracing a call path
type TraceCallPathParams struct {
	// From is the full name of the starting function/method
	From string `json:"from"`
	// To is the full name of the ending function/method
	To string `json:"to"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// ArchitectureAnalysisParams represents the parameters for architecture analysis
type ArchitectureAnalysisParams struct {
	// AnalysisType specifies the type of architecture analysis to perform
	AnalysisType string `json:"analysisType"`
	// Workspace is the workspace to query (required)
	Workspace string `json:"workspace"`
}

// executeCypherQuery executes a Cypher query and returns raw Neo4j records
func (s *Server) executeCypherQuery(ctx context.Context, query string, params map[string]any, explain bool) ([]*neo4j.Record, error) {
	// Create Neo4j session
	neo4jSession := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer neo4jSession.Close(ctx)

	var actualQuery string
	if explain {
		actualQuery = "EXPLAIN " + query
	} else {
		actualQuery = query
	}

	// Execute query
	result, err := neo4jSession.Run(ctx, actualQuery, params)
	if err != nil {
		return nil, fmt.Errorf("cypher query failed: %w", err)
	}

	// Collect results
	records, err := result.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect results: %w", err)
	}

	return records, nil
}

// parseNeo4jResults converts Neo4j records to JSON bytes
func parseNeo4jResults(records []*neo4j.Record) ([]byte, error) {
	var results []map[string]any

	for _, record := range records {
		recordMap := make(map[string]any)
		for j, key := range record.Keys {
			value := record.Values[j]
			// Convert Neo4j types to proper Go types
			recordMap[key] = convertNeo4jValue(value)
		}
		results = append(results, recordMap)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	return jsonData, nil
}

// handleCypherQuery orchestrates Cypher query execution and result formatting
func (s *Server) handleCypherQuery(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CypherQueryParams]) (*mcp.CallToolResultFor[any], error) {
	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	// Ensure workspace parameter is in the query parameters
	if params.Arguments.Parameters == nil {
		params.Arguments.Parameters = make(map[string]any)
	}
	params.Arguments.Parameters["workspace"] = params.Arguments.Workspace

	// Execute query
	records, err := s.executeCypherQuery(ctx, params.Arguments.Query, params.Arguments.Parameters, params.Arguments.Explain)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil
	}

	// Parse results
	jsonData, err := parseNeo4jResults(records)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil
	}

	// Success response
	summary := fmt.Sprintf("Query executed successfully. Found %d records.", len(records))
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, nil
}

// convertNeo4jValue properly converts Neo4j driver values to Go values
func convertNeo4jValue(value any) any {
	switch v := value.(type) {
	case []any:
		// Convert array elements recursively
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertNeo4jValue(item)
		}
		return result
	case map[string]any:
		// Convert map values recursively
		result := make(map[string]any)
		for k, val := range v {
			result[k] = convertNeo4jValue(val)
		}
		return result
	default:
		// Return the value as-is for primitive types
		return v
	}
}

// handleNaturalQuery converts natural language to Cypher queries with hybrid search
func (s *Server) handleNaturalQuery(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[NaturalQueryParams]) (*mcp.CallToolResultFor[any], error) {
	s.logger.Debug("Processing natural language query", "question", params.Arguments.Question)

	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	// Try hybrid search if embedding client is available
	if s.embeddingsGenerator != nil {
		s.logger.Debug("Using hybrid search with embeddings")
		return s.handleHybridSearch(ctx, session, params)
	}

	// Fallback to pattern-based Cypher generation
	s.logger.Debug("Using pattern-based query generation (no embeddings)")
	return s.handlePatternBasedQuery(ctx, session, params)
}

// handleHybridSearch combines vector similarity with graph traversal
func (s *Server) handleHybridSearch(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[NaturalQueryParams]) (*mcp.CallToolResultFor[any], error) {
	question := params.Arguments.Question

	// Generate embedding for the user's question
	queryEmbedding, err := s.embeddingsGenerator.CreateEmbedding(question)
	if err != nil {
		s.logger.Debug("Failed to create embedding, falling back to pattern-based search", "error", err)
		return s.handlePatternBasedQuery(ctx, session, params)
	}

	// Hybrid search query: combine vector similarity with graph patterns
	hybridQuery := `
		// First, find semantically similar nodes using vector search
		CALL db.index.vector.queryNodes('code_embeddings_index', 10, $queryEmbedding) 
		YIELD node as similarNode, score
		WHERE similarNode.workspace = $workspace
		
		// Extract only the fields we need (exclude embeddings)
		WITH {
		  id: similarNode.id,
		  label: similarNode.label,
		  type: similarNode.type,
		  package: similarNode.package,
		  full_name: similarNode.full_name,
		  summary: similarNode.semantic_summary,
		  signature: similarNode.signature,
		  complexity: similarNode.complexity,
		  visibility: similarNode.visibility
		} as cleanNode, score
		
		// Then expand to related nodes through graph traversal
		OPTIONAL MATCH (n:CodeNode {id: cleanNode.id, workspace: $workspace})-[r:RELATES_TO]-(relatedNode:CodeNode {workspace: $workspace})
		
		// Collect results with relevance scoring
		WITH cleanNode, score, 
		     collect(DISTINCT {
		       id: relatedNode.id,
		       label: relatedNode.label,
		       type: relatedNode.type,
		       relationship: r.type
		     }) as related,
		     count(relatedNode) as relatedCount
		
		// Return hybrid results ordered by semantic similarity and connectivity
		RETURN {
		  id: cleanNode.id,
		  label: cleanNode.label,
		  type: cleanNode.type,
		  package: cleanNode.package,
		  full_name: cleanNode.full_name,
		  summary: cleanNode.summary,
		  signature: cleanNode.signature,
		  complexity: cleanNode.complexity,
		  visibility: cleanNode.visibility,
		  similarity_score: score,
		  related_nodes: related,
		  connectivity: relatedCount
		} as result
		ORDER BY score DESC, relatedCount DESC
		LIMIT 20
	`

	// Execute hybrid search
	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query: hybridQuery,
			Parameters: map[string]any{
				"queryEmbedding": queryEmbedding,
			},
			Workspace: params.Arguments.Workspace,
		},
	}

	result, err := s.handleCypherQuery(ctx, session, cypherParams)
	if err != nil {
		// If hybrid search fails, fallback to pattern-based search
		s.logger.Debug("Hybrid search failed, falling back to pattern-based search", "error", err)
		return s.handlePatternBasedQuery(ctx, session, params)
	}

	// Add hybrid search explanation to response
	var content []mcp.Content
	content = append(content, &mcp.TextContent{Text: fmt.Sprintf("Hybrid search results for: \"%s\" (combining semantic similarity with graph connectivity)", question)})
	content = append(content, result.Content...)

	return &mcp.CallToolResultFor[any]{
		IsError: result.IsError,
		Content: content,
	}, nil
}

// handlePatternBasedQuery uses pattern matching for query generation (fallback)
func (s *Server) handlePatternBasedQuery(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[NaturalQueryParams]) (*mcp.CallToolResultFor[any], error) {
	// Convert natural language to Cypher using pattern matching
	cypherQuery, err := s.naturalLanguageToCypher(params.Arguments.Question, params.Arguments.Context, params.Arguments.Workspace)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to parse natural language query: %v", err)},
			},
		}, nil
	}

	s.logger.Debug("Generated Cypher query", "query", cypherQuery)

	// Execute the generated Cypher query
	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query:      cypherQuery,
			Parameters: make(map[string]any),
			Explain:    false,
			Workspace:  params.Arguments.Workspace,
		},
	}

	result, err := s.handleCypherQuery(ctx, session, cypherParams)
	if err != nil {
		return result, err
	}

	// Add the generated Cypher query to the response
	var content []mcp.Content
	content = append(content, &mcp.TextContent{Text: fmt.Sprintf("Pattern-based query: %s", cypherQuery)})
	content = append(content, result.Content...)

	return &mcp.CallToolResultFor[any]{
		IsError: result.IsError,
		Content: content,
	}, nil
}

// handleImpactAnalysis analyzes the impact of changing a code component
func (s *Server) handleImpactAnalysis(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ImpactAnalysisParams]) (*mcp.CallToolResultFor[any], error) {
	s.logger.Debug("Analyzing impact", "nodeId", params.Arguments.NodeID, "changeType", params.Arguments.ChangeType)

	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	maxDepth := params.Arguments.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 5 // Default depth
	}

	// Build impact analysis query based on change type
	var cypherQuery string
	switch strings.ToLower(params.Arguments.ChangeType) {
	case "delete":
		// Find nodes that depend on the target node (would be affected by deletion)
		cypherQuery = `
			MATCH (n:CodeNode {workspace: $workspace})
			WHERE n.label = $nodeId OR n.id = $nodeId OR n.full_name = $nodeId
			MATCH (dependent:CodeNode {workspace: $workspace})-[r:RELATES_TO]->(n)
			WHERE r.type IN ['calls', 'uses', 'implements', 'returns', 'embeds', 
			                 'constructs', 'parameter_type', 'imports']
			RETURN DISTINCT dependent.id as affectedNode,
			       dependent.type as nodeType, 
			       dependent.label as name,
			       dependent.package as package,
			       r.type as dependency_type,
			       CASE r.type
			         WHEN 'calls' THEN 'function_calls_deleted'
			         WHEN 'implements' THEN 'implements_deleted_interface'
			         WHEN 'embeds' THEN 'embeds_deleted_type'
			         WHEN 'imports' THEN 'imports_deleted_package'
			         WHEN 'returns' THEN 'returns_deleted_type'
			         WHEN 'constructs' THEN 'constructs_deleted_type'
			         WHEN 'parameter_type' THEN 'parameter_uses_deleted_type'
			         ELSE 'depends_on_deleted_node'
			       END as impact_type
			ORDER BY dependent.type, r.type
			LIMIT 50
		`
	default:
		// General impact analysis - find all related nodes
		cypherQuery = `
			MATCH (n:CodeNode {workspace: $workspace})
			WHERE n.id = $nodeId OR n.full_name = $nodeId OR n.label = $nodeId
			MATCH (related:CodeNode {workspace: $workspace})-[r:RELATES_TO]-(n)
			WHERE r.type IN ['calls', 'uses', 'implements', 'has_method', 'method_of', 
			                 'returns', 'embeds', 'constructs', 'handles_error', 
			                 'spawns_goroutine', 'parameter_type', 'imports']
			RETURN DISTINCT related.id as relatedNode,
			       related.type as nodeType,
			       related.label as name, 
			       related.package as package,
			       r.type as relationship_type
			ORDER BY related.type
			LIMIT 50
		`
	}

	// Execute the impact analysis query
	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query: cypherQuery,
			Parameters: map[string]any{
				"nodeId": params.Arguments.NodeID,
			},
			Workspace: params.Arguments.Workspace,
		},
	}

	result, err := s.handleCypherQuery(ctx, session, cypherParams)
	if err != nil {
		return result, err
	}

	// Add impact analysis summary
	summary := fmt.Sprintf("Impact analysis for %s (change type: %s, max depth: %d)",
		params.Arguments.NodeID,
		params.Arguments.ChangeType,
		maxDepth)

	var content []mcp.Content
	content = append(content, &mcp.TextContent{Text: summary})
	content = append(content, result.Content...)

	return &mcp.CallToolResultFor[any]{
		IsError: result.IsError,
		Content: content,
	}, nil
}

// handlePatternDetection finds patterns in the codebase
func (s *Server) handlePatternDetection(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[PatternDetectionParams]) (*mcp.CallToolResultFor[any], error) {
	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	var cypherQuery string

	switch strings.ToLower(params.Arguments.PatternType) {
	case "duplicate":
		cypherQuery = strings.TrimSpace(duplicateFunctionsQuery)
	case "usage":
		cypherQuery = strings.TrimSpace(usageFunctionsQuery)
	case "error_handlers":
		cypherQuery = strings.TrimSpace(errorHandlersQuery)
	case "goroutine_spawners":
		cypherQuery = strings.TrimSpace(goroutineSpawnersQuery)
	case "type_constructors":
		cypherQuery = strings.TrimSpace(typeConstructorsQuery)
	case "embedded_types":
		cypherQuery = strings.TrimSpace(embeddedTypesQuery)
	default:
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Unsupported pattern type. Available: duplicate, usage, error_handlers, goroutine_spawners, type_constructors, embedded_types"},
			},
		}, nil
	}

	// Execute the pattern detection query
	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query:      cypherQuery,
			Parameters: make(map[string]any),
			Workspace:  params.Arguments.Workspace,
		},
	}

	result, err := s.handleCypherQuery(ctx, session, cypherParams)
	if err != nil {
		return result, err
	}

	// Add pattern detection summary
	summary := fmt.Sprintf("Pattern detection results for type: %s", params.Arguments.PatternType)

	var content []mcp.Content
	content = append(content, &mcp.TextContent{Text: summary})
	content = append(content, result.Content...)

	return &mcp.CallToolResultFor[any]{
		IsError: result.IsError,
		Content: content,
	}, nil
}

// handleFindImplementers finds all types that implement a given interface
func (s *Server) handleFindImplementers(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[FindImplementersParams]) (*mcp.CallToolResultFor[any], error) {
	s.logger.Debug("Finding implementers", "interface", params.Arguments.InterfaceName)

	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query: strings.TrimSpace(findImplementersQuery),
			Parameters: map[string]any{
				"name": params.Arguments.InterfaceName,
			},
			Workspace: params.Arguments.Workspace,
		},
	}

	result, err := s.handleCypherQuery(ctx, session, cypherParams)
	if err != nil {
		return result, err
	}

	// Add summary
	summary := fmt.Sprintf("Implementers for interface '%s':", params.Arguments.InterfaceName)
	var content []mcp.Content
	content = append(content, &mcp.TextContent{Text: summary})
	content = append(content, result.Content...)

	return &mcp.CallToolResultFor[any]{
		IsError: result.IsError,
		Content: content,
	}, nil
}

// handleTraceCallPath finds the call path between two functions/methods
func (s *Server) handleTraceCallPath(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[TraceCallPathParams]) (*mcp.CallToolResultFor[any], error) {
	s.logger.Debug("Tracing call path", "from", params.Arguments.From, "to", params.Arguments.To)

	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query: strings.TrimSpace(traceCallPathQuery),
			Parameters: map[string]any{
				"from": params.Arguments.From,
				"to":   params.Arguments.To,
			},
			Workspace: params.Arguments.Workspace,
		},
	}

	result, err := s.handleCypherQuery(ctx, session, cypherParams)
	if err != nil {
		return result, err
	}

	summary := fmt.Sprintf("Call path from '%s' to '%s':", params.Arguments.From, params.Arguments.To)
	var content []mcp.Content
	content = append(content, &mcp.TextContent{Text: summary})
	content = append(content, result.Content...)

	return &mcp.CallToolResultFor[any]{
		IsError: result.IsError,
		Content: content,
	}, nil
}

// handleArchitectureAnalysis performs structural analysis of the codebase
func (s *Server) handleArchitectureAnalysis(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ArchitectureAnalysisParams]) (*mcp.CallToolResultFor[any], error) {
	s.logger.Debug("Performing architecture analysis", "type", params.Arguments.AnalysisType)

	// Validate workspace parameter
	if params.Arguments.Workspace == "" {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Workspace parameter is required"},
			},
		}, nil
	}

	var cypherQuery string
	switch strings.ToLower(params.Arguments.AnalysisType) {
	case "layers":
		cypherQuery = `
			MATCH (n:CodeNode {workspace: $workspace})
			WHERE n.type IN ['struct', 'interface', 'function']
			WITH n.package as layer, count(n) as nodeCount
			RETURN layer, nodeCount
			ORDER BY nodeCount DESC
		`
	case "patterns":
		cypherQuery = `
			MATCH (f:CodeNode {type: 'function', workspace: $workspace})-[r:RELATES_TO {type: 'constructs'}]->(t:CodeNode {type: 'struct', workspace: $workspace})
			RETURN 'Factory Pattern' as pattern, count(*) as instances
		`
	default:
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Unsupported analysis type. Available: layers, patterns"},
			},
		}, nil
	}

	cypherParams := &mcp.CallToolParamsFor[CypherQueryParams]{
		Arguments: CypherQueryParams{
			Query:     cypherQuery,
			Workspace: params.Arguments.Workspace,
		},
	}

	return s.handleCypherQuery(ctx, session, cypherParams)
}

// naturalLanguageToCypher converts natural language queries to Cypher
func (s *Server) naturalLanguageToCypher(question, context, workspace string) (string, error) {
	question = strings.ToLower(strings.TrimSpace(question))
	// Context could be used for more sophisticated query generation in the future
	_ = context

	// Extract search patterns like "with X in their name" or "containing X"
	var namePattern string
	patterns := []string{
		"with (\\w+) in their name",
		"containing (\\w+)",
		"named (\\w+)",
		"that have (\\w+)",
		"includes (\\w+)",
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if matches := re.FindStringSubmatch(question); len(matches) > 1 {
			namePattern = matches[1]
			break
		}
	}

	switch {
	// Complex function/method queries
	case strings.Contains(question, "complex") && (strings.Contains(question, "function") || strings.Contains(question, "method")):
		return `
			MATCH (f:CodeNode {workspace: $workspace})
			WHERE f.type IN ['function', 'method'] 
			  AND f.complexity > 5
			RETURN f.label as name, f.package, f.complexity, f.type
			ORDER BY f.complexity DESC
			LIMIT 20
		`, nil

	// Interface queries
	case strings.Contains(question, "interface"):
		// Use extracted name pattern if available
		if namePattern != "" {
			return fmt.Sprintf(`
				MATCH (i:CodeNode {type: 'interface', workspace: $workspace})
				WHERE toLower(i.label) CONTAINS '%s'
				RETURN i.label as interfaceName, i.package, i.full_name
				ORDER BY i.label
				LIMIT 20
			`, namePattern), nil
		}

		// Check for specific name patterns
		for _, pattern := range []string{"checkout", "repository", "service", "client", "handler", "manager"} {
			if strings.Contains(question, pattern) {
				return fmt.Sprintf(`
					MATCH (i:CodeNode {type: 'interface', workspace: $workspace})
					WHERE toLower(i.label) CONTAINS '%s'
					RETURN i.label as interfaceName, i.package, i.full_name
					ORDER BY i.label
					LIMIT 20
				`, pattern), nil
			}
		}

		if strings.Contains(question, "implement") {
			// Find types that implement interfaces
			return `
				MATCH (t:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'implements'}]->(i:CodeNode {type: 'interface', workspace: $workspace})
				RETURN t.label as implementer, t.type as implementerType, i.label as interface, t.package
				ORDER BY i.label, t.label
				LIMIT 50
			`, nil
		}
		// List all interfaces
		return `
			MATCH (i:CodeNode {type: 'interface', workspace: $workspace})
			RETURN i.label as interfaceName, i.package
			ORDER BY i.label
		`, nil

	// Package queries
	case strings.Contains(question, "package"):
		if strings.Contains(question, "import") {
			// Show package import relationships
			return `
				MATCH (p1:CodeNode {type: 'package', workspace: $workspace})-[r:RELATES_TO {type: 'imports'}]->(p2:CodeNode {type: 'package', workspace: $workspace})
				RETURN p1.label as importer, p2.label as imported
				ORDER BY p1.label
				LIMIT 50
			`, nil
		}
		// List all packages
		return `
			MATCH (p:CodeNode {type: 'package', workspace: $workspace})
			RETURN p.label as packageName, p.package as path
			ORDER BY p.label
		`, nil

	// Function/method call queries
	case strings.Contains(question, "call") || strings.Contains(question, "calls"):
		if strings.Contains(question, "who calls") || strings.Contains(question, "callers") {
			// Find callers of functions
			return `
				MATCH (caller:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'calls'}]->(callee:CodeNode {workspace: $workspace})
				WHERE callee.type IN ['function', 'method']
				RETURN callee.label as function, collect(DISTINCT caller.label) as callers, count(DISTINCT caller) as callerCount
				ORDER BY callerCount DESC
				LIMIT 20
			`, nil
		}
		// Show call relationships
		return `
			MATCH (f1:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'calls'}]->(f2:CodeNode {workspace: $workspace})
			WHERE f1.type IN ['function', 'method'] AND f2.type IN ['function', 'method']
			RETURN f1.label as caller, f2.label as callee, f1.package
			ORDER BY f1.package, f1.label
			LIMIT 50
		`, nil

	// Embedding queries
	case strings.Contains(question, "embed") || strings.Contains(question, "embeds"):
		return `
			MATCH (t1:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'embeds'}]->(t2:CodeNode {workspace: $workspace})
			RETURN t1.label as embedder, t1.type as embedderType, t2.label as embedded, t2.type as embeddedType
			ORDER BY t1.label
			LIMIT 50
		`, nil

	// Error handling queries
	case strings.Contains(question, "error") && strings.Contains(question, "handle"):
		return `
			MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'handles_error'}]->(e:CodeNode {workspace: $workspace})
			RETURN f.label as function, f.type, e.label as errorType, f.package
			ORDER BY f.package, f.label
			LIMIT 50
		`, nil

	// Goroutine queries
	case strings.Contains(question, "goroutine") || strings.Contains(question, "concurrent"):
		return `
			MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'spawns_goroutine'}]->(g:CodeNode {workspace: $workspace})
			RETURN f.label as spawner, f.type, g.label as goroutineFunc, f.package
			ORDER BY f.package, f.label
			LIMIT 50
		`, nil

	// Return type queries
	case strings.Contains(question, "return") && strings.Contains(question, "type"):
		return `
			MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'returns'}]->(t:CodeNode {workspace: $workspace})
			WHERE f.type IN ['function', 'method']
			RETURN f.label as function, t.label as returnType, f.package
			ORDER BY t.label, f.label
			LIMIT 50
		`, nil

	// Parameter type queries
	case strings.Contains(question, "parameter") && strings.Contains(question, "type"):
		return `
			MATCH (p:CodeNode {type: 'parameter', workspace: $workspace})-[r:RELATES_TO {type: 'parameter_type'}]->(t:CodeNode {workspace: $workspace})
			RETURN p.label as parameter, t.label as paramType, p.package
			ORDER BY t.label, p.label
			LIMIT 50
		`, nil

	// Method relationships
	case strings.Contains(question, "method"):
		if strings.Contains(question, "has") || strings.Contains(question, "with") {
			// Find types with methods
			return `
				MATCH (t:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'has_method'}]->(m:CodeNode {type: 'method', workspace: $workspace})
				RETURN t.label as type, t.type as nodeType, collect(m.label) as methods, count(m) as methodCount
				ORDER BY methodCount DESC
				LIMIT 20
			`, nil
		}
		// List all methods
		return `
			MATCH (m:CodeNode {type: 'method', workspace: $workspace})
			RETURN m.label as methodName, m.package, m.complexity
			ORDER BY m.complexity DESC
			LIMIT 50
		`, nil

	// Construction/instantiation queries
	case strings.Contains(question, "construct") || strings.Contains(question, "create") || strings.Contains(question, "instantiate"):
		return `
			MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'constructs'}]->(t:CodeNode {workspace: $workspace})
			RETURN f.label as constructor, f.type, t.label as constructedType, f.package
			ORDER BY t.label, f.label
			LIMIT 50
		`, nil

	// Usage queries
	case strings.Contains(question, "use") || strings.Contains(question, "uses"):
		return `
			MATCH (n1:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'uses'}]->(n2:CodeNode {workspace: $workspace})
			RETURN n1.label as user, n1.type as userType, n2.label as used, n2.type as usedType, n1.package
			ORDER BY n2.label, n1.label
			LIMIT 50
		`, nil

	// Most connected nodes
	case strings.Contains(question, "connected") || strings.Contains(question, "popular"):
		return `
			MATCH (n:CodeNode {workspace: $workspace})-[r:RELATES_TO]-(other:CodeNode {workspace: $workspace})
			WITH n, count(DISTINCT other) as connections
			RETURN n.label as node, n.type, n.package, connections
			ORDER BY connections DESC
			LIMIT 20
		`, nil

	// Error types
	case strings.Contains(question, "error"):
		return `
			MATCH (e:CodeNode {workspace: $workspace})
			WHERE e.label CONTAINS 'error' OR e.label CONTAINS 'Error'
			RETURN e.label as errorType, e.type, e.package
			ORDER BY e.label
			LIMIT 50
		`, nil

	default:
		// Generic query - show node type distribution
		return `
			MATCH (n:CodeNode {workspace: $workspace})
			RETURN n.type as nodeType, count(n) as count
			ORDER BY count DESC
		`, nil
	}
}

// ListWorkspacesParams represents parameters for listing workspaces
type ListWorkspacesParams struct{}

// handleListWorkspaces lists all available workspaces in the graph database
func (s *Server) handleListWorkspaces(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ListWorkspacesParams]) (*mcp.CallToolResultFor[any], error) {
	s.logger.Debug("Listing available workspaces")

	// Query to get all unique workspaces with node counts
	cypherQuery := `
		MATCH (n:CodeNode)
		WHERE n.workspace IS NOT NULL
		WITH n.workspace as workspace, count(n) as nodeCount
		RETURN workspace, nodeCount
		ORDER BY workspace
	`

	// Execute query
	records, err := s.executeCypherQuery(ctx, cypherQuery, map[string]any{}, false)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list workspaces: %v", err)},
			},
		}, nil
	}

	// Parse results
	var workspaces []map[string]any
	for _, record := range records {
		workspace, _ := record.Get("workspace")
		nodeCount, _ := record.Get("nodeCount")

		workspaces = append(workspaces, map[string]any{
			"workspace": workspace,
			"nodeCount": nodeCount,
		})
	}

	// Format response
	jsonData, err := json.MarshalIndent(workspaces, "", "  ")
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to format results: %v", err)},
			},
		}, nil
	}

	summary := fmt.Sprintf("Found %d workspace(s) in the graph database", len(workspaces))
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, nil
}
