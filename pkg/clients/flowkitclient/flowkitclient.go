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

package flowkitclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ansys/snps-ai-sharedtypes/pkg/clients"
	"github.com/ansys/snps-ai-sharedtypes/pkg/flowkitgrpc"
	"github.com/ansys/snps-ai-sharedtypes/pkg/logging"
	"github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/snps-ai-sharedtypes/pkg/typeconverters"
	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// HealthCheck checks the health of the external function server
// This function is used to check if the external function server is running and reachable
//
// Parameters:
//   - url: the URL of the external function server
//   - apiKey: the API key to authenticate with the external function server
//
// Returns:
//   - err: an error message if the gRPC call fails
func HealthCheck(url string, apiKey string) (err error) {
	// Set up a connection to the server.
	c, conn, err := createClient(url, apiKey)
	if err != nil {
		return fmt.Errorf("unable to connect to external function gRPC: %v", err)
	}
	defer conn.Close()

	// Create a context with a cancel
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call HealthCheck
	_, err = c.HealthCheck(ctxWithCancel, &flowkitgrpc.HealthRequest{})
	if err != nil {
		return fmt.Errorf("error in external function gRPC HealthCheck: %v", err)
	}

	return nil
}

// GetVersion retrieves the version of the external function server
// This function is used to get the version of the external function server
//
// Parameters:
//   - url: the URL of the external function server
//   - apiKey: the API key to authenticate with the external function server
//
// Returns:
//   - version: the version of the external function server
//   - err: an error message if the gRPC call fails
func GetVersion(url string, apiKey string) (version string, err error) {
	// Set up a connection to the server.
	c, conn, err := createClient(url, apiKey)
	if err != nil {
		return "", fmt.Errorf("unable to connect to external function gRPC: %v", err)
	}
	defer conn.Close()

	// Create a context with a cancel
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call GetVersion
	resp, err := c.GetVersion(ctxWithCancel, &flowkitgrpc.VersionRequest{})
	if err != nil {
		return "", fmt.Errorf("error in external function gRPC GetVersion: %v", err)
	}

	return resp.Version, nil
}

// Global variable to store the available functions, types and categories
var AvailableFunctions map[string]*sharedtypes.FunctionDefinition
var AvailableTypes map[string]bool
var AvailableCategories map[string]bool

// ListFunctionsAndSaveToInteralStates calls the ListFunctions gRPC and saves the functions to internal states
// This function is used to get the list of available functions from the external function server
// and save them to internal states
//
// Parameters:
//   - url: the URL of the external function server
//   - apiKey: the API key to authenticate with the external function server
//
// Returns:
//   - error: an error message if the gRPC call fails
func ListFunctionsAndSaveToInteralStates(url string, apiKey string) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occurred in ListFunctionsAndSaveToInteralStates: %v", r)
		}
	}()

	// Set up a connection to the server.
	c, conn, err := createClient(url, apiKey)
	if err != nil {
		return fmt.Errorf("unable to connect to external function gRPC: %v", err)
	}
	defer conn.Close()

	// Create a context with a cancel
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call ListFunctions
	listResp, err := c.ListFunctions(ctxWithCancel, &flowkitgrpc.ListFunctionsRequest{})
	if err != nil {
		return fmt.Errorf("error in external function gRPC ListFunctions: %v", err)
	}

	// Save the functions to internal states
	for _, function := range listResp.Functions {
		// convert inputs and outputs
		inputs := []sharedtypes.FunctionInput{}
		for _, inputParam := range function.Input {
			// check if options is nil
			if inputParam.Options == nil {
				inputParam.Options = []string{}
			}
			inputs = append(inputs, sharedtypes.FunctionInput{
				Name:    inputParam.Name,
				Type:    inputParam.Type,
				GoType:  inputParam.GoType,
				Options: inputParam.Options,
			})
		}
		outputs := []sharedtypes.FunctionOutput{}
		for _, outputParam := range function.Output {
			outputs = append(outputs, sharedtypes.FunctionOutput{
				Name:   outputParam.Name,
				Type:   outputParam.Type,
				GoType: outputParam.GoType,
			})
		}

		// Save the function to internal states
		AvailableFunctions[function.Name] = &sharedtypes.FunctionDefinition{
			Name:             function.Name,
			FlowkitUrl:       url,
			ApiKey:           apiKey,
			DisplayName:      function.DisplayName,
			Description:      function.Description,
			Category:         function.Category,
			DeprecatedParams: function.DeprecatedParams,
			Inputs:           inputs,
			Outputs:          outputs,
			Type:             "go",
		}
		// add the category to available categories
		if AvailableCategories != nil && function.Category != "" {
			AvailableCategories[function.Category] = true
		}
	}

	// Save the available types to internal states
	AvailableTypes = make(map[string]bool)
	for _, goType := range typeconverters.GetSupportedTypes() {
		AvailableTypes[goType] = true
	}

	return nil
}

// ResponseType is a type that represents the type of response from flowkit
type ResponseType string

const (
	Response         ResponseType = "response"
	DisableInterrupt ResponseType = "disable_interrupt"
	EnableInterrupt  ResponseType = "enable_interrupt"
	InfoMessage      ResponseType = "info_message"
	StatusMessage    ResponseType = "status_message"
	GetApproval      ResponseType = "get_approval"
)

// FunctionType is a type that represents the type of function to run
type FunctionType string

const (
	Run    FunctionType = "run"
	Stream FunctionType = "stream"
)

// OpenApprovals is a global variable that keeps track of the open approvals for each function
type OpenApproval struct {
	FunctionType FunctionType
	Channel      *chan string
}

var OpenApprovals = map[string]OpenApproval{}
var OpenApprovalsLock = sync.RWMutex{}

// ApprovalRequest is a struct that represents the approval request from flowkit
type ApprovalRequest struct {
	InstructionId   string   `json:"instruction_id"`
	ApprovalText    string   `json:"approval_text"`
	ApprovalOptions []string `json:"approval_options"`
}

// ApprovalResponse is a struct that represents the approval response from the client
type ApprovalResponse struct {
	InstructionId    string `json:"instruction_id"`
	ApprovalResponse string `json:"approval_response"`
}

// RunFunction calls the RunFunction gRPC and returns the outputs
// This function is used to run an external function
//
// Parameters:
//   - functionName: the name of the function to run
//   - inputs: the inputs to the function
//   - responseChannel: a channel to send messages to the client
//
// Returns:
//   - map[string]sharedtypes.FilledInputOutput: the outputs of the function
//   - error: an error message if the gRPC call fails
func RunFunction(ctx *logging.ContextMap, functionName string, inputs map[string]sharedtypes.FilledInputOutput, responseChannel chan sharedtypes.ClientResponse, conversationHistory *[]sharedtypes.ConversationHistoryMessage, workflowRunLock *sync.RWMutex) (outputs map[string]sharedtypes.FilledInputOutput, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occurred in RunFunction of function '%v': %v", functionName, r)
		}
	}()

	// Get function definition
	functionDef, ok := AvailableFunctions[functionName]
	if !ok {
		return nil, fmt.Errorf("function '%s' not found in available functions", functionName)
	}

	// Set up a connection to the server.
	c, conn, err := createClient(functionDef.FlowkitUrl, functionDef.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to external function gRPC: %v", err)
	}
	defer conn.Close()

	// Create a context with a cancel
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get logging metadata from context
	ctxWithMetadata, err := logging.CreateMetaDataFromCtx(ctx, ctxWithCancel)
	if err != nil {
		return nil, fmt.Errorf("error adding metadata: %v", err)
	}

	// Convert inputs to gRPC format based on order from function definition
	grpcInputs := []*flowkitgrpc.FunctionInput{}
	for _, inputDef := range functionDef.Inputs {
		// create grpc input
		grpcInput := &flowkitgrpc.FunctionInput{
			Name:   inputDef.Name,
			GoType: inputDef.GoType,
		}

		// Get the input value
		value, ok := inputs[inputDef.Name]
		if ok {
			// found: convert value to string
			stringValue, exists, err := typeconverters.ConvertGivenTypeToString(value.Value, inputDef.GoType)
			if err != nil {
				return nil, fmt.Errorf("error converting input '%s' for function '%v' to string: %v", inputDef.Name, functionName, err)
			}
			if !exists {
				return nil, fmt.Errorf("type '%s' does not exist in typeconverters.ConvertGivenTypeToString", inputDef.Name)
			}
			grpcInput.Value = stringValue

		} else {
			// input discrepancy, set to null value
			grpcInput.Value = ""
		}

		// Append the grpc input to the list
		grpcInputs = append(grpcInputs, grpcInput)
	}

	// create a channel to send messages to the server
	responseChannelToServer := make(chan *flowkitgrpc.FunctionInputs)
	defer close(responseChannelToServer)

	// track pending approval instruction IDs for cleanup on error
	pendingApprovalIDs := []string{}

	// Call RunFunction (bidirectional streaming)
	stream, err := c.RunFunction(ctxWithMetadata)
	if err != nil {
		return nil, fmt.Errorf("error in external function gRPC RunFunction for function '%v': %v", functionName, err)
	}

	// Send the initial message with function inputs
	err = stream.Send(&flowkitgrpc.FunctionInputs{
		Name:   functionName,
		Inputs: grpcInputs,
	})
	if err != nil {
		return nil, fmt.Errorf("error sending initial message in RunFunction for function '%v': %v", functionName, err)
	}

	// Launch goroutine to forward messages from responseChannelToServer to the server
	go func() {
		for msg := range responseChannelToServer {
			if err := stream.Send(msg); err != nil {
				logging.Log.Errorf(ctx, "error sending message to server in RunFunction for function '%v': %v", functionName, err)
				return
			}
		}
	}()

	// cleanup function to remove pending approvals on error
	cleanupApprovals := func() {
		OpenApprovalsLock.Lock()
		defer OpenApprovalsLock.Unlock()
		for _, id := range pendingApprovalIDs {
			if openApproval, ok := OpenApprovals[id]; ok {
				close(*openApproval.Channel)
				delete(OpenApprovals, id)
			}
		}
	}

	// Receive the stream from the server
	var runResp *flowkitgrpc.FunctionOutputs
out:
	for {
		res, err := stream.Recv()
		if err != nil && err != io.EOF {
			err := fmt.Errorf("error receiving stream for RunFunction '%v': %v", functionName, err)
			logging.Log.Error(ctx, err)
			cleanupApprovals()
			return nil, err
		}
		if err == io.EOF {
			err := fmt.Errorf("received EOF before the final response for RunFunction '%v'", functionName)
			logging.Log.Error(ctx, err)
			cleanupApprovals()
			return nil, err
		}

		// Handle the response based on its type
		switch ResponseType(res.Type) {
		case Response:
			// default case: asign runResp and break the loop
			runResp = res
			break out
		case DisableInterrupt, EnableInterrupt:
			// send a message to the response channel to disable interrupts
			responseChannel <- sharedtypes.ClientResponse{
				Type: string(res.Type),
			}
		case InfoMessage, StatusMessage:
			// send a message to the response channel with the info or status message
			instructionId := strings.ReplaceAll(uuid.New().String(), "-", "")
			responseChannel <- sharedtypes.ClientResponse{
				InstructionId: instructionId,
				Type:          string(res.Type),
				ChatData:      res.Message,
			}
			// append info message to conversation history
			if ResponseType(res.Type) == InfoMessage {
				func() {
					workflowRunLock.Lock()
					defer workflowRunLock.Unlock()
					// append message to conversation history
					workflowMessage := sharedtypes.ConversationHistoryMessage{
						MessageId: instructionId,
						Role:      "assistant",
						Content:   res.Message,
					}
					*conversationHistory = append(*conversationHistory, workflowMessage)
				}()
			}
		case GetApproval:
			// create a channel to receive the approval response from the client
			approvalResponseChannel := make(chan string, 1)

			// track for cleanup on error
			pendingApprovalIDs = append(pendingApprovalIDs, res.InstructionId)

			// add instruction ID to the response channel to get approval from the client
			func() {
				OpenApprovalsLock.Lock()
				defer OpenApprovalsLock.Unlock()
				OpenApprovals[res.InstructionId] = OpenApproval{
					FunctionType: Run,
					Channel:      &approvalResponseChannel,
				}
			}()

			// send a message to the response channel with the approval request
			logging.Log.Debugf(ctx, "Sending approval request for instruction ID '%v' to client; Approval Text: '%v'; Approval Options '%v'", res.InstructionId, res.ApprovalText, res.ApprovalOptions)
			responseChannel <- sharedtypes.ClientResponse{
				InstructionId:   res.InstructionId,
				Type:            string(res.Type),
				ApprovalText:    res.ApprovalText,
				ApprovalOptions: res.ApprovalOptions,
			}

			// launch go routine to wait for the approval response from the client and send it to the server
			go func() {
				defer close(approvalResponseChannel)
				// check if free_text us allowed for the approval request
				freeTextAllowed := false
				for _, option := range res.ApprovalOptions {
					if option == "free_text" {
						freeTextAllowed = true
						break
					}
				}

				// wait for the approval response from the client
				var approvalResponse string
				for {
					approvalResponse = <-approvalResponseChannel
					logging.Log.Debugf(ctx, "Received approval response '%v' for instruction ID '%v'", approvalResponse, res.InstructionId)

					// verify response if free_text is not allowed
					if !freeTextAllowed {
						validResponse := false
						for _, option := range res.ApprovalOptions {
							if approvalResponse == option {
								validResponse = true
								break
							}
						}

						// send error message to client if response is not valid and wait for a new response
						if !validResponse {
							responseChannel <- sharedtypes.ClientResponse{
								Type:     "error_message",
								ChatData: fmt.Sprintf("Invalid approval response '%v' for instruction ID '%v'. Please choose from the available options: %v", approvalResponse, res.InstructionId, res.ApprovalOptions),
							}
						} else {
							// valid response, break the loop
							break
						}
					} else {
						// any response is valid if free_text is allowed, break the loop
						break
					}
				}

				// send the approval server reponse channel
				responseChannelToServer <- &flowkitgrpc.FunctionInputs{
					Type:             "approval_response",
					InstructionId:    res.InstructionId,
					ApprovalResponse: approvalResponse,
				}

				// remove the instruction ID from the OpenApprovals map
				OpenApprovalsLock.Lock()
				defer OpenApprovalsLock.Unlock()
				delete(OpenApprovals, res.InstructionId)
			}()

		default:
			// unknown type
			err := fmt.Errorf("unknown response type '%v' received from RunFunction '%v'", res.Type, functionName)
			logging.Log.Error(ctx, err)
			cleanupApprovals()
			return nil, err
		}
	}

	// Close the send direction of the stream
	stream.CloseSend()

	// wait for EOF so grpc trailers are available
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logging.Log.Warnf(ctx, "error draining RunFunction stream for '%v' before reading trailers: %v", functionName, err)
				break
			}
		}
		// no additional response expected, but log if any unexpected responses are received
		logging.Log.Warnf(ctx, "unexpected response received from RunFunction '%v' while draining stream: %v", functionName, resp)
	}

	// Update logging context with token counts from response trailers
	responseTrailer := stream.Trailer()
	if values := responseTrailer.Get("snps-ai-logging-context"); len(values) > 0 {
		var body []map[string]interface{}
		if err := json.Unmarshal([]byte(values[0]), &body); err == nil && len(body) > 0 {
			tokenKeys := []logging.ContextKey{
				logging.InputTokenCount,
				logging.OutputTokenCount,
				logging.CachedTokenCount,
				logging.ReasoningTokenCount,
			}
			for _, key := range tokenKeys {
				if val, ok := body[0][string(key)]; ok {
					ctx.Set(key, val)
				}
			}
		}
	}

	// convert outputs to map[string]sharedtypes.FilledInputOutput
	outputs = map[string]sharedtypes.FilledInputOutput{}
	for _, output := range runResp.Outputs {
		// convert value to Go type
		value, exists, err := typeconverters.ConvertStringToGivenType(output.Value, output.GoType)
		if err != nil {
			return nil, fmt.Errorf("error converting output '%s' with value '%v' for function '%v' to Go type: %v", output.Name, output.Value, functionName, err)
		}
		if !exists {
			return nil, fmt.Errorf("type '%s' does not exist in typeconverters.ConvertStringToGivenType", output.Name)
		}

		// Save the output to the map
		outputs[output.Name] = sharedtypes.FilledInputOutput{
			Name:   output.Name,
			GoType: output.GoType,
			Value:  value,
		}
	}

	return outputs, nil
}

// StreamFunction calls the StreamFunction gRPC and returns a channel to stream the outputs
// and an interrupt channel to send interrupts to the server.
// This function is used to stream the outputs of an external function
//
// Parameters:
//   - functionName: the name of the function to run
//   - inputs: the inputs to the function
//
// Returns:
//   - *chan string: a channel to stream the output from the server
//   - *chan string: an interrupt channel to send messages to the server
//   - error: an error message if the gRPC call fails
func StreamFunction(ctx *logging.ContextMap, functionName string, inputs map[string]sharedtypes.FilledInputOutput) (channel *chan string, interruptChannel *chan string, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in StreamFunction for function '%v': %v", functionName, r)
		}
	}()

	// Get function definition
	functionDef, ok := AvailableFunctions[functionName]
	if !ok {
		return nil, nil, fmt.Errorf("function '%s' not found in available functions", functionName)
	}

	// Set up a connection to the server.
	c, conn, err := createClient(functionDef.FlowkitUrl, functionDef.ApiKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to external function gRPC: %v", err)
	}

	// Create a context with a cancel
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	// get logging metadata from context
	ctxWithMetadata, err := logging.CreateMetaDataFromCtx(ctx, ctxWithCancel)
	if err != nil {
		conn.Close()
		cancel()
		return nil, nil, fmt.Errorf("error adding metadata: %v", err)
	}

	// Convert inputs to gRPC format based on order from function definition
	grpcInputs := []*flowkitgrpc.FunctionInput{}
	for _, inputDef := range functionDef.Inputs {
		// create grpc input
		grpcInput := &flowkitgrpc.FunctionInput{
			Name:   inputDef.Name,
			GoType: inputDef.GoType,
		}

		// Get the input value
		value, ok := inputs[inputDef.Name]
		if ok {
			// found: convert value to string
			stringValue, exists, err := typeconverters.ConvertGivenTypeToString(value.Value, inputDef.GoType)
			if err != nil {
				conn.Close()
				cancel()
				return nil, nil, fmt.Errorf("error converting input '%s' for function '%v' to string: %v", inputDef.Name, functionName, err)
			}
			if !exists {
				conn.Close()
				cancel()
				return nil, nil, fmt.Errorf("type '%s' does not exist in typeconverters.ConvertGivenTypeToString", inputDef.Name)
			}
			grpcInput.Value = stringValue

		} else {
			// input discrepancy, set to null value
			grpcInput.Value = ""
		}

		// Append the grpc input to the list
		grpcInputs = append(grpcInputs, grpcInput)
	}

	// Call StreamFunction (bidirectional)
	stream, err := c.StreamFunction(ctxWithMetadata)
	if err != nil {
		conn.Close()
		cancel()
		return nil, nil, fmt.Errorf("error in external function gRPC StreamFunction for function '%v': %v", functionName, err)
	}

	// Send the initial message with function inputs
	err = stream.Send(&flowkitgrpc.StreamInput{
		Name:   functionName,
		Inputs: grpcInputs,
	})
	if err != nil {
		conn.Close()
		cancel()
		return nil, nil, fmt.Errorf("error sending initial message in StreamFunction for function '%v': %v", functionName, err)
	}

	// Create channels
	streamChannel := make(chan string, 400)
	interruptCh := make(chan string, 400)

	// Receive the stream from the server
	go receiveStreamFromServer(ctx, stream, &streamChannel, conn, cancel, functionName)

	// Send interrupts to the server
	go sendInterruptsToServer(ctx, stream, &interruptCh, functionName)

	return &streamChannel, &interruptCh, nil
}

// receiveStreamFromServer receives the stream from the server and sends it to the channel
//
// Parameters:
//   - stream: the stream from the server
//   - streamChannel: the channel to send the stream to
func receiveStreamFromServer(ctx *logging.ContextMap, stream flowkitgrpc.ExternalFunctions_StreamFunctionClient, streamChannel *chan string, conn *grpc.ClientConn, cancel context.CancelFunc, functionName string) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic occured in receiveStreamFromServer for function '%v': %v", functionName, r)
		}
	}()

	// Receive the stream from the server
	for {
		res, err := stream.Recv()
		if err != nil && err != io.EOF {
			logging.Log.Errorf(ctx, "error receiving stream for function '%v': %v", functionName, err)
			*streamChannel <- fmt.Sprintf("$&$error$&$:$&$%v$&$", err)
			break
		}

		// Send the stream to the channel
		*streamChannel <- res.Value

		// end if isLast is true
		if res.IsLast {
			break
		}
	}

	// Close the channel
	conn.Close()
	cancel()
	close(*streamChannel)
}

// sendInterruptsToServer listens to the interrupt channel and sends messages to the server via gRPC
//
// Parameters:
//   - stream: the bidirectional stream to the server
//   - interruptChannel: the channel to receive interrupt messages from
//   - functionName: the name of the function (for logging)
func sendInterruptsToServer(ctx *logging.ContextMap, stream flowkitgrpc.ExternalFunctions_StreamFunctionClient, interruptChannel *chan string, functionName string) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic occured in sendInterruptsToServer for function '%v': %v", functionName, r)
		}
	}()

	// Listen to the interrupt channel and send messages to the server
	for msg := range *interruptChannel {
		switch {
		case strings.Contains(msg, "$&$approval_response$&$"):
			// remove $&$approval_response$&$ from the start of the message
			approvalResponseJson := strings.TrimPrefix(msg, "$&$approval_response$&$")
			// unmarshal the approval response
			var approvalResponse ApprovalResponse
			err := json.Unmarshal([]byte(approvalResponseJson), &approvalResponse)
			if err != nil {
				logging.Log.Errorf(ctx, "error unmarshalling approval response for function '%v': %v", functionName, err)
				continue
			}
			// send the approval response to the server
			err = stream.Send(&flowkitgrpc.StreamInput{
				InstructionId:    approvalResponse.InstructionId,
				Type:             "approval_response",
				ApprovalResponse: approvalResponse.ApprovalResponse,
			})
			if err != nil {
				logging.Log.Errorf(ctx, "error sending approval response for function '%v': %v", functionName, err)
				return
			}
		default:
			err := stream.Send(&flowkitgrpc.StreamInput{
				Type:      "interrupt",
				Interrupt: msg,
			})
			if err != nil {
				logging.Log.Errorf(ctx, "error sending interrupt for function '%v': %v", functionName, err)
				return
			}
		}
	}

	// Close the send direction when the interrupt channel is closed
	stream.CloseSend()
}

// createClient creates a client to the external functions gRPC
//
// Returns:
//   - client: the client to the external functions gRPC
//   - connection: the connection to the external functions gRPC
//   - err: an error message if the client creation fails
func createClient(url string, apiKey string) (client flowkitgrpc.ExternalFunctionsClient, connection *grpc.ClientConn, err error) {
	// Extract the scheme (http or https) from the EXTERNALFUNCTIONS_ENDPOINT
	var scheme string
	var address string
	switch {
	case strings.HasPrefix(url, "https://"):
		scheme = "https"
		address = strings.TrimPrefix(url, scheme+"://")
	case strings.HasPrefix(url, "http://"):
		scheme = "http"
		address = strings.TrimPrefix(url, scheme+"://")
	default:
		// legacy support for endpoint definition without http or https in front
		scheme = "http"
		address = url
	}

	// Get gRPC dial options
	opts, err := clients.GetGrpcDialOptions(scheme)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get gRPC dial options: %v", err)
	}

	// Add the API key if it is set
	if apiKey != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(apiKeyInterceptor(apiKey)))
	}

	// Set max message size to 1GB
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)))

	// Set up a connection to the server
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to external function gRPC: %v", err)
	}

	// Return the client
	c := flowkitgrpc.NewExternalFunctionsClient(conn)
	return c, conn, nil
}

// apiKeyInterceptor is a gRPC client interceptor that adds an API key to the context metadata
// This interceptor is used to add the API key to the context metadata for all gRPC calls
//
// Parameters:
//   - apiKey: the API key to add to the context metadata
//
// Returns:
//   - grpc.UnaryClientInterceptor: the interceptor that adds the API key to the context metadata
func apiKeyInterceptor(apiKey string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Get existing metadata from context (if any)
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			// No existing metadata, create new
			md = metadata.MD{}
		} else {
			// Copy the metadata to avoid modifying the original
			md = md.Copy()
		}

		// Add API key to the existing metadata (this preserves other keys)
		md.Set("x-api-key", apiKey)

		// Create new context with MERGED metadata
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Invoke the RPC with the modified context
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
