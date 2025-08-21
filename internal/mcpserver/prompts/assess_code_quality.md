# Assess Code Quality

Please perform a comprehensive code quality assessment{{#if focus_area}} focusing on {{focus_area}}{{/if}}:

1. **Find complex functions** - Identify functions needing refactoring:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method']{{#if focus_area}} AND f.package CONTAINS '{{focus_area}}'{{/if}} AND f.complexity > 10 RETURN f.label, f.package, f.complexity ORDER BY f.complexity DESC LIMIT 20",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

2. **Identify duplicate code** - Find copy-paste patterns:

   ```json
   Use tool: find_patterns
   Arguments: {
     "patternType": "duplicate"{{#if focus_area}},
     "filter": {"package": "{{focus_area}}"}{{/if}},
     "workspace": "{{workspace}}"
   }
   ```

3. **Find unused code** - Detect dead code:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (n:CodeNode) WHERE n.type IN ['function', 'struct', 'interface']{{#if focus_area}} AND n.package CONTAINS '{{focus_area}}'{{/if}} AND n.visibility = 'private' AND NOT (n)<-[:RELATES_TO {type: 'calls'}]-() AND NOT (n)<-[:RELATES_TO {type: 'constructs'}]-() AND n.label <> 'init' RETURN n.type, n.label, n.package ORDER BY n.package, n.label",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

4. **Check for god objects** - Find overly complex structs:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (s:CodeNode {type: 'struct'})-[r:RELATES_TO {type: 'has_field'}]->(f:CodeNode) {{#if focus_area}}WHERE s.package CONTAINS '{{focus_area}}' {{/if}}WITH s, count(f) as fieldCount WHERE fieldCount > 10 OPTIONAL MATCH (s)<-[m:RELATES_TO {type: 'has_method'}]-(method:CodeNode) WITH s, fieldCount, count(method) as methodCount WHERE (fieldCount + methodCount) > 15 RETURN s.label, s.package, fieldCount, methodCount, (fieldCount + methodCount) as totalComplexity ORDER BY totalComplexity DESC",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

5. **Functions with too many parameters**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'has_parameter'}]->(p:CodeNode) WHERE f.type IN ['function', 'method']{{#if focus_area}} AND f.package CONTAINS '{{focus_area}}'{{/if}} WITH f, count(p) as paramCount WHERE paramCount > 5 RETURN f.label, f.package, paramCount, f.signature ORDER BY paramCount DESC",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

6. **Error handling patterns** - Check error handling quality:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Show me functions that return errors but might have poor error handling. Look for functions that return error type but have low complexity{{#if focus_area}} in {{focus_area}}{{/if}}",
     "context": "Checking error handling quality",
     "workspace": "{{workspace}}"
   }
   ```

Based on the results, provide:

- Prioritized list of refactoring targets
- Specific improvement recommendations
- Estimated effort for each improvement
- Quick wins vs long-term improvements
