# Find Code Patterns

Find all instances of {{pattern}}{{#if package_filter}} in {{package_filter}}{{/if}}:

1. **Search for patterns**:

   ```json
   Use tool: find_patterns
   Arguments: {
     "patternType": "{{pattern}}"{{#if package_filter}},
     "filter": {"package": "{{package_filter}}"}{{/if}},
     "workspace": "{{workspace}}"
   }
   ```

2. **Find similar implementations**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find functions that implement {{pattern}} patterns{{#if package_filter}} in {{package_filter}}{{/if}}. Show different approaches used",
     "context": "Looking for implementation patterns",
     "workspace": "{{workspace}}"
   }
   ```

Provide examples and best practice recommendations.
