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

package wiki

// GetDatabaseInfoRequest asks for a single database's metadata by name.
type GetDatabaseInfoRequest struct {
	Name string `json:"name"`
}

// CreateDatabaseRequest creates a database with a required description.
type CreateDatabaseRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeleteDatabaseRequest drops a database by name.
type DeleteDatabaseRequest struct {
	Name string `json:"name"`
}

// ModelOptions mirrors the snps-ai-llm model configuration passed through on a request; kept local so this client stays dependency-light.
type ModelOptions struct {
	FrequencyPenalty     *float32 `json:"frequencyPenalty,omitempty"`
	MaxTokens            *int32   `json:"maxTokens,omitempty"`
	PresencePenalty      *float32 `json:"presencePenalty,omitempty"`
	Stop                 []string `json:"stop,omitempty"`
	Temperature          *float32 `json:"temperature,omitempty"`
	TopP                 *float32 `json:"topP,omitempty"`
	ReasoningEffort      *string  `json:"reasoningEffort,omitempty"`
	ReasoningSummary     *string  `json:"reasoningSummary,omitempty"`
	Verbosity            *string  `json:"verbosity,omitempty"`
	ThinkingMode         *string  `json:"thinkingMode,omitempty"`
	ThinkingBudgetTokens *int64   `json:"thinkingBudgetTokens,omitempty"`
	ThinkingDisplayMode  *string  `json:"thinkingDisplayMode,omitempty"`
}

// ImportFolderRequest imports a folder of documents into a database.
type ImportFolderRequest struct {
	Database        string        `json:"database"`
	FolderPath      string        `json:"folder_path"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// IngestFolderRequest ingests a folder of documents into a database.
type IngestFolderRequest struct {
	Database        string        `json:"database"`
	FolderPath      string        `json:"folder_path"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// IngestFileRequest ingests a single file into a database.
type IngestFileRequest struct {
	Database        string        `json:"database"`
	FilePath        string        `json:"file_path"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// IngestTextRequest ingests inline text under a page name into a database.
type IngestTextRequest struct {
	Database        string        `json:"database"`
	Name            string        `json:"name"`
	Content         string        `json:"content"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// RemoveContentRequest removes content matching a natural-language request.
type RemoveContentRequest struct {
	Database        string        `json:"database"`
	Request         string        `json:"request"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// SaveRequest writes a database snapshot archive to a server-side path.
type SaveRequest struct {
	Database   string `json:"database"`
	OutputPath string `json:"output_path"`
}

// SetAnswerPromptRequest sets the answer prompt for a database.
type SetAnswerPromptRequest struct {
	Database     string `json:"database"`
	AnswerPrompt string `json:"answer_prompt"`
}

// QueryRequest asks a question against a database.
type QueryRequest struct {
	Database        string        `json:"database"`
	Query           string        `json:"query"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// ResumeRequest restores a database from a snapshot archive on a server-side path.
type ResumeRequest struct {
	Database    string `json:"database"`
	ArchivePath string `json:"archive_path"`
}

// Tokens is the token accounting the server reports for an LLM-backed operation.
type Tokens struct {
	Input         int64 `json:"input"`
	Output        int64 `json:"output"`
	CacheRead     int64 `json:"cache_read"`
	CacheCreation int64 `json:"cache_creation"`
	Estimated     bool  `json:"estimated"`
}

// Stats is the resulting wiki size plus the number of pages a mutation touched.
type Stats struct {
	Files   int `json:"files"`
	Lines   int `json:"lines"`
	Words   int `json:"words"`
	Bytes   int `json:"bytes"`
	Updated int `json:"updated"`
}

// MutationResponse is returned by every operation that changes wiki content.
type MutationResponse struct {
	Stats    Stats  `json:"stats"`
	Revision string `json:"revision"`
	Tokens   Tokens `json:"tokens"`
}

// QueryResponse is the typed answer to a query.
type QueryResponse struct {
	Answer   string `json:"answer"`
	Revision string `json:"revision"`
	Tokens   Tokens `json:"tokens"`
}

// DatabaseInfo identifies a database and carries its description and current revision.
type DatabaseInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Revision    string `json:"revision"`
}

// DeletedResponse acknowledges a dropped database.
type DeletedResponse struct {
	Database string `json:"database"`
}
