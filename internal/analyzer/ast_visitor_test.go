package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// mockImporter implements types.Importer for testing
type mockImporter struct{}

func (m *mockImporter) Import(path string) (*types.Package, error) {
	// Return mock packages for common imports
	switch path {
	case "fmt":
		pkg := types.NewPackage("fmt", "fmt")
		// Add Println to the package
		scope := pkg.Scope()
		// Println(a ...any) (n int, err error)
		anyType := types.NewInterfaceType(nil, nil)
		printlnType := types.NewSignatureType(nil, nil, nil,
			types.NewTuple(types.NewParam(token.NoPos, pkg, "a", types.NewSlice(anyType))),
			types.NewTuple(
				types.NewVar(token.NoPos, pkg, "n", types.Typ[types.Int]),
				types.NewVar(token.NoPos, pkg, "err", types.Universe.Lookup("error").Type())),
			true)
		scope.Insert(types.NewFunc(token.NoPos, pkg, "Println", printlnType))
		return pkg, nil
	case "context":
		pkg := types.NewPackage("context", "context")
		// Add Context interface
		contextType := types.NewNamed(
			types.NewTypeName(token.NoPos, pkg, "Context", nil),
			types.NewInterfaceType(nil, nil),
			nil,
		)
		pkg.Scope().Insert(contextType.Obj())
		return pkg, nil
	default:
		return nil, nil
	}
}

func TestASTVisitor(t *testing.T) {
	// Create test Go source
	src := `
package main

type Config struct {
	Name string
	Port int
}

func main() {
	// test function
}

func (c *Config) String() string {
	return c.Name
}
`

	// Parse the source
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Create a mock package with proper type checking
	conf := types.Config{
		Importer: &mockImporter{},
		Error: func(err error) {
			// Ignore import errors during tests
		},
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	// Type check to populate the info
	typePkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("Type checking failed: %v", err)
	}

	pkg := &packages.Package{
		PkgPath:   "test",
		Name:      "main",
		Syntax:    []*ast.File{f},
		Fset:      fset,
		Types:     typePkg,
		TypesInfo: info,
	}

	// Create graph and node builder
	graph := NewGraph()
	nodeBuilder := NewNodeBuilder(graph)

	// Create relationship analyzer (even if not used in this test)
	relationshipAnalyzer := NewRelationshipAnalyzer(graph, nodeBuilder)

	// Create and run visitor
	visitor := NewASTVisitor(pkg, nodeBuilder, relationshipAnalyzer)
	visitor.BuildNodes()

	// Verify nodes were created
	if len(graph.Nodes) == 0 {
		t.Error("Expected nodes to be created")
	}

	// Check for specific nodes
	nodeTypes := make(map[string]int)
	for _, node := range graph.Nodes {
		nodeTypes[node.Type]++
	}

	// We should have at least:
	// - 1 package node
	// - 1 struct (Config)
	// - 2 functions (main and String method)
	// - 2 fields (Name and Port)
	if nodeTypes[NodeTypePackage] == 0 {
		t.Error("Expected at least one package node")
	}
	if nodeTypes[NodeTypeStruct] == 0 {
		t.Error("Expected at least one struct node")
	}
	if nodeTypes[NodeTypeFunction] == 0 && nodeTypes[NodeTypeMethod] == 0 {
		t.Error("Expected at least one function or method node")
	}
	if nodeTypes[NodeTypeField] == 0 {
		t.Error("Expected at least one field node")
	}
}

func TestProcessTypeDeclarations(t *testing.T) {
	src := `
package test

type StringAlias string

type InterfaceExample interface {
	Method() error
}

type StructExample struct {
	Field1 string
	Field2 int
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Run type checker to populate Types information
	typesInfo := &types.Info{
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Scopes: make(map[ast.Node]*types.Scope),
	}

	conf := &types.Config{
		Importer: nil, // Use default importer
	}

	typePkg, err := conf.Check("test", fset, []*ast.File{f}, typesInfo)
	if err != nil {
		t.Fatalf("Type checking failed: %v", err)
	}

	pkg := &packages.Package{
		PkgPath:   "test",
		Name:      "test",
		Syntax:    []*ast.File{f},
		Fset:      fset,
		TypesInfo: typesInfo,
		Types:     typePkg,
	}

	graph := NewGraph()
	nodeBuilder := NewNodeBuilder(graph)
	relationshipAnalyzer := NewRelationshipAnalyzer(graph, nodeBuilder)

	visitor := NewASTVisitor(pkg, nodeBuilder, relationshipAnalyzer)
	visitor.BuildNodes()

	// Debug: Print all nodes created
	t.Logf("Total nodes created: %d", len(graph.Nodes))
	for _, node := range graph.Nodes {
		t.Logf("Node: Type=%s, Label=%s, ID=%s", node.Type, node.Label, node.ID)
	}

	// Check that struct and interface declarations were processed
	// Note: Type aliases like "type StringAlias string" are not created as nodes
	// since they're not structs or interfaces
	var foundInterface, foundStruct bool
	for _, node := range graph.Nodes {
		switch node.Label {
		case "InterfaceExample":
			foundInterface = true
			if node.Type != NodeTypeInterface {
				t.Errorf("Expected InterfaceExample to be NodeTypeInterface, got %s", node.Type)
			}
		case "StructExample":
			foundStruct = true
			if node.Type != NodeTypeStruct {
				t.Errorf("Expected StructExample to be NodeTypeStruct, got %s", node.Type)
			}
		}
	}

	if !foundInterface {
		t.Error("InterfaceExample type not found")
	}
	if !foundStruct {
		t.Error("StructExample type not found")
	}

	// Verify that fields were created for the struct
	fieldCount := 0
	for _, node := range graph.Nodes {
		if node.Type == NodeTypeField {
			fieldCount++
		}
	}
	if fieldCount != 2 {
		t.Errorf("Expected 2 fields, got %d", fieldCount)
	}
}

func TestProcessFunctionDeclarations(t *testing.T) {
	src := `
package test

func PublicFunction() error {
	return nil
}

func privateFunction() {
}

func (r *Receiver) Method() {
}

func init() {
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Type check to populate the info
	conf := types.Config{
		Importer: &mockImporter{},
		Error: func(err error) {
			// Ignore import errors during tests
		},
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	typePkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil && !strings.Contains(err.Error(), "Receiver") {
		// Ignore error about undefined Receiver type
		t.Fatalf("Type checking failed: %v", err)
	}

	pkg := &packages.Package{
		PkgPath:   "test",
		Name:      "test",
		Syntax:    []*ast.File{f},
		Fset:      fset,
		Types:     typePkg,
		TypesInfo: info,
	}

	graph := NewGraph()
	nodeBuilder := NewNodeBuilder(graph)
	relationshipAnalyzer := NewRelationshipAnalyzer(graph, nodeBuilder)

	visitor := NewASTVisitor(pkg, nodeBuilder, relationshipAnalyzer)
	visitor.BuildNodes()

	// Check functions
	funcCount := 0
	methodCount := 0
	for _, node := range graph.Nodes {
		switch node.Type {
		case NodeTypeFunction:
			funcCount++
			if node.Label == "PublicFunction" && node.Visibility != VisibilityPublic {
				t.Error("PublicFunction should have public visibility")
			}
			if node.Label == "privateFunction" && node.Visibility != VisibilityPrivate {
				t.Error("privateFunction should have private visibility")
			}
		case NodeTypeMethod:
			methodCount++
		}
	}

	if funcCount < 2 { // At least PublicFunction and privateFunction
		t.Errorf("Expected at least 2 functions, got %d", funcCount)
	}

	if funcCount < 2 { // At least PublicFunction and privateFunction
		t.Errorf("Expected at least 2 functions, got %d", funcCount)
	}
	if methodCount < 1 { // At least Method
		t.Errorf("Expected at least 1 method, got %d", methodCount)
	}
}

func TestProcessStructFields(t *testing.T) {
	src := `
package test

type MyStruct struct {
	PublicField  string
	privateField int
	EmbeddedType
	*EmbeddedPointer
}

type EmbeddedType struct{}
type EmbeddedPointer struct{}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Type check to populate the info
	conf := types.Config{
		Importer: &mockImporter{},
		Error: func(err error) {
			// Ignore import errors during tests
		},
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	typePkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("Type checking failed: %v", err)
	}

	pkg := &packages.Package{
		PkgPath:   "test",
		Name:      "test",
		Syntax:    []*ast.File{f},
		Fset:      fset,
		Types:     typePkg,
		TypesInfo: info,
	}

	graph := NewGraph()
	nodeBuilder := NewNodeBuilder(graph)
	relationshipAnalyzer := NewRelationshipAnalyzer(graph, nodeBuilder)

	visitor := NewASTVisitor(pkg, nodeBuilder, relationshipAnalyzer)
	visitor.BuildNodes()

	// Check fields
	fieldCount := 0
	var publicField, privateField, embeddedField *EnhancedNode

	for i := range graph.Nodes {
		node := &graph.Nodes[i]
		if node.Type == NodeTypeField {
			fieldCount++
			switch node.Label {
			case "PublicField":
				publicField = node
			case "privateField":
				privateField = node
			case "EmbeddedType", "EmbeddedPointer":
				embeddedField = node
			}
		}
	}

	if fieldCount < 3 { // At least PublicField, privateField, and one embedded
		t.Errorf("Expected at least 3 fields, got %d", fieldCount)
	}

	if publicField != nil && publicField.Visibility != VisibilityPublic {
		t.Error("PublicField should have public visibility")
	}
	if privateField != nil && privateField.Visibility != VisibilityPrivate {
		t.Error("privateField should have private visibility")
	}
	if embeddedField == nil {
		t.Error("Expected embedded field to be processed")
	}
}

func TestProcessFunctionParameters(t *testing.T) {
	src := `
package test

import "context"

type Receiver struct{}

func FunctionWithParams(name string, age int, options ...string) error {
	return nil
}

func (r *Receiver) MethodWithParams(ctx context.Context, data []byte) {
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Type check to populate the info
	conf := types.Config{
		Importer: &mockImporter{},
		Error: func(err error) {
			// Ignore import errors during tests
		},
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	typePkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil && !strings.Contains(err.Error(), "context") {
		// Ignore import errors
		t.Fatalf("Type checking failed: %v", err)
	}

	pkg := &packages.Package{
		PkgPath:   "test",
		Name:      "test",
		Syntax:    []*ast.File{f},
		Fset:      fset,
		Types:     typePkg,
		TypesInfo: info,
	}

	graph := NewGraph()
	nodeBuilder := NewNodeBuilder(graph)
	relationshipAnalyzer := NewRelationshipAnalyzer(graph, nodeBuilder)

	visitor := NewASTVisitor(pkg, nodeBuilder, relationshipAnalyzer)
	visitor.BuildNodes()

	// Check parameters
	paramCount := 0
	paramNames := make(map[string]bool)

	for _, node := range graph.Nodes {
		if node.Type == NodeTypeParameter {
			paramCount++
			paramNames[node.Label] = true
		}
	}

	// Should have parameters: name, age, options, ctx, data
	expectedParams := []string{"name", "age", "options", "ctx", "data"}
	for _, param := range expectedParams {
		if !paramNames[param] {
			t.Errorf("Expected parameter %s not found", param)
		}
	}

	if paramCount < len(expectedParams) {
		t.Errorf("Expected at least %d parameters, got %d", len(expectedParams), paramCount)
	}
}
