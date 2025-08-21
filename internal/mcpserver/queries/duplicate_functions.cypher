MATCH (f:CodeNode {type: 'function', workspace: $workspace})
WITH f.label as functionName, collect(DISTINCT f.package) as packages
WHERE size(packages) > 1
RETURN functionName, packages, size(packages) as duplicateCount
ORDER BY duplicateCount DESC, functionName
