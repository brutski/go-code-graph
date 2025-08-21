#!/bin/bash
# test-mcp-tools.sh - Comprehensive MCP tools testing script with workspace support

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MCP_SERVER="$PROJECT_ROOT/bin/mcp-server"
NEO4J_URI="bolt://localhost:7687"
NEO4J_USER="neo4j"
NEO4J_PASSWORD="codeGraph123"
EMBEDDING_PROVIDER="bedrock"
EMBEDDING_MODEL="amazon.titan-embed-text-v2:0"
AWS_REGION="us-east-1"
LOG_LEVEL="info"  # Change to debug for more output

# Default workspace - can be overridden by command line argument
DEFAULT_WORKSPACE="go-code-graph"
WORKSPACE=${1:-$DEFAULT_WORKSPACE}

echo "======================================"
echo "MCP Tools Testing Script"
echo "Using workspace: $WORKSPACE"
echo "======================================"
echo

# Function to call MCP tool
call_mcp_tool() {
    local tool_name="$1"
    local tool_args="$2"
    local skip_workspace="$3"
    
    # Add workspace parameter if not skipped and not already present
    if [ "$skip_workspace" != "true" ] && [ "$tool_name" != "list_workspaces" ] && [ "$tool_name" != "analyze_workspace" ]; then
        if ! echo "$tool_args" | grep -q '"workspace"'; then
            tool_args=$(echo "$tool_args" | sed "s/}$/,\"workspace\":\"$WORKSPACE\"}/")
        fi
    fi
    
    echo -e "${BLUE}Testing tool: ${YELLOW}$tool_name${NC}"
    echo -e "Arguments: $tool_args"
    echo "---"
    
    result=$(
        (echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'; \
         sleep 0.5; \
         echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool_name\",\"arguments\":$tool_args}}")| \
        NEO4J_URI="$NEO4J_URI" \
        NEO4J_USER="$NEO4J_USER" \
        NEO4J_PASSWORD="$NEO4J_PASSWORD" \
        EMBEDDING_PROVIDER="$EMBEDDING_PROVIDER" \
        EMBEDDING_MODEL="$EMBEDDING_MODEL" \
        AWS_REGION="$AWS_REGION" \
        LOG_LEVEL="$LOG_LEVEL" \
        "$MCP_SERVER" 2>&1
    )
    
    # Check if successful
    if echo "$result" | grep -q '"error"'; then
        echo -e "${RED}✗ Failed${NC}"
        echo "$result" | grep -A5 -B5 '"error"' | sed 's/^/  /'
    elif echo "$result" | grep -q '"result"'; then
        echo -e "${GREEN}✓ Success${NC}"
        # Pretty print the result
        echo "$result" | grep '"result"' | python3 -m json.tool 2>/dev/null || echo "$result" | grep '"result"'
    else
        echo -e "${YELLOW}⚠ Unexpected response${NC}"
        echo "$result" | tail -20
    fi
    echo
    echo "=========================================="
    echo
}

# Function to run a diagnostic query
run_diagnostic() {
    local description="$1"
    local query="$2"
    
    echo -e "${BLUE}Diagnostic: ${YELLOW}$description${NC}"
    call_mcp_tool "cypher_query" "{\"query\":\"$query\",\"parameters\":{}}"
}

# First, check if workspaces exist
echo -e "${BLUE}=== Checking Available Workspaces ===${NC}"
call_mcp_tool "list_workspaces" "{}" "true"

# Optional: Analyze a workspace if needed
echo -e "${BLUE}=== Setup: Ensure workspace exists ===${NC}"
echo "To analyze a new workspace, run:"
echo "./scripts/test-mcp-tool.sh analyze_workspace '$WORKSPACE' '{\"workspacePath\":\"/path/to/code\",\"workspaceName\":\"$WORKSPACE\",\"incremental\":false}'"
echo

# Check if Neo4j has data for this workspace
echo -e "${BLUE}=== Checking Neo4j Connection and Data ===${NC}"
run_diagnostic "Count all nodes in workspace" "MATCH (n {workspace: '$WORKSPACE'}) RETURN count(n) as totalNodes"
run_diagnostic "Count by node type" "MATCH (n {workspace: '$WORKSPACE'}) RETURN labels(n)[0] as type, count(n) as count ORDER BY count DESC"
run_diagnostic "List interfaces" "MATCH (i:CodeNode {type: 'interface', workspace: '$WORKSPACE'}) RETURN i.label as name, i.package LIMIT 10"

# Test 1: natural_query
echo -e "${BLUE}=== Test 1: natural_query ===${NC}"
call_mcp_tool "natural_query" '{"question":"What are the most complex functions?","context":"Looking for refactoring candidates"}'
call_mcp_tool "natural_query" '{"question":"Which structs have the most fields?"}'
call_mcp_tool "natural_query" '{"question":"Show me unused functions"}'
call_mcp_tool "natural_query" '{"question":"What interfaces exist in the codebase?"}'

# Test 2: cypher_query
echo -e "${BLUE}=== Test 2: cypher_query ===${NC}"
call_mcp_tool "cypher_query" "{\"query\":\"MATCH (f:CodeNode {type: 'function', workspace: '$WORKSPACE'}) WHERE f.complexity > 5 RETURN f.label, f.package, f.complexity ORDER BY f.complexity DESC LIMIT 5\",\"parameters\":{}}"
call_mcp_tool "cypher_query" "{\"query\":\"MATCH (p1:CodeNode {type: 'package', workspace: '$WORKSPACE'})-[:RELATES_TO {type: 'imports'}]->(p2:CodeNode {type: 'package'}) RETURN p1.name, p2.name LIMIT 5\",\"parameters\":{}}"

# Test 3: find_patterns
echo -e "${BLUE}=== Test 3: find_patterns ===${NC}"
call_mcp_tool "find_patterns" '{"patternType":"duplicate"}'
call_mcp_tool "find_patterns" '{"patternType":"usage"}'
call_mcp_tool "find_patterns" '{"patternType":"error_handlers"}'
call_mcp_tool "find_patterns" '{"patternType":"goroutine_spawners"}'

# Test 4: find_implementers
echo -e "${BLUE}=== Test 4: find_implementers ===${NC}"
# First find an interface
echo "Finding available interfaces..."
run_diagnostic "Find any interface" "MATCH (i:CodeNode {type: 'interface', workspace: '$WORKSPACE'}) RETURN i.label LIMIT 5"
# Then test with a common interface name
call_mcp_tool "find_implementers" '{"interfaceName":"error"}'

# Test 5: detect_architecture
echo -e "${BLUE}=== Test 5: detect_architecture ===${NC}"
call_mcp_tool "detect_architecture" '{"analysisType":"layers"}'
call_mcp_tool "detect_architecture" '{"analysisType":"patterns"}'

# Test 6: analyze_impact (needs a valid node ID)
echo -e "${BLUE}=== Test 6: analyze_impact ===${NC}"
echo "Finding a function to analyze..."
run_diagnostic "Get a function ID" "MATCH (f:CodeNode {type: 'function', workspace: '$WORKSPACE'}) RETURN f.id, f.label LIMIT 1"
echo "Note: Use the function ID from above with analyze_impact tool"

# Test 7: trace_call_path
echo -e "${BLUE}=== Test 7: trace_call_path ===${NC}"
echo "Finding functions for path tracing..."
run_diagnostic "Find main functions" "MATCH (f:CodeNode {type: 'function', label: 'main', workspace: '$WORKSPACE'}) RETURN f.label, f.package LIMIT 5"
call_mcp_tool "trace_call_path" '{"from":"main","to":"init"}'

echo
echo "======================================"
echo "Testing Complete!"
echo "======================================"
echo
echo "Summary:"
echo "- Workspace tested: $WORKSPACE"
echo "- To test a different workspace, run: $0 <workspace_name>"
echo "- To analyze a new codebase first, use the analyze_workspace tool"
echo
echo "Common issues:"
echo "1. If no data found, analyze a workspace first"
echo "2. Some tools need valid node IDs from the specific workspace"
echo "3. Interface names must exist in the analyzed codebase"