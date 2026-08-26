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
	"fmt"

	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/anthropics/anthropic-sdk-go"
)

// ConvertMCPToAnthropicFormat converts MCP tools to Anthropic tool definition format.
//
// Parameters:
//
//	ctx: The logging context map.
//	mcpTools: Array of MCP tool definitions (typed MCPTool structs).
//
// Returns:
//
//	[]anthropic.ToolUnionParam: Anthropic formatted tools.
//	[]error: List of errors (empty for typed input, kept for API compatibility).
func ConvertMCPToAnthropicFormat(
	ctx *logging.ContextMap,
	mcpTools []sharedtypes.MCPTool,
) ([]anthropic.ToolUnionParam, []error) {
	var tools []anthropic.ToolUnionParam

	for _, mcpTool := range mcpTools {
		if mcpTool.Name == "" {
			logging.Log.Warnf(ctx, "Skipping tool with empty name")
			continue
		}

		if mcpTool.Description == "" {
			logging.Log.Warnf(ctx, "Tool '%s': missing description (recommended for better LLM understanding)", mcpTool.Name)
		}

		// Use provided inputSchema or create empty one as fallback
		inputSchema := mcpTool.InputSchema
		if inputSchema == nil {
			logging.Log.Warnf(ctx, "Tool '%s': missing 'inputSchema' (LLM may not understand parameters)", mcpTool.Name)
			inputSchema = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		} else if _, has := inputSchema["properties"]; !has {
			inputSchema["properties"] = map[string]interface{}{}
		}

		// Build Anthropic input schema
		schemaParam := anthropic.ToolInputSchemaParam{
			Properties: inputSchema["properties"],
		}

		// Extract "required" field if present
		if req, ok := inputSchema["required"]; ok {
			if reqSlice, ok := req.([]interface{}); ok {
				required := make([]string, 0, len(reqSlice))
				for _, r := range reqSlice {
					if s, ok := r.(string); ok {
						required = append(required, s)
					}
				}
				schemaParam.Required = required
			}
		}

		tool := anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        SanitizeToolName(mcpTool.Name),
				Description: anthropic.String(mcpTool.Description),
				InputSchema: schemaParam,
			},
		}
		tools = append(tools, tool)
		logging.Log.Debugf(ctx, "Converted MCP tool '%s' to Anthropic format", mcpTool.Name)
	}

	if len(tools) > 0 {
		logging.Log.Infof(ctx, "Converted %d MCP tools to Anthropic format", len(tools))
	}

	return tools, nil
}

// ConvertAnthropicToolCallsToSharedTypes converts Anthropic content blocks containing
// tool_use blocks to shared ToolCall format.
//
// Parameters:
//
//	ctx: The logging context map.
//	contentBlocks: Array of Anthropic response content blocks.
//
// Returns:
//
//	[]sharedtypes.ToolCall: Shared format tool calls.
//	[]error: List of errors for tool calls that were skipped during conversion.
func ConvertAnthropicToolCallsToSharedTypes(
	ctx *logging.ContextMap,
	contentBlocks []anthropic.ContentBlockUnion,
) ([]sharedtypes.ToolCall, []error) {
	var toolCalls []sharedtypes.ToolCall
	var errors []error

	for i, block := range contentBlocks {
		// Skip non-tool-use blocks (text, thinking, etc.)
		if block.Type != "tool_use" {
			continue
		}

		// Parse input - handle empty or zero-parameter tools
		var args map[string]interface{}
		if len(block.Input) == 0 || string(block.Input) == "{}" {
			args = map[string]interface{}{}
			logging.Log.Debugf(ctx, "Tool call at index %d (ID: %s, Name: %s) has no arguments (zero-parameter tool)", i, block.ID, block.Name)
		} else {
			if err := json.Unmarshal(block.Input, &args); err != nil {
				parseErr := fmt.Errorf("failed to parse tool call at index %d (ID: %s, Name: %s): %w, raw input: %s",
					i, block.ID, block.Name, err, string(block.Input))
				errors = append(errors, parseErr)
				logging.Log.Errorf(ctx, "Failed to parse Anthropic tool call at index %d (ID: %s, Name: %s): %v, raw input: %s, skipping tool call",
					i, block.ID, block.Name, err, string(block.Input))
				continue
			}
		}

		toolCalls = append(toolCalls, sharedtypes.ToolCall{
			ID:    block.ID,
			Type:  "function",
			Name:  block.Name,
			Input: args,
		})
	}

	if len(toolCalls) > 0 {
		logging.Log.Infof(ctx, "Converted %d Anthropic tool calls to shared format", len(toolCalls))
	}
	if len(errors) > 0 {
		logging.Log.Errorf(ctx, "Failed to convert %d tool calls (see detailed errors above)", len(errors))
	}

	return toolCalls, errors
}

// ConvertSharedTypesToAnthropicToolCalls converts shared ToolCall format to Anthropic
// content blocks for conversation history reconstruction.
//
// Parameters:
//
//	ctx: The logging context map.
//	toolCalls: Array of shared format tool calls.
//
// Returns:
//
//	[]anthropic.ContentBlockParamUnion: Anthropic formatted content blocks.
//	[]error: List of errors for tool calls that failed conversion.
func ConvertSharedTypesToAnthropicToolCalls(
	ctx *logging.ContextMap,
	toolCalls []sharedtypes.ToolCall,
) ([]anthropic.ContentBlockParamUnion, []error) {
	var blocks []anthropic.ContentBlockParamUnion
	var errors []error

	for i, tc := range toolCalls {
		// Serialize arguments back to JSON
		argsJSON, err := json.Marshal(tc.Input)
		if err != nil {
			parseErr := fmt.Errorf("failed to serialize tool call arguments at index %d (ID: %s, Name: %s): %w",
				i, tc.ID, tc.Name, err)
			errors = append(errors, parseErr)
			logging.Log.Errorf(ctx, "Failed to serialize tool call at index %d (ID: %s, Name: %s): %v, skipping",
				i, tc.ID, tc.Name, err)
			continue
		}

		block := anthropic.NewToolUseBlock(tc.ID, json.RawMessage(argsJSON), tc.Name)
		blocks = append(blocks, block)
	}

	if len(blocks) > 0 {
		logging.Log.Debugf(ctx, "Converted %d shared tool calls to Anthropic format", len(blocks))
	}

	return blocks, errors
}
