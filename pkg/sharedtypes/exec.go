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
	"time"
)

// ExecRequest represents the requests that can be sent to snps-ai-exec
type ExecRequest struct {
	Type                 string                       `json:"type"`   // "code", "flowkit"
	Action               string                       `json:"action"` // type "code":"execute", "append", "cancel", "status"; for type "flowkit": "<functionName>"
	InstructionGuid      string                       `json:"instructionGuid"`
	ExecutionInstruction *ExecutionInstruction        `json:"executionInstruction"` // only for type "code"
	Inputs               map[string]FilledInputOutput `json:"inputs"`               // only for type "flowkit"
}

// ExecutionInstruction contain an array of strings that represent the code to be executed in snps-ai-exec
type ExecutionInstruction struct {
	CodeType       string   `json:"codeType"` // "python", "bash"
	Code           []string `json:"code"`
	VenvExecutable string   `json:"venvExecutable"`
}

// ExecResponse represents the response that snps-ai-exec sends back
type ExecResponse struct {
	Type             string                       `json:"type"` // "status", "flowkit", "file", "error"
	InstructionGuid  string                       `json:"instructionGuid"`
	Error            *ErrorResponse               `json:"error,omitempty"`
	ExecutionDetails *ExecutionDetails            `json:"executionDetails,omitempty"`
	FileDetails      *FileDetails                 `json:"fileDetails,omitempty"`
	Outputs          map[string]FilledInputOutput `json:"outputs"` // only for type "flowkit"
}

// ExecutionDetails represents the details of the execution
type ExecutionDetails struct {
	InstructionGuid     string        `json:"instructionGuid"`
	ClientGuid          string        `json:"clientGuid"`
	StartTime           time.Time     `json:"startTime"`
	TimeoutAt           time.Time     `json:"timeoutAt"`
	Response            string        `json:"response"`
	Status              string        `json:"status"` // "started", "running", "completed", "failed"
	LastResponseDiff    string        `json:"lastResponseDiff"`
	InterruptionChannel chan bool     `json:"-"`
	StdinChannel        chan []string `json:"-"`
	Cancelled           bool          `json:"-"`
}

// FileDetails contain parts of a file that is being sent
type FileDetails struct {
	FileName        string `json:"fileName"`
	FileSize        int64  `json:"fileSize"`
	FileChunkNumber int    `json:"fileChunkNumber"`
	FileChunk       []byte `json:"fileChunk"`
	IsLastChunk     bool   `json:"isLastChunk"`
}
