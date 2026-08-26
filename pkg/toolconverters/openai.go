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

// Package toolconverters provides utilities for converting MCP tool definitions to various LLM provider formats.
package toolconverters

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// sanitizeToolName converts tool names to OpenAI-compatible format.
func SanitizeToolName(name string) string {
	sanitized := strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile(`[^a-zA-Z0-9_.-]`)
	return reg.ReplaceAllString(sanitized, "_")
}

// ConvertMCPToOpenAIFormat converts MCP tools to OpenAI function calling format.
//
// Parameters:
//
//	ctx: The logging context map.
//	mcpTools: Array of MCP tool definitions (typed MCPTool structs).
//
// Returns:
//
//	[]openai.ChatCompletionToolUnionParam: OpenAI formatted tools.
//	[]error: List of errors (empty for typed input, kept for API compatibility).
func ConvertMCPToOpenAIFormat(
	ctx *logging.ContextMap,
	mcpTools []sharedtypes.MCPTool,
) ([]openai.ChatCompletionToolUnionParam, []error) {
	var openaiTools []openai.ChatCompletionToolUnionParam

	for _, mcpTool := range mcpTools {
		// Validate name (required field)
		if mcpTool.Name == "" {
			logging.Log.Warnf(ctx, "Skipping tool with empty name")
			continue
		}

		// Warn if description is missing (recommended but not required)
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
		} else {
			// Ensure 'properties' field exists - Azure OpenAI requires it for object schemas
			if _, hasProperties := inputSchema["properties"]; !hasProperties {
				inputSchema["properties"] = map[string]interface{}{}
			}
		}

		// Convert to OpenAI format
		functionDef := shared.FunctionDefinitionParam{
			Name:        SanitizeToolName(mcpTool.Name),
			Description: openai.String(mcpTool.Description),
			Parameters:  shared.FunctionParameters(inputSchema),
		}

		openaiTool := openai.ChatCompletionFunctionTool(functionDef)
		openaiTools = append(openaiTools, openaiTool)
		logging.Log.Debugf(ctx, "Converted MCP tool '%s' to OpenAI format", mcpTool.Name)
	}

	if len(openaiTools) > 0 {
		logging.Log.Infof(ctx, "Converted %d MCP tools to OpenAI format", len(openaiTools))
	}

	return openaiTools, nil
}

// ConvertOpenAIToolCallsToSharedTypes converts OpenAI SDK tool calls to shared ToolCall format.
//
// Parameters:
//
//	ctx: The logging context map.
//	openaiToolCalls: Array of OpenAI tool call responses.
//
// Returns:
//
//	[]sharedtypes.ToolCall: Shared format tool calls.
//	[]error: List of errors for tool calls that were skipped during conversion.
func ConvertOpenAIToolCallsToSharedTypes(
	ctx *logging.ContextMap,
	openaiToolCalls []openai.ChatCompletionMessageToolCallUnion,
) ([]sharedtypes.ToolCall, []error) {
	var toolCalls []sharedtypes.ToolCall
	var errors []error

	for i, tc := range openaiToolCalls {
		// Parse arguments - handle empty string as empty object (zero-parameter tool)
		var args map[string]interface{}
		if tc.Function.Arguments == "" {
			// Empty arguments string represents a tool with no parameters
			args = map[string]interface{}{}
			logging.Log.Debugf(ctx, "Tool call at index %d (ID: %s, Name: %s) has no arguments (zero-parameter tool)", i, tc.ID, tc.Function.Name)
		} else {
			// Parse non-empty arguments
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				parseErr := fmt.Errorf("failed to parse tool call at index %d (ID: %s, Name: %s): %w, raw arguments: %s",
					i, tc.ID, tc.Function.Name, err, tc.Function.Arguments)
				errors = append(errors, parseErr)
				logging.Log.Errorf(ctx, "Failed to parse tool call at index %d (ID: %s, Name: %s): %v, raw arguments: %s, skipping tool call",
					i, tc.ID, tc.Function.Name, err, tc.Function.Arguments)
				continue
			}
		}

		// Only append valid tool calls
		toolCalls = append(toolCalls, sharedtypes.ToolCall{
			ID:    tc.ID,
			Type:  string(tc.Type),
			Name:  tc.Function.Name,
			Input: args,
		})
	}

	if len(toolCalls) > 0 {
		logging.Log.Infof(ctx, "Converted %d OpenAI tool calls to shared format", len(toolCalls))
	}
	if len(errors) > 0 {
		logging.Log.Errorf(ctx, "Failed to convert %d out of %d tool calls (see detailed errors above)", len(errors), len(openaiToolCalls))
	}

	return toolCalls, errors
}

// ConvertSharedTypesToOpenAIToolCalls converts shared ToolCall format to OpenAI SDK tool call params.
// This is used when reconstructing conversation history with assistant messages that made tool calls.
//
// Parameters:
//
//	ctx: The logging context map.
//	toolCalls: Array of shared format tool calls.
//
// Returns:
//
//	[]openai.ChatCompletionMessageToolCallUnionParam: OpenAI formatted tool call params.
//	[]error: List of errors for tool calls that failed conversion.
func ConvertSharedTypesToOpenAIToolCalls(
	ctx *logging.ContextMap,
	toolCalls []sharedtypes.ToolCall,
) ([]openai.ChatCompletionMessageToolCallUnionParam, []error) {
	var openaiToolCalls []openai.ChatCompletionMessageToolCallUnionParam
	var errors []error

	for i, tc := range toolCalls {
		// Serialize arguments back to JSON string
		argsJSON, err := json.Marshal(tc.Input)
		if err != nil {
			parseErr := fmt.Errorf("failed to serialize tool call arguments at index %d (ID: %s, Name: %s): %w",
				i, tc.ID, tc.Name, err)
			errors = append(errors, parseErr)
			logging.Log.Errorf(ctx, "Failed to serialize tool call at index %d (ID: %s, Name: %s): %v, skipping",
				i, tc.ID, tc.Name, err)
			continue
		}

		openaiToolCall := openai.ChatCompletionMessageToolCallUnionParam{
			OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
				ID: tc.ID,
				Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
					Name:      tc.Name,
					Arguments: string(argsJSON),
				},
			},
		}
		openaiToolCalls = append(openaiToolCalls, openaiToolCall)
	}

	if len(openaiToolCalls) > 0 {
		logging.Log.Debugf(ctx, "Converted %d shared tool calls to OpenAI format", len(openaiToolCalls))
	}

	return openaiToolCalls, errors
}
