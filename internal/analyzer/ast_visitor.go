package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// ASTVisitor traverses the AST and builds nodes/relationships
type ASTVisitor struct {
	pkg                  *packages.Package
	nodeBuilder          *NodeBuilder
	relationshipAnalyzer *RelationshipAnalyzer
	fileDecls            map[string][]ast.Decl // Track declarations per file
	currentFile          *ast.File
}

// NewASTVisitor creates a new AST visitor
func NewASTVisitor(pkg *packages.Package, nodeBuilder *NodeBuilder, relAnalyzer *RelationshipAnalyzer) *ASTVisitor {
	return &ASTVisitor{
		pkg:                  pkg,
		nodeBuilder:          nodeBuilder,
		relationshipAnalyzer: relAnalyzer,
		fileDecls:            make(map[string][]ast.Decl),
	}
}

// BuildNodes builds all nodes for the package
func (v *ASTVisitor) BuildNodes() {
	// Create package node first
	v.nodeBuilder.CreatePackageNode(v.pkg)

	// Process type declarations first (structs, interfaces)
	v.processTypeDeclarations()

	// Process functions and methods
	v.processFunctionDeclarations()

	// Process fields and parameters (requires types and functions to exist)
	v.processFieldsAndParameters()

	// Note: Relationships are analyzed in Phase 2 after all nodes are created
	// and the node index is built. Do NOT analyze relationships here.
}

// processTypeDeclarations processes all type declarations in the package
func (v *ASTVisitor) processTypeDeclarations() {
	scope := v.pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			// Find the corresponding AST TypeSpec
			typeSpec := v.findTypeSpec(typeName.Name())
			v.nodeBuilder.CreateTypeNode(v.pkg, typeName, typeSpec)
		}
	}
}

// processFunctionDeclarations processes all function declarations
func (v *ASTVisitor) processFunctionDeclarations() {
	for _, file := range v.pkg.Syntax {
		v.currentFile = file
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				v.nodeBuilder.CreateFunctionNode(v.pkg, fn)
				// Note: Function relationships are analyzed in Phase 2 after all nodes are created
			}
			return true
		})
	}
}

// processFieldsAndParameters creates field and parameter nodes
func (v *ASTVisitor) processFieldsAndParameters() {
	// Process struct fields
	scope := v.pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			if structType, ok := typeName.Type().Underlying().(*types.Struct); ok {
				v.processStructFields(typeName, structType)
			}
		}
	}

	// Process function parameters
	for _, file := range v.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				v.processFunctionParameters(fn)
			}
			return true
		})
	}
}

// processStructFields creates field nodes for a struct
func (v *ASTVisitor) processStructFields(typeName *types.TypeName, structType *types.Struct) {
	structNodeID := GenerateNodeID(NodeTypeStruct, v.pkg.PkgPath, typeName.Name())
	structNode := v.findNodeByID(structNodeID)
	if structNode == nil {
		return
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)

		// Find field position
		position := v.findFieldPosition(typeName.Name(), field.Name(), i)

		v.nodeBuilder.CreateFieldNode(structNode, field, tag, position)
	}
}

// processFunctionParameters creates parameter nodes for a function
func (v *ASTVisitor) processFunctionParameters(fn *ast.FuncDecl) {
	if fn.Name == nil || fn.Type.Params == nil {
		return
	}

	funcName := fn.Name.Name
	receiverType := ""

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		if t, ok := v.pkg.TypesInfo.Types[fn.Recv.List[0].Type]; ok {
			receiverType = types.TypeString(t.Type, nil)
		}
	}

	var funcNodeID string
	if receiverType != "" {
		funcNodeID = GenerateNodeID(NodeTypeMethod, v.pkg.PkgPath, receiverType+"."+funcName)
	} else {
		funcNodeID = GenerateNodeID(NodeTypeFunction, v.pkg.PkgPath, funcName)
	}

	funcNode := v.findNodeByID(funcNodeID)
	if funcNode == nil {
		return
	}

	// Process type parameters (generics)
	if fn.Type.TypeParams != nil {
		for _, typeParam := range fn.Type.TypeParams.List {
			for _, name := range typeParam.Names {
				v.nodeBuilder.CreateGenericNode(v.pkg, funcNode, name.Name)
			}
		}
	}

	// Process regular parameters
	paramIndex := 0
	for _, paramGroup := range fn.Type.Params.List {
		paramType := ""
		if t, ok := v.pkg.TypesInfo.Types[paramGroup.Type]; ok {
			paramType = types.TypeString(t.Type, nil)
		}

		for _, name := range paramGroup.Names {
			pos := v.pkg.Fset.Position(name.Pos())
			position := Position{
				Filename: pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Offset:   pos.Offset,
			}

			v.nodeBuilder.CreateParameterNode(funcNode, name.Name, paramType, paramIndex, position)
			paramIndex++
		}
	}
}

// findTypeSpec finds the AST TypeSpec for a given type name
func (v *ASTVisitor) findTypeSpec(typeName string) *ast.TypeSpec {
	for _, file := range v.pkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.Name == typeName {
							return typeSpec
						}
					}
				}
			}
		}
	}
	return nil
}

// findFieldPosition finds the position of a field in a struct
func (v *ASTVisitor) findFieldPosition(typeName, fieldName string, fieldIndex int) Position {
	typeSpec := v.findTypeSpec(typeName)
	if typeSpec == nil {
		return Position{}
	}

	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		if structType.Fields != nil && structType.Fields.List != nil {
			currentIndex := 0
			for _, fieldGroup := range structType.Fields.List {
				for _, name := range fieldGroup.Names {
					if name.Name == fieldName && currentIndex == fieldIndex {
						pos := v.pkg.Fset.Position(name.Pos())
						return Position{
							Filename: pos.Filename,
							Line:     pos.Line,
							Column:   pos.Column,
							Offset:   pos.Offset,
						}
					}
					currentIndex++
				}
			}
		}
	}

	return Position{}
}

// findNodeByID finds a node by its ID
func (v *ASTVisitor) findNodeByID(nodeID string) *Node {
	for i := range v.nodeBuilder.graph.Nodes {
		if v.nodeBuilder.graph.Nodes[i].ID == nodeID {
			return &v.nodeBuilder.graph.Nodes[i]
		}
	}
	return nil
}
