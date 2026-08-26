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

package flowkitpythonclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ansys/snps-ai-sharedtypes/pkg/clients"
	"github.com/ansys/snps-ai-sharedtypes/pkg/clients/flowkitclient"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/snps-ai-sharedtypes/pkg/typeconverters"
)

// ListFunctionsAndSaveToInteralStates calls the FlowKit-Python API and saves the functions to internal states
// This function is used to get the list of available functions from the external function server
// and save them to internal states
//
// Returns:
//   - error: an error message if the API call fails
func ListFunctionsAndSaveToInteralStates(url string, apiKey string) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in ListFunctionsAndSaveToInteralStates: %v", r)
		}
	}()

	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorMessage := fmt.Errorf("error creating GET request: %v", err)
		return errorMessage
	}

	// Add the required header
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Create a client and make the request
	client, err := clients.GetHttpClient()
	if err != nil {
		errorMessage := fmt.Errorf("error getting HTTP client with cert: %v", err)
		return errorMessage
	}
	resp, err := client.Do(req)
	if err != nil {
		errorMessage := fmt.Errorf("error making GET request: %v", err)
		return errorMessage
	}
	defer resp.Body.Close()

	// Check if the status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		errorMessage := fmt.Errorf("error: received non-200 status code: %d", resp.StatusCode)
		return errorMessage
	}

	// Read the body of the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMessage := fmt.Errorf("error reading response body: %v", err)
		return errorMessage
	}

	// Unmarshal the JSON response into the struct
	var listResp []sharedtypes.FlowKitPythonFunction
	err = json.Unmarshal(body, &listResp)
	if err != nil {
		errorMessage := fmt.Errorf("error unmarshalling JSON response: %v", err)
		return errorMessage
	}

	// Save the functions to internal states
	for _, function := range listResp {
		// convert inputs and outputs
		inputs := []sharedtypes.FunctionInput{}
		for _, inputParam := range function.Inputs {
			// check if options is nil
			if inputParam.Options == nil {
				inputParam.Options = []string{}
			}

			// conert type to go type
			inputParam.GoType, err = typeconverters.JSONToGo(inputParam.Type)
			if err != nil {
				return err
			}

			// append the input to the list
			inputs = append(inputs, sharedtypes.FunctionInput{
				Name:    inputParam.Name,
				Type:    inputParam.Type,
				GoType:  inputParam.GoType,
				Options: inputParam.Options,
			})
		}
		outputs := []sharedtypes.FunctionOutput{}
		for _, outputParam := range function.Outputs {
			// convert type to go type
			outputParam.GoType, err = typeconverters.JSONToGo(outputParam.Type)
			if err != nil {
				return err
			}

			// append the output to the list
			outputs = append(outputs, sharedtypes.FunctionOutput{
				Name:   outputParam.Name,
				Type:   outputParam.Type,
				GoType: outputParam.GoType,
			})
		}

		// Save the function to internal states
		flowkitclient.AvailableFunctions[function.Name] = &sharedtypes.FunctionDefinition{
			Name:        function.Name,
			FlowkitUrl:  url,
			ApiKey:      apiKey,
			Description: function.Description,
			DisplayName: function.DisplayName,
			Category:    function.Category,
			Inputs:      inputs,
			Outputs:     outputs,
			Type:        "python",
			Path:        function.Path,
		}
		// add the category to available categories
		if flowkitclient.AvailableCategories != nil && function.Category != "" {
			flowkitclient.AvailableCategories[function.Category] = true
		}
	}

	return nil
}

// RunFunction calls the external function server and returns the outputs
// This function is used to run an external function
//
// Parameters:
//   - functionPath: the path of the function to run
//   - inputs: the inputs to the function
//   - outputDefinition: the definition of the outputs
//
// Returns:
//   - map[string]sharedtypes.FilledInputOutput: the outputs of the function
//   - error: an error message if the API call fails
func RunFunction(functionName string, inputs map[string]sharedtypes.FilledInputOutput) (outputs map[string]sharedtypes.FilledInputOutput, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in RunFunction: %v", r)
		}
	}()

	// Get function definition
	functionDefinition := flowkitclient.AvailableFunctions[functionName]

	// Create input dict
	inputDict := map[string]interface{}{}
	for _, value := range inputs {
		inputDict[value.Name] = value.Value
	}

	// Create outputs go type map
	outputGoTypes := map[string]string{}
	for _, output := range functionDefinition.Outputs {
		outputGoTypes[output.Name] = output.GoType
	}

	// Add inputDict as the request body
	reqBody, err := json.Marshal(inputDict)
	if err != nil {
		errorMessage := fmt.Errorf("error marshalling inputDict: %v", err)
		return nil, errorMessage
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", functionDefinition.FlowkitUrl+functionDefinition.Path, bytes.NewBuffer(reqBody))
	if err != nil {
		errorMessage := fmt.Errorf("error creating POST request: %v", err)
		return nil, errorMessage
	}

	// Add the required header
	req.Header.Set("api-key", functionDefinition.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Create a client and make the request
	client, err := clients.GetHttpClient()
	if err != nil {
		errorMessage := fmt.Errorf("error getting HTTP client with cert: %v", err)
		return nil, errorMessage
	}
	resp, err := client.Do(req)
	if err != nil {
		errorMessage := fmt.Errorf("error making POST request: %v", err)
		return nil, errorMessage
	}
	defer resp.Body.Close()

	// Check if the status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		// Read the body of the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errorMessage := fmt.Errorf("error reading response body: %v", err)
			return nil, errorMessage
		}
		errorMessage := fmt.Errorf("error: received non-200 status code: %d, error message: %v", resp.StatusCode, string(body))
		return nil, errorMessage
	}

	// Read the body of the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMessage := fmt.Errorf("error reading response body: %v", err)
		return nil, errorMessage
	}

	// Unmarshal the JSON response into the struct
	outputDict := map[string]interface{}{}
	err = json.Unmarshal(body, &outputDict)
	if err != nil {
		errorMessage := fmt.Errorf("error unmarshalling JSON response: %v", err)
		return nil, errorMessage
	}

	// Create the outputs map
	outputs = map[string]sharedtypes.FilledInputOutput{}
	for key, value := range outputDict {
		outputs[key] = sharedtypes.FilledInputOutput{
			Name:   key,
			Value:  value,
			GoType: outputGoTypes[key],
		}
	}

	return outputs, nil
}
