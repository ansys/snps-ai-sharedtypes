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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ansys/snps-ai-sharedtypes/pkg/graphdb"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/snps-ai-sharedtypes/pkg/wiki"
	"github.com/coder/websocket"
)

// TypeConverter holds functions to convert to and from a given Go type
type TypeConverter struct {
	FromString func(value string) (interface{}, error)
	ToString   func(value interface{}) (string, error)
}

// typeRegistry maps Go type names to their converters
var typeRegistry map[string]TypeConverter

// init initializes the type registry with supported types and their converters
// it is called automatically when the package is imported
// when adding new sharedtypes, add them here for conversion support
func init() {
	typeRegistry = map[string]TypeConverter{
		// Primitive types
		"string": {
			FromString: func(value string) (interface{}, error) { return value, nil },
			ToString:   func(value interface{}) (string, error) { return value.(string), nil },
		},
		"float32": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				f, err := strconv.ParseFloat(value, 32)
				return float32(f), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32), nil
			},
		},
		"float64": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				return strconv.ParseFloat(value, 64)
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatFloat(value.(float64), 'f', -1, 64), nil
			},
		},
		"int": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				return strconv.Atoi(value)
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.Itoa(value.(int)), nil
			},
		},
		"int8": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseInt(value, 10, 8)
				return int8(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatInt(int64(value.(int8)), 10), nil
			},
		},
		"int16": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseInt(value, 10, 16)
				return int16(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatInt(int64(value.(int16)), 10), nil
			},
		},
		"int32": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseInt(value, 10, 32)
				return int32(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatInt(int64(value.(int32)), 10), nil
			},
		},
		"int64": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				return strconv.ParseInt(value, 10, 64)
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatInt(value.(int64), 10), nil
			},
		},
		"uint": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseUint(value, 10, 64)
				return uint(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatUint(uint64(value.(uint)), 10), nil
			},
		},
		"uint8": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseUint(value, 10, 8)
				return uint8(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatUint(uint64(value.(uint8)), 10), nil
			},
		},
		"uint16": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseUint(value, 10, 16)
				return uint16(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatUint(uint64(value.(uint16)), 10), nil
			},
		},
		"uint32": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				v, err := strconv.ParseUint(value, 10, 32)
				return uint32(v), err
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatUint(uint64(value.(uint32)), 10), nil
			},
		},
		"uint64": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "0"
				}
				return strconv.ParseUint(value, 10, 64)
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatUint(value.(uint64), 10), nil
			},
		},
		"bool": {
			FromString: func(value string) (interface{}, error) {
				if value == "" {
					value = "false"
				}
				return strconv.ParseBool(value)
			},
			ToString: func(value interface{}) (string, error) {
				return strconv.FormatBool(value.(bool)), nil
			},
		},

		// Interface types
		"interface{}": interfaceConverter(),
		"any":         interfaceConverter(),
		"json":        interfaceConverter(),

		// Slice types - JSON based
		"[]interface{}": jsonSliceConverter[[]interface{}](),
		"[]string":      jsonSliceConverter[[]string](),
		"[][]string":    jsonSliceConverter[[][]string](),
		"[]float32":     jsonSliceConverter[[]float32](),
		"[]float64":     jsonSliceConverter[[]float64](),
		"[]int":         jsonSliceConverter[[]int](),
		"[]bool":        jsonSliceConverter[[]bool](),
		"[]byte":        jsonSliceConverter[[]byte](),
		"[][]float32":   jsonSliceConverter[[][]float32](),

		// Channel types (special handling - always nil/empty)
		"*chan string":      chanConverter[chan string](),
		"*chan interface{}": chanConverter[chan interface{}](),
		"*Conn":             chanConverter[websocket.Conn](),

		// Map types - JSON based
		"map[string]string":                   jsonMapConverter[map[string]string](),
		"map[string]float64":                  jsonMapConverter[map[string]float64](),
		"map[string]int":                      jsonMapConverter[map[string]int](),
		"map[string]bool":                     jsonMapConverter[map[string]bool](),
		"map[string][]string":                 jsonMapConverter[map[string][]string](),
		"map[string]map[string]string":        jsonMapConverter[map[string]map[string]string](),
		"map[string]interface{}":              jsonMapConverter[map[string]interface{}](),
		"map[string]any":                      jsonMapConverter[map[string]interface{}](),
		"map[uint]float32":                    jsonMapConverter[map[uint]float32](),
		"map[string]AedtElementContextsTuple": jsonMapConverter[map[string]sharedtypes.AedtElementContextsTuple](),

		// Slice of maps - JSON based
		"[]map[string]string":      jsonSliceConverter[[]map[string]string](),
		"[]map[uint]float32":       jsonSliceConverter[[]map[uint]float32](),
		"[]map[string]interface{}": jsonSliceConverter[[]map[string]interface{}](),
		"[]map[string]any":         jsonSliceConverter[[]map[string]interface{}](),

		// Custom types - graphdb
		"ParameterMap": jsonMapConverter[graphdb.ParameterMap](),

		// Custom types - wiki (structs)
		"Page":                 jsonMapConverter[wiki.Page](),
		"Scope":                jsonMapConverter[wiki.Scope](),
		"DiffReport":           jsonMapConverter[wiki.DiffReport](),
		"ResolveResponse":      jsonMapConverter[wiki.ResolveResponse](),
		"ScopeDeletedResponse": jsonMapConverter[wiki.ScopeDeletedResponse](),

		// Custom types - wiki (slices)
		"[]Scope":          jsonSliceConverter[[]wiki.Scope](),
		"[]PageEntry":      jsonSliceConverter[[]wiki.PageEntry](),
		"[]Version":        jsonSliceConverter[[]wiki.Version](),
		"[]Conflict":       jsonSliceConverter[[]wiki.Conflict](),
		"[]ChangeSetEntry": jsonSliceConverter[[]wiki.ChangeSetEntry](),

		// Custom types - sharedtypes (structs)
		"DbArrayFilter":            jsonMapConverter[sharedtypes.DbArrayFilter](),
		"DbFilters":                jsonMapConverter[sharedtypes.DbFilters](),
		"Feedback":                 jsonMapConverter[sharedtypes.Feedback](),
		"ModelOptions":             jsonMapConverter[sharedtypes.ModelOptions](),
		"MCPConfig":                jsonMapConverter[sharedtypes.MCPConfig](),
		"MCPTool":                  jsonMapConverter[sharedtypes.MCPTool](),
		"ToolCall":                 jsonMapConverter[sharedtypes.ToolCall](),
		"ToolResult":               jsonMapConverter[sharedtypes.ToolResult](),
		"SlashCommand":             jsonMapConverter[sharedtypes.SlashCommand](),
		"DiscoverySimulationInput": jsonMapConverter[sharedtypes.DiscoverySimulationInput](),
		"DiscoveryDimensions":      jsonMapConverter[sharedtypes.DiscoveryDimensions](),

		// Custom types - sharedtypes (slices)
		"[]DbJsonFilter":                   jsonSliceConverter[[]sharedtypes.DbJsonFilter](),
		"[]DbResponse":                     jsonSliceConverter[[]sharedtypes.DbResponse](),
		"[]HistoricMessage":                jsonSliceConverter[[]sharedtypes.HistoricMessage](),
		"[]AnsysGPTDefaultFields":          jsonSliceConverter[[]sharedtypes.AnsysGPTDefaultFields](),
		"[]ACSSearchResponse":              jsonSliceConverter[[]sharedtypes.ACSSearchResponse](),
		"[]AnsysGPTCitation":               jsonSliceConverter[[]sharedtypes.AnsysGPTCitation](),
		"[]AnsysGPTRetrieverModuleChunk":   jsonSliceConverter[[]sharedtypes.AnsysGPTRetrieverModuleChunk](),
		"[]DbData":                         jsonSliceConverter[[]sharedtypes.DbData](),
		"[]CodeGenerationElement":          jsonSliceConverter[[]sharedtypes.CodeGenerationElement](),
		"[]CodeGenerationExample":          jsonSliceConverter[[]sharedtypes.CodeGenerationExample](),
		"[]CodeGenerationUserGuideSection": jsonSliceConverter[[]sharedtypes.CodeGenerationUserGuideSection](),
		"[]MaterialLlmCriterion":           jsonSliceConverter[[]sharedtypes.MaterialLlmCriterion](),
		"[]MaterialCriterionWithGuid":      jsonSliceConverter[[]sharedtypes.MaterialCriterionWithGuid](),
		"[]MaterialAttribute":              jsonSliceConverter[[]sharedtypes.MaterialAttribute](),
		"[]MCPConfig":                      jsonSliceConverter[[]sharedtypes.MCPConfig](),
		"[]MCPTool":                        jsonSliceConverter[[]sharedtypes.MCPTool](),
		"[]ToolCall":                       jsonSliceConverter[[]sharedtypes.ToolCall](),
		"[]ToolResult":                     jsonSliceConverter[[]sharedtypes.ToolResult](),
		"[]SlashCommand":                   jsonSliceConverter[[]sharedtypes.SlashCommand](),
		"[]DiscoveryMaterial":              jsonSliceConverter[[]sharedtypes.DiscoveryMaterial](),
		"[]DiscoveryBoundaryCondition":     jsonSliceConverter[[]sharedtypes.DiscoveryBoundaryCondition](),
		"[]DiscoveryMonitors":              jsonSliceConverter[[]sharedtypes.DiscoveryMonitors](),
		"[]DiscoveryAttachment":            jsonSliceConverter[[]sharedtypes.DiscoveryAttachment](),
		"[]AedtApiDbResponse":              jsonSliceConverter[[]sharedtypes.AedtApiDbResponse](),
		"[]AedtCodeGenerationExample":      jsonSliceConverter[[]sharedtypes.AedtCodeGenerationExample](),
		"[]AedtCodeGenerationElement":      jsonSliceConverter[[]sharedtypes.AedtCodeGenerationElement](),
		"[][]AedtApiDbResponse":            jsonSliceConverter[[][]sharedtypes.AedtApiDbResponse](), // batch responses
		"[][]DbResponse":                   jsonSliceConverter[[][]sharedtypes.DbResponse](),        // batch responses
	}
}

// interfaceConverter creates a converter for interface{}/any types
func interfaceConverter() TypeConverter {
	return TypeConverter{
		FromString: func(value string) (interface{}, error) {
			if value == "" || value == "null" {
				return nil, nil
			}
			trimmed := strings.TrimSpace(value)

			// Check if it looks like a JSON object or array
			if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
				(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) ||
				(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) {
				var output interface{}
				err := json.Unmarshal([]byte(value), &output)
				return output, err
			} else if trimmed == "true" || trimmed == "false" {
				return trimmed == "true", nil
			} else if num, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if intNum, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
					return intNum, nil
				}
				return num, nil
			}
			return value, nil
		},
		ToString: func(value interface{}) (string, error) {
			if s, ok := value.(string); ok {
				return s, nil
			}
			output, err := json.Marshal(value)
			return string(output), err
		},
	}
}

// jsonSliceConverter creates a converter for slice types that use JSON serialization
func jsonSliceConverter[T any]() TypeConverter {
	return TypeConverter{
		FromString: func(value string) (interface{}, error) {
			if value == "" {
				value = "[]"
			}
			var output T
			err := json.Unmarshal([]byte(value), &output)
			return output, err
		},
		ToString: func(value interface{}) (string, error) {
			output, err := json.Marshal(value)
			return string(output), err
		},
	}
}

// jsonMapConverter creates a converter for map/struct types that use JSON serialization
func jsonMapConverter[T any]() TypeConverter {
	return TypeConverter{
		FromString: func(value string) (interface{}, error) {
			if value == "" {
				value = "{}"
			}
			var output T
			err := json.Unmarshal([]byte(value), &output)
			return output, err
		},
		ToString: func(value interface{}) (string, error) {
			output, err := json.Marshal(value)
			return string(output), err
		},
	}
}

// chanConverter creates a converter for channel pointer types (always nil/empty)
func chanConverter[T any]() TypeConverter {
	return TypeConverter{
		FromString: func(value string) (interface{}, error) {
			var output *T
			return output, nil
		},
		ToString: func(value interface{}) (string, error) {
			return "", nil
		},
	}
}

// JSONToGo converts a JSON data type to a Go data type.
//
// Parameters:
//
//	jsonType: The JSON data type to convert.
//
// Returns:
//
//	string: The Go data type.
//	error: An error if the JSON data type is not supported.
func JSONToGo(jsonType string) (string, error) {
	// Handle array types
	if strings.HasPrefix(jsonType, "array<") && strings.HasSuffix(jsonType, ">") {
		elementType := jsonType[6 : len(jsonType)-1]
		arrayType, err := JSONToGo(elementType)
		if err != nil {
			return "", err
		}

		return "[]" + arrayType, nil
	}

	// Handle dictionary types
	if strings.HasPrefix(jsonType, "dict[") && strings.HasSuffix(jsonType, "]") {
		// Extract the inner types of the dictionary
		inner := jsonType[5 : len(jsonType)-1]
		parts := strings.Split(inner, "][")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid dictionary type: %s", jsonType)
		}

		keyType := parts[0]
		valueType := parts[1]

		// Convert the value type using JSONToGo
		goValueType, err := JSONToGo(valueType)
		if err != nil {
			return "", err
		}

		// Go maps always have string keys
		if keyType != "string" {
			return "", fmt.Errorf("unsupported key type for Go maps: %s (only string keys are allowed)", keyType)
		}

		return fmt.Sprintf("map[string]%s", goValueType), nil
	}

	switch {
	case jsonType == "string":
		return "string", nil
	case jsonType == "string(binary)":
		return "[]byte", nil
	case jsonType == "number":
		return "float64", nil // Default to float64 for general numeric values
	case jsonType == "integer":
		return "int", nil
	case jsonType == "boolean":
		return "bool", nil
	case jsonType == "Any" || jsonType == "any" || jsonType == "object":
		return "interface{}", nil
	default:
		return "", fmt.Errorf("not supported JSON type: %s", jsonType)
	}
}

// GoToJSON converts a Go data type to a JSON data type.
//
// Parameters:
//
//	goType: The Go data type to convert.
//
// Returns:
//
//	string: The JSON data type.
func GoToJSON(goType string) string {
	if strings.HasPrefix(goType, "[]") && goType != "[]byte" {
		elementType := goType[2:]
		return "array<" + GoToJSON(elementType) + ">"
	}

	// Handle maps (map[string]T)
	if strings.HasPrefix(goType, "map[string]") {
		// Extract the value type (after "map[string]")
		valueType := goType[len("map[string]"):]
		return "dict[string][" + GoToJSON(valueType) + "]"
	}

	switch goType {
	case "string":
		return "string"
	case "float32", "float64":
		return "number"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "bool":
		return "boolean"
	case "[]byte":
		return "string(binary)"
	default:
		return "object"
	}
}

// GetSupportedTypes returns a list of all Go types supported by ConvertStringToGivenType.
//
// Returns:
// - []string: a slice containing all supported Go type names
func GetSupportedTypes() []string {
	types := make([]string, 0, len(typeRegistry))
	for goType := range typeRegistry {
		types = append(types, goType)
	}
	return types
}

// ConvertStringToGivenType converts a string to a given Go type.
//
// Parameters:
// - value: a string containing the value to convert
// - goType: a string containing the Go type to convert to
//
// Returns:
// - output: an interface containing the converted value
// - exists: a bool indicating whether the conversion was successful
// - err: an error containing the error message
func ConvertStringToGivenType(value string, goType string) (output interface{}, exists bool, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in ConvertStringToGivenType: %v", r)
		}
	}()

	converter, ok := typeRegistry[goType]
	if !ok {
		return nil, false, nil
	}

	result, err := converter.FromString(value)
	return result, true, err
}

// ConvertGivenTypeToString converts a given Go type to a string.
//
// Parameters:
// - value: an interface containing the value to convert
// - goType: a string containing the Go type to convert from
//
// Returns:
// - output: a string containing the converted value
// - exists: a bool indicating whether the conversion was successful
// - err: an error containing the error message
func ConvertGivenTypeToString(value interface{}, goType string) (output string, exists bool, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in ConvertGivenTypeToString: %v", r)
		}
	}()

	converter, ok := typeRegistry[goType]
	if !ok {
		return "", false, nil
	}

	result, err := converter.ToString(value)
	return result, true, err
}

// DeepCopy deep copies the source interface to the destination interface.
//
// Parameters:
// - src: an interface containing the source
// - dst: an interface containing the destination
//
// Returns:
// - err: an error containing the error message
func DeepCopy(src, dst interface{}) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in DeepCopy: %v", r)
		}
	}()

	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dst)
}
