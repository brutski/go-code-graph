# Tools vs Prompts: When to Use Each

## Understanding the Difference

### Tools 🔧

Individual, focused operations for specific queries

Tools are the fundamental building blocks of the Go Code Graph MCP server. Each tool performs a specific operation:

- **Single purpose** - Each tool does one thing well
- **Direct control** - You specify exactly what you want
- **Flexible** - Can be combined in any way you need
- **Exploratory** - Great for ad-hoc queries and discovery

### Prompts 📋

Pre-defined workflows that chain multiple tools together

Prompts are structured templates that guide you through complex analysis tasks:

- **Multi-step workflows** - Combine multiple tools in the right order
- **Best practices built-in** - Use optimal queries and patterns
- **Comprehensive analysis** - Cover all important aspects of a task
- **Guided experience** - Show exactly which tools to use and how

## When to Use Tools

Use **individual tools** when you:

### 1. **Know exactly what you're looking for**

```text
Example: "What functions does ProcessOrder call?"
→ Use: cypher_query with a specific query
```

### 2. **Want to explore the codebase interactively**

```text
Example: "Let me see what's in the payment package"
→ Use: natural_query to explore step by step
```

### 3. **Need a quick answer to a specific question**

```text
Example: "What implements the Repository interface?"
→ Use: find_implementers
```

### 4. **Are building your own custom workflow**

```text
Example: Creating a custom analysis for your specific needs
→ Use: Multiple tools in your own sequence
```

### 5. **Want fine-grained control**

```text
Example: "Show me only functions with complexity > 20 that call database functions"
→ Use: cypher_query with custom query
```

## When to Use Prompts

Use **prompts** when you:

### 1. **Starting a new analysis task**

```text
Example: "I need to understand this new codebase"
→ Use: analyze_new_codebase prompt
```

### 2. **Want comprehensive coverage**

```text
Example: "Assess the overall code quality"
→ Use: assess_code_quality prompt
```

### 3. **Following best practices**

```text
Example: "Plan a safe refactoring"
→ Use: plan_refactoring prompt
```

### 4. **Need structured analysis**

```text
Example: "Prepare for a code review"
→ Use: prepare_code_review prompt
```

### 5. **Want to ensure nothing is missed**

```text
Example: "Analyze security implications"
→ Use: security_analysis prompt
```

## Comparison Chart

| Aspect | Tools | Prompts |
|--------|-------|---------|
| **Purpose** | Single operation | Complete workflow |
| **Control** | Direct, specific | Guided, comprehensive |
| **Flexibility** | Mix and match freely | Pre-defined sequence |
| **Best for** | Exploration, specific queries | Standard tasks, best practices |
| **Learning curve** | Need to know which tool | Just provide parameters |
| **Coverage** | What you ask for | Comprehensive analysis |

## Examples: Same Task, Different Approaches

### Task: Understanding a function's impact

**Using Tools (exploratory approach):**

```text
1. First, let me find the function:
   → natural_query: "Where is ProcessOrder defined?"

2. Check its complexity:
   → cypher_query: "MATCH (f:Function {name: 'ProcessOrder'}) RETURN f.complexity"

3. See who calls it:
   → analyze_impact: "ProcessOrder" with changeType: "modify"

4. Check for tests:
   → natural_query: "What tests cover ProcessOrder?"
```

**Using Prompts (structured approach):**

```text
Use analyze_change_impact with:
- component_name: ProcessOrder
- change_type: modify

→ Automatically runs all necessary tools in the right order
→ Provides comprehensive impact analysis
→ Includes test coverage, interfaces, and migration strategy
```

## Combined Approach

The most effective approach often combines both:

1. **Start with a prompt** for comprehensive analysis
2. **Use individual tools** to dive deeper into specific findings
3. **Create custom workflows** using tools for your specific needs

### Example Workflow

```text
1. Run analyze_new_codebase prompt → Get overview
2. Notice high complexity in payment module
3. Use natural_query tool → "Show me the most complex functions in payment"
4. Use trace_call_path tool → Understand specific execution flows
5. Run plan_refactoring prompt → Create improvement plan
```

## Quick Decision Guide

**Use Tools when:**

- ❓ You have a specific question
- 🔍 You're exploring interactively  
- 🎯 You need precise control
- 🔧 You're building custom analysis
- 💡 You know what you're looking for

**Use Prompts when:**

- 📋 You want comprehensive analysis
- 🎓 You're following best practices
- 🆕 You're starting a standard task
- ✅ You want complete coverage
- 🗺️ You need guided workflow

## Conclusion

- **Tools** = Building blocks for specific queries and custom exploration
- **Prompts** = Pre-built workflows following best practices

Both are valuable and complementary. Prompts help you get started quickly and ensure comprehensive analysis, while tools give you the flexibility to explore and answer specific questions. Use prompts for standard workflows and tools for custom exploration.
