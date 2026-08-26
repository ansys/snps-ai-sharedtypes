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

package graphdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/ansys/snps-ai-sharedtypes/pkg/clients"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
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

// Update the api key of the client.
func (client *Client) WithApiKey(apiKey string) *Client {
	client.apiKey = apiKey
	return client
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "*/*")
	if client.apiKey != "" {
		req.Header.Set("api-key", client.apiKey)
	}
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func (client Client) GetHealth() (bool, error) {
	url, err := url.JoinPath(client.address, "health")
	if err != nil {
		return false, err
	}
	resp, err := client.get(url)
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

type getVersionResponse struct {
	Status             string `json:"status"`
	Version            string `json:"version"`
	KuzuVersion        string `json:"kuzu_version"`
	KuzuStorageVersion int    `json:"kuzu_storage_version"`
}

func (client Client) GetVersion() (getVersionResponse, error) {
	url, err := url.JoinPath(client.address, "health")
	if err != nil {
		return getVersionResponse{}, err
	}
	resp, err := client.get(url)
	if err != nil {
		return getVersionResponse{}, err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return getVersionResponse{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	var r getVersionResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return getVersionResponse{}, err
	}

	return r, nil
}

type getDatabasesResponse struct {
	Databases []string `json:"databases"`
}

func (client Client) GetDatabases() ([]string, error) {
	url, err := url.JoinPath(client.address, "databases")
	if err != nil {
		return nil, err
	}
	resp, err := client.get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	var r getDatabasesResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return r.Databases, nil
}

const createDbPutMinVer = "v1.2.14"

func (client Client) CreateDatabase(name string) error {
	ver, err := client.GetVersion()
	if err != nil {
		return err
	}

	if semverComp(ver.Version, createDbPutMinVer) >= 0 {
		return client.createDatabasePut(name)
	} else {
		return client.createDatabasePost(name)
	}
}

func (client Client) createDatabasePost(name string) error {
	client.logger.Warn("The `POST /databases` method for creating a new DB is deprecated. Upgrade your snps-ai-graphdb server to use the newer `PUT /databases/{name}` method")

	u, err := url.JoinPath(client.address, "databases")
	if err != nil {
		return err
	}

	resp, err := client.post(u, map[string]any{"name": name, "in_memory": false})
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

func (client Client) createDatabasePut(name string) error {
	u, err := url.JoinPath(client.address, "databases", name)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, u, nil)
	if err != nil {
		return err
	}
	if client.apiKey != "" {
		req.Header.Set("api-key", client.apiKey)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

func (client Client) DeleteDatabase(name string) error {
	u, err := url.JoinPath(client.address, "databases", name)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	if client.apiKey != "" {
		req.Header.Set("api-key", client.apiKey)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}
	return nil
}

const getSchemaMinVer = "v1.2.14"

func (client Client) GetSchema(name string) (string, error) {
	ver, err := client.GetVersion()
	if err != nil {
		return "", err
	}
	if semverComp(ver.Version, getSchemaMinVer) < 0 {
		return "", fmt.Errorf("the `GET /databases/{name}/schema` endpoint was introduced in %s, but the current server is %s", createDbPutMinVer, ver.Version)
	}

	url, err := url.JoinPath(client.address, "databases", name, "schema")
	if err != nil {
		return "", err
	}
	resp, err := client.get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type cypherQueryResponse[T any] struct {
	Result []T `json:"result"`
}

func CypherQueryReadGeneric[T any](client *Client, db string, cypher string, parameters Parameters) ([]T, error) {
	u, err := url.JoinPath(client.address, "databases", db, "read")
	if err != nil {
		return nil, err
	}

	var params map[string]Value
	if parameters != nil {
		params, err = parameters.AsParameters()
		if err != nil {
			return nil, err
		}
	}

	resp, err := client.post(u, map[string]any{"cypher": cypher, "parameters": params})
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		return nil, fmt.Errorf("unexpected status code: %v %q", resp.StatusCode, body)
	}

	var r cypherQueryResponse[T]
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return r.Result, nil
}

func (client *Client) CypherQueryRead(db string, cypher string, parameters Parameters) ([]map[string]any, error) {
	return CypherQueryReadGeneric[map[string]any](client, db, cypher, parameters)
}

func CypherQueryWriteGeneric[T any](client *Client, db string, cypher string, parameters Parameters) ([]T, error) {
	u, err := url.JoinPath(client.address, "databases", db, "write")
	if err != nil {
		return nil, err
	}

	var params map[string]Value
	if parameters != nil {
		params, err = parameters.AsParameters()
		if err != nil {
			return nil, err
		}
	}

	resp, err := client.post(u, map[string]any{"cypher": cypher, "parameters": params})
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		return nil, fmt.Errorf("unexpected status code: %v %q", resp.StatusCode, body)
	}

	var r cypherQueryResponse[T]
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return r.Result, nil
}

func (client *Client) CypherQueryWrite(db string, cypher string, parameters Parameters) ([]map[string]any, error) {
	return CypherQueryWriteGeneric[map[string]any](client, db, cypher, parameters)
}

type ParameterMap map[string]Value

func (pm *ParameterMap) UnmarshalJSON(data []byte) error {
	var intermediate map[string]valueUnmarshalHelper
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return err
	}

	paramMap := make(map[string]Value, len(intermediate))
	for k, v := range intermediate {
		paramMap[k] = v.Value
	}
	*pm = ParameterMap(paramMap)
	return nil
}

type Parameters interface {
	AsParameters() (map[string]Value, error)
}

func (params ParameterMap) AsParameters() (map[string]Value, error) {
	return params, nil
}

type graphDbExportOpts struct {
	Format      string `json:"format,omitempty"`
	Compression string `json:"compression,omitempty"`
}

type AaliGraphDbExportOpt interface {
	apply(*graphDbExportOpts)
}

type WithFormatParquet struct{}

func (WithFormatParquet) apply(opts *graphDbExportOpts) {
	opts.Format = "parquet"
}

type WithFormatCsv struct{}

func (WithFormatCsv) apply(opts *graphDbExportOpts) {
	opts.Format = "csv"
}

type WithCompressionDefault struct{}

func (WithCompressionDefault) apply(opts *graphDbExportOpts) {
	opts.Compression = "default"
}

type WithCompressionFast struct{}

func (WithCompressionFast) apply(opts *graphDbExportOpts) {
	opts.Compression = "fast"
}

type WithCompressionBest struct{}

func (WithCompressionBest) apply(opts *graphDbExportOpts) {
	opts.Compression = "best"
}

func (client *Client) ExportDatabase(name string, dst io.Writer, opts ...AaliGraphDbExportOpt) error {
	exportOpts := &graphDbExportOpts{}
	for _, opt := range opts {
		opt.apply(exportOpts)
	}

	u, err := url.JoinPath(client.address, "databases", name, "export")
	if err != nil {
		return err
	}

	resp, err := client.post(u, exportOpts)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		return fmt.Errorf("unexpected status code: %v %q", resp.StatusCode, body)
	}

	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) ImportDatabase(name string, src io.Reader) error {
	u, err := url.JoinPath(client.address, "databases", name, "import")
	if err != nil {
		return err
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormField("file")
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, src); err != nil {
		return nil
	}
	if err = w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u, &b)
	if err != nil {
		return err
	}
	if client.apiKey != "" {
		req.Header.Set("api-key", client.apiKey)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			client.logger.Warn("could not close body")
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		return fmt.Errorf("unexpected status code: %v %q", resp.StatusCode, body)
	}

	return nil
}

// A helper method for comparing semvers.
//
// It does the same thing as [semver.Compare], excepts it prepends a "v" prefix if the provided versions don't have them.
// This is since Go's semver requires `v*` even though that's not really part of the semver spec.
func semverComp(v string, w string) int {
	if !strings.HasPrefix(v, "v") {
		v = fmt.Sprintf("v%s", v)
	}
	if !strings.HasPrefix(w, "v") {
		w = fmt.Sprintf("v%s", w)
	}
	return semver.Compare(v, w)
}
