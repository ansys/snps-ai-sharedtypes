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
	Replace         bool          `json:"replace,omitempty"`
	Description     string        `json:"description,omitempty"`
	Scope           string        `json:"scope,omitempty"`
	Author          string        `json:"author,omitempty"`
	Message         string        `json:"message,omitempty"`
	Priority        bool          `json:"priority,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// IngestFolderRequest ingests a folder of documents into a database.
type IngestFolderRequest struct {
	Database        string        `json:"database"`
	FolderPath      string        `json:"folder_path"`
	Scope           string        `json:"scope,omitempty"`
	Author          string        `json:"author,omitempty"`
	Message         string        `json:"message,omitempty"`
	Priority        bool          `json:"priority,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// IngestFileRequest ingests a single file into a database.
type IngestFileRequest struct {
	Database        string        `json:"database"`
	FilePath        string        `json:"file_path"`
	Type            string        `json:"type,omitempty"`
	Description     string        `json:"description,omitempty"`
	Language        string        `json:"language,omitempty"`
	Scope           string        `json:"scope,omitempty"`
	Author          string        `json:"author,omitempty"`
	Message         string        `json:"message,omitempty"`
	Priority        bool          `json:"priority,omitempty"`
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
	Type            string        `json:"type,omitempty"`
	Description     string        `json:"description,omitempty"`
	Language        string        `json:"language,omitempty"`
	Scope           string        `json:"scope,omitempty"`
	Author          string        `json:"author,omitempty"`
	Message         string        `json:"message,omitempty"`
	Priority        bool          `json:"priority,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// RemoveContentRequest removes content matching a natural-language request.
type RemoveContentRequest struct {
	Database        string        `json:"database"`
	Request         string        `json:"request"`
	Scope           string        `json:"scope,omitempty"`
	Author          string        `json:"author,omitempty"`
	Message         string        `json:"message,omitempty"`
	Priority        bool          `json:"priority,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// SaveRequest writes a database snapshot archive to a server-side path.
type SaveRequest struct {
	Database    string `json:"database"`
	ArchivePath string `json:"archive_path"`
}

// SetAnswerPromptRequest sets the answer prompt for a database.
type SetAnswerPromptRequest struct {
	Database     string `json:"database"`
	AnswerPrompt string `json:"answer_prompt"`
}

// SetSchemaRequest sets the schema of a database from inline text or a server-side file.
type SetSchemaRequest struct {
	Database string `json:"database"`
	Schema   string `json:"schema"`
	FilePath string `json:"file_path"`
}

// QueryRequest asks a question against a database.
type QueryRequest struct {
	Database        string        `json:"database"`
	Query           string        `json:"query"`
	Scope           string        `json:"scope,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// ResumeRequest restores a database from a snapshot archive on a server-side path.
type ResumeRequest struct {
	Database        string        `json:"database"`
	ArchivePath     string        `json:"archive_path"`
	Replace         bool          `json:"replace,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// RemovePageRequest removes one page by its path.
type RemovePageRequest struct {
	Database        string        `json:"database"`
	Path            string        `json:"path"`
	Scope           string        `json:"scope,omitempty"`
	Author          string        `json:"author,omitempty"`
	Message         string        `json:"message,omitempty"`
	Priority        bool          `json:"priority,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// ListScopesRequest lists the scopes of a database.
type ListScopesRequest struct {
	Database string `json:"database"`
}

// CreateScopeRequest creates a scope under a parent scope, root when the parent is empty.
type CreateScopeRequest struct {
	Database string `json:"database"`
	Parent   string `json:"parent,omitempty"`
	Name     string `json:"name"`
}

// DeleteScopeRequest removes a scope and everything under it.
type DeleteScopeRequest struct {
	Database string `json:"database"`
	Scope    string `json:"scope"`
}

// ListPagesRequest lists the pages a scope sees, root when the scope is empty.
type ListPagesRequest struct {
	Database string `json:"database"`
	Scope    string `json:"scope,omitempty"`
}

// GetPageRequest reads one page as a scope sees it, root when the scope is empty.
type GetPageRequest struct {
	Database string `json:"database"`
	Path     string `json:"path"`
	Scope    string `json:"scope,omitempty"`
}

// HistoryRequest lists the recorded versions of one page as a scope sees it, root when the scope is empty.
type HistoryRequest struct {
	Database string `json:"database"`
	Path     string `json:"path"`
	Scope    string `json:"scope,omitempty"`
}

// DiffRequest compares one page between a source and a target scope against the version they last shared, root when a scope is empty.
type DiffRequest struct {
	Database    string `json:"database"`
	Path        string `json:"path"`
	SourceScope string `json:"source_scope,omitempty"`
	TargetScope string `json:"target_scope,omitempty"`
}

// ListConflictsRequest lists the open conflicts of a scope, root when the scope is empty and every scope for "-".
type ListConflictsRequest struct {
	Database string `json:"database"`
	Scope    string `json:"scope,omitempty"`
}

// ResolveConflictRequest settles one conflict of a page in a scope.
type ResolveConflictRequest struct {
	Database        string        `json:"database"`
	Scope           string        `json:"scope"`
	Path            string        `json:"path"`
	Author          string        `json:"author"`
	Resolution      string        `json:"resolution"`
	Content         string        `json:"content,omitempty"`
	Message         string        `json:"message,omitempty"`
	ModelCategory   []string      `json:"model_category,omitempty"`
	ModelIDs        []string      `json:"model_ids,omitempty"`
	ModelOptions    *ModelOptions `json:"model_options,omitempty"`
	InstructionGuid string        `json:"instruction_guid,omitempty"`
}

// ListChangeSetsRequest lists the operations a scope recorded, root when the scope is empty and every scope for "-", from an RFC3339 since when set.
type ListChangeSetsRequest struct {
	Database string `json:"database"`
	Scope    string `json:"scope,omitempty"`
	Since    string `json:"since,omitempty"`
}

// Tokens is the token accounting the server reports for an LLM-backed operation.
type Tokens struct {
	Input         int64 `json:"input"`
	Output        int64 `json:"output"`
	CacheRead     int64 `json:"cache_read"`
	CacheCreation int64 `json:"cache_creation"`
	Estimated     bool  `json:"estimated"`
}

// Snapshot is a point-in-time measure of wiki content size.
type Snapshot struct {
	Files int `json:"files"`
	Lines int `json:"lines"`
	Words int `json:"words"`
	Bytes int `json:"bytes"`
}

// Stats is the resulting wiki size plus the number of pages a mutation touched.
type Stats struct {
	Files   int `json:"files"`
	Lines   int `json:"lines"`
	Words   int `json:"words"`
	Bytes   int `json:"bytes"`
	Updated int `json:"updated"`
	// Before is the size ahead of the write, absent on a write of the whole wiki.
	Before *Snapshot `json:"before,omitempty"`
}

// ScopePage names one page as one scope holds it.
type ScopePage struct {
	Scope string `json:"scope"`
	Path  string `json:"path"`
}

// PageReason names one page and a reason attached to it.
type PageReason struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// PageChange is one page a write added, updated or removed.
type PageChange struct {
	Path   string `json:"path"`
	Change string `json:"change"`
}

// Propagation reports where a write landed below, which scopes wait and why it had priority.
type Propagation struct {
	Updated    []ScopePage  `json:"updated"`
	Pending    []ScopePage  `json:"pending"`
	Conflicts  []ScopePage  `json:"conflicts"`
	Stale      []ScopePage  `json:"stale"`
	Breaking   []PageReason `json:"breaking"`
	Unassessed []string     `json:"unassessed"`
}

// ChangeSet is what one operation recorded: the id its versions link to and its summary.
type ChangeSet struct {
	ID      int64  `json:"id"`
	Summary string `json:"summary"`
}

// Replacement is the content a whole-wiki write took the place of.
type Replacement struct {
	Pages  int `json:"pages"`
	Scopes int `json:"scopes"`
}

// ScopeCount is how many pages one scope below root holds of its own.
type ScopeCount struct {
	Scope string `json:"scope"`
	Pages int    `json:"pages"`
}

// MutationResponse is returned by every operation that changes wiki content.
type MutationResponse struct {
	Stats Stats `json:"stats"`
	// Pages is absent on a write of the whole wiki.
	Pages       []PageChange `json:"pages,omitempty"`
	Revision    string       `json:"revision"`
	Tokens      Tokens       `json:"tokens"`
	Propagation Propagation  `json:"propagation"`
	ChangeSet   ChangeSet    `json:"change_set"`
	Skipped     []PageReason `json:"skipped,omitempty"`
	Replaced    Replacement  `json:"replaced,omitzero"`
	Scopes      []ScopeCount `json:"scopes,omitempty"`
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

// Scope is one scope of the wiki's hierarchy with the level it sits on.
type Scope struct {
	Path  string `json:"path"`
	Level string `json:"level"`
}

// ScopeDeletedResponse acknowledges a removed scope and counts what went with it.
type ScopeDeletedResponse struct {
	Scope    string `json:"scope"`
	Scopes   int    `json:"scopes"`
	Versions int    `json:"versions"`
}

// PageEntry names one page a scope sees, with the one-line description the index holds for it.
type PageEntry struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Language    string `json:"language"`
	Standard    bool   `json:"standard"`
}

// Page is one page as a scope sees it; an image's content is the base64 of its bytes.
type Page struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Scope    string `json:"scope"`
	Type     string `json:"type"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

// Version is one recorded change of a page: the scope it was written in, by whom, when and why.
type Version struct {
	Scope       string `json:"scope"`
	Author      string `json:"author"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	Deleted     bool   `json:"deleted"`
	// UpstreamScope is the scope a merged version came from, root as ""; nil on a direct write.
	UpstreamScope *string `json:"upstream_scope,omitempty"`
}

// Change is one contiguous edit an author made relative to the base text.
type Change struct {
	Author      string   `json:"author"`
	Type        string   `json:"type"`
	BaseStart   int      `json:"base_start"`
	AuthorStart int      `json:"author_start"`
	Removed     []string `json:"removed"`
	Added       []string `json:"added"`
	Conflict    bool     `json:"conflict"`
}

// ChangeReport is the result of comparing two edited versions of a common base text.
type ChangeReport struct {
	Changes     []Change `json:"changes"`
	Merged      string   `json:"merged"`
	HasConflict bool     `json:"has_conflict"`
}

// DiffReport is what changed between two scopes' versions of one page, and who wrote each of them.
type DiffReport struct {
	Source Version `json:"source"`
	Target Version `json:"target"`
	ChangeReport
}

// AffectedPage is one page a proposal names, and what the model says has to change there.
type AffectedPage struct {
	Path   string `json:"path"`
	Change string `json:"change"`
}

// Conflict is one propagation a scope could not take.
type Conflict struct {
	Path     string  `json:"path"`
	Scope    string  `json:"scope"`
	Upstream Version `json:"upstream"`
	Own      Version `json:"own"`
	// Proposal is empty on a script, an image, or a refusal.
	Proposal string         `json:"proposal"`
	Affected []AffectedPage `json:"affected"`
	ChangeReport
}

// ResolveResponse acknowledges a settled conflict, what the refresh spent and where it landed below.
type ResolveResponse struct {
	Path     string `json:"path"`
	Scope    string `json:"scope"`
	Revision string `json:"revision"`
	Tokens   Tokens `json:"tokens"`
	// Inherited reports a scope returned to reading from above.
	Inherited   bool        `json:"inherited"`
	Propagation Propagation `json:"propagation"`
}

// ChangeSetEntry is one recorded operation as a caller reads it back.
type ChangeSetEntry struct {
	ID        int64    `json:"id"`
	Action    string   `json:"action"`
	Scope     string   `json:"scope"`
	Author    string   `json:"author"`
	CreatedAt string   `json:"created_at"`
	Summary   string   `json:"summary"`
	Pages     []string `json:"pages"`
}
