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

package toolconverters

import (
	"testing"

	"github.com/ansys/snps-ai-sharedtypes/pkg/config"
	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/openai/openai-go/v2"
)

func initTestLogger() {
	testConfig := &config.Config{LOG_LEVEL: "debug"}
	config.GlobalConfig = testConfig
	logging.InitLogger(testConfig)
}
func TestSanitizeToolName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"spaces to underscores", "List Running Products", "List_Running_Products"},
		{"already valid", "valid_name", "valid_name"},
		{"dashes preserved", "get-data", "get-data"},
		{"dots preserved", "file.read", "file.read"},
		{"special chars removed", "get@data!", "get_data_"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeToolName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeToolName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
func TestConvertMCPToOpenAIFormat(t *testing.T) {
	initTestLogger()
	ctx := &logging.ContextMap{}

	tests := []struct {
		name      string
		tools     []sharedtypes.MCPTool
		wantCount int
	}{
		{
			name:      "empty list",
			tools:     []sharedtypes.MCPTool{},
			wantCount: 0,
		},
		{
			name: "single tool",
			tools: []sharedtypes.MCPTool{
				{Name: "List Running Products", Description: "Lists products"},
			},
			wantCount: 1,
		},
		{
			name: "tool with schema",
			tools: []sharedtypes.MCPTool{
				{
					Name:        "Start Product",
					Description: "Starts a product",
					InputSchema: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "empty name skipped",
			tools: []sharedtypes.MCPTool{
				{Name: "", Description: "no name"},
				{Name: "valid", Description: "has name"},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ConvertMCPToOpenAIFormat(ctx, tt.tools)
			if len(result) != tt.wantCount {
				t.Errorf("got %d tools, want %d", len(result), tt.wantCount)
			}
		})
	}
}
func TestConvertOpenAIToolCallsToSharedTypes(t *testing.T) {
	initTestLogger()
	ctx := &logging.ContextMap{}

	tests := []struct {
		name       string
		toolCalls  []openai.ChatCompletionMessageToolCallUnion
		wantCount  int
		wantErrors int
	}{
		{
			name:       "empty list",
			toolCalls:  []openai.ChatCompletionMessageToolCallUnion{},
			wantCount:  0,
			wantErrors: 0,
		},
		{
			name: "valid tool call",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID:   "call_123",
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "List_Running_Products",
						Arguments: `{"filter": "MAPDL"}`,
					},
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "empty arguments (zero-param tool)",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID:   "call_456",
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "no_params_tool",
						Arguments: "",
					},
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "invalid JSON skipped",
			toolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID:   "call_valid",
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "tool1",
						Arguments: `{"valid": "json"}`,
					},
				},
				{
					ID:   "call_invalid",
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "tool2",
						Arguments: `{invalid json`,
					},
				},
			},
			wantCount:  1,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := ConvertOpenAIToolCallsToSharedTypes(ctx, tt.toolCalls)
			if len(result) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result), tt.wantCount)
			}
			if len(errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(errs), tt.wantErrors)
			}
		})
	}
}
func TestConvertSharedTypesToOpenAIToolCalls(t *testing.T) {
	initTestLogger()
	ctx := &logging.ContextMap{}

	tests := []struct {
		name       string
		toolCalls  []sharedtypes.ToolCall
		wantCount  int
		wantErrors int
	}{
		{
			name:       "empty list",
			toolCalls:  []sharedtypes.ToolCall{},
			wantCount:  0,
			wantErrors: 0,
		},
		{
			name: "single tool call",
			toolCalls: []sharedtypes.ToolCall{
				{
					ID:    "call_123",
					Type:  "function",
					Name:  "test_tool",
					Input: map[string]interface{}{"param": "value"},
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "empty input (zero-param tool)",
			toolCalls: []sharedtypes.ToolCall{
				{
					ID:    "call_456",
					Type:  "function",
					Name:  "no_params",
					Input: map[string]interface{}{},
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "multiple tool calls",
			toolCalls: []sharedtypes.ToolCall{
				{
					ID:    "call_1",
					Type:  "function",
					Name:  "tool1",
					Input: map[string]interface{}{"a": "b"},
				},
				{
					ID:    "call_2",
					Type:  "function",
					Name:  "tool2",
					Input: map[string]interface{}{"x": float64(42)},
				},
			},
			wantCount:  2,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := ConvertSharedTypesToOpenAIToolCalls(ctx, tt.toolCalls)
			if len(result) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result), tt.wantCount)
			}
			if len(errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(errs), tt.wantErrors)
			}
		})
	}
}
func TestToolCallsRoundtrip(t *testing.T) {
	initTestLogger()
	ctx := &logging.ContextMap{}

	original := []sharedtypes.ToolCall{
		{
			ID:   "call_roundtrip",
			Type: "function",
			Name: "test_tool",
			Input: map[string]interface{}{
				"stringArg": "value",
				"numberArg": float64(42),
				"boolArg":   true,
			},
		},
	}

	// Convert to OpenAI format
	openaiCalls, errs1 := ConvertSharedTypesToOpenAIToolCalls(ctx, original)
	if len(errs1) > 0 {
		t.Fatalf("ToOpenAI errors: %v", errs1)
	}

	// Convert OpenAI params to response type for roundtrip
	var responseToolCalls []openai.ChatCompletionMessageToolCallUnion
	for _, param := range openaiCalls {
		responseToolCalls = append(responseToolCalls, openai.ChatCompletionMessageToolCallUnion{
			ID:   param.OfFunction.ID,
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      param.OfFunction.Function.Name,
				Arguments: param.OfFunction.Function.Arguments,
			},
		})
	}

	// Convert back to shared types
	restored, errs2 := ConvertOpenAIToolCallsToSharedTypes(ctx, responseToolCalls)
	if len(errs2) > 0 {
		t.Fatalf("FromOpenAI errors: %v", errs2)
	}

	// Verify roundtrip preserved data
	if len(restored) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(restored))
	}

	if restored[0].ID != original[0].ID {
		t.Errorf("ID mismatch: got %q, want %q", restored[0].ID, original[0].ID)
	}
	if restored[0].Name != original[0].Name {
		t.Errorf("Name mismatch: got %q, want %q", restored[0].Name, original[0].Name)
	}
	if restored[0].Input["stringArg"] != original[0].Input["stringArg"] {
		t.Errorf("stringArg mismatch: got %v, want %v", restored[0].Input["stringArg"], original[0].Input["stringArg"])
	}
	if restored[0].Input["numberArg"] != original[0].Input["numberArg"] {
		t.Errorf("numberArg mismatch: got %v, want %v", restored[0].Input["numberArg"], original[0].Input["numberArg"])
	}
	if restored[0].Input["boolArg"] != original[0].Input["boolArg"] {
		t.Errorf("boolArg mismatch: got %v, want %v", restored[0].Input["boolArg"], original[0].Input["boolArg"])
	}
}
