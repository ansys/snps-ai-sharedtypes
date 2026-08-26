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
	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/openai/openai-go/v2"
)

// ConvertMCPToAzureFormat converts MCP tool definitions to Azure OpenAI format.
// Azure OpenAI Service uses the same SDK and format as OpenAI, so this function
// wraps the OpenAI converter for clarity and future flexibility.
func ConvertMCPToAzureFormat(
	ctx *logging.ContextMap,
	mcpTools []sharedtypes.MCPTool,
) ([]openai.ChatCompletionToolUnionParam, []error) {
	// Azure uses the same format as OpenAI
	return ConvertMCPToOpenAIFormat(ctx, mcpTools)
}

// ConvertAzureToolCallsToSharedTypes converts Azure OpenAI tool call responses
// to the shared ToolCall format used throughout the Synopsys AI Tools framework.
// Azure OpenAI Service uses the same response format as OpenAI.
func ConvertAzureToolCallsToSharedTypes(
	ctx *logging.ContextMap,
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
) ([]sharedtypes.ToolCall, []error) {
	// Azure uses the same format as OpenAI
	return ConvertOpenAIToolCallsToSharedTypes(ctx, toolCalls)
}
