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
	"encoding/json"
	"testing"

	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/anthropics/anthropic-sdk-go"
)

func TestConvertMCPToAnthropicFormat(t *testing.T) {
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
						"type": "object",
						"properties": map[string]interface{}{
							"product_name": map[string]interface{}{
								"type":        "string",
								"description": "Name of the product",
							},
						},
						"required": []interface{}{"product_name"},
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
		{
			name: "tool without properties gets default",
			tools: []sharedtypes.MCPTool{
				{
					Name:        "no_props_tool",
					Description: "Tool without properties key",
					InputSchema: map[string]interface{}{
						"type": "object",
					},
				},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ConvertMCPToAnthropicFormat(ctx, tt.tools)
			if len(result) != tt.wantCount {
				t.Errorf("got %d tools, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestConvertAnthropicToolCallsToSharedTypes(t *testing.T) {
	initTestLogger()
	ctx := &logging.ContextMap{}

	tests := []struct {
		name       string
		blocks     []anthropic.ContentBlockUnion
		wantCount  int
		wantErrors int
	}{
		{
			name:       "empty list",
			blocks:     []anthropic.ContentBlockUnion{},
			wantCount:  0,
			wantErrors: 0,
		},
		{
			name: "valid tool call",
			blocks: []anthropic.ContentBlockUnion{
				{
					Type:  "tool_use",
					ID:    "toolu_123",
					Name:  "List_Running_Products",
					Input: json.RawMessage(`{"filter": "MAPDL"}`),
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "empty input (zero-param tool)",
			blocks: []anthropic.ContentBlockUnion{
				{
					Type:  "tool_use",
					ID:    "toolu_456",
					Name:  "no_params_tool",
					Input: json.RawMessage(`{}`),
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "text blocks skipped",
			blocks: []anthropic.ContentBlockUnion{
				{
					Type: "text",
					Text: "Let me check that for you.",
				},
				{
					Type:  "tool_use",
					ID:    "toolu_789",
					Name:  "get_weather",
					Input: json.RawMessage(`{"city": "Zagreb"}`),
				},
			},
			wantCount:  1,
			wantErrors: 0,
		},
		{
			name: "invalid JSON skipped",
			blocks: []anthropic.ContentBlockUnion{
				{
					Type:  "tool_use",
					ID:    "toolu_valid",
					Name:  "tool1",
					Input: json.RawMessage(`{"valid": "json"}`),
				},
				{
					Type:  "tool_use",
					ID:    "toolu_invalid",
					Name:  "tool2",
					Input: json.RawMessage(`{invalid json`),
				},
			},
			wantCount:  1,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := ConvertAnthropicToolCallsToSharedTypes(ctx, tt.blocks)
			if len(result) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result), tt.wantCount)
			}
			if len(errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(errs), tt.wantErrors)
			}
		})
	}
}

func TestConvertSharedTypesToAnthropicToolCalls(t *testing.T) {
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
					ID:    "toolu_123",
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
					ID:    "toolu_456",
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
					ID:    "toolu_1",
					Type:  "function",
					Name:  "tool1",
					Input: map[string]interface{}{"a": "b"},
				},
				{
					ID:    "toolu_2",
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
			result, errs := ConvertSharedTypesToAnthropicToolCalls(ctx, tt.toolCalls)
			if len(result) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result), tt.wantCount)
			}
			if len(errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(errs), tt.wantErrors)
			}
		})
	}
}

func TestAnthropicToolCallsRoundtrip(t *testing.T) {
	initTestLogger()
	ctx := &logging.ContextMap{}

	original := []sharedtypes.ToolCall{
		{
			ID:   "toolu_roundtrip",
			Type: "function",
			Name: "test_tool",
			Input: map[string]interface{}{
				"stringArg": "value",
				"numberArg": float64(42),
				"boolArg":   true,
			},
		},
	}

	// Convert to Anthropic format (shared → Anthropic content blocks)
	anthropicBlocks, errs1 := ConvertSharedTypesToAnthropicToolCalls(ctx, original)
	if len(errs1) > 0 {
		t.Fatalf("ToAnthropic errors: %v", errs1)
	}
	if len(anthropicBlocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(anthropicBlocks))
	}

	// Simulate response by creating ContentBlockUnion from the param
	// In real usage, this comes from the Anthropic API response
	argsJSON, _ := json.Marshal(original[0].Input)
	responseBlocks := []anthropic.ContentBlockUnion{
		{
			Type:  "tool_use",
			ID:    original[0].ID,
			Name:  original[0].Name,
			Input: json.RawMessage(argsJSON),
		},
	}

	// Convert back to shared types (Anthropic response → shared)
	restored, errs2 := ConvertAnthropicToolCallsToSharedTypes(ctx, responseBlocks)
	if len(errs2) > 0 {
		t.Fatalf("FromAnthropic errors: %v", errs2)
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
