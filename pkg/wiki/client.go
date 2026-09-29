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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ansys/snps-ai-sharedtypes/pkg/clients"
	"go.uber.org/zap"
)

type Client struct {
	address    string
	apiKey     string
	logger     *zap.Logger
	httpClient *http.Client
}

func NewClient(address string, apiKey string, httpClient *http.Client) (*Client, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	defer logger.Sync() //nolint:errcheck
	return &Client{address, apiKey, logger, httpClient}, nil
}

func DefaultClient(address string, apiKey string) (*Client, error) {
	client, err := clients.GetHttpClient()
	if err != nil {
		return nil, fmt.Errorf("error getting HTTP client with cert: %v", err)
	}
	return NewClient(address, apiKey, client)
}

// WithApiKey replaces the api key.
func (client *Client) WithApiKey(apiKey string) *Client {
	client.apiKey = apiKey
	return client
}

// route builds the full URL for an /api/v1 operation.
func (client Client) route(op string) (string, error) {
	return url.JoinPath(client.address, "api", "v1", op)
}

func (client Client) get(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	if client.apiKey != "" {
		req.Header.Set("api-key", client.apiKey)
	}

	return client.httpClient.Do(req)
}

func (client Client) post(u string, body any) (*http.Response, error) {
	jsonReq, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "*/*")
	if client.apiKey != "" {
		req.Header.Set("api-key", client.apiKey)
	}

	return client.httpClient.Do(req)
}

// postJSON posts body to an /api/v1 operation and decodes the JSON response into T.
func postJSON[T any](client Client, op string, body any) (T, error) {
	var zero T
	u, err := client.route(op)
	if err != nil {
		return zero, err
	}
	resp, err := client.post(u, body)
	if err != nil {
		return zero, err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return zero, fmt.Errorf("unexpected status code: %v %q", resp.StatusCode, errBody)
	}

	var r T
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return zero, err
	}

	return r, nil
}

func (client Client) GetHealth() (bool, error) {
	u, err := client.route("health")
	if err != nil {
		return false, err
	}
	resp, err := client.get(u)
	if err != nil {
		return false, err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return true, nil
}

func (client Client) ListDatabases() ([]DatabaseInfo, error) {
	return postJSON[[]DatabaseInfo](client, "list_databases", struct{}{})
}

func (client Client) GetDatabaseInfo(req GetDatabaseInfoRequest) (DatabaseInfo, error) {
	return postJSON[DatabaseInfo](client, "get_database_info", req)
}

func (client Client) CreateDatabase(req CreateDatabaseRequest) (DatabaseInfo, error) {
	return postJSON[DatabaseInfo](client, "create_database", req)
}

func (client Client) DeleteDatabase(req DeleteDatabaseRequest) (DeletedResponse, error) {
	return postJSON[DeletedResponse](client, "delete_database", req)
}

func (client Client) ImportFolder(req ImportFolderRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "import_folder", req)
}

func (client Client) IngestFolder(req IngestFolderRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "ingest_folder", req)
}

func (client Client) IngestFile(req IngestFileRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "ingest_file", req)
}

func (client Client) IngestText(req IngestTextRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "ingest_text", req)
}

func (client Client) RemoveContent(req RemoveContentRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "remove_content", req)
}

func (client Client) Save(req SaveRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "save", req)
}

func (client Client) SetAnswerPrompt(req SetAnswerPromptRequest) (DatabaseInfo, error) {
	return postJSON[DatabaseInfo](client, "set_answer_prompt", req)
}

func (client Client) SetSchema(req SetSchemaRequest) (DatabaseInfo, error) {
	return postJSON[DatabaseInfo](client, "set_schema", req)
}

func (client Client) Query(req QueryRequest) (QueryResponse, error) {
	return postJSON[QueryResponse](client, "query", req)
}

func (client Client) Resume(req ResumeRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "resume", req)
}

func (client Client) RemovePage(req RemovePageRequest) (MutationResponse, error) {
	return postJSON[MutationResponse](client, "remove_page", req)
}

func (client Client) ListScopes(req ListScopesRequest) ([]Scope, error) {
	return postJSON[[]Scope](client, "list_scopes", req)
}

func (client Client) CreateScope(req CreateScopeRequest) (Scope, error) {
	return postJSON[Scope](client, "create_scope", req)
}

func (client Client) DeleteScope(req DeleteScopeRequest) (ScopeDeletedResponse, error) {
	return postJSON[ScopeDeletedResponse](client, "delete_scope", req)
}

func (client Client) ListPages(req ListPagesRequest) ([]PageEntry, error) {
	return postJSON[[]PageEntry](client, "list_pages", req)
}

func (client Client) GetPage(req GetPageRequest) (Page, error) {
	return postJSON[Page](client, "get_page", req)
}

func (client Client) History(req HistoryRequest) ([]Version, error) {
	return postJSON[[]Version](client, "history", req)
}

func (client Client) Diff(req DiffRequest) (DiffReport, error) {
	return postJSON[DiffReport](client, "diff", req)
}

func (client Client) ListConflicts(req ListConflictsRequest) ([]Conflict, error) {
	return postJSON[[]Conflict](client, "list_conflicts", req)
}

func (client Client) ResolveConflict(req ResolveConflictRequest) (ResolveResponse, error) {
	return postJSON[ResolveResponse](client, "resolve_conflict", req)
}

func (client Client) ListChangeSets(req ListChangeSetsRequest) ([]ChangeSetEntry, error) {
	return postJSON[[]ChangeSetEntry](client, "list_change_sets", req)
}
