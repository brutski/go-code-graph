#!/bin/bash
# test-mcp-tool.sh - Test MCP tools easily

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

TOOL_NAME=${1:-"list_workspaces"}
WORKSPACE=${2:-"go-code-graph"}
TOOL_ARGS=${3:-'{}'}

# If it's list_workspaces, don't add workspace parameter
if [ "$TOOL_NAME" = "list_workspaces" ]; then
    FINAL_ARGS="$TOOL_ARGS"
elif [ "$TOOL_NAME" = "analyze_workspace" ]; then
    # For analyze_workspace, use the provided args as-is
    FINAL_ARGS="$TOOL_ARGS"
else
    # For other tools, add workspace parameter if not present
    if echo "$TOOL_ARGS" | grep -q '"workspace"'; then
        FINAL_ARGS="$TOOL_ARGS"
    else
        # Insert workspace parameter into the JSON
        FINAL_ARGS=$(echo "$TOOL_ARGS" | sed "s/}$/,\"workspace\":\"$WORKSPACE\"}/")
    fi
fi

echo "Testing tool: $TOOL_NAME"
echo "Workspace: $WORKSPACE"
echo "Arguments: $FINAL_ARGS"
echo "---"

(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'; \
 sleep 0.5; \
 echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"$TOOL_NAME\",\"arguments\":$FINAL_ARGS}}")| \
NEO4J_URI="bolt://localhost:7687" \
NEO4J_USER="neo4j" \
NEO4J_PASSWORD="codeGraph123" \
EMBEDDING_PROVIDER="bedrock" \
EMBEDDING_MODEL="amazon.titan-embed-text-v2:0" \
AWS_REGION="us-east-1" \
LOG_LEVEL="debug" \
"$PROJECT_ROOT/bin/mcp-server"
