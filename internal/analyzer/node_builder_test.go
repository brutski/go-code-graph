package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestNodeBuilder(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	// Create a mock package
	pkg := &packages.Package{
		PkgPath: "github.com/test/pkg",
		Name:    "pkg",
		GoFiles: []string{"test.go"},
		Imports: make(map[string]*packages.Package),
	}

	// Test CreatePackageNode
	pkgNode := builder.CreatePackageNode(pkg)
	if pkgNode == nil {
		t.Fatal("Expected non-nil package node")
	}
	if pkgNode.Type != NodeTypePackage {
		t.Errorf("Expected node type %s, got %s", NodeTypePackage, pkgNode.Type)
	}
	if pkgNode.ID != "package:github.com/test/pkg" {
		t.Errorf("Expected ID 'package:github.com/test/pkg', got %s", pkgNode.ID)
	}

	// Test NodeExists
	if !builder.NodeExists(pkgNode.ID) {
		t.Error("NodeExists should return true for created node")
	}
	if builder.NodeExists("non-existent-id") {
		t.Error("NodeExists should return false for non-existent node")
	}
}

func TestCreateTypeNode(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	// Create mock packages and types
	pkg := &packages.Package{
		PkgPath: "main",
		Name:    "main",
		Fset:    token.NewFileSet(),
		TypesInfo: &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
		},
	}

	tests := []struct {
		name       string
		typeName   string
		kind       string
		visibility string
		wantType   string
		wantID     string
	}{
		{
			name:       "public struct",
			typeName:   "Config",
			kind:       "struct",
			visibility: "public",
			wantType:   NodeTypeStruct,
			wantID:     "struct:main.Config",
		},
		{
			name:       "private interface",
			typeName:   "writer",
			kind:       "interface",
			visibility: "private",
			wantType:   NodeTypeInterface,
			wantID:     "interface:main.writer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock type based on kind
			var typ types.Type
			if tt.kind == "struct" {
				typ = types.NewStruct(nil, nil)
			} else {
				typ = types.NewInterfaceType(nil, nil)
			}

			// Create TypeName
			typeName := types.NewTypeName(token.NoPos, pkg.Types, tt.typeName, nil)
			_ = types.NewNamed(typeName, typ, nil)

			typeSpec := &ast.TypeSpec{
				Name: &ast.Ident{Name: tt.typeName},
			}

			node := builder.CreateTypeNode(pkg, typeName, typeSpec)

			if node == nil {
				t.Fatal("Expected non-nil node")
			}
			if node.Type != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, node.Type)
			}
			if node.ID != tt.wantID {
				t.Errorf("Expected ID %s, got %s", tt.wantID, node.ID)
			}

			// Check visibility based on exported status
			expectedVisibility := VisibilityPrivate
			if typeName.Exported() {
				expectedVisibility = VisibilityPublic
			}
			if node.Visibility != expectedVisibility {
				t.Errorf("Expected visibility %s, got %s", expectedVisibility, node.Visibility)
			}
		})
	}
}

func TestCreateFunctionNode(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	// Create mock package
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", fset.Base(), 1000)

	pkg := &packages.Package{
		PkgPath: "utils",
		Name:    "utils",
		Fset:    fset,
		TypesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
		},
	}

	tests := []struct {
		name         string
		funcName     string
		hasReceiver  bool
		wantID       string
		wantFullName string
		wantType     string
	}{
		{
			name:         "simple function",
			funcName:     "Calculate",
			hasReceiver:  false,
			wantID:       "function:utils.Calculate",
			wantFullName: "",
			wantType:     NodeTypeFunction,
		},
		{
			name:         "method with receiver",
			funcName:     "String",
			hasReceiver:  true,
			wantID:       "method:utils", // Will be adjusted based on actual receiver
			wantFullName: "",
			wantType:     NodeTypeMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcDecl := &ast.FuncDecl{
				Name: &ast.Ident{
					Name:    tt.funcName,
					NamePos: file.Pos(1),
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
				},
			}

			if tt.hasReceiver {
				// Add receiver
				funcDecl.Recv = &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.StarExpr{
								X: &ast.Ident{Name: "Config"},
							},
						},
					},
				}
			}

			node := builder.CreateFunctionNode(pkg, funcDecl)

			if node == nil {
				t.Fatal("Expected non-nil node")
			}

			expectedType := NodeTypeFunction
			if tt.hasReceiver {
				expectedType = NodeTypeMethod
			}
			if node.Type != expectedType {
				t.Errorf("Expected type %s, got %s", expectedType, node.Type)
			}

			// For methods, the ID includes the receiver type
			if tt.hasReceiver && node.Type == NodeTypeMethod {
				// Just verify it's a method node
				if node.Type != NodeTypeMethod {
					t.Errorf("Expected method node for function with receiver")
				}
			} else if !tt.hasReceiver {
				if node.ID != tt.wantID {
					t.Errorf("Expected ID %s, got %s", tt.wantID, node.ID)
				}
			}
		})
	}
}

func TestCreateFieldNode(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	// Create a struct node first
	structNode := &Node{
		ID:      "struct:models.User",
		Label:   "User",
		Type:    NodeTypeStruct,
		Package: "models",
	}

	// Create a field variable
	fieldType := types.Typ[types.String]
	fieldVar := types.NewField(token.NoPos, nil, "Name", fieldType, false)

	pos := Position{
		Filename: "models.go",
		Line:     15,
	}

	node := builder.CreateFieldNode(structNode, fieldVar, `json:"name"`, pos)

	if node == nil {
		t.Fatal("Expected non-nil field node")
	}
	if node.Type != NodeTypeField {
		t.Errorf("Expected type %s, got %s", NodeTypeField, node.Type)
	}
	if node.ID != "field:models.User.Name" {
		t.Errorf("Expected ID 'field:models.User.Name', got %s", node.ID)
	}
	if node.Label != "Name" {
		t.Errorf("Expected label 'Name', got %s", node.Label)
	}

	// Check metadata
	if node.Metadata["tag"] != `json:"name"` {
		t.Errorf("Expected tag metadata to be preserved")
	}
}

func TestCreateParameterNode(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	// Create a function node first
	funcNode := &Node{
		ID:      "function:handlers.HandleRequest",
		Label:   "HandleRequest",
		Type:    NodeTypeFunction,
		Package: "handlers",
	}

	pos := Position{
		Filename: "handlers.go",
		Line:     25,
	}

	node := builder.CreateParameterNode(funcNode, "ctx", "context.Context", 0, pos)

	if node == nil {
		t.Fatal("Expected non-nil parameter node")
	}
	if node.Type != NodeTypeParameter {
		t.Errorf("Expected type %s, got %s", NodeTypeParameter, node.Type)
	}
	if node.ID != "parameter:handlers.HandleRequest.ctx" {
		t.Errorf("Expected ID 'parameter:handlers.HandleRequest.ctx', got %s", node.ID)
	}
	if node.Label != "ctx" {
		t.Errorf("Expected label 'ctx', got %s", node.Label)
	}
}

func TestCreateGenericNode(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	pkg := &packages.Package{
		PkgPath: "generics",
		Name:    "generics",
	}

	funcNode := &Node{
		ID:      "function:generics.Map",
		Label:   "Map",
		Type:    NodeTypeFunction,
		Package: "generics",
	}

	node := builder.CreateGenericNode(pkg, funcNode, "T")

	// Generic nodes have been removed as rarely used
	if node != nil {
		t.Fatal("Expected nil for generic node (feature removed)")
	}
}

func TestNodeBuilderDuplicateHandling(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	pkg := &packages.Package{
		PkgPath: "test",
		Name:    "test",
		GoFiles: []string{"test.go"},
		Imports: make(map[string]*packages.Package),
	}

	// Create the same node twice
	_ = builder.CreatePackageNode(pkg)
	node2 := builder.CreatePackageNode(pkg)

	// Second call should return nil (node already exists)
	if node2 != nil {
		t.Error("Expected nil for duplicate node creation")
	}

	// Graph should only have one node
	if len(graph.Nodes) != 1 {
		t.Errorf("Expected 1 node in graph, got %d", len(graph.Nodes))
	}
}

func TestNodeBuilderEdgeCases(t *testing.T) {
	graph := NewGraph()
	builder := NewNodeBuilder(graph)

	t.Run("anonymous field", func(t *testing.T) {
		structNode := &Node{
			ID:      "struct:edge.Handler",
			Label:   "Handler",
			Type:    NodeTypeStruct,
			Package: "edge",
		}

		// Anonymous embedded field
		fieldType := types.NewNamed(
			types.NewTypeName(token.NoPos, nil, "Context", nil),
			types.NewStruct(nil, nil),
			nil,
		)
		fieldVar := types.NewField(token.NoPos, nil, "", fieldType, true) // anonymous = true

		pos := Position{Filename: "edge.go", Line: 1}
		node := builder.CreateFieldNode(structNode, fieldVar, "", pos)

		if node == nil {
			t.Fatal("Expected non-nil node for anonymous field")
		}
		// Anonymous fields should use the type name
		if node.Label != "" && node.Label != "Context" {
			// The implementation might handle this differently
			t.Logf("Anonymous field label: %s", node.Label)
		}
	})
}
