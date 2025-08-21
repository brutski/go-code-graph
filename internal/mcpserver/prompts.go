package mcpserver

import (
	"context"
	_ "embed"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Embed all prompt templates
var (
	//go:embed prompts/analyze_new_codebase.md
	analyzeNewCodebasePrompt string

	//go:embed prompts/assess_code_quality.md
	assessCodeQualityPrompt string

	//go:embed prompts/analyze_change_impact.md
	analyzeChangeImpactPrompt string

	//go:embed prompts/debug_execution_flow.md
	debugExecutionFlowPrompt string

	//go:embed prompts/plan_refactoring.md
	planRefactoringPrompt string

	//go:embed prompts/understand_architecture.md
	understandArchitecturePrompt string

	//go:embed prompts/analyze_interface_usage.md
	analyzeInterfaceUsagePrompt string

	//go:embed prompts/find_code_patterns.md
	findCodePatternsPrompt string

	//go:embed prompts/analyze_dependencies.md
	analyzeDependenciesPrompt string

	//go:embed prompts/find_performance_issues.md
	findPerformanceIssuesPrompt string

	//go:embed prompts/analyze_test_coverage.md
	analyzeTestCoveragePrompt string

	//go:embed prompts/plan_api_migration.md
	planAPIMigrationPrompt string

	//go:embed prompts/prepare_code_review.md
	prepareCodeReviewPrompt string

	//go:embed prompts/onboard_new_developer.md
	onboardNewDeveloperPrompt string

	//go:embed prompts/security_analysis.md
	securityAnalysisPrompt string
)

// registerPrompts registers all MCP prompts for better user experience
func (s *Server) registerPrompts() {
	// Register each prompt with its handler
	prompts := []struct {
		name        string
		title       string
		description string
		args        []*mcp.PromptArgument
		template    string
	}{
		{
			name:        "analyze_new_codebase",
			title:       "Analyze New Codebase",
			description: "Analyze a Go codebase for the first time with comprehensive overview",
			args: []*mcp.PromptArgument{
				{Name: "workspace_path", Description: "Path to the Go project directory", Required: true},
				{Name: "workspace_name", Description: "Name for the workspace", Required: false},
				{Name: "include_external", Description: "External packages to include (comma-separated)", Required: false},
			},
			template: analyzeNewCodebasePrompt,
		},
		{
			name:        "assess_code_quality",
			title:       "Assess Code Quality",
			description: "Perform comprehensive code quality assessment with actionable insights",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "focus_area", Description: "Specific package or component to focus on", Required: false},
			},
			template: assessCodeQualityPrompt,
		},
		{
			name:        "understand_architecture",
			title:       "Understand Architecture",
			description: "Get a comprehensive understanding of the codebase architecture",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "component", Description: "Specific component or service to analyze", Required: false},
			},
			template: understandArchitecturePrompt,
		},
		{
			name:        "analyze_change_impact",
			title:       "Analyze Change Impact",
			description: "Analyze the impact of changing a specific component",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "component_name", Description: "Name of the function, struct, or interface to change", Required: true},
				{Name: "change_type", Description: "Type of change: signature, delete, modify, rename", Required: true},
			},
			template: analyzeChangeImpactPrompt,
		},
		{
			name:        "debug_execution_flow",
			title:       "Debug Execution Flow",
			description: "Trace and understand execution flow for debugging purposes",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "entry_point", Description: "Starting function or handler", Required: true},
				{Name: "target_point", Description: "Target function or suspected problem area", Required: true},
			},
			template: debugExecutionFlowPrompt,
		},
		{
			name:        "analyze_interface_usage",
			title:       "Analyze Interface Usage",
			description: "Understand how interfaces are used throughout the codebase",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "interface_name", Description: "Name of the interface to analyze", Required: true},
			},
			template: analyzeInterfaceUsagePrompt,
		},
		{
			name:        "plan_refactoring",
			title:       "Plan Refactoring",
			description: "Get detailed guidance for refactoring a specific component",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "target", Description: "Function, package, or component to refactor", Required: true},
				{Name: "goal", Description: "Refactoring goal (e.g., reduce complexity, extract interface)", Required: true},
			},
			template: planRefactoringPrompt,
		},
		{
			name:        "find_code_patterns",
			title:       "Find Code Patterns",
			description: "Search for specific code patterns or implementations",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "pattern", Description: "Pattern to search for (e.g., error handling, HTTP handlers)", Required: true},
				{Name: "package_filter", Description: "Limit search to specific package", Required: false},
			},
			template: findCodePatternsPrompt,
		},
		{
			name:        "analyze_dependencies",
			title:       "Analyze Dependencies",
			description: "Analyze dependencies between packages or components",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "source_package", Description: "Source package to analyze", Required: true},
				{Name: "include_external", Description: "Include external dependencies (true/false)", Required: false},
			},
			template: analyzeDependenciesPrompt,
		},
		{
			name:        "find_performance_issues",
			title:       "Find Performance Issues",
			description: "Identify potential performance bottlenecks",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "focus_area", Description: "Specific area to analyze", Required: false},
			},
			template: findPerformanceIssuesPrompt,
		},
		{
			name:        "analyze_test_coverage",
			title:       "Analyze Test Coverage",
			description: "Analyze test coverage and identify gaps",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "package", Description: "Package to analyze", Required: true},
			},
			template: analyzeTestCoveragePrompt,
		},
		{
			name:        "plan_api_migration",
			title:       "Plan API Migration",
			description: "Plan migration from one API/interface to another",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "old_api", Description: "Current API or interface being used", Required: true},
				{Name: "new_api", Description: "New API or interface to migrate to", Required: true},
			},
			template: planAPIMigrationPrompt,
		},
		{
			name:        "prepare_code_review",
			title:       "Prepare Code Review",
			description: "Prepare for code review by analyzing changes",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "changed_components", Description: "List of changed functions/structs (comma-separated)", Required: true},
			},
			template: prepareCodeReviewPrompt,
		},
		{
			name:        "onboard_new_developer",
			title:       "Onboard New Developer",
			description: "Get an onboarding overview for a new developer",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "focus_area", Description: "Specific area the developer will work on", Required: false},
			},
			template: onboardNewDeveloperPrompt,
		},
		{
			name:        "security_analysis",
			title:       "Security Analysis",
			description: "Analyze security-sensitive code paths",
			args: []*mcp.PromptArgument{
				{Name: "workspace", Description: "Workspace to analyze", Required: true},
				{Name: "entry_points", Description: "Type of entry points to analyze (HTTP, API, CLI)", Required: true},
			},
			template: securityAnalysisPrompt,
		},
	}

	// Register all prompts
	for _, p := range prompts {
		// Create a copy of the template for this closure
		template := p.template
		promptName := p.name
		s.mcpServer.AddPrompt(&mcp.Prompt{
			Name:        promptName,
			Title:       p.title,
			Description: p.description,
			Arguments:   p.args,
		}, func(ctx context.Context, session *mcp.ServerSession, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
			return s.handleTemplatePrompt(ctx, session, params, template)
		})
	}

	s.logger.Debug("Registered MCP prompts", "count", len(prompts))
}

// handleTemplatePrompt processes a template-based prompt
func (s *Server) handleTemplatePrompt(ctx context.Context, session *mcp.ServerSession, params *mcp.GetPromptParams, template string) (*mcp.GetPromptResult, error) {
	// Replace template variables
	content := template
	if params.Arguments != nil {
		for key, value := range params.Arguments {
			// Handle basic replacements
			content = strings.ReplaceAll(content, "{{"+key+"}}", value)

			// Handle conditional blocks {{#if key}}...{{/if}}
			ifStart := "{{#if " + key + "}}"
			ifEnd := "{{/if}}"
			for {
				start := strings.Index(content, ifStart)
				if start == -1 {
					break
				}
				end := strings.Index(content[start:], ifEnd)
				if end == -1 {
					break
				}
				end += start + len(ifEnd)

				if value != "" {
					// Keep the content between if tags
					content = content[:start] + content[start+len(ifStart):end-len(ifEnd)] + content[end:]
				} else {
					// Remove the entire if block
					content = content[:start] + content[end:]
				}
			}

			// Handle unless blocks {{#unless key}}...{{/unless}}
			unlessStart := "{{#unless " + key + "}}"
			unlessEnd := "{{/unless}}"
			for {
				start := strings.Index(content, unlessStart)
				if start == -1 {
					break
				}
				end := strings.Index(content[start:], unlessEnd)
				if end == -1 {
					break
				}
				end += start + len(unlessEnd)

				if value == "" {
					// Keep the content between unless tags
					content = content[:start] + content[start+len(unlessStart):end-len(unlessEnd)] + content[end:]
				} else {
					// Remove the entire unless block
					content = content[:start] + content[end:]
				}
			}
		}
	}

	// Clean up any remaining template variables that weren't provided
	content = cleanupTemplate(content)

	return &mcp.GetPromptResult{
		Description: "Generated prompt with tool usage examples",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}

// cleanupTemplate removes any remaining template variables
func cleanupTemplate(content string) string {
	// Remove simple variables
	for {
		start := strings.Index(content, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(content[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2
		content = content[:start] + content[end:]
	}

	return content
}
