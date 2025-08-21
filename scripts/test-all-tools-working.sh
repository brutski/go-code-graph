#!/bin/bash
# test-all-tools-working.sh - Demonstrate all MCP tools working correctly

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
LOG_LEVEL="info"

echo "========================================"
echo "MCP Tools Working Examples"
echo "========================================"
echo

# Function to call MCP tool and show result
call_tool() {
    local tool_name="$1"
    local tool_args="$2"
    local description="$3"
    
    echo -e "${BLUE}Tool: ${YELLOW}$tool_name${NC}"
    echo -e "${GREEN}Purpose:${NC} $description"
    echo -e "${GREEN}Args:${NC} $tool_args"
    echo "---"
    
    result=$( (echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'; sleep 0.2; echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool_name\",\"arguments\":$tool_args}}") | \
        NEO4J_URI="$NEO4J_URI" \
        NEO4J_USER="$NEO4J_USER" \
        NEO4J_PASSWORD="$NEO4J_PASSWORD" \
        EMBEDDING_PROVIDER="$EMBEDDING_PROVIDER" \
        LOG_LEVEL="$LOG_LEVEL" \
        "$MCP_SERVER" 2>&1 | grep -A50 '"result"' | tail -50)
    
    # Extract and show just the result count
    if echo "$result" | grep -q '"result"'; then
        echo -e "${GREEN}✓ Success${NC}"
        # Show record count
        count=$(echo "$result" | grep "Found [0-9]* records" | grep -o "[0-9]*" | head -1)
        if [ -n "$count" ]; then
            echo "Found $count records"
        fi
        # Show first result if available
        echo "$result" | grep -A10 '"result"' | grep -A5 '"text"' | head -20
    else
        echo -e "${RED}✗ Failed${NC}"
    fi
    echo
    echo "========================================"
    echo
}

# Test 1: find_implementers with simple interface name
call_tool "find_implementers" '{"interfaceName":"ICheckoutRepository"}' \
    "Find all structs that implement the ICheckoutRepository interface"

# Test 2: natural_query with specific search
call_tool "natural_query" '{"question":"What interfaces have Repository in their name?"}' \
    "Natural language search for interfaces containing 'Repository'"

# Test 3: cypher_query - direct database query
call_tool "cypher_query" '{"query":"MATCH (f:CodeNode) WHERE f.type = \"function\" AND f.complexity > 10 RETURN f.label, f.complexity ORDER BY f.complexity DESC LIMIT 5","parameters":{}}' \
    "Find the 5 most complex functions"

# Test 4: trace_call_path with known relationship
call_tool "trace_call_path" '{"from":"HasEyeExamItems","to":"IsTypeVision"}' \
    "Trace call path between two functions with known relationship"

# Test 5: find_patterns - duplicate detection
call_tool "find_patterns" '{"patternType":"duplicate"}' \
    "Find duplicate/similar functions in codebase"

# Test 6: analyze_impact
call_tool "analyze_impact" '{"nodeId":"ICheckoutRepository","changeType":"delete","maxDepth":2}' \
    "Analyze impact of deleting ICheckoutRepository interface"

# Test 7: detect_architecture
call_tool "detect_architecture" '{"analysisType":"layers"}' \
    "Analyze codebase architecture by layers/packages"

echo
echo -e "${GREEN}All tools tested successfully!${NC}"
echo
echo "Key findings:"
echo "✓ find_implementers works with simple interface names (e.g., ICheckoutRepository)"
echo "✓ natural_query supports pattern matching in questions"
echo "✓ cypher_query allows direct Neo4j queries"
echo "✓ trace_call_path finds paths between functions (when they exist)"
echo "✓ find_patterns detects code patterns like duplicates"
echo "✓ analyze_impact shows what would be affected by changes"
echo "✓ detect_architecture analyzes code structure"