package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// NodeBuilder handles creation of all node types
type NodeBuilder struct {
	graph   *Graph
	nodeIDs map[string]bool
}

// NewNodeBuilder creates a new node builder
func NewNodeBuilder(graph *Graph) *NodeBuilder {
	return &NodeBuilder{
		graph:   graph,
		nodeIDs: make(map[string]bool),
	}
}

// NodeExists checks if a node already exists
func (nb *NodeBuilder) NodeExists(nodeID string) bool {
	return nb.nodeIDs[nodeID]
}

// CreatePackageNode creates a package node
func (nb *NodeBuilder) CreatePackageNode(pkg *packages.Package) *Node {
	nodeID := GenerateNodeID(NodeTypePackage, "", pkg.PkgPath)
	if nb.nodeIDs[nodeID] {
		return nil
	}

	node := Node{
		ID:         nodeID,
		Label:      pkg.Name,
		Type:       NodeTypePackage,
		Package:    pkg.PkgPath,
		FullName:   pkg.PkgPath,
		Size:       len(pkg.GoFiles), // Size based on number of files
		Level:      0,
		Visibility: "public",
		Metadata: map[string]interface{}{
			"path":    pkg.PkgPath,
			"files":   len(pkg.GoFiles),
			"imports": len(pkg.Imports),
		},
	}

	nb.graph.AddNode(node)
	nb.nodeIDs[nodeID] = true
	return &node
}

// CreateTypeNode creates a struct or interface node
func (nb *NodeBuilder) CreateTypeNode(pkg *packages.Package, typeName *types.TypeName, typeSpec *ast.TypeSpec) *Node {
	var nodeType string
	var size int
	var typeInfo TypeInfo

	switch underlying := typeName.Type().Underlying().(type) {
	case *types.Interface:
		nodeType = NodeTypeInterface
		size = underlying.NumMethods()*3 + 5
		typeInfo = TypeInfoFromGoTypes(typeName.Type())
	case *types.Struct:
		nodeType = NodeTypeStruct
		size = underlying.NumFields()*2 + 8
		typeInfo = TypeInfoFromGoTypes(typeName.Type())
	default:
		// Skip other types
		return nil
	}

	nodeID := GenerateNodeID(nodeType, pkg.PkgPath, typeName.Name())
	if nb.nodeIDs[nodeID] {
		return nil
	}

	// Extract position
	position := Position{}
	if typeSpec != nil {
		pos := pkg.Fset.Position(typeSpec.Pos())
		position = Position{
			Filename: pos.Filename,
			Line:     pos.Line,
			Column:   pos.Column,
			Offset:   pos.Offset,
		}
	}

	// Extract documentation if available
	var documentation string
	if typeSpec != nil && typeSpec.Doc != nil {
		documentation = ExtractDocumentation(typeSpec.Doc)
	}

	node := Node{
		ID:            nodeID,
		Label:         typeName.Name(),
		Type:          nodeType,
		Package:       pkg.PkgPath,
		FullName:      fmt.Sprintf("%s.%s", pkg.PkgPath, typeName.Name()),
		Size:          size,
		Level:         1,
		Position:      position,
		Documentation: documentation,
		Visibility: func() string {
			if typeName.Exported() {
				return VisibilityPublic
			}
			return VisibilityPrivate
		}(),
		TypeInfo: typeInfo,
		Metadata: map[string]interface{}{
			"exported": typeName.Exported(),
			"kind":     nodeType,
		},
	}

	nb.graph.AddNode(node)
	nb.nodeIDs[nodeID] = true
	return &node
}

// CreateFunctionNode creates a function or method node
func (nb *NodeBuilder) CreateFunctionNode(pkg *packages.Package, fn *ast.FuncDecl) *Node {
	if fn.Name == nil {
		return nil
	}

	funcName := fn.Name.Name
	nodeType := NodeTypeFunction
	var receiverType string

	// Check if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		nodeType = NodeTypeMethod
		if t, ok := pkg.TypesInfo.Types[fn.Recv.List[0].Type]; ok {
			receiverType = types.TypeString(t.Type, nil)
		}
	}

	// Generate appropriate node ID
	nodeID := GenerateNodeID(nodeType, pkg.PkgPath, funcName)
	if receiverType != "" {
		nodeID = GenerateNodeID(nodeType, pkg.PkgPath, receiverType+"."+funcName)
	}

	if nb.nodeIDs[nodeID] {
		return nil
	}

	// Calculate complexity
	complexity := calculateComplexity(fn)

	// Extract position
	pos := pkg.Fset.Position(fn.Pos())
	position := Position{
		Filename: pos.Filename,
		Line:     pos.Line,
		Column:   pos.Column,
		Offset:   pos.Offset,
	}

	// Extract signature
	signature := extractFunctionSignature(pkg, fn)

	// Extract documentation
	documentation := ExtractDocumentation(fn.Doc)

	node := Node{
		ID:            nodeID,
		Label:         funcName,
		Type:          nodeType,
		Package:       pkg.PkgPath,
		FullName:      fmt.Sprintf("%s.%s", pkg.PkgPath, funcName),
		Size:          complexity + 5,
		Level:         2,
		Position:      position,
		Complexity:    complexity,
		Signature:     signature,
		Documentation: documentation,
		Visibility: func() string {
			if fn.Name.IsExported() {
				return VisibilityPublic
			}
			return VisibilityPrivate
		}(),
		Metadata: map[string]interface{}{
			"exported":   fn.Name.IsExported(),
			"receiver":   receiverType,
			"complexity": complexity,
		},
	}

	nb.graph.AddNode(node)
	nb.nodeIDs[nodeID] = true

	// For methods, create has_method edge from struct to method
	if nodeType == NodeTypeMethod && receiverType != "" {
		// Clean receiver type (remove pointer prefix and package qualifiers)
		cleanReceiverType := receiverType
		cleanReceiverType = strings.TrimPrefix(cleanReceiverType, "*")
		// Remove package qualifier if present
		if idx := strings.LastIndex(cleanReceiverType, "."); idx >= 0 {
			cleanReceiverType = cleanReceiverType[idx+1:]
		}

		// Try to find the struct node
		structNodeID := GenerateNodeID(NodeTypeStruct, pkg.PkgPath, cleanReceiverType)
		if nb.nodeIDs[structNodeID] {
			edge := Edge{
				Source: structNodeID,
				Target: nodeID,
				Type:   EdgeTypeHasMethod,
				Weight: 1,
			}
			nb.graph.AddEdge(edge)
		}
	}

	return &node
}

// CreateFieldNode creates a field node for a struct
func (nb *NodeBuilder) CreateFieldNode(structNode *Node, field *types.Var, tag string, position Position) *Node {
	fieldNodeID := GenerateNodeID(NodeTypeField, structNode.Package,
		fmt.Sprintf("%s.%s", structNode.Label, field.Name()))

	if nb.nodeIDs[fieldNodeID] {
		return nil
	}

	node := Node{
		ID:       fieldNodeID,
		Label:    field.Name(),
		Type:     NodeTypeField,
		Package:  structNode.Package,
		FullName: fmt.Sprintf("%s.%s.%s", structNode.Package, structNode.Label, field.Name()),
		Size:     3,
		Level:    2,
		Position: position,
		Visibility: func() string {
			if field.Exported() {
				return VisibilityPublic
			}
			return VisibilityPrivate
		}(),
		Metadata: map[string]interface{}{
			"exported":  field.Exported(),
			"embedded":  field.Embedded(),
			"tag":       tag,
			"type_name": field.Type().String(),
		},
	}

	nb.graph.AddNode(node)
	nb.nodeIDs[fieldNodeID] = true

	// Create HAS_FIELD relationship
	edge := Edge{
		Source: structNode.ID,
		Target: fieldNodeID,
		Type:   EdgeTypeHasField,
		Weight: 1,
	}
	nb.graph.AddEdge(edge)

	return &node
}

// CreateParameterNode creates a parameter node for a function
func (nb *NodeBuilder) CreateParameterNode(funcNode *Node, paramName string, paramType string,
	index int, position Position,
) *Node {
	if paramName == "" {
		paramName = fmt.Sprintf("param%d", index)
	}

	paramNodeID := GenerateNodeID(NodeTypeParameter, funcNode.Package,
		fmt.Sprintf("%s.%s", funcNode.Label, paramName))

	if nb.nodeIDs[paramNodeID] {
		return nil
	}

	node := Node{
		ID:         paramNodeID,
		Label:      paramName,
		Type:       NodeTypeParameter,
		Package:    funcNode.Package,
		FullName:   fmt.Sprintf("%s.%s.%s", funcNode.Package, funcNode.Label, paramName),
		Size:       2,
		Level:      3,
		Position:   position,
		Visibility: "private", // Parameters are function-scoped
		Metadata: map[string]interface{}{
			"index":     index,
			"type_name": paramType,
		},
	}

	nb.graph.AddNode(node)
	nb.nodeIDs[paramNodeID] = true

	// Create HAS_PARAMETER relationship
	edge := Edge{
		Source: funcNode.ID,
		Target: paramNodeID,
		Type:   EdgeTypeHasParameter,
		Weight: 1,
	}
	nb.graph.AddEdge(edge)

	return &node
}

// CreateGenericNode - generic type parameter support removed
func (nb *NodeBuilder) CreateGenericNode(pkg *packages.Package, funcNode *Node, name string) *Node {
	// Generic nodes (NodeTypeGeneric) have been removed as rarely used
	// Returning nil to maintain API compatibility
	return nil
}

// CreateBuiltinTypeNode creates a node for builtin types like string, int, error, etc.
func (nb *NodeBuilder) CreateBuiltinTypeNode(typeName string, isInterface bool) *Node {
	// Determine node type
	nodeType := NodeTypeStruct
	if isInterface {
		nodeType = NodeTypeInterface
	}

	nodeID := GenerateNodeID(nodeType, "builtin", typeName)
	if nb.nodeIDs[nodeID] {
		return nil
	}

	node := Node{
		ID:         nodeID,
		Label:      typeName,
		Type:       nodeType,
		Package:    "builtin",
		FullName:   typeName,
		Size:       5, // Small size for builtin types
		Level:      1,
		Visibility: VisibilityPublic,
		TypeInfo: TypeInfo{
			Kind:       "builtin",
			IsExported: true,
		},
		Metadata: map[string]interface{}{
			"builtin": true,
			"kind":    nodeType,
		},
	}

	nb.graph.AddNode(node)
	nb.nodeIDs[nodeID] = true
	return &node
}

// EnsureBuiltinTypes creates nodes for common builtin types
func (nb *NodeBuilder) EnsureBuiltinTypes() {
	// Create builtin package node
	builtinPkgID := GenerateNodeID(NodeTypePackage, "", "builtin")
	if !nb.nodeIDs[builtinPkgID] {
		builtinPkg := Node{
			ID:         builtinPkgID,
			Label:      "builtin",
			Type:       NodeTypePackage,
			Package:    "builtin",
			FullName:   "builtin",
			Size:       1,
			Level:      0,
			Visibility: VisibilityPublic,
			Metadata: map[string]interface{}{
				"path":        "builtin",
				"description": "Go builtin types",
			},
		}
		nb.graph.AddNode(builtinPkg)
		nb.nodeIDs[builtinPkgID] = true
	}

	// Common builtin types (as structs)
	builtinTypes := []string{
		"bool", "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune", "float32", "float64", "complex64", "complex128",
	}

	for _, typeName := range builtinTypes {
		nb.CreateBuiltinTypeNode(typeName, false)
	}

	// Builtin interfaces
	builtinInterfaces := []string{"error", "any"}
	for _, typeName := range builtinInterfaces {
		nb.CreateBuiltinTypeNode(typeName, true)
	}
}
