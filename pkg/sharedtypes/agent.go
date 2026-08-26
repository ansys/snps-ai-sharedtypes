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

package sharedtypes

import (
	"encoding/json"

	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
)

// Message represents the JSON message you are expecting
type SessionContext struct {
	WorkflowId string            `json:"workflow_id,omitempty"` // Workflow ID, only relevant if "workflow_endpoint" is "custom"
	Variables  map[string]string `json:"variables,omitempty"`   // Variables to be passed to the workflow
	// Exec connection
	SessionType string `json:"session_type"`      // Type of session: "workflow", "exec"; default is "workflow" if not provided
	ExecId      string `json:"exec_id,omitempty"` // Unique identifier of connecting Exec, only relevant if "session_type" is "exec"
	// Authentication
	JwtToken string `json:"jwt_token"` // JWT token for authentication for "session_type" "workflow" (optional)
	ApiKey   string `json:"api_key"`   // API key for authentication, used if "session_type" is "exec" or can be used if "session_type" is "workflow" for authentication
	// Snapshot logic
	SnapshotId     string `json:"snapshot_id,omitempty"`     // Snapshot ID, only relevant if "session_type" is "workflow"; if defined, the given snapshot will retrived from the database
	WorkflowRunId  string `json:"workflow_run_id,omitempty"` // Workflow run ID, only relevant if "session_type" is "workflow"; if defined, mandatory if "snapshot_id" is defined in order to retrieve the snapshot from the database
	UserId         string `json:"user_id,omitempty"`         // User ID, only relevant if "session_type" is "workflow"; if defined, mandatory if "snapshot_id" is defined in order to retrieve the snapshot from the database
	StoreSnapshots bool   `json:"store_snapshots,omitempty"` // Store snapshots, only relevant if "session_type" is "workflow"; if true, all taken snapshots will be stored in the database
	// Model selection
	ChatModelId string `json:"chat_model_id,omitempty"` // Chat model ID, only relevant if "session_type" is "workflow"; if defined, the given chat model will be used for the workflow run
	// Other
	IpAddress string `json:"ip_address,omitempty"` // IP address of the client, required for some auth_types
}

// ClientRequest is a structure that contains the instruction ID, input, and client ID.
type ClientRequest struct {
	InstructionId     string            `json:"instruction_id"`
	Type              string            `json:"type"` // "message", "get_variable_values", "set_variable_values", "keepalive", "take_snapshot", "load_snapshot", "get_slash_commands", "feedback", "get_conversation_title", "tool_validation", "code_execution", "parallel_code_execution", "edit_assitant_message", "set_chat_model", "approval_response"
	Input             string            `json:"input"`
	Images            []string          `json:"images,omitempty"`
	VariableValues    map[string]string `json:"variable_values,omitempty"`
	SnapshotId        string            `json:"snapshot_id,omitempty"` // Snapshot ID of the snapshot taken or loaded
	Feedback          WorkflowFeedback  `json:"feedback,omitempty"`    // Feedback for the Conversation
	AcceptToolCall    bool              `json:"accept_tool_call,omitempty"`
	UpdatedLlmMessage string            `json:"updated_llm_message,omitempty"` // Updated LLM message after user code change for code execution or edit_assistant_message
	MessageId         string            `json:"message_id,omitempty"`          // Message ID of the message to be edited for the edit_assitant_message case
	ChatModelId       string            `json:"chat_model_id,omitempty"`       // Chat model ID for setting the chat model for the workflow run
}

// ClientResponse is a structure that contains the instruction ID, type of response, output, and client ID.
type ClientResponse struct {
	// Common properties
	InstructionId string `json:"instruction_id"`
	Type          string `json:"type"` // "message", "stream", "info_message", "info_stream", "error", "info", "varaible_values", "snapshot_taken", "snapshot_loaded", "slash_commands", "feedback_received", "conversation_title", "get_tool_validation", "code_execution_response", "assistant_message_edited", "status_message", "status_stream", "chat_model_set", "disable_interrupt", "enable_interrupt", "get_approval"

	// Chat Interface properties
	IsLast           bool   `json:"is_last,omitempty"`
	Position         uint32 `json:"position,omitempty"`
	ChatData         string `json:"chat_data,omitempty"`
	CodeValidation   string `json:"code_validation,omitempty"` // unvalidated, valID, warning, invalid
	InputTokenCount  int    `json:"input_token_count,omitempty"`
	OutputTokenCount int    `json:"output_token_count,omitempty"`
	// Optional usage details forwarded from the provider when available. Omitted when zero so
	// existing clients are unaffected (backward compatible).
	CachedTokenCount    int    `json:"cached_token_count,omitempty"`
	ReasoningTokenCount int    `json:"reasoning_token_count,omitempty"`
	Context             string `json:"context,omitempty"`

	// Variable values properties
	VariableValues map[string]string `json:"variable_values,omitempty"`

	// Snapshot properties
	SnapshotId string `json:"snapshot_id,omitempty"` // Snapshot ID of the snapshot taken or loaded

	// Slash commands properties
	CommandsByCategory []SlashCommandCategory `json:"commands_by_category,omitempty"` // List of @ commands containing their corresponding slash commands

	// Conversation title properties
	ConversationTitle string `json:"conversation_title,omitempty"`

	// Tool validation properties
	ToolCall *ToolCall `json:"tool_call,omitempty"`

	// Code execution properties
	CodeExecutionAllowed bool `json:"code_execution_allowed,omitempty"`

	// Images properties
	Images []string `json:"images,omitempty"`

	// Stream Interruption properties
	StreamInterruptionAllowed bool `json:"stream_interruption_allowed,omitempty"`

	// Approval properties
	ApprovalText    string   `json:"approval_text,omitempty"`
	ApprovalOptions []string `json:"approval_options,omitempty"`

	// Error properties
	Error *ErrorResponse `json:"error,omitempty"`

	// Info properties
	InfoMessage *string `json:"info_message,omitempty"`
}

// ConversationHistoryMessage is a structure that contains the message ID, role, content, and images of a conversation history message.
type ConversationHistoryMessage struct {
	MessageId           string   `json:"message_id"`
	Role                string   `json:"role"`
	Content             string   `json:"content"`
	Images              []string `json:"images"` // image in base64 format
	PositiveFeedback    bool     `json:"positive_feedback"`
	NegativeFeedback    bool     `json:"negative_feedback"`
	FeedbackText        string   `json:"feedback_text,omitempty"` // Optional feedback text
	InputTokenCount     int      `json:"input_token_count,omitempty"`
	OutputTokenCount    int      `json:"output_token_count,omitempty"`
	CachedTokenCount    int      `json:"cached_token_count,omitempty"`
	ReasoningTokenCount int      `json:"reasoning_token_count,omitempty"`
}

// Feedback is a structure that contains the conversation history, message ID, and feedback options of a workflow feedback.
type Feedback struct {
	ConversationHistory []ConversationHistoryMessage `json:"conversation"`
	MessageId           string                       `json:"message_id"`
	AddPositive         bool                         `json:"add_positive"`
	AddNegative         bool                         `json:"add_negative"`
	RemovePositive      bool                         `json:"remove_positive"`
	RemoveNegative      bool                         `json:"remove_negative"`
	FeedbackText        string                       `json:"feedback_text,omitempty"` // Optional feedback text
}

// WorkflowFeedback is a structure that contains the message ID, add positive, add negative, remove positive, and remove negative of a workflow feedback.
type WorkflowFeedback struct {
	MessageId      string `json:"message_id"`
	AddPositive    bool   `json:"add_positive"`
	AddNegative    bool   `json:"add_negative"`
	RemovePositive bool   `json:"remove_positive"`
	RemoveNegative bool   `json:"remove_negative"`
	FeedbackText   string `json:"feedback_text,omitempty"` // Optional feedback text
}

// SlashCommandCategory is a structure that contains the name, description, and list of commands in a slash command category.
type SlashCommandCategory struct {
	Name        string              `json:"name" yaml:"name"`               // Name of the category
	Description string              `json:"description" yaml:"description"` // Description of the category
	Commands    []SlashCommandAgent `json:"commands" yaml:"commands"`       // List of commands in the category
}

// SlashCommand is a structure that contains the name and description of a slash command.
type SlashCommandAgent struct {
	Name        string `json:"name" yaml:"name"`               // Name of the command
	Description string `json:"description" yaml:"description"` // Description of the command
}

// SetSessionContext sets the SessionContext struct from the JSON payload
//
// Parameters:
//   - msg: the JSON payload
//
// Returns:
//   - SessionContext: the SessionContext struct
func ExtractSessionContext(ctx *logging.ContextMap, msg []byte) (SessionContext, error) {
	var SessionContext SessionContext

	// Unmarshal the JSON payload into your struct
	if err := json.Unmarshal(msg, &SessionContext); err != nil {
		logging.Log.Error(ctx, "Error decoding JSON:", err)
		return SessionContext, err
	} else {
		return SessionContext, nil
	}
}
