#!/bin/bash
# setup-test-workspace.sh - Setup a test workspace for MCP tools testing

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "======================================"
echo "Setup Test Workspace for MCP Tools"
echo "======================================"
echo

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Default values
DEFAULT_WORKSPACE="go-code-graph"
DEFAULT_PATH="$PROJECT_ROOT"

WORKSPACE_NAME=${1:-$DEFAULT_WORKSPACE}
WORKSPACE_PATH=${2:-$DEFAULT_PATH}

echo -e "${BLUE}Workspace Name:${NC} $WORKSPACE_NAME"
echo -e "${BLUE}Workspace Path:${NC} $WORKSPACE_PATH"
echo

# Check if workspace already exists
echo -e "${BLUE}=== Checking Existing Workspaces ===${NC}"
"$SCRIPT_DIR/test-mcp-tool.sh" list_workspaces

echo
echo -e "${BLUE}=== Analyzing Workspace ===${NC}"
echo "This will analyze the codebase and import it into Neo4j..."
echo

# Analyze the workspace
"$SCRIPT_DIR/test-mcp-tool.sh" analyze_workspace "$WORKSPACE_NAME" "{\"workspacePath\":\"$WORKSPACE_PATH\",\"workspaceName\":\"$WORKSPACE_NAME\",\"incremental\":false}"

echo
echo -e "${GREEN}Setup complete!${NC}"
echo
echo "You can now run the test scripts with this workspace:"
echo "  ./scripts/test-mcp-tools.sh $WORKSPACE_NAME"
echo "  ./scripts/test-mcp-tool.sh <tool_name> $WORKSPACE_NAME '<arguments>'"