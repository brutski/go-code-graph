// Find all functions that spawn goroutines
MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'spawns_goroutine'}]->(g:CodeNode {workspace: $workspace})
WHERE f.type IN ['function', 'method']
RETURN DISTINCT f.full_name as spawner,
       f.package as package,
       collect(DISTINCT g.label) as goroutineFunctions,
       count(DISTINCT g) as goroutineCount,
       f.complexity as spawnerComplexity
ORDER BY goroutineCount DESC, f.full_name