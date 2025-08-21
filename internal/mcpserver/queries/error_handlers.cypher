// Find all functions that handle errors and what error types they handle
MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'handles_error'}]->(e:CodeNode {workspace: $workspace})
WHERE f.type IN ['function', 'method']
RETURN DISTINCT f.full_name as errorHandler,
       f.package as package,
       collect(DISTINCT e.label) as handledErrors,
       count(DISTINCT e) as errorTypeCount
ORDER BY errorTypeCount DESC, f.full_name