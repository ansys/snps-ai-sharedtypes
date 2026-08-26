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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captured records what the fake wiki server saw on the last request.
type captured struct {
	method string
	path   string
	apiKey string
	body   []byte
}

// newTestClient stands up a fake wiki REST server that records the request and replies with status and respBody.
func newTestClient(t *testing.T, apiKey string, status int, respBody string, rec *captured) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.apiKey = r.Header.Get("api-key")
		rec.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, apiKey, srv.Client())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func bodyMap(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode request body %q: %v", raw, err)
	}
	return m
}

func TestHealth(t *testing.T) {
	var rec captured
	c := newTestClient(t, "test-key", http.StatusOK, `{"status":"ok"}`, &rec)

	ok, err := c.GetHealth()
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !ok {
		t.Error("GetHealth() = false, want true")
	}
	if rec.method != http.MethodGet {
		t.Errorf("method = %s, want GET", rec.method)
	}
	if rec.path != "/api/v1/health" {
		t.Errorf("path = %s, want /api/v1/health", rec.path)
	}
	if rec.apiKey != "test-key" {
		t.Errorf("api-key header = %q, want test-key", rec.apiKey)
	}
}

func TestHealth_Error(t *testing.T) {
	var rec captured
	c := newTestClient(t, "test-key", http.StatusInternalServerError, "", &rec)

	ok, err := c.GetHealth()
	if err == nil {
		t.Fatal("GetHealth() error = nil, want non-nil on 500")
	}
	if ok {
		t.Error("GetHealth() = true, want false on 500")
	}
}

// TestOperationRouting drives every POST operation and checks path, method, and the api-key header.
func TestOperationRouting(t *testing.T) {
	cases := []struct {
		name string
		path string
		resp string
		call func(*Client) error
	}{
		{"list_databases", "/api/v1/list_databases", "[]", func(c *Client) error { _, err := c.ListDatabases(); return err }},
		{"get_database_info", "/api/v1/get_database_info", "{}", func(c *Client) error {
			_, err := c.GetDatabaseInfo(GetDatabaseInfoRequest{Name: "db"})
			return err
		}},
		{"create_database", "/api/v1/create_database", "{}", func(c *Client) error {
			_, err := c.CreateDatabase(CreateDatabaseRequest{Name: "db", Description: "d"})
			return err
		}},
		{"delete_database", "/api/v1/delete_database", "{}", func(c *Client) error {
			_, err := c.DeleteDatabase(DeleteDatabaseRequest{Name: "db"})
			return err
		}},
		{"import_folder", "/api/v1/import_folder", "{}", func(c *Client) error {
			_, err := c.ImportFolder(ImportFolderRequest{Database: "db", FolderPath: "/p"})
			return err
		}},
		{"ingest_folder", "/api/v1/ingest_folder", "{}", func(c *Client) error {
			_, err := c.IngestFolder(IngestFolderRequest{Database: "db", FolderPath: "/p"})
			return err
		}},
		{"ingest_file", "/api/v1/ingest_file", "{}", func(c *Client) error {
			_, err := c.IngestFile(IngestFileRequest{Database: "db", FilePath: "/p/f.md"})
			return err
		}},
		{"ingest_text", "/api/v1/ingest_text", "{}", func(c *Client) error {
			_, err := c.IngestText(IngestTextRequest{Database: "db", Name: "n", Content: "c"})
			return err
		}},
		{"remove_content", "/api/v1/remove_content", "{}", func(c *Client) error {
			_, err := c.RemoveContent(RemoveContentRequest{Database: "db", Request: "r"})
			return err
		}},
		{"save", "/api/v1/save", "{}", func(c *Client) error {
			_, err := c.Save(SaveRequest{Database: "db", OutputPath: "/o.zip"})
			return err
		}},
		{"set_answer_prompt", "/api/v1/set_answer_prompt", "{}", func(c *Client) error {
			_, err := c.SetAnswerPrompt(SetAnswerPromptRequest{Database: "db", AnswerPrompt: "p"})
			return err
		}},
		{"query", "/api/v1/query", "{}", func(c *Client) error {
			_, err := c.Query(QueryRequest{Database: "db", Query: "q"})
			return err
		}},
		{"resume", "/api/v1/resume", "{}", func(c *Client) error {
			_, err := c.Resume(ResumeRequest{Database: "db", ArchivePath: "/a.zip"})
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var rec captured
			c := newTestClient(t, "test-key", http.StatusOK, tc.resp, &rec)
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if rec.method != http.MethodPost {
				t.Errorf("method = %s, want POST", rec.method)
			}
			if rec.path != tc.path {
				t.Errorf("path = %s, want %s", rec.path, tc.path)
			}
			if rec.apiKey != "test-key" {
				t.Errorf("api-key header = %q, want test-key", rec.apiKey)
			}
		})
	}
}

func TestApiKeyOmittedWhenEmpty(t *testing.T) {
	var rec captured
	c := newTestClient(t, "", http.StatusOK, "[]", &rec)
	if _, err := c.ListDatabases(); err != nil {
		t.Fatalf("ListDatabases: %v", err)
	}
	if rec.apiKey != "" {
		t.Errorf("api-key header = %q, want empty when no key configured", rec.apiKey)
	}
}

func TestQuery(t *testing.T) {
	var rec captured
	resp := `{"answer":"the answer","revision":"0123456789ab","tokens":{"input":10,"output":20,"cache_read":1,"cache_creation":2,"estimated":true}}`
	c := newTestClient(t, "test-key", http.StatusOK, resp, &rec)

	got, err := c.Query(QueryRequest{Database: "kb", Query: "what?", ModelCategory: []string{"fast"}, InstructionGuid: "guid-1"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	m := bodyMap(t, rec.body)
	if m["database"] != "kb" || m["query"] != "what?" {
		t.Errorf("request body = %s, want database/query set", rec.body)
	}
	if m["instruction_guid"] != "guid-1" {
		t.Errorf("instruction_guid = %v, want guid-1", m["instruction_guid"])
	}
	mc, ok := m["model_category"].([]any)
	if !ok || len(mc) != 1 || mc[0] != "fast" {
		t.Errorf("model_category = %v, want [fast]", m["model_category"])
	}

	if got.Answer != "the answer" {
		t.Errorf("Answer = %q, want %q", got.Answer, "the answer")
	}
	if got.Revision != "0123456789ab" {
		t.Errorf("Revision = %q, want 0123456789ab", got.Revision)
	}
	want := Tokens{Input: 10, Output: 20, CacheRead: 1, CacheCreation: 2, Estimated: true}
	if got.Tokens != want {
		t.Errorf("Tokens = %+v, want %+v", got.Tokens, want)
	}
}

func TestIngestText_MutationResponse(t *testing.T) {
	var rec captured
	resp := `{"stats":{"files":3,"lines":40,"words":120,"bytes":900,"updated":1},"revision":"abcabcabcabc","tokens":{"input":5,"output":6,"cache_read":0,"cache_creation":0,"estimated":false}}`
	c := newTestClient(t, "test-key", http.StatusOK, resp, &rec)

	got, err := c.IngestText(IngestTextRequest{Database: "kb", Name: "note.md", Content: "fact"})
	if err != nil {
		t.Fatalf("IngestText: %v", err)
	}

	m := bodyMap(t, rec.body)
	if m["database"] != "kb" || m["name"] != "note.md" || m["content"] != "fact" {
		t.Errorf("request body = %s, want database/name/content set", rec.body)
	}

	wantStats := Stats{Files: 3, Lines: 40, Words: 120, Bytes: 900, Updated: 1}
	if got.Stats != wantStats {
		t.Errorf("Stats = %+v, want %+v", got.Stats, wantStats)
	}
	if got.Revision != "abcabcabcabc" {
		t.Errorf("Revision = %q, want abcabcabcabc", got.Revision)
	}
	if got.Tokens.Input != 5 || got.Tokens.Output != 6 || got.Tokens.Estimated {
		t.Errorf("Tokens = %+v", got.Tokens)
	}
}

// TestOptionalLLMFieldsOmitted proves the optional routing fields drop out of the wire body when unset.
func TestOptionalLLMFieldsOmitted(t *testing.T) {
	var rec captured
	c := newTestClient(t, "test-key", http.StatusOK, "{}", &rec)

	if _, err := c.Query(QueryRequest{Database: "kb", Query: "q"}); err != nil {
		t.Fatalf("Query: %v", err)
	}
	body := string(rec.body)
	if strings.Contains(body, "model_category") {
		t.Errorf("model_category must be omitted when unset: %s", body)
	}
	if strings.Contains(body, "instruction_guid") {
		t.Errorf("instruction_guid must be omitted when unset: %s", body)
	}
	if strings.Contains(body, "model_ids") {
		t.Errorf("model_ids must be omitted when unset: %s", body)
	}
	if strings.Contains(body, "model_options") {
		t.Errorf("model_options must be omitted when unset: %s", body)
	}
}

// TestModelIDsAndOptionsOnWire proves the model id list and options object serialize into the request body when set.
func TestModelIDsAndOptionsOnWire(t *testing.T) {
	maxTok := int32(256)
	var rec captured
	c := newTestClient(t, "test-key", http.StatusOK, "{}", &rec)

	if _, err := c.Query(QueryRequest{
		Database:     "kb",
		Query:        "q",
		ModelIDs:     []string{"gpt-4o"},
		ModelOptions: &ModelOptions{MaxTokens: &maxTok},
	}); err != nil {
		t.Fatalf("Query: %v", err)
	}
	m := bodyMap(t, rec.body)
	ids, ok := m["model_ids"].([]any)
	if !ok || len(ids) != 1 || ids[0] != "gpt-4o" {
		t.Errorf("model_ids = %v, want [gpt-4o]", m["model_ids"])
	}
	opts, ok := m["model_options"].(map[string]any)
	if !ok || opts["maxTokens"] != float64(256) {
		t.Errorf("model_options.maxTokens = %v, want 256", m["model_options"])
	}
}

func TestListDatabases(t *testing.T) {
	var rec captured
	resp := `[{"name":"kb","description":"knowledge","revision":"aaa"},{"name":"kb2","description":"more","revision":"bbb"}]`
	c := newTestClient(t, "test-key", http.StatusOK, resp, &rec)

	got, err := c.ListDatabases()
	if err != nil {
		t.Fatalf("ListDatabases: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0] != (DatabaseInfo{Name: "kb", Description: "knowledge", Revision: "aaa"}) {
		t.Errorf("got[0] = %+v", got[0])
	}
}

func TestCreateDatabase(t *testing.T) {
	var rec captured
	resp := `{"name":"kb","description":"knowledge","revision":"rev1"}`
	c := newTestClient(t, "test-key", http.StatusOK, resp, &rec)

	got, err := c.CreateDatabase(CreateDatabaseRequest{Name: "kb", Description: "knowledge"})
	if err != nil {
		t.Fatalf("CreateDatabase: %v", err)
	}
	m := bodyMap(t, rec.body)
	if m["name"] != "kb" || m["description"] != "knowledge" {
		t.Errorf("request body = %s, want name/description set", rec.body)
	}
	if got != (DatabaseInfo{Name: "kb", Description: "knowledge", Revision: "rev1"}) {
		t.Errorf("got = %+v", got)
	}
}

func TestDeleteDatabase(t *testing.T) {
	var rec captured
	c := newTestClient(t, "test-key", http.StatusOK, `{"database":"kb"}`, &rec)

	got, err := c.DeleteDatabase(DeleteDatabaseRequest{Name: "kb"})
	if err != nil {
		t.Fatalf("DeleteDatabase: %v", err)
	}
	if bodyMap(t, rec.body)["name"] != "kb" {
		t.Errorf("request body = %s, want name set", rec.body)
	}
	if got.Database != "kb" {
		t.Errorf("Database = %q, want kb", got.Database)
	}
}

func TestErrorStatusCarriesBody(t *testing.T) {
	var rec captured
	c := newTestClient(t, "test-key", http.StatusNotFound, `{"error":"database \"ghost\" not found"}`, &rec)

	_, err := c.GetDatabaseInfo(GetDatabaseInfoRequest{Name: "ghost"})
	if err == nil {
		t.Fatal("error = nil, want non-nil on 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want it to carry status 404", err)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to carry the response body", err)
	}
}
