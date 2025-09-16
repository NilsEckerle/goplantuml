package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestParseTypeParameters(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []TypeParameter
	}{
		{
			name: "Single type parameter with any constraint",
			code: `package test

type Model[T any] struct {}`,
			expected: []TypeParameter{
				{Name: "T", Constraints: "any"},
			},
		},
		{
			name: "Multiple type parameters with different constraints",
			code: `package test

type Container[T any, U comparable] struct {}`,
			expected: []TypeParameter{
				{Name: "T", Constraints: "any"},
				{Name: "U", Constraints: "comparable"},
			},
		},
		{
			name: "Type parameter with interface constraint",
			code: `package test

type Processor[T interface{ Process() }] struct {}`,
			expected: []TypeParameter{
				{Name: "T", Constraints: "Process <font color=blue>func</font>()"},
			},
		},
		{
			name: "Type parameter with union constraint",
			code: `package test

type Number[T int|float64] struct {}`,
			expected: []TypeParameter{
				{Name: "T", Constraints: "int|float64"},
			},
		},
		{
			name: "Type parameter with tilde constraint",
			code: `package test

type Numeric[T ~int] struct {}`,
			expected: []TypeParameter{
				{Name: "T", Constraints: "~int"},
			},
		},
		{
			name: "Complex constraint with multiple types",
			code: `package test

type Complex[T ~int|~float64, U interface{ String() string }] struct {}`,
			expected: []TypeParameter{
				{Name: "T", Constraints: "~int|~float64"},
				{Name: "U", Constraints: "String <font color=blue>func</font>() string"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			var typeSpec *ast.TypeSpec
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							typeSpec = ts
							break
						}
					}
				}
			}

			if typeSpec == nil {
				t.Fatal("No TypeSpec found")
			}

			result := parseTypeParameters(typeSpec.TypeParams, map[string]string{})
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d type parameters, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].Name != expected.Name {
					t.Errorf("Type parameter %d: expected name %s, got %s", i, expected.Name, result[i].Name)
				}
				if result[i].Constraints != expected.Constraints {
					t.Errorf("Type parameter %d: expected constraints %s, got %s", i, expected.Constraints, result[i].Constraints)
				}
			}
		})
	}
}

func TestStringifyConstraint(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "any constraint",
			code: `package test

type T any`,
			expected: "any",
		},
		{
			name: "comparable constraint",
			code: `package test

type T comparable`,
			expected: "comparable",
		},
		{
			name: "union constraint",
			code: `package test

type T interface{ int|float64 }`,
			expected: "int|float64",
		},
		{
			name: "tilde constraint",
			code: `package test

type T interface{ ~int }`,
			expected: "~int",
		},
		{
			name: "interface constraint",
			code: `package test

type T interface{ String() string }`,
			expected: "String <font color=blue>func</font>() string",
		},
		{
			name: "complex union with tilde",
			code: `package test

type T interface{ ~int|~float64 }`,
			expected: "~int|~float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			var typeSpec *ast.TypeSpec
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							typeSpec = ts
							break
						}
					}
				}
			}

			if typeSpec == nil {
				t.Fatal("No TypeSpec found")
			}

			result := stringifyConstraint(typeSpec.Type, map[string]string{})
			if result != tt.expected {
				t.Errorf("Expected constraint %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetFieldTypeWithGenericInstantiations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "Generic instantiation with single parameter",
			code: `package test

var x Model[int]`,
			expected: "Model",
		},
		{
			name: "Generic instantiation with multiple parameters",
			code: `package test

var x Container[string, int]`,
			expected: "Container",
		},
		{
			name: "Generic instantiation with selector expression",
			code: `package test

var x pkg.Generic[string]`,
			expected: "pkg.Generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			var valueSpec *ast.ValueSpec
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if vs, ok := spec.(*ast.ValueSpec); ok {
							valueSpec = vs
							break
						}
					}
				}
			}

			if valueSpec == nil {
				t.Fatal("No ValueSpec found")
			}

			result, _ := getFieldType(valueSpec.Type, map[string]string{})
			// Remove package constant prefix for comparison
			result = strings.TrimPrefix(result, "{packageName}")
			if result != tt.expected {
				t.Errorf("Expected type %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStructIsGenericParamType(t *testing.T) {
	st := &Struct{
		TypeParameters: []TypeParameter{
			{Name: "T", Constraints: "any"},
			{Name: "U", Constraints: "comparable"},
		},
	}

	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{
			name:     "Type parameter T",
			param:    "{packageName}T",
			expected: true,
		},
		{
			name:     "Type parameter U",
			param:    "{packageName}U",
			expected: true,
		},
		{
			name:     "Non-type parameter",
			param:    "{packageName}String",
			expected: false,
		},
		{
			name:     "Primitive type",
			param:    "int",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := st.isGenericParamType(tt.param)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRenderStructureWithGenerics(t *testing.T) {
	// Create a generic struct
	st := &Struct{
		PackageName: "test",
		Type:        "class",
		Fields: []*Field{
			{Name: "Data", Type: "[]T"},
		},
		TypeParameters: []TypeParameter{
			{Name: "T", Constraints: "any"},
		},
	}

	str := &LineStringBuilder{}
	composition := &LineStringBuilder{}
	extends := &LineStringBuilder{}
	aggregations := &LineStringBuilder{}
	params := &LineStringBuilder{}
	emittedTypeParamClass := map[string]struct{}{}
	emittedParamLink := map[string]struct{}{}

	parser := getEmptyParser("test")
	parser.renderStructureWithGenerics(st, "test", "Model", str, composition, extends, aggregations, params, emittedTypeParamClass, emittedParamLink)

	result := str.String()

	// Check that the class is rendered with alias and type parameters in stereotype
	expectedClass := `class "Model" as Model_generic_T <<[T]>> {`
	if !strings.Contains(result, expectedClass) {
		t.Errorf("Expected class declaration %s not found in: %s", expectedClass, result)
	}

	// Check that type parameter class is rendered
	if !strings.Contains(result, `class "T" <<type parameter>> {`) {
		t.Errorf("Type parameter class not found in: %s", result)
	}

	if !strings.Contains(result, `constraints: any`) {
		t.Errorf("Type parameter constraints not found in: %s", result)
	}

	// Check that param relationship is added
	paramResult := params.String()
	expectedLink := `"T" <-- "param" "Model_generic_T"`
	if !strings.Contains(paramResult, expectedLink) {
		t.Errorf("Expected param link %s not found in: %s", expectedLink, paramResult)
	}
}

func TestRenderStructureWithGenericsDeduplication(t *testing.T) {
	// Create two generic structs with the same type parameter
	st1 := &Struct{
		PackageName: "test",
		Type:        "class",
		TypeParameters: []TypeParameter{
			{Name: "T", Constraints: "any"},
		},
	}

	st2 := &Struct{
		PackageName: "test",
		Type:        "class",
		TypeParameters: []TypeParameter{
			{Name: "T", Constraints: "any"},
		},
	}

	str := &LineStringBuilder{}
	composition := &LineStringBuilder{}
	extends := &LineStringBuilder{}
	aggregations := &LineStringBuilder{}
	params := &LineStringBuilder{}
	emittedTypeParamClass := map[string]struct{}{}
	emittedParamLink := map[string]struct{}{}

	parser := getEmptyParser("test")

	// Render first struct
	parser.renderStructureWithGenerics(st1, "test", "Model1", str, composition, extends, aggregations, params, emittedTypeParamClass, emittedParamLink)

	// Render second struct
	parser.renderStructureWithGenerics(st2, "test", "Model2", str, composition, extends, aggregations, params, emittedTypeParamClass, emittedParamLink)

	result := str.String()

	// Count occurrences of type parameter class definition
	typeParamClassCount := strings.Count(result, `class "T" <<type parameter>> {`)
	if typeParamClassCount != 1 {
		t.Errorf("Expected type parameter class to be defined once, found %d times", typeParamClassCount)
	}

	// Count occurrences of param links
	paramLinkCount := strings.Count(params.String(), `"T" <-- "param"`)
	if paramLinkCount != 2 {
		t.Errorf("Expected 2 param links, found %d", paramLinkCount)
	}
}

func TestGenericFieldAggregationExclusion(t *testing.T) {
	// Create a generic struct with a field using the type parameter
	st := &Struct{
		PackageName: "test",
		Type:        "class",
		TypeParameters: []TypeParameter{
			{Name: "T", Constraints: "any"},
		},
		Aggregations: map[string]struct{}{},
	}

	// Create a field that uses the type parameter
	field := &ast.Field{
		Names: []*ast.Ident{{Name: "Data"}},
		Type:  &ast.Ident{Name: "T"},
	}

	// Add the field
	st.AddField(field, map[string]string{})

	// Check that the type parameter is not added to aggregations
	if len(st.Aggregations) != 0 {
		t.Errorf("Expected no aggregations, got %v", st.Aggregations)
	}
}

func TestGenericMethodReceiver(t *testing.T) {
	// Test that generic method receivers are handled correctly
	code := `
package test

type Model[T any] struct {
    Data T
}

func (m *Model[T]) Process() T {
    return m.Data
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	parser := getEmptyParser("test")

	// Process the file declarations
	for _, decl := range file.Decls {
		parser.parseFileDeclarations(decl)
	}

	// Check that the struct was created with type parameters
	structs := parser.structure["test"]
	if structs == nil {
		t.Fatal("No structs found in test package")
	}

	model, exists := structs["Model"]
	if !exists {
		t.Fatal("Model struct not found")
	}

	if len(model.TypeParameters) != 1 {
		t.Errorf("Expected 1 type parameter, got %d", len(model.TypeParameters))
	}

	if model.TypeParameters[0].Name != "T" {
		t.Errorf("Expected type parameter name T, got %s", model.TypeParameters[0].Name)
	}

	if model.TypeParameters[0].Constraints != "any" {
		t.Errorf("Expected type parameter constraint 'any', got %s", model.TypeParameters[0].Constraints)
	}

	// Check that the method was added
	if len(model.Functions) != 1 {
		t.Errorf("Expected 1 method, got %d", len(model.Functions))
	}

	if model.Functions[0].Name != "Process" {
		t.Errorf("Expected method name 'Process', got %s", model.Functions[0].Name)
	}
}
