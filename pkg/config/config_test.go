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

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// TestDefineOptionalProperties tests the defineOptionalProperties function
func TestDefineOptionalProperties(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name                  string
		initialConfig         Config
		optionalDefaultValues map[string]interface{}
		expectedConfig        Config
		expectError           bool
	}{
		{
			name: "Set some optional defaults",
			initialConfig: Config{
				LOG_LEVEL:                  "",
				LOCAL_LOGS:                 false,
				NUMBER_OF_WORKFLOW_WORKERS: 0,
				WEBSERVER_PORT:             "",
				SERVICE_NAME:               "",
			},
			optionalDefaultValues: map[string]interface{}{
				"LOG_LEVEL":                  "info",
				"LOCAL_LOGS":                 true,
				"NUMBER_OF_WORKFLOW_WORKERS": 5,
				"WEBSERVER_PORT":             "8080",
				"SERVICE_NAME":               "SnpsAiService",
			},
			expectedConfig: Config{
				LOG_LEVEL:                  "info",
				LOCAL_LOGS:                 true,
				NUMBER_OF_WORKFLOW_WORKERS: 5,
				WEBSERVER_PORT:             "8080",
				SERVICE_NAME:               "SnpsAiService",
			},
			expectError: false,
		},
		{
			name: "Partial defaults applied",
			initialConfig: Config{
				LOG_LEVEL:                  "debug",
				LOCAL_LOGS:                 false,
				NUMBER_OF_WORKFLOW_WORKERS: 0,
				WEBSERVER_PORT:             "",
				SERVICE_NAME:               "ExistingService",
			},
			optionalDefaultValues: map[string]interface{}{
				"LOG_LEVEL":                  "info",
				"LOCAL_LOGS":                 true,
				"NUMBER_OF_WORKFLOW_WORKERS": 10,
				"WEBSERVER_PORT":             "9090",
				"SERVICE_NAME":               "SnpsAiService",
			},
			expectedConfig: Config{
				LOG_LEVEL:                  "debug",
				LOCAL_LOGS:                 true,
				NUMBER_OF_WORKFLOW_WORKERS: 10,
				WEBSERVER_PORT:             "9090",
				SERVICE_NAME:               "ExistingService",
			},
			expectError: false,
		},
		{
			name: "No changes needed",
			initialConfig: Config{
				LOG_LEVEL:                  "warn",
				LOCAL_LOGS:                 true,
				NUMBER_OF_WORKFLOW_WORKERS: 8,
				WEBSERVER_PORT:             "8081",
				SERVICE_NAME:               "CustomService",
			},
			optionalDefaultValues: map[string]interface{}{
				"LOG_LEVEL":                  "info",
				"LOCAL_LOGS":                 false,
				"NUMBER_OF_WORKFLOW_WORKERS": 5,
				"WEBSERVER_PORT":             "8080",
				"SERVICE_NAME":               "SnpsAiService",
			},
			expectedConfig: Config{
				LOG_LEVEL:                  "warn",
				LOCAL_LOGS:                 true,
				NUMBER_OF_WORKFLOW_WORKERS: 8,
				WEBSERVER_PORT:             "8081",
				SERVICE_NAME:               "CustomService",
			},
			expectError: false,
		},
		{
			name: "Type mismatch error",
			initialConfig: Config{
				LOG_LEVEL: "",
			},
			optionalDefaultValues: map[string]interface{}{
				"LOG_LEVEL": 123, // Wrong type - should be string
			},
			expectedConfig: Config{},
			expectError:    true,
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the initial config for each test case
			config := tt.initialConfig

			// Call the function under test
			err := defineOptionalProperties(&config, tt.optionalDefaultValues)

			// Check for expected error
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Compare the result with the expected config
			if !reflect.DeepEqual(config, tt.expectedConfig) {
				t.Errorf("Expected config %+v, got %+v", tt.expectedConfig, config)
			}
		})
	}
}

// TestIsZeroValue tests the isZeroValue function
func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "Empty string",
			value:    "",
			expected: true,
		},
		{
			name:     "Non-empty string",
			value:    "test",
			expected: false,
		},
		{
			name:     "Zero int",
			value:    0,
			expected: true,
		},
		{
			name:     "Non-zero int",
			value:    42,
			expected: false,
		},
		{
			name:     "False bool",
			value:    false,
			expected: true,
		},
		{
			name:     "True bool",
			value:    true,
			expected: false,
		},
		{
			name:     "Nil slice",
			value:    []string(nil),
			expected: true,
		},
		{
			name:     "Empty slice",
			value:    []string{},
			expected: false,
		},
		{
			name:     "Non-empty slice",
			value:    []string{"a", "b"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			result := isZeroValue(v)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for value %v", tt.expected, result, tt.value)
			}
		})
	}
}

// TestValidateConfig tests the ValidateConfig function
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name               string
		config             Config
		requiredProperties []string
		expectError        bool
		errorContains      string
	}{
		{
			name: "All required properties present",
			config: Config{
				LOG_LEVEL:    "info",
				SERVICE_NAME: "TestService",
			},
			requiredProperties: []string{"LOG_LEVEL", "SERVICE_NAME"},
			expectError:        false,
		},
		{
			name: "Missing required property",
			config: Config{
				LOG_LEVEL: "info",
			},
			requiredProperties: []string{"LOG_LEVEL", "SERVICE_NAME"},
			expectError:        true,
			errorContains:      "SERVICE_NAME",
		},
		{
			name: "Zero value for required property",
			config: Config{
				LOG_LEVEL:    "",
				SERVICE_NAME: "TestService",
			},
			requiredProperties: []string{"LOG_LEVEL", "SERVICE_NAME"},
			expectError:        true,
			errorContains:      "LOG_LEVEL",
		},
		{
			name: "No required properties",
			config: Config{
				LOG_LEVEL: "",
			},
			requiredProperties: []string{},
			expectError:        false,
		},
		{
			name: "Invalid property name",
			config: Config{
				LOG_LEVEL: "info",
			},
			requiredProperties: []string{"INVALID_PROPERTY"},
			expectError:        true,
			errorContains:      "INVALID_PROPERTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config, tt.requiredProperties)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestGetGlobalConfigAsJSON tests the GetGlobalConfigAsJSON function
func TestGetGlobalConfigAsJSON(t *testing.T) {
	// Save original GlobalConfig and restore after test
	originalConfig := GlobalConfig
	defer func() { GlobalConfig = originalConfig }()

	tests := []struct {
		name           string
		config         *Config
		expectedFields map[string]interface{}
	}{
		{
			name: "Valid config to JSON",
			config: &Config{
				LOG_LEVEL:    "info",
				SERVICE_NAME: "TestService",
				LOCAL_LOGS:   true,
			},
			expectedFields: map[string]interface{}{
				"LOGLEVEL":    "info",
				"SERVICENAME": "TestService",
				"LOCALLOGS":   true,
			},
		},
		{
			name:           "Empty config to JSON",
			config:         &Config{},
			expectedFields: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GlobalConfig = tt.config
			jsonStr := GetGlobalConfigAsJSON()

			if jsonStr == "" {
				t.Errorf("Expected non-empty JSON string")
				return
			}

			// Parse JSON to verify structure
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(jsonStr), &parsed)
			if err != nil {
				t.Errorf("Failed to parse JSON: %v", err)
				return
			}

			// Verify expected fields
			for key, expectedValue := range tt.expectedFields {
				actualValue, exists := parsed[key]
				if !exists {
					t.Errorf("Expected field '%s' not found in JSON", key)
					continue
				}
				if !reflect.DeepEqual(actualValue, expectedValue) {
					t.Errorf("Field '%s': expected %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestTimeToString tests the timeToString function
func TestTimeToString(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "Specific date",
			time:     time.Date(2025, 12, 19, 15, 30, 45, 123456789, time.UTC),
			expected: "2025-12-19 15:30:45.123",
		},
		{
			name:     "Start of year",
			time:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2025-01-01 00:00:00.000",
		},
		{
			name:     "End of year",
			time:     time.Date(2025, 12, 31, 23, 59, 59, 999999999, time.UTC),
			expected: "2025-12-31 23:59:59.999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeToString(tt.time)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestHandleLegacyPortDefinition tests the HandleLegacyPortDefinition function
func TestHandleLegacyPortDefinition(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		legacyPort    string
		expected      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Address provided",
			address:     "0.0.0.0:8080",
			legacyPort:  "9090",
			expected:    "0.0.0.0:8080",
			expectError: false,
		},
		{
			name:        "Only legacy port provided",
			address:     "",
			legacyPort:  "9090",
			expected:    "0.0.0.0:9090",
			expectError: false,
		},
		{
			name:          "Neither provided",
			address:       "",
			legacyPort:    "",
			expected:      "",
			expectError:   true,
			errorContains: "both address and legacy port are empty",
		},
		{
			name:        "Custom address format",
			address:     "127.0.0.1:3000",
			legacyPort:  "",
			expected:    "127.0.0.1:3000",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HandleLegacyPortDefinition(tt.address, tt.legacyPort)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

// TestReadYaml tests the readYaml function
func TestReadYaml(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		fileContent   string
		fileName      string
		expectError   bool
		errorContains string
		expectedField string
		expectedValue interface{}
	}{
		{
			name: "Valid YAML file",
			fileContent: `LOG_LEVEL: "info"
SERVICE_NAME: "TestService"
LOCAL_LOGS: true
NUMBER_OF_WORKFLOW_WORKERS: 5`,
			fileName:      "valid_config.yaml",
			expectError:   false,
			expectedField: "LOG_LEVEL",
			expectedValue: "info",
		},
		{
			name:          "Missing file",
			fileContent:   "",
			fileName:      "nonexistent.yaml",
			expectError:   true,
			errorContains: "missing from directory",
		},
		{
			name:          "Invalid YAML content",
			fileContent:   "invalid: yaml: content: [unclosed",
			fileName:      "invalid.yaml",
			expectError:   true,
			errorContains: "incorrect content",
		},
		{
			name:          "Empty file",
			fileContent:   "",
			fileName:      "empty.yaml",
			expectError:   false,
			expectedField: "LOG_LEVEL",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.fileName)

			// Create file only if content is provided and not testing missing file
			if tt.fileContent != "" || !contains(tt.name, "Missing") {
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
				if err != nil && !tt.expectError {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			config := Config{}
			result, err := readYaml(filePath, config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify expected field value
				if tt.expectedField != "" {
					resultValue := reflect.ValueOf(result)
					fieldValue := resultValue.FieldByName(tt.expectedField)
					if fieldValue.IsValid() {
						actualValue := fieldValue.Interface()
						if !reflect.DeepEqual(actualValue, tt.expectedValue) {
							t.Errorf("Field '%s': expected %v, got %v", tt.expectedField, tt.expectedValue, actualValue)
						}
					}
				}
			}
		})
	}
}

// TestInitGlobalConfigFromFile tests the InitGlobalConfigFromFile function
func TestInitGlobalConfigFromFile(t *testing.T) {
	// Save original GlobalConfig and restore after test
	originalConfig := GlobalConfig
	defer func() { GlobalConfig = originalConfig }()

	tempDir := t.TempDir()

	tests := []struct {
		name                  string
		fileContent           string
		fileName              string
		requiredProperties    []string
		optionalDefaultValues map[string]interface{}
		expectError           bool
	}{
		{
			name: "Valid config file with defaults",
			fileContent: `LOG_LEVEL: "debug"
SERVICE_NAME: "TestService"`,
			fileName:           "config.yaml",
			requiredProperties: []string{"LOG_LEVEL"},
			optionalDefaultValues: map[string]interface{}{
				"LOCAL_LOGS": true,
			},
			expectError: false,
		},
		{
			name:                  "Missing config file",
			fileContent:           "",
			fileName:              "missing.yaml",
			requiredProperties:    []string{},
			optionalDefaultValues: map[string]interface{}{},
			expectError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.fileName)

			if tt.fileContent != "" {
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			err := InitGlobalConfigFromFile(filePath, tt.requiredProperties, tt.optionalDefaultValues)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if GlobalConfig == nil {
					t.Errorf("GlobalConfig should not be nil")
				}
			}
		})
	}
}

// TestWriteStringToFile tests the writeStringToFile function
func TestWriteStringToFile(t *testing.T) {
	// Save current directory and change to temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "Write simple string",
			data:     "Test error message",
			expected: "Test error message",
		},
		{
			name:     "Write empty string",
			data:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing error.log
			os.Remove("error.log")

			err := writeStringToFile(tt.data)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Read the file and verify content
			content, err := os.ReadFile("error.log")
			if err != nil {
				t.Errorf("Failed to read error.log: %v", err)
				return
			}

			if !contains(string(content), tt.expected) {
				t.Errorf("Expected content to contain '%s', got: %s", tt.expected, string(content))
			}

			// Clean up
			os.Remove("error.log")
		})
	}
}

// TestWriteInterfaceToFile tests the writeInterfaceToFile function
func TestWriteInterfaceToFile(t *testing.T) {
	// Save current directory and change to temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name        string
		data        interface{}
		expectError bool
	}{
		{
			name:        "Write map",
			data:        map[string]string{"key": "value"},
			expectError: false,
		},
		{
			name:        "Write struct",
			data:        Config{LOG_LEVEL: "info"},
			expectError: false,
		},
		{
			name:        "Write string",
			data:        "simple string",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing error.log
			os.Remove("error.log")

			err := writeInterfaceToFile(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file exists
				_, err := os.Stat("error.log")
				if err != nil {
					t.Errorf("error.log was not created: %v", err)
				}
			}

			// Clean up
			os.Remove("error.log")
		})
	}
}

// TestApplySecretValueToConfig tests the applySecretValueToConfig function
// which is the core logic of InitGlobalConfigFromAzureKeyVault
func TestApplySecretValueToConfig(t *testing.T) {
	tests := []struct {
		name          string
		secretName    string
		secretValue   string
		setupConfig   func() *Config
		validate      func(t *testing.T, config *Config)
		expectError   bool
		errorContains string
	}{
		// String type
		{
			name:        "Set string field",
			secretName:  "LOGLEVEL",
			secretValue: "debug",
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				if config.LOG_LEVEL != "debug" {
					t.Errorf("Expected LOG_LEVEL 'debug', got '%v'", config.LOG_LEVEL)
				}
			},
		},
		{
			name:        "Set string field with escape sequences",
			secretName:  "LOGLEVEL",
			secretValue: `line1\nline2`,
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				if config.LOG_LEVEL != "line1\nline2" {
					t.Errorf("Expected LOG_LEVEL with newline, got '%v'", config.LOG_LEVEL)
				}
			},
		},
		// Bool type
		{
			name:        "Set bool field true",
			secretName:  "LOCALLOGS",
			secretValue: "true",
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				if !config.LOCAL_LOGS {
					t.Errorf("Expected LOCAL_LOGS true, got false")
				}
			},
		},
		{
			name:        "Set bool field empty string",
			secretName:  "LOCALLOGS",
			secretValue: "",
			setupConfig: func() *Config { return &Config{LOCAL_LOGS: true} },
			validate: func(t *testing.T, config *Config) {
				if config.LOCAL_LOGS {
					t.Errorf("Expected LOCAL_LOGS false, got true")
				}
			},
		},
		{
			name:          "Set bool field invalid value",
			secretName:    "LOCALLOGS",
			secretValue:   "notabool",
			setupConfig:   func() *Config { return &Config{} },
			expectError:   true,
			errorContains: "strconv.ParseBool",
		},
		// Int type
		{
			name:        "Set int field",
			secretName:  "NUMBEROFWORKFLOWWORKERS",
			secretValue: "42",
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				if config.NUMBER_OF_WORKFLOW_WORKERS != 42 {
					t.Errorf("Expected NUMBER_OF_WORKFLOW_WORKERS 42, got %v", config.NUMBER_OF_WORKFLOW_WORKERS)
				}
			},
		},
		{
			name:        "Set int field empty string",
			secretName:  "NUMBEROFWORKFLOWWORKERS",
			secretValue: "",
			setupConfig: func() *Config { return &Config{NUMBER_OF_WORKFLOW_WORKERS: 10} },
			validate: func(t *testing.T, config *Config) {
				if config.NUMBER_OF_WORKFLOW_WORKERS != 0 {
					t.Errorf("Expected NUMBER_OF_WORKFLOW_WORKERS 0, got %v", config.NUMBER_OF_WORKFLOW_WORKERS)
				}
			},
		},
		{
			name:          "Set int field invalid value",
			secretName:    "NUMBEROFWORKFLOWWORKERS",
			secretValue:   "notanint",
			setupConfig:   func() *Config { return &Config{} },
			expectError:   true,
			errorContains: "strconv.Atoi",
		},
		// []string type
		{
			name:        "Set string slice field",
			secretName:  "PRIVATEWORKFLOWSFOLDERS",
			secretValue: `["folder1","folder2","folder3"]`,
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				expected := []string{"folder1", "folder2", "folder3"}
				if !reflect.DeepEqual(config.PRIVATE_WORKFLOWS_FOLDERS, expected) {
					t.Errorf("Expected PRIVATE_WORKFLOWS_FOLDERS %v, got %v", expected, config.PRIVATE_WORKFLOWS_FOLDERS)
				}
			},
		},
		{
			name:        "Set string slice field empty string",
			secretName:  "PRIVATEWORKFLOWSFOLDERS",
			secretValue: "",
			setupConfig: func() *Config { return &Config{PRIVATE_WORKFLOWS_FOLDERS: []string{"old"}} },
			validate: func(t *testing.T, config *Config) {
				if len(config.PRIVATE_WORKFLOWS_FOLDERS) != 0 {
					t.Errorf("Expected empty slice, got %v", config.PRIVATE_WORKFLOWS_FOLDERS)
				}
			},
		},
		{
			name:          "Set string slice field invalid JSON",
			secretName:    "PRIVATEWORKFLOWSFOLDERS",
			secretValue:   "not json",
			setupConfig:   func() *Config { return &Config{} },
			expectError:   true,
			errorContains: "json.Unmarshal []string",
		},
		// []int type
		{
			name:        "Set int slice field",
			secretName:  "ANSYSAUTHORIZATIONACCEPTEDPERSONAS",
			secretValue: `[1,2,3]`,
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				expected := []int{1, 2, 3}
				if !reflect.DeepEqual(config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS, expected) {
					t.Errorf("Expected ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS %v, got %v", expected, config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS)
				}
			},
		},
		{
			name:        "Set int slice field single element",
			secretName:  "ANSYSAUTHORIZATIONACCEPTEDPERSONAS",
			secretValue: `[42]`,
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				expected := []int{42}
				if !reflect.DeepEqual(config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS, expected) {
					t.Errorf("Expected ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS %v, got %v", expected, config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS)
				}
			},
		},
		{
			name:        "Set int slice field empty string",
			secretName:  "ANSYSAUTHORIZATIONACCEPTEDPERSONAS",
			secretValue: "",
			setupConfig: func() *Config { return &Config{ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS: []int{1, 2}} },
			validate: func(t *testing.T, config *Config) {
				if len(config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS) != 0 {
					t.Errorf("Expected empty slice, got %v", config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS)
				}
			},
		},
		{
			name:        "Set int slice field empty array",
			secretName:  "ANSYSAUTHORIZATIONACCEPTEDPERSONAS",
			secretValue: `[]`,
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				if len(config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS) != 0 {
					t.Errorf("Expected empty slice, got %v", config.ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS)
				}
			},
		},
		{
			name:          "Set int slice field invalid JSON",
			secretName:    "ANSYSAUTHORIZATIONACCEPTEDPERSONAS",
			secretValue:   `["not","ints"]`,
			setupConfig:   func() *Config { return &Config{} },
			expectError:   true,
			errorContains: "json.Unmarshal []int",
		},
		// map[string]string type
		{
			name:        "Set map field",
			secretName:  "WORKFLOWCONFIGVARIABLES",
			secretValue: `{"key1":"val1","key2":"val2"}`,
			setupConfig: func() *Config { return &Config{} },
			validate: func(t *testing.T, config *Config) {
				expected := map[string]string{"key1": "val1", "key2": "val2"}
				if !reflect.DeepEqual(config.WORKFLOW_CONFIG_VARIABLES, expected) {
					t.Errorf("Expected WORKFLOW_CONFIG_VARIABLES %v, got %v", expected, config.WORKFLOW_CONFIG_VARIABLES)
				}
			},
		},
		{
			name:        "Set map field empty string",
			secretName:  "WORKFLOWCONFIGVARIABLES",
			secretValue: "",
			setupConfig: func() *Config { return &Config{WORKFLOW_CONFIG_VARIABLES: map[string]string{"old": "val"}} },
			validate: func(t *testing.T, config *Config) {
				if len(config.WORKFLOW_CONFIG_VARIABLES) != 0 {
					t.Errorf("Expected empty map, got %v", config.WORKFLOW_CONFIG_VARIABLES)
				}
			},
		},
		// Non-matching secret name
		{
			name:        "Non-matching secret name leaves config unchanged",
			secretName:  "NONEXISTENTSECRET",
			secretValue: "somevalue",
			setupConfig: func() *Config { return &Config{LOG_LEVEL: "info"} },
			validate: func(t *testing.T, config *Config) {
				if config.LOG_LEVEL != "info" {
					t.Errorf("Expected LOG_LEVEL 'info' unchanged, got '%v'", config.LOG_LEVEL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()
			err := applySecretValueToConfig(config, tt.secretName, tt.secretValue)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
