package mcpserver

import (
	"context"
	"log/slog"

	"github.com/brutski/go-code-graph/internal/analyzer"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Server represents the Code Graph MCP server
type Server struct {
	neo4jDriver         neo4j.DriverWithContext
	embeddingsGenerator *analyzer.EmbeddingsGenerator // Use interface from embeddings package
	mcpServer           *mcp.Server
	logger              *slog.Logger
}

// NewServer creates a new Code Graph MCP server with provided clients
func NewServer(neo4jDriver neo4j.DriverWithContext, embeddingsGenerator *analyzer.EmbeddingsGenerator, logger *slog.Logger) (*Server, error) {
	logger.Debug("Initializing MCP server")

	// Create MCP server
	implementation := &mcp.Implementation{
		Name:    "code-graph-mcp-server",
		Version: "1.0.0",
	}

	mcpServer := mcp.NewServer(implementation, nil)

	server := &Server{
		neo4jDriver:         neo4jDriver,
		embeddingsGenerator: embeddingsGenerator,
		mcpServer:           mcpServer,
		logger:              logger,
	}

	// Register tools
	server.registerTools()

	// Register prompts
	server.registerPrompts()

	// Register resources
	server.registerResources()

	return server, nil
}

// Run starts the MCP server with the given transport
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.mcpServer.Run(ctx, transport)
}

// Close closes the server and Neo4j connection
func (s *Server) Close() {
	ctx := context.Background()
	s.neo4jDriver.Close(ctx)
}

// registerTools registers all MCP tools
func (s *Server) registerTools() {
	// Register Cypher query tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cypher_query",
		Description: "Execute a Cypher query against the code graph database",
	}, s.handleCypherQuery)

	// Register natural language query tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "natural_query",
		Description: "Ask questions about the codebase in natural language",
	}, s.handleNaturalQuery)

	// Register impact analysis tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "analyze_impact",
		Description: "Analyze the impact of changing a specific code component",
	}, s.handleImpactAnalysis)

	// Register pattern detection tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "find_patterns",
		Description: "Find code patterns, duplicates, or anti-patterns",
	}, s.handlePatternDetection)

	// Register workspace analysis tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "analyze_workspace",
		Description: "Analyze a workspace/repository and import it to the graph database",
	}, s.handleAnalyzeWorkspace)

	// Register find implementers tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "find_implementers",
		Description: "Find all types that implement a given interface",
	}, s.handleFindImplementers)

	// Register trace call path tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "trace_call_path",
		Description: "Find the call path between two functions/methods",
	}, s.handleTraceCallPath)

	// Register architecture analysis tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "detect_architecture",
		Description: "Detect architectural patterns and violations",
	}, s.handleArchitectureAnalysis)

	// Register list workspaces tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_workspaces",
		Description: "List all available workspaces in the graph database",
	}, s.handleListWorkspaces)

	// Additional tools can be registered here as handlers are implemented
}

// registerResources registers all MCP resources
func (s *Server) registerResources() {
	// We'll add resource registration here later
	s.logger.Debug("Resources will be registered in future implementation")
}

// Public wrapper methods for testing

// HandleNaturalQuery is a public wrapper for the natural query handler (testing only)
func (s *Server) HandleNaturalQuery(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[NaturalQueryParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handleNaturalQuery(ctx, session, params)
}

// HandleArchitectureAnalysis is a public wrapper for the architecture analysis handler (testing only)
func (s *Server) HandleArchitectureAnalysis(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ArchitectureAnalysisParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handleArchitectureAnalysis(ctx, session, params)
}

// HandlePatternDetection is a public wrapper for the pattern detection handler (testing only)
func (s *Server) HandlePatternDetection(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[PatternDetectionParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handlePatternDetection(ctx, session, params)
}

// HandleFindImplementers is a public wrapper for the find implementers handler (testing only)
func (s *Server) HandleFindImplementers(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[FindImplementersParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handleFindImplementers(ctx, session, params)
}

// HandleTraceCallPath is a public wrapper for the trace call path handler (testing only)
func (s *Server) HandleTraceCallPath(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[TraceCallPathParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handleTraceCallPath(ctx, session, params)
}

// HandleAnalyzeImpact is a public wrapper for the impact analysis handler (testing only)
func (s *Server) HandleAnalyzeImpact(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ImpactAnalysisParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handleImpactAnalysis(ctx, session, params)
}

// HandleCypherQuery is a public wrapper for the cypher query handler (testing only)
func (s *Server) HandleCypherQuery(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CypherQueryParams]) (*mcp.CallToolResultFor[any], error) {
	return s.handleCypherQuery(ctx, session, params)
}
