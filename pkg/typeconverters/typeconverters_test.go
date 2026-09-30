// Copyright (C) 2025 - 2026 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package typeconverters

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/snps-ai-sharedtypes/pkg/wiki"
)

func TestJSONToGo(t *testing.T) {
	tests := []struct {
		jsonType  string
		goType    string
		expectErr bool
	}{
		{"string", "string", false},
		{"string(binary)", "[]byte", false},
		{"number", "float64", false},
		{"integer", "int", false},
		{"boolean", "bool", false},
		{"array<string>", "[]string", false},
		{"array<number>", "[]float64", false},
		{"array<integer>", "[]int", false},
		{"array<boolean>", "[]bool", false},
		{"array<object>", "[]interface{}", false},
		{"array<any>", "[]interface{}", false},
		{"array<Any>", "[]interface{}", false},
		{"any", "interface{}", false},
		{"Any", "interface{}", false},
		{"object", "interface{}", false},
		{"dict[string][Any]", "map[string]interface{}", false},
		{"dict[string][any]", "map[string]interface{}", false},
		{"dict[string][object]", "map[string]interface{}", false},
		{"dict[string][string]", "map[string]string", false},
		{"dict[string][integer]", "map[string]int", false},
		{"dict[string][number]", "map[string]float64", false},
		{"unsupportedType", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.jsonType, func(t *testing.T) {
			got, err := JSONToGo(tt.jsonType)
			if (err != nil) != tt.expectErr {
				t.Errorf("JSONToGo(%s) error = %v, expectErr %v", tt.jsonType, err, tt.expectErr)
				return
			}
			if got != tt.goType {
				t.Errorf("JSONToGo(%s) = %v, want %v", tt.jsonType, got, tt.goType)
			}
		})
	}
}

func TestGoToJSON(t *testing.T) {
	tests := []struct {
		goType   string
		jsonType string
	}{
		{"string", "string"},
		{"[]byte", "string(binary)"},
		{"float32", "number"},
		{"float64", "number"},
		{"int", "integer"},
		{"int8", "integer"},
		{"int16", "integer"},
		{"int32", "integer"},
		{"int64", "integer"},
		{"uint", "integer"},
		{"uint8", "integer"},
		{"uint16", "integer"},
		{"uint32", "integer"},
		{"uint64", "integer"},
		{"bool", "boolean"},
		{"map[string]string", "dict[string][string]"},
		{"map[string]int", "dict[string][integer]"},
		{"map[string]float64", "dict[string][number]"},
		{"map[string]interface{}", "dict[string][object]"},
		{"interface{}", "object"},
		{"[]string", "array<string>"},
		{"[]float64", "array<number>"},
		{"[]int", "array<integer>"},
		{"unknown", "object"},
	}

	for _, test := range tests {
		result := GoToJSON(test.goType)
		if result != test.jsonType {
			t.Errorf("GoToJSON(%q) = %q; want %q", test.goType, result, test.jsonType)
		}
	}
}

func TestConvertStringToGivenType(t *testing.T) {
	tests := []struct {
		value       string
		goType      string
		expected    interface{}
		expectedErr error
	}{
		{"42", "int", 42, nil},
		{"true", "bool", true, nil},
		{"3.14", "float64", 3.14, nil},
		{`["a","b","c"]`, "[]string", []string{"a", "b", "c"}, nil},
		{"", "[]int", []int{}, nil},
		{"{\"key\":\"value\"}", "map[string]string", map[string]string{"key": "value"}, nil},
		{"", "map[string]float64", map[string]float64{}, nil},
		{"{}", "map[string]bool", map[string]bool{}, nil},
		{"[]", "[]DbJsonFilter", []sharedtypes.DbJsonFilter{}, nil},
		{"", "*chan string", (*chan string)(nil), nil},
		{`{"serverURL":"http://localhost:8080","transport":"http","authToken":"secret123","timeout":30}`, "MCPConfig", sharedtypes.MCPConfig{ServerURL: "http://localhost:8080", Transport: "http", AuthToken: "secret123", Timeout: 30}, nil},
		{`[{"serverURL":"http://localhost:8080","transport":"http","authToken":"secret123","timeout":30}]`, "[]MCPConfig", []sharedtypes.MCPConfig{{ServerURL: "http://localhost:8080", Transport: "http", AuthToken: "secret123", Timeout: 30}}, nil},
		{`[{"name":"test_tool","description":"A test tool","inputSchema":{"type":"object","properties":{"param1":{"type":"string"}}}}]`, "[]MCPTool", []sharedtypes.MCPTool{{Name: "test_tool", Description: "A test tool", InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"param1": map[string]interface{}{"type": "string"}}}}}, nil},
		{"", "[]MCPTool", []sharedtypes.MCPTool{}, nil},
		// Add more test cases as needed for each supported type
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.goType, test.value), func(t *testing.T) {
			output, _, err := ConvertStringToGivenType(test.value, test.goType)
			if err != nil && test.expectedErr == nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if err == nil && test.expectedErr != nil {
				t.Errorf("Expected error: %v, got nil", test.expectedErr)
			}
			if fmt.Sprintf("%v", output) != fmt.Sprintf("%v", test.expected) {
				t.Errorf("Expected output: %v, got: %v", test.expected, output)
			}
		})
	}
}

func TestConvertStringToGivenType_UnsupportedType(t *testing.T) {
	output, exists, err := ConvertStringToGivenType("value", "UnsupportedType")
	if exists {
		t.Errorf("Expected exists to be false for unsupported type")
	}
	if output != nil {
		t.Errorf("Expected output to be nil for unsupported type, got: %v", output)
	}
	if err != nil {
		t.Errorf("Expected no error for unsupported type, got: %v", err)
	}
}

func TestConvertGivenTypeToString(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		goType    string
		expected  string
		expectErr bool
	}{
		// Primitive types
		{"string", "hello", "string", "hello", false},
		{"int", 42, "int", "42", false},
		{"int8", int8(8), "int8", "8", false},
		{"int16", int16(16), "int16", "16", false},
		{"int32", int32(32), "int32", "32", false},
		{"int64", int64(64), "int64", "64", false},
		{"uint", uint(42), "uint", "42", false},
		{"uint8", uint8(8), "uint8", "8", false},
		{"uint16", uint16(16), "uint16", "16", false},
		{"uint32", uint32(32), "uint32", "32", false},
		{"uint64", uint64(64), "uint64", "64", false},
		{"float32", float32(3.14), "float32", "3.14", false},
		{"float64", float64(3.14159), "float64", "3.14159", false},
		{"bool_true", true, "bool", "true", false},
		{"bool_false", false, "bool", "false", false},

		// Interface types
		{"interface_string", "test", "interface{}", "test", false},
		{"interface_map", map[string]interface{}{"key": "value"}, "interface{}", `{"key":"value"}`, false},
		{"any_string", "test", "any", "test", false},

		// Slice types
		{"[]string", []string{"a", "b", "c"}, "[]string", `["a","b","c"]`, false},
		{"[]int", []int{1, 2, 3}, "[]int", `[1,2,3]`, false},
		{"[]float64", []float64{1.1, 2.2}, "[]float64", `[1.1,2.2]`, false},
		{"[]bool", []bool{true, false}, "[]bool", `[true,false]`, false},
		{"[]interface{}", []interface{}{"a", 1}, "[]interface{}", `["a",1]`, false},

		// Map types
		{"map[string]string", map[string]string{"key": "value"}, "map[string]string", `{"key":"value"}`, false},
		{"map[string]int", map[string]int{"num": 42}, "map[string]int", `{"num":42}`, false},
		{"map[string]float64", map[string]float64{"pi": 3.14}, "map[string]float64", `{"pi":3.14}`, false},
		{"map[string]bool", map[string]bool{"flag": true}, "map[string]bool", `{"flag":true}`, false},
		{"map[string]interface{}", map[string]interface{}{"key": "value"}, "map[string]interface{}", `{"key":"value"}`, false},

		// Channel types (always empty string)
		{"*chan string", (*chan string)(nil), "*chan string", "", false},
		{"*chan interface{}", (*chan interface{})(nil), "*chan interface{}", "", false},

		// Custom sharedtypes
		{"MCPConfig", sharedtypes.MCPConfig{ServerURL: "http://localhost", Transport: "http"}, "MCPConfig", `{"serverURL":"http://localhost","transport":"http","authToken":"","timeout":0}`, false},
		{"[]MCPConfig", []sharedtypes.MCPConfig{{ServerURL: "http://localhost"}}, "[]MCPConfig", `[{"serverURL":"http://localhost","transport":"","authToken":"","timeout":0}]`, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, exists, err := ConvertGivenTypeToString(test.value, test.goType)
			if !exists {
				t.Errorf("Expected exists to be true for type %s", test.goType)
			}
			if test.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", test.name)
			}
			if !test.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", test.name, err)
			}
			if output != test.expected {
				t.Errorf("Expected output: %s, got: %s", test.expected, output)
			}
		})
	}
}

func TestConvertGivenTypeToString_UnsupportedType(t *testing.T) {
	output, exists, err := ConvertGivenTypeToString("value", "UnsupportedType")
	if exists {
		t.Errorf("Expected exists to be false for unsupported type")
	}
	if output != "" {
		t.Errorf("Expected output to be empty for unsupported type, got: %s", output)
	}
	if err != nil {
		t.Errorf("Expected no error for unsupported type, got: %v", err)
	}
}

func TestGetSupportedTypes(t *testing.T) {
	types := GetSupportedTypes()

	// Verify we get a non-empty list
	if len(types) == 0 {
		t.Error("Expected GetSupportedTypes to return a non-empty list")
	}

	// Verify some expected types are present
	expectedTypes := []string{
		"string",
		"int",
		"float64",
		"bool",
		"[]string",
		"map[string]string",
		"interface{}",
		"any",
		"MCPConfig",
		"[]MCPTool",
	}

	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	for _, expected := range expectedTypes {
		if !typeSet[expected] {
			t.Errorf("Expected type %s to be in supported types", expected)
		}
	}
}

func TestGetSupportedTypes_AllTypesHaveConverters(t *testing.T) {
	// Verify that all types returned by GetSupportedTypes actually work with the converters
	types := GetSupportedTypes()

	for _, goType := range types {
		t.Run(goType, func(t *testing.T) {
			// Test that ConvertStringToGivenType recognizes the type
			_, exists, _ := ConvertStringToGivenType("", goType)
			if !exists {
				t.Errorf("Type %s returned by GetSupportedTypes but not recognized by ConvertStringToGivenType", goType)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that converting to string and back gives the same value
	tests := []struct {
		name   string
		value  interface{}
		goType string
	}{
		{"string", "hello world", "string"},
		{"int", 42, "int"},
		{"float64", 3.14, "float64"},
		{"bool", true, "bool"},
		{"[]string", []string{"a", "b", "c"}, "[]string"},
		{"[]int", []int{1, 2, 3}, "[]int"},
		{"map[string]string", map[string]string{"key": "value"}, "map[string]string"},
		{"Page", wiki.Page{Path: "meshing/sizing.md", Content: "use 1 mm"}, "Page"},
		{"[]Scope", []wiki.Scope{{Path: "eng", Level: "department"}}, "[]Scope"},
		{"[]PageEntry", []wiki.PageEntry{{Path: "meshing/sizing.md", Standard: false}}, "[]PageEntry"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert to string
			str, exists, err := ConvertGivenTypeToString(test.value, test.goType)
			if !exists || err != nil {
				t.Fatalf("Failed to convert %v to string: exists=%v, err=%v", test.value, exists, err)
			}

			// Convert back from string
			result, exists, err := ConvertStringToGivenType(str, test.goType)
			if !exists || err != nil {
				t.Fatalf("Failed to convert string back to %s: exists=%v, err=%v", test.goType, exists, err)
			}

			// Compare
			if !reflect.DeepEqual(test.value, result) {
				t.Errorf("Round trip failed for %s: original=%v, result=%v", test.goType, test.value, result)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	type TestData struct {
		Name string
		Age  int
	}

	src := TestData{Name: "John", Age: 30}
	dst := new(TestData)

	err := DeepCopy(src, dst)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(src, *dst) {
		t.Errorf("deep copy failed, got: %v, want: %v", *dst, src)
	}
}
