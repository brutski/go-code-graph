// Find call paths between functions/methods
// If simple names are provided, find all possible paths
// If full names are provided, find specific path
MATCH (start:CodeNode {workspace: $workspace}), (end:CodeNode {workspace: $workspace})
WHERE (start.full_name = $from OR start.label = $from)
  AND (end.full_name = $to OR end.label = $to)
  AND start.type IN ['function', 'method']
  AND end.type IN ['function', 'method']
WITH start, end
MATCH path = shortestPath(
    (start)-[r:RELATES_TO*1..10]->(end)
)
WHERE all(rel in r WHERE rel.type = 'calls')
  AND all(n in nodes(path) WHERE n.workspace = $workspace)
RETURN DISTINCT 
       start.full_name as fromFunction,
       end.full_name as toFunction,
       [n in nodes(path) | n.full_name] as callPath,
       [n in nodes(path) | n.label] as callNames,
       length(path) as pathLength
ORDER BY pathLength ASC
LIMIT 5
