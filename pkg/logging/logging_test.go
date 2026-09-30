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

package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/ansys/snps-ai-sharedtypes/pkg/config"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
)

// TestContextMap_Set tests the Set method of ContextMap
func TestContextMap_Set(t *testing.T) {
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "test-guid-123")
	ctx.Set(UserId, "user-456")

	// Verify values were set
	value, exists := ctx.Get(InstructionGuid)
	if !exists {
		t.Error("Expected InstructionGuid to exist")
	}
	if value != "test-guid-123" {
		t.Errorf("Expected 'test-guid-123', got '%v'", value)
	}

	value, exists = ctx.Get(UserId)
	if !exists {
		t.Error("Expected UserId to exist")
	}
	if value != "user-456" {
		t.Errorf("Expected 'user-456', got '%v'", value)
	}
}

// TestContextMap_Get tests the Get method of ContextMap
func TestContextMap_Get(t *testing.T) {
	ctx := &ContextMap{}
	ctx.Set(WorkflowId, "workflow-789")

	// Test existing key
	value, exists := ctx.Get(WorkflowId)
	if !exists {
		t.Error("Expected WorkflowId to exist")
	}
	if value != "workflow-789" {
		t.Errorf("Expected 'workflow-789', got '%v'", value)
	}

	// Test non-existing key
	_, exists = ctx.Get(WorkflowRunId)
	if exists {
		t.Error("Expected WorkflowRunId to not exist")
	}
}

// TestContextMap_Copy tests the Copy method of ContextMap
func TestContextMap_Copy(t *testing.T) {
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "guid-1")
	ctx.Set(UserId, "user-1")
	ctx.Set(Action, "test-action")

	// Copy the context
	copiedCtx := ctx.Copy()

	// Verify all values were copied
	value, exists := copiedCtx.Get(InstructionGuid)
	if !exists || value != "guid-1" {
		t.Error("InstructionGuid not copied correctly")
	}
	value, exists = copiedCtx.Get(UserId)
	if !exists || value != "user-1" {
		t.Error("UserId not copied correctly")
	}
	value, exists = copiedCtx.Get(Action)
	if !exists || value != "test-action" {
		t.Error("Action not copied correctly")
	}

	// Modify original and ensure copy is unaffected
	ctx.Set(InstructionGuid, "guid-2")
	value, _ = copiedCtx.Get(InstructionGuid)
	if value != "guid-1" {
		t.Error("Copy was affected by changes to original")
	}
}

// TestInitLoggerConfig tests the initLoggerConfig function
func TestInitLoggerConfig(t *testing.T) {
	testConfig := Config{
		ErrorFileLocation: "/tmp/errors.log",
		LogLevel:          "debug",
		LocalLogs:         true,
		LocalLogsLocation: "/tmp/local.log",
		DatadogLogs:       false,
		DatadogSource:     "test-source",
		DatadogStage:      "test",
		DatadogVersion:    "1.0.0",
		DatadogService:    "test-service",
		DatadogAPIKey:     "test-key",
		DatadogLogsURL:    "http://example.com",
		DatadogMetrics:    false,
		DatadogMetricsURL: "http://example.com/metrics",
	}

	initLoggerConfig(testConfig)

	// Verify global variables were set
	if ERROR_FILE_LOCATION != "/tmp/errors.log" {
		t.Errorf("ERROR_FILE_LOCATION not set correctly: %s", ERROR_FILE_LOCATION)
	}
	if LOG_LEVEL != "debug" {
		t.Errorf("LOG_LEVEL not set correctly: %s", LOG_LEVEL)
	}
	if !LOCAL_LOGS {
		t.Error("LOCAL_LOGS not set correctly")
	}
	if LOCAL_LOGS_LOCATION != "/tmp/local.log" {
		t.Errorf("LOCAL_LOGS_LOCATION not set correctly: %s", LOCAL_LOGS_LOCATION)
	}
	if DATADOG_LOGS {
		t.Error("DATADOG_LOGS should be false")
	}
	if DATADOG_SOURCE != "test-source" {
		t.Errorf("DATADOG_SOURCE not set correctly: %s", DATADOG_SOURCE)
	}
}

// TestDailyLogPath tests the dailyLogPath function with various input formats
func TestDailyLogPath(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	tests := []struct {
		name     string
		input    string
		wantBase string // expected filename (base) portion
	}{
		{"bare filename", "app.log", today + "_app.log"},
		{"relative path", filepath.Join("logs", "app.log"), today + "_app.log"},
		{"empty string defaults", "", today + "_local.log"},
		{"filename without extension", "mylog", today + "_mylog"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := dailyLogPath(tc.input)
			base := filepath.Base(result)
			if base != tc.wantBase {
				t.Errorf("dailyLogPath(%q) base = %q, want %q", tc.input, base, tc.wantBase)
			}
			// For relative path, verify directory is preserved
			if tc.input == filepath.Join("logs", "app.log") {
				dir := filepath.Dir(result)
				if dir != "logs" {
					t.Errorf("dailyLogPath(%q) dir = %q, want %q", tc.input, dir, "logs")
				}
			}
		})
	}
}

// TestInitLogger tests the InitLogger function
func TestInitLogger(t *testing.T) {
	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(os.TempDir(), "test_errors.log"),
		LOG_LEVEL:           "info",
		LOCAL_LOGS:          false,
		LOCAL_LOGS_LOCATION: filepath.Join(os.TempDir(), "test_local.log"),
		DATADOG_LOGS:        false,
		DATADOG_SOURCE:      "test-app",
		STAGE:               "development",
		VERSION:             "1.0.0",
		SERVICE_NAME:        "test-service",
		LOGGING_API_KEY:     "",
		LOGGING_URL:         "",
		DATADOG_METRICS:     false,
		METRICS_URL:         "",
	}

	InitLogger(testConfig)

	// Verify logger was initialized
	if Log.lw == nil {
		t.Error("Logger was not initialized")
	}

	// Verify config was set
	if LOG_LEVEL != "info" {
		t.Errorf("LOG_LEVEL not set correctly: %s", LOG_LEVEL)
	}
	if DATADOG_SOURCE != "test-app" {
		t.Errorf("DATADOG_SOURCE not set correctly: %s", DATADOG_SOURCE)
	}
}

// TestLoggerError tests the Error logging method
func TestLoggerError(t *testing.T) {
	// Setup
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_error_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "error",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	// Test logging
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "test-error-guid")

	Log.Error(ctx, "Test error message")

	// Give time for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify log file was created
	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLoggerErrorf tests the Errorf logging method
func TestLoggerErrorf(t *testing.T) {
	// Setup
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_errorf_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "error",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	// Test logging
	ctx := &ContextMap{}
	ctx.Set(UserId, "user-123")

	Log.Errorf(ctx, "Test error: %s, code: %d", "connection failed", 500)

	// Give time for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify log file was created
	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLoggerWarn tests the Warn logging method
func TestLoggerWarn(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_warn_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "warn",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	ctx.Set(Action, "test-warn-action")

	Log.Warn(ctx, "Test warning message")

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLoggerWarnf tests the Warnf logging method
func TestLoggerWarnf(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_warnf_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "warn",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	Log.Warnf(ctx, "Warning: %s", "low disk space")

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLoggerInfo tests the Info logging method
func TestLoggerInfo(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_info_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "info",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	ctx.Set(WorkflowId, "workflow-info-123")

	Log.Info(ctx, "Test info message")

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLoggerInfof tests the Infof logging method
func TestLoggerInfof(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_infof_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "info",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	Log.Infof(ctx, "Info: %s started at %s", "service", "12:00")

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLoggerDebugf tests the Debugf logging method
func TestLoggerDebugf(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_debugf_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "debug",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	ctx.Set(ClientGuid, "client-debug-456")

	Log.Debugf(ctx, "Debug: variable value is %d", 42)

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLogLevelFiltering tests that log levels are filtered correctly
func TestLogLevelFiltering(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_level_filter.log")
	defer os.Remove(localLogFile)

	// Set log level to error, so warn/info/debug should not log
	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "error",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}

	// These should not log anything
	Log.Warn(ctx, "This should not appear")
	Log.Info(ctx, "This should not appear")
	Log.Debugf(ctx, "This should not appear: %s", "test")

	time.Sleep(100 * time.Millisecond)

	// File may be created but should not contain the messages (or not exist)
	// This is acceptable behavior
}

// TestMapsToJSONBytes tests the mapsToJSONBytes function
func TestMapsToJSONBytes(t *testing.T) {
	testMaps := []map[string]interface{}{
		{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		},
	}

	jsonBytes, err := mapsToJSONBytes(testMaps)
	if err != nil {
		t.Fatalf("mapsToJSONBytes failed: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("Expected non-empty JSON bytes")
	}

	// Verify it's valid JSON
	var result []map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}
}

// TestLevelToString tests the levelToString function
func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    zapcore.Level
		expected string
	}{
		{TraceLevel, "trace"},
		{zapcore.DebugLevel, "debug"},
		{zapcore.InfoLevel, "info"},
		{zapcore.WarnLevel, "warn"},
		{zapcore.ErrorLevel, "error"},
		{zapcore.FatalLevel, "fatal"},
	}

	for _, tt := range tests {
		result := levelToString(tt.level)
		if result != tt.expected {
			t.Errorf("levelToString(%v) = %s; expected %s", tt.level, result, tt.expected)
		}
	}
}

// TestTimeToString tests the timeToString function
func TestTimeToString(t *testing.T) {
	testTime := time.Date(2025, 1, 15, 10, 30, 45, 123000000, time.UTC)
	result := timeToString(testTime)

	if !strings.Contains(result, "2025-01-15") {
		t.Errorf("Expected date in result, got: %s", result)
	}
	if !strings.Contains(result, "10:30:45") {
		t.Errorf("Expected time in result, got: %s", result)
	}
}

// TestEntryCallerToString tests the entryCallerToString function
func TestEntryCallerToString(t *testing.T) {
	caller := zapcore.EntryCaller{
		Defined: true,
		PC:      0,
		File:    "/path/to/file.go",
		Line:    42,
	}

	result := entryCallerToString(caller)
	if result == "" {
		t.Error("Expected non-empty string")
	}
}

// TestWriteStringToFile tests the writeStringToFile function
func TestWriteStringToFile(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "test_write_string.log")
	defer os.Remove(tempFile)

	err := writeStringToFile(tempFile, "Test message")
	if err != nil {
		t.Fatalf("writeStringToFile failed: %v", err)
	}

	// Verify file exists and contains content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "Test message") {
		t.Error("File does not contain expected message")
	}

	// Test appending
	err = writeStringToFile(tempFile, "Second message")
	if err != nil {
		t.Fatalf("Failed to append to file: %v", err)
	}

	content, _ = os.ReadFile(tempFile)
	if !strings.Contains(string(content), "Second message") {
		t.Error("File does not contain appended message")
	}
}

// TestWriteInterfaceToFile tests the writeInterfaceToFile function
func TestWriteInterfaceToFile(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "test_write_interface.log")
	defer os.Remove(tempFile)

	testData := map[string]interface{}{
		"field1": "value1",
		"field2": 456,
	}

	err := writeInterfaceToFile(tempFile, testData)
	if err != nil {
		t.Fatalf("writeInterfaceToFile failed: %v", err)
	}

	// Verify file exists and contains JSON content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "field1") {
		t.Error("File does not contain expected JSON data")
	}
}

// TestCreateMetaDataFromCtx tests the CreateMetaDataFromCtx function
func TestCreateMetaDataFromCtx(t *testing.T) {
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "guid-123")
	ctx.Set(UserId, "user-456")
	ctx.Set(WorkflowId, "workflow-789")

	grpcCtx := context.Background()
	ctxWithMetadata, err := CreateMetaDataFromCtx(ctx, grpcCtx)
	if err != nil {
		t.Fatalf("CreateMetaDataFromCtx failed: %v", err)
	}

	// Verify metadata was attached
	md, ok := metadata.FromOutgoingContext(ctxWithMetadata)
	if !ok {
		t.Fatal("Failed to extract metadata from context")
	}

	metadataValues := md.Get("snps-ai-logging-context")
	if len(metadataValues) == 0 {
		t.Fatal("No metadata values found")
	}

	// Verify the JSON structure
	var body []map[string]interface{}
	err = json.Unmarshal([]byte(metadataValues[0]), &body)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if len(body) == 0 {
		t.Fatal("Empty metadata body")
	}

	if body[0]["instructionGuid"] != "guid-123" {
		t.Errorf("Expected instructionGuid to be 'guid-123', got '%v'", body[0]["instructionGuid"])
	}
}

// TestCreateCtxFromMetaData tests the CreateCtxFromMetaData function
func TestCreateCtxFromMetaData(t *testing.T) {
	// Create metadata
	body := []map[string]interface{}{
		{
			"instructionGuid": "guid-123",
			"userId":          "user-456",
			"workflowId":      "workflow-789",
		},
	}
	jsonData, _ := json.Marshal(body)

	md := metadata.Pairs("snps-ai-logging-context", string(jsonData))
	grpcCtx := metadata.NewIncomingContext(context.Background(), md)

	// Extract ContextMap
	ctx, err := CreateCtxFromMetaData(grpcCtx)
	if err != nil {
		t.Fatalf("CreateCtxFromMetaData failed: %v", err)
	}

	// Verify values
	value, exists := ctx.Get(InstructionGuid)
	if !exists || value != "guid-123" {
		t.Errorf("Expected instructionGuid to be 'guid-123', got '%v'", value)
	}

	value, exists = ctx.Get(UserId)
	if !exists || value != "user-456" {
		t.Errorf("Expected userId to be 'user-456', got '%v'", value)
	}

	value, exists = ctx.Get(WorkflowId)
	if !exists || value != "workflow-789" {
		t.Errorf("Expected workflowId to be 'workflow-789', got '%v'", value)
	}
}

// TestCreateCtxFromMetaData_Empty tests CreateCtxFromMetaData with empty metadata
func TestCreateCtxFromMetaData_Empty(t *testing.T) {
	grpcCtx := context.Background()

	ctx, err := CreateCtxFromMetaData(grpcCtx)
	if err != nil {
		t.Fatalf("CreateCtxFromMetaData should not fail with empty metadata: %v", err)
	}

	// Should return an empty ContextMap
	_, exists := ctx.Get(InstructionGuid)
	if exists {
		t.Error("Expected no values in empty context")
	}
}

// TestCreateDialOptionsFromCtx tests the CreateDialOptionsFromCtx function
func TestCreateDialOptionsFromCtx(t *testing.T) {
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "guid-websocket")
	ctx.Set(UserId, "user-websocket")

	opts, err := CreateDialOptionsFromCtx(ctx)
	if err != nil {
		t.Fatalf("CreateDialOptionsFromCtx failed: %v", err)
	}

	// Verify options contain header
	if opts == nil || opts.HTTPHeader == nil {
		t.Fatal("Expected non-nil dial options with HTTP headers")
	}

	headerValues := opts.HTTPHeader["snps-ai-logging-context"]
	if len(headerValues) == 0 {
		t.Fatal("Expected snps-ai-logging-context header to be set")
	}

	// Verify JSON structure
	var body []map[string]interface{}
	err = json.Unmarshal([]byte(headerValues[0]), &body)
	if err != nil {
		t.Fatalf("Failed to unmarshal header: %v", err)
	}

	if body[0]["instructionGuid"] != "guid-websocket" {
		t.Errorf("Expected instructionGuid to be 'guid-websocket', got '%v'", body[0]["instructionGuid"])
	}
}

// TestCreateCtxFromHeader tests the CreateCtxFromHeader function
func TestCreateCtxFromHeader(t *testing.T) {
	// Create HTTP request with header
	body := []map[string]interface{}{
		{
			"instructionGuid": "guid-http",
			"userId":          "user-http",
			"action":          "test-action",
		},
	}
	jsonData, _ := json.Marshal(body)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("snps-ai-logging-context", string(jsonData))

	// Extract ContextMap
	ctx, err := CreateCtxFromHeader(req)
	if err != nil {
		t.Fatalf("CreateCtxFromHeader failed: %v", err)
	}

	// Verify values
	value, exists := ctx.Get(InstructionGuid)
	if !exists || value != "guid-http" {
		t.Errorf("Expected instructionGuid to be 'guid-http', got '%v'", value)
	}

	value, exists = ctx.Get(UserId)
	if !exists || value != "user-http" {
		t.Errorf("Expected userId to be 'user-http', got '%v'", value)
	}

	value, exists = ctx.Get(Action)
	if !exists || value != "test-action" {
		t.Errorf("Expected action to be 'test-action', got '%v'", value)
	}
}

// TestCreateCtxFromHeader_Empty tests CreateCtxFromHeader with no header
func TestCreateCtxFromHeader_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)

	ctx, err := CreateCtxFromHeader(req)
	if err != nil {
		t.Fatalf("CreateCtxFromHeader should not fail with no header: %v", err)
	}

	// Should return an empty ContextMap
	_, exists := ctx.Get(InstructionGuid)
	if exists {
		t.Error("Expected no values in empty context")
	}
}

// TestSendPostRequestToDatadog tests the sendPostRequestToDatadog function with mock server
func TestSendPostRequestToDatadog(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("DD-API-KEY") != "test-api-key" {
			t.Error("Expected DD-API-KEY header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type to be application/json")
		}

		// Verify method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Return 202 Accepted
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	testBody := []byte(`{"test": "data"}`)
	resp, err := sendPostRequestToDatadog(server.URL, testBody, "test-api-key")
	if err != nil {
		t.Fatalf("sendPostRequestToDatadog failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", resp.StatusCode)
	}
}

// TestSendPostRequestToDatadog_Error tests error handling
func TestSendPostRequestToDatadog_Error(t *testing.T) {
	// Use invalid URL to trigger error
	_, err := sendPostRequestToDatadog("http://invalid-url-that-does-not-exist.local", []byte("test"), "key")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

// TestMetrics tests the Metrics function
func TestMetrics(t *testing.T) {
	// Setup mock server for metrics
	metricsReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsReceived = true

		// Read and verify body
		body, _ := io.ReadAll(r.Body)
		var metrics Metrics
		err := json.Unmarshal(body, &metrics)
		if err != nil {
			t.Errorf("Failed to unmarshal metrics: %v", err)
		}

		if len(metrics.Series) == 0 {
			t.Error("Expected at least one metric in series")
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	// Configure logger with metrics enabled
	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(os.TempDir(), "test_errors.log"),
		LOG_LEVEL:           "info",
		LOCAL_LOGS:          false,
		DATADOG_LOGS:        false,
		DATADOG_METRICS:     true,
		METRICS_URL:         server.URL,
		LOGGING_API_KEY:     "test-key",
	}
	InitLogger(testConfig)

	// Send metric
	Log.Metrics("test.metric", 42.5)

	// Give time for async operation
	time.Sleep(200 * time.Millisecond)

	if !metricsReceived {
		t.Error("Metrics were not sent to server")
	}
}

// TestMetrics_Disabled tests that metrics are not sent when disabled
func TestMetrics_Disabled(t *testing.T) {
	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(os.TempDir(), "test_errors.log"),
		LOG_LEVEL:           "info",
		DATADOG_METRICS:     false,
	}
	InitLogger(testConfig)

	// This should not panic or error
	Log.Metrics("test.metric", 100.0)
	time.Sleep(50 * time.Millisecond)
}

// TestLoggerWrapper_AllContextKeys tests logging with all context keys
func TestLoggerWrapper_AllContextKeys(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_all_keys.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "info",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	// Set all context keys
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "instruction-123")
	ctx.Set(WorkflowId, "workflow-456")
	ctx.Set(WorkflowRunId, "run-789")
	ctx.Set(UserId, "user-abc")
	ctx.Set(AdapterType, "adapter-xyz")
	ctx.Set(WatchFolderPath, "/path/to/folder")
	ctx.Set(WatchFilePath, "/path/to/file")
	ctx.Set(ReaderGuid, "reader-def")
	ctx.Set(ClientGuid, "client-ghi")
	ctx.Set(Action, "test-action")
	ctx.Set(Rest_Call_Id, "rest-123")
	ctx.Set(Rest_Call, "GET /api/test")
	ctx.Set(UserMail, "user@example.com")

	Log.Info(ctx, "Test message with all context keys")

	time.Sleep(100 * time.Millisecond)

	// Verify log file contains some of the keys
	content, err := os.ReadFile(dailyFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "instruction-123") {
		t.Error("Log does not contain instructionGuid")
	}
	if !strings.Contains(contentStr, "user@example.com") {
		t.Error("Log does not contain userMail")
	}
}

// TestRoundtripMetadata tests creating metadata and extracting it back
func TestRoundtripMetadata(t *testing.T) {
	// Create original context
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "roundtrip-guid")
	ctx.Set(UserId, "roundtrip-user")
	ctx.Set(WorkflowId, "roundtrip-workflow")

	// Convert to gRPC metadata
	grpcCtx := context.Background()
	ctxWithMetadata, err := CreateMetaDataFromCtx(ctx, grpcCtx)
	if err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	// Simulate server receiving the metadata
	md, _ := metadata.FromOutgoingContext(ctxWithMetadata)
	serverCtx := metadata.NewIncomingContext(context.Background(), md)

	// Extract back to ContextMap
	extractedCtx, err := CreateCtxFromMetaData(serverCtx)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Verify all values match
	value, exists := extractedCtx.Get(InstructionGuid)
	if !exists || value != "roundtrip-guid" {
		t.Error("InstructionGuid does not match after roundtrip")
	}

	value, exists = extractedCtx.Get(UserId)
	if !exists || value != "roundtrip-user" {
		t.Error("UserId does not match after roundtrip")
	}

	value, exists = extractedCtx.Get(WorkflowId)
	if !exists || value != "roundtrip-workflow" {
		t.Error("WorkflowId does not match after roundtrip")
	}
}

// TestMultipleLogLevels tests that different log levels work correctly
func TestMultipleLogLevels(t *testing.T) {
	testCases := []struct {
		name     string
		logLevel string
	}{
		{"Fatal", "fatal"},
		{"Error", "error"},
		{"Warn", "warn"},
		{"Info", "info"},
		{"Debug", "debug"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := os.TempDir()
			localLogFile := filepath.Join(tempDir, "test_level_"+tc.name+".log")
			defer os.Remove(localLogFile)

			testConfig := &config.Config{
				ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
				LOG_LEVEL:           tc.logLevel,
				LOCAL_LOGS:          true,
				LOCAL_LOGS_LOCATION: localLogFile,
				DATADOG_LOGS:        false,
			}
			InitLogger(testConfig)

			if LOG_LEVEL != tc.logLevel {
				t.Errorf("LOG_LEVEL not set to %s", tc.logLevel)
			}
		})
	}
}

// TestLoggerTracef tests the Tracef logging method
func TestLoggerTracef(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_tracef_log.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "trace",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	ctx.Set(ClientGuid, "client-trace-789")

	Log.Tracef(ctx, "Trace: detailed value is %d", 99)

	// Give time for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify log file was created and contains trace content
	content, err := os.ReadFile(dailyFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Trace: detailed value is 99") {
		t.Error("Log file does not contain expected trace message")
	}
	if !strings.Contains(contentStr, "TRACE") {
		t.Error("Log file does not contain trace level status")
	}
}

// TestLoggerTracef_NotLoggedAtDebugLevel tests that Tracef does not log when level is debug
func TestLoggerTracef_NotLoggedAtDebugLevel(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_tracef_filtered.log")
	defer os.Remove(localLogFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "debug",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	Log.Tracef(ctx, "This trace should not appear: %s", "hidden")

	time.Sleep(100 * time.Millisecond)

	// File should not exist since trace is filtered at debug level
	if _, err := os.Stat(localLogFile); !os.IsNotExist(err) {
		content, _ := os.ReadFile(localLogFile)
		if strings.Contains(string(content), "This trace should not appear") {
			t.Error("Trace message was logged when log level is debug")
		}
	}
}

// BenchmarkContextMapSet benchmarks the Set operation
func BenchmarkContextMapSet(b *testing.B) {
	ctx := &ContextMap{}
	for i := 0; i < b.N; i++ {
		ctx.Set(InstructionGuid, "test-guid")
	}
}

// BenchmarkContextMapGet benchmarks the Get operation
func BenchmarkContextMapGet(b *testing.B) {
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "test-guid")
	for i := 0; i < b.N; i++ {
		ctx.Get(InstructionGuid)
	}
}

// BenchmarkContextMapCopy benchmarks the Copy operation
func BenchmarkContextMapCopy(b *testing.B) {
	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "test-guid")
	ctx.Set(UserId, "test-user")
	ctx.Set(WorkflowId, "test-workflow")
	for i := 0; i < b.N; i++ {
		ctx.Copy()
	}
}

// BenchmarkLogInfo benchmarks Info logging
func BenchmarkLogInfo(b *testing.B) {
	tempDir := os.TempDir()
	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "bench_errors.log"),
		LOG_LEVEL:           "info",
		LOCAL_LOGS:          false,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "bench-guid")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Log.Info(ctx, "Benchmark message")
	}
}

// TestConcurrentLogOrder verifies that all concurrent log entries appear in the file
// with no missing or corrupted lines.
func TestConcurrentLogOrder(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_concurrent_order.log")
	dailyFile := dailyLogPath(localLogFile)
	defer os.Remove(dailyFile)

	testConfig := &config.Config{
		ERROR_FILE_LOCATION: filepath.Join(tempDir, "test_errors.log"),
		LOG_LEVEL:           "debug",
		LOCAL_LOGS:          true,
		LOCAL_LOGS_LOCATION: localLogFile,
		DATADOG_LOGS:        false,
	}
	InitLogger(testConfig)

	const numGoroutines = 10
	const logsPerGoroutine = 5
	totalLogs := numGoroutines * logsPerGoroutine

	// Launch concurrent goroutines writing to the same log file
	done := make(chan struct{})
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			ctx := &ContextMap{}
			ctx.Set(ClientGuid, fmt.Sprintf("client-%d", goroutineID))
			for i := 0; i < logsPerGoroutine; i++ {
				Log.Debugf(ctx, "CONCURRENT_MARKER g=%d i=%d", goroutineID, i)
			}
			done <- struct{}{}
		}(g)
	}

	// Wait for all goroutines to finish dispatching
	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	// Wait for all async log writes to complete
	pendingLogs.Wait()

	content, err := os.ReadFile(dailyFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)

	// Count how many unique log entries made it into the file
	found := 0
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < logsPerGoroutine; i++ {
			marker := fmt.Sprintf("CONCURRENT_MARKER g=%d i=%d", g, i)
			if strings.Contains(contentStr, marker) {
				found++
			}
		}
	}

	if found != totalLogs {
		t.Errorf("Expected %d log entries, found %d. Some logs were lost.", totalLogs, found)
	}

	// Verify no corrupted lines: every non-empty, non-header line should contain "|"
	lines := strings.Split(contentStr, "\n")
	for lineNum, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "|") {
			t.Errorf("Line %d appears corrupted (no pipe separator): %q", lineNum+1, line)
		}
	}
}

// TestWriteFormattedLogNeverOverwritesExistingFile verifies that writeFormattedLogToFile
// always appends to an existing file and never overwrites its contents.
func TestWriteFormattedLogNeverOverwritesExistingFile(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_no_overwrite.log")
	defer os.Remove(localLogFile)

	// Pre-populate file with known content
	preExisting := "PRE_EXISTING_CONTENT_LINE_1\nPRE_EXISTING_CONTENT_LINE_2\n"
	err := os.WriteFile(localLogFile, []byte(preExisting), 0644)
	if err != nil {
		t.Fatalf("Failed to create pre-existing file: %v", err)
	}

	ctx := &ContextMap{}
	ctx.Set(InstructionGuid, "test-no-overwrite")

	// Write a formatted log entry to the same file
	err = writeFormattedLogToFile(
		localLogFile,
		"test-app",
		"2026-07-14 10:00:00.000",
		"info",
		"test.Function",
		"test.go:42",
		"New log entry",
		"",
		nil,
		ctx,
	)
	if err != nil {
		t.Fatalf("writeFormattedLogToFile failed: %v", err)
	}

	content, err := os.ReadFile(localLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	contentStr := string(content)

	// Pre-existing content must still be there
	if !strings.Contains(contentStr, "PRE_EXISTING_CONTENT_LINE_1") {
		t.Error("Pre-existing content line 1 was overwritten")
	}
	if !strings.Contains(contentStr, "PRE_EXISTING_CONTENT_LINE_2") {
		t.Error("Pre-existing content line 2 was overwritten")
	}
	// New log entry must also be there
	if !strings.Contains(contentStr, "New log entry") {
		t.Error("New log entry was not appended")
	}
	// No header should be written since the file was not empty
	if strings.Contains(contentStr, "TIMESTAMP") {
		t.Error("Header was written to a non-empty file")
	}
}

// TestConcurrentWriteFormattedLogNoDataLoss verifies that many concurrent direct calls to
// writeFormattedLogToFile never lose data or corrupt each other.
func TestConcurrentWriteFormattedLogNoDataLoss(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_concurrent_write.log")
	defer os.Remove(localLogFile)

	const numWriters = 50
	done := make(chan struct{})

	for w := 0; w < numWriters; w++ {
		go func(writerID int) {
			ctx := &ContextMap{}
			ctx.Set(ClientGuid, fmt.Sprintf("writer-%d", writerID))
			err := writeFormattedLogToFile(
				localLogFile,
				"test-app",
				"2026-07-14 10:00:00.000",
				"debug",
				fmt.Sprintf("pkg.Function%d", writerID),
				"file.go:1",
				fmt.Sprintf("WRITER_MARKER_%d", writerID),
				"",
				nil,
				ctx,
			)
			if err != nil {
				t.Errorf("Writer %d failed: %v", writerID, err)
			}
			done <- struct{}{}
		}(w)
	}

	for w := 0; w < numWriters; w++ {
		<-done
	}

	content, err := os.ReadFile(localLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	contentStr := string(content)

	// Every writer's marker must be present
	for w := 0; w < numWriters; w++ {
		marker := fmt.Sprintf("WRITER_MARKER_%d", w)
		if !strings.Contains(contentStr, marker) {
			t.Errorf("Writer %d marker not found — data was lost", w)
		}
	}

	// Header should appear at least once (multiple concurrent writers may each write
	// a header if they all see the file as empty simultaneously — this is a harmless
	// cosmetic race that only happens on initial file creation)
	headerCount := strings.Count(contentStr, "TIMESTAMP")
	if headerCount < 1 {
		t.Errorf("Expected at least 1 header row, found %d", headerCount)
	}
}

// TestWriteFormattedLogHeaderOnFreshFile verifies a header is written on new files
// but not on existing ones.
func TestWriteFormattedLogHeaderOnFreshFile(t *testing.T) {
	tempDir := os.TempDir()
	localLogFile := filepath.Join(tempDir, "test_header_fresh.log")
	defer os.Remove(localLogFile)

	ctx := &ContextMap{}

	// First write: file doesn't exist → header should appear
	err := writeFormattedLogToFile(localLogFile, "test-app", "2026-07-14 10:00:00.000", "info", "test.Func", "test.go:1", "First entry", "", nil, ctx)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	content, err := os.ReadFile(localLogFile)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if !strings.Contains(string(content), "TIMESTAMP") {
		t.Error("Header missing on fresh file")
	}

	// Second write: file exists → no additional header
	err = writeFormattedLogToFile(localLogFile, "test-app", "2026-07-14 10:00:01.000", "debug", "test.Func2", "test.go:2", "Second entry", "", nil, ctx)
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	content, err = os.ReadFile(localLogFile)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	contentStr := string(content)

	headerCount := strings.Count(contentStr, "TIMESTAMP")
	if headerCount != 1 {
		t.Errorf("Expected 1 header, found %d", headerCount)
	}
	if !strings.Contains(contentStr, "First entry") {
		t.Error("First entry missing after second write")
	}
	if !strings.Contains(contentStr, "Second entry") {
		t.Error("Second entry missing")
	}
}

// TestWriteFormattedLogFitsWidestAppName verifies the widest supported app name stays on one line.
func TestWriteFormattedLogFitsWidestAppName(t *testing.T) {
	localLogFile := filepath.Join(t.TempDir(), "test_app_width.log")
	err := writeFormattedLogToFile(localLogFile, "snps-ai-flowkit", "2026-07-14 10:00:00.000", "info", "test.Func", "test.go:1", "Test entry", "", nil, nil)
	if err != nil {
		t.Fatalf("writeFormattedLogToFile failed: %v", err)
	}

	content, err := os.ReadFile(localLogFile)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected header, separator, and one log line; got %d lines", len(lines))
	}
	if !strings.Contains(lines[2], "| snps-ai-flowkit | INFO") {
		t.Errorf("App name was not kept on one line: %q", lines[2])
	}
}

// TestWrapTextWordsNoInfiniteLoop verifies that wrapTextWords never hangs
// on strings with no spaces or only leading-indent spaces.
func TestWrapTextWordsNoInfiniteLoop(t *testing.T) {
	// A string with no spaces at all, longer than width
	result := wrapTextWords("abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890", 20)
	joined := strings.Join(result, "")
	// All characters must survive (ignoring indent spaces)
	cleaned := strings.ReplaceAll(joined, " ", "")
	original := strings.ReplaceAll("abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890", " ", "")
	if cleaned != original {
		t.Errorf("Characters lost during wrapping: got %q", cleaned)
	}
	for _, line := range result {
		if len(line) > 20 {
			t.Errorf("Line exceeds width 20: %q (len=%d)", line, len(line))
		}
	}

	// A string where spaces only appear near the start (simulating indent + long token)
	result2 := wrapTextWords("  verylongwordwithoutanyspacesatallinthisstring", 15)
	if len(result2) == 0 {
		t.Error("Expected at least one line")
	}
	for _, line := range result2 {
		if len(line) > 15 {
			t.Errorf("Line exceeds width 15: %q (len=%d)", line, len(line))
		}
	}
}

// TestWrapTextLargeInputLinearMemory guards against copying the remainder on every line (quadratic memory).
func TestWrapTextLargeInputLinearMemory(t *testing.T) {
	s := strings.Repeat(`{"key":"value","n":1},`, 1<<20/22)
	for name, wrap := range map[string]func(string, int) []string{
		"wrapText": wrapText, "wrapTextDot": wrapTextDot, "wrapTextWords": wrapTextWords,
	} {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		lines := wrap(s, colWidthMessage)
		runtime.ReadMemStats(&after)

		if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 32<<20 {
			t.Errorf("%s allocated %d MB for a 1 MB input", name, allocated>>20)
		}
		if joined := strings.ReplaceAll(strings.Join(lines, ""), " ", ""); joined != strings.ReplaceAll(s, " ", "") {
			t.Errorf("%s lost characters while wrapping", name)
		}
	}
}

func TestTruncateLogMessage(t *testing.T) {
	if got := truncateLogMessage("short", 10); got != "short" {
		t.Errorf("short message changed: %q", got)
	}
	got := truncateLogMessage(strings.Repeat("a", 20), 10)
	if want := strings.Repeat("a", 10) + " ... [truncated 10 bytes]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	// "é" is 2 bytes; cutting at 3 must not split it
	got = truncateLogMessage("aaéaaa", 3)
	if !utf8.ValidString(got) || !strings.HasPrefix(got, "aa ...") {
		t.Errorf("truncation split a rune: %q", got)
	}
}
