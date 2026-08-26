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

// FunctionDefinition is a struct that contains the id, name, description, package, inputs and outputs of a function
type FunctionDefinition struct {
	Name             string           `json:"name" yaml:"name"`
	FlowkitUrl       string           `json:"flowkit_url" yaml:"flowkit_url"` // URL of the connected FlowKit server containing this function
	ApiKey           string           `json:"api_key" yaml:"api_key"`         // API key for the FlowKit server
	DisplayName      string           `json:"display_name" yaml:"display_name"`
	Description      string           `json:"description" yaml:"description"`
	Category         string           `json:"category" yaml:"category"` // "data_extraction", "generic", "knowledge_db", "llm_handler", "ansys_gpt"
	Type             string           `json:"type" yaml:"type"`         // "go", "python"
	Path             string           `json:"path" yaml:"path"`         // only for python functions
	Inputs           []FunctionInput  `json:"inputs" yaml:"inputs"`
	Outputs          []FunctionOutput `json:"outputs" yaml:"outputs"`
	DeprecatedParams []string         `json:"deprecated_params" yaml:"deprecated_params"` // list of deprecated parameter names
}

// FlowKitPythonFunction is a struct that contains the name, path, description, inputs, outputs and definitions of a FlowKit-Python function
type FlowKitPythonFunction struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	DisplayName string           `json:"display_name"`
	Inputs      []FunctionInput  `json:"inputs"`
	Outputs     []FunctionOutput `json:"outputs"`
	Definitions interface{}      `json:"definitions"`
}

// FunctionDefinitionShort is equivalent to FunctionDefinition but without the API key (used for snps-ai-agent rest API)
type FunctionDefinitionShort struct {
	Name             string           `json:"name" yaml:"name"`
	FlowkitUrl       string           `json:"flowkit_url" yaml:"flowkit_url"` // URL of the connected FlowKit server containing this function
	DisplayName      string           `json:"display_name" yaml:"display_name"`
	Description      string           `json:"description" yaml:"description"`
	Category         string           `json:"category" yaml:"category"` // "data_extraction", "generic", "knowledge_db", "llm_handler", "ansys_gpt"
	Type             string           `json:"type" yaml:"type"`         // "go", "python"
	Path             string           `json:"path" yaml:"path"`         // only for python functions
	Inputs           []FunctionInput  `json:"inputs" yaml:"inputs"`
	Outputs          []FunctionOutput `json:"outputs" yaml:"outputs"`
	DeprecatedParams []string         `json:"deprecated_params" yaml:"deprecated_params"` // list of deprecated parameter names
}

// FunctionInput is a struct that contains the name, type, go type and options of a function input
type FunctionInput struct {
	Name    string   `json:"name" yaml:"name"`
	Type    string   `json:"type" yaml:"type"` // string, number, boolean, json
	GoType  string   `json:"go_type" yaml:"go_type"`
	Options []string `json:"options" yaml:"options"` // only applicable if not empty
}

// FunctionOutput is a struct that contains the name, type and go type of a function output
type FunctionOutput struct {
	Name   string `json:"name" yaml:"name"`
	Type   string `json:"type" yaml:"type"` // string, number, boolean, json
	GoType string `json:"go_type" yaml:"go_type"`
}

// FilledInputOutput is a struct that contains the name, go type and value of a filled input/output
type FilledInputOutput struct {
	Name   string      `json:"name" yaml:"name"`
	GoType string      `json:"go_type" yaml:"go_type"`
	Value  interface{} `json:"value" yaml:"value"`
}
