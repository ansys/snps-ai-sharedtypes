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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/ansys/snps-ai-sharedtypes/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
)

///////////////////////////////////
// Create Context
///////////////////////////////////

// Set function sets ContextKeys equal to any value
func (ctx *ContextMap) Set(key ContextKey, value interface{}) {
	ctx.data.Store(key, value)
}

// Get function retrieves the value for a ContextKey
//
// Parameters:
//   - key: The ContextKey for which to retrieve the value.
//
// Returns:
//   - interface{}: The value associated with the specified ContextKey.
//   - bool: A boolean indicating whether the ContextKey exists.
func (ctx *ContextMap) Get(key ContextKey) (interface{}, bool) {
	return ctx.data.Load(key)
}

// Copy function copies the current contextMap so new uses of Set do not overwrite existing values
//
// Returns:
//   - *ContextMap: A copy of the current ContextMap.
func (ctx *ContextMap) Copy() *ContextMap {
	newCtx := &ContextMap{}
	ctx.data.Range(func(key, value interface{}) bool {
		newCtx.data.Store(key, value)
		return true
	})
	return newCtx
}

///////////////////////////////////
// Create Logger
///////////////////////////////////

// InitLogger initializes the global logger.
//
// The function creates a new zap logger with the specified configuration and sets the global logger variable to the new logger.
//
// Parameters:
//   - GlobalConfig: The global configuration from the config package.
func InitLogger(GlobalConfig *config.Config) {

	// Create a new zap logger with the specified configuration
	config := zap.NewProductionConfig()
	config.Level.SetLevel(TraceLevel)
	option := zap.AddCallerSkip(1)
	config.EncoderConfig.FunctionKey = "func"
	config.EncoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		if l == TraceLevel {
			enc.AppendString("trace")
		} else {
			zapcore.LowercaseLevelEncoder(l, enc)
		}
	}
	temp, _ := config.Build(option)
	Log = loggerWrapper{lw: temp}

	// Set the global configuration variables for the logging package
	initLoggerConfig(Config{
		ErrorFileLocation: GlobalConfig.ERROR_FILE_LOCATION,
		LogLevel:          GlobalConfig.LOG_LEVEL,
		LocalLogs:         GlobalConfig.LOCAL_LOGS,
		LocalLogsLocation: GlobalConfig.LOCAL_LOGS_LOCATION,
		DatadogLogs:       GlobalConfig.DATADOG_LOGS,
		DatadogSource:     GlobalConfig.DATADOG_SOURCE,
		DatadogStage:      GlobalConfig.STAGE,
		DatadogVersion:    GlobalConfig.VERSION,
		DatadogService:    GlobalConfig.SERVICE_NAME,
		DatadogAPIKey:     GlobalConfig.LOGGING_API_KEY,
		DatadogLogsURL:    GlobalConfig.LOGGING_URL,
		DatadogMetrics:    GlobalConfig.DATADOG_METRICS,
		DatadogMetricsURL: GlobalConfig.METRICS_URL,
	})
}

// initLoggerConfig initializes the global configuration variables for the logging package.
//
// The function sets the global configuration variables to the values specified in the provided Config struct.
//
// Parameters:
//   - config: The Config struct containing the configuration values to set.
func initLoggerConfig(config Config) {
	if AppName != "" {
		APP_NAME = AppName
	} else if config.DatadogService != "" {
		APP_NAME = config.DatadogService
	} else {
		APP_NAME = "unknown"
	}
	ERROR_FILE_LOCATION = config.ErrorFileLocation
	LOG_LEVEL = config.LogLevel
	LOCAL_LOGS = config.LocalLogs
	LOCAL_LOGS_LOCATION = config.LocalLogsLocation
	DATADOG_LOGS = config.DatadogLogs
	DATADOG_SOURCE = config.DatadogSource
	DATADOG_STAGE = config.DatadogStage
	DATADOG_VERSION = config.DatadogVersion
	DATADOG_SERVICE_NAME = config.DatadogService
	DATADOG_API_KEY = config.DatadogAPIKey
	DATADOG_LOGS_URL = config.DatadogLogsURL
	DATADOG_METRICS = config.DatadogMetrics
	DATADOG_METRICS_URL = config.DatadogMetricsURL
}

///////////////////////////////////
// Logging functions
///////////////////////////////////

// Fatal logs a message with Fatal level and terminates the program.
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - args: The log message.
func (logger *loggerWrapper) Fatal(ctx *ContextMap, args ...interface{}) {
	entry := logger.lw.Check(zapcore.FatalLevel, fmt.Sprint(args...))
	if entry != nil {
		sendLogs(
			ctx,
			entry.Level,
			entry.Time,
			entry.Message,
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}

	message := "Program terminated with Fatal Error:"
	pan := writeStringToFile(ERROR_FILE_LOCATION, message)
	if pan != nil {
		panic(pan)
	}
	pan = writeInterfaceToFile(ERROR_FILE_LOCATION, fmt.Sprint(args...))
	if pan != nil {
		panic(pan)
	}

	// Wait for all pending async log writes to finish before exiting
	pendingLogs.Wait()
	logger.lw.Fatal(fmt.Sprint(args...))
}

// Fatalf logs a formatted message with Fatal level and terminates the program.
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - format: The format of the log message.
//   - args: The log message.
func (logger *loggerWrapper) Fatalf(ctx *ContextMap, format string, args ...interface{}) {
	entry := logger.lw.Check(zapcore.FatalLevel, format)
	if entry != nil {
		sendLogs(
			ctx,
			entry.Level,
			entry.Time,
			fmt.Sprintf(format, args...),
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}

	message := "Program terminated with Fatal Error:"
	pan := writeStringToFile(ERROR_FILE_LOCATION, message)
	if pan != nil {
		panic(pan)
	}
	pan = writeInterfaceToFile(ERROR_FILE_LOCATION, fmt.Sprintf(format, args...))
	if pan != nil {
		panic(pan)
	}

	// Wait for all pending async log writes to finish before exiting
	pendingLogs.Wait()
	fields := []zap.Field{zap.Any("Arguments", args)}
	logger.lw.Fatal(fmt.Sprintf(format, args...), fields...)
}

// Error logs a message with Error level if the global log level is not set to "fatal".
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - args: The log message.
func (logger *loggerWrapper) Error(ctx *ContextMap, args ...interface{}) {
	if LOG_LEVEL == "fatal" {
		return
	}

	logger.lw.Error(fmt.Sprint(args...))

	entry := logger.lw.Check(zapcore.ErrorLevel, fmt.Sprint(args...))
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			entry.Message,
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Errorf logs a formatted message with Error level if the global log level is not set to "fatal".
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - format: The format of the log message.
//   - args: The log message.
func (logger *loggerWrapper) Errorf(ctx *ContextMap, format string, args ...interface{}) {
	if LOG_LEVEL == "fatal" {
		return
	}

	fields := []zap.Field{zap.Any("Arguments", args)}
	logger.lw.Error(fmt.Sprintf(format, args...), fields...)

	entry := logger.lw.Check(zapcore.ErrorLevel, format)
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			fmt.Sprintf(format, args...),
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Warn logs a message with Error level if the global log level is not set to "error".
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - args: The log message.
func (logger *loggerWrapper) Warn(ctx *ContextMap, args ...interface{}) {
	if (LOG_LEVEL == "fatal") || (LOG_LEVEL == "error") {
		return
	}

	logger.lw.Warn(fmt.Sprint(args...))

	entry := logger.lw.Check(zapcore.WarnLevel, fmt.Sprint(args...))
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			entry.Message,
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Warnf logs a message with Error level if the global log level is not set to "error".
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - format: The format of the log message.
//   - args: The log message.
func (logger *loggerWrapper) Warnf(ctx *ContextMap, format string, args ...interface{}) {
	if (LOG_LEVEL == "fatal") || (LOG_LEVEL == "error") {
		return
	}

	fields := []zap.Field{zap.Any("Arguments", args)}
	logger.lw.Warn(fmt.Sprintf(format, args...), fields...)

	entry := logger.lw.Check(zapcore.WarnLevel, format)
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			fmt.Sprintf(format, args...),
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Info logs a message with Error level if the global log level is not set to "warn".
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - args: The log message.
func (logger *loggerWrapper) Info(ctx *ContextMap, args ...interface{}) {
	if (LOG_LEVEL == "fatal") || (LOG_LEVEL == "error") || (LOG_LEVEL == "warn") {
		return
	}

	logger.lw.Info(fmt.Sprint(args...))

	entry := logger.lw.Check(zapcore.InfoLevel, fmt.Sprint(args...))
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			entry.Message,
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Infof logs a message with Error level if the global log level is not set to "warn".
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - format: The format of the log message.
//   - args: The log message.
func (logger *loggerWrapper) Infof(ctx *ContextMap, format string, args ...interface{}) {
	if (LOG_LEVEL == "fatal") || (LOG_LEVEL == "error") || (LOG_LEVEL == "warn") {
		return
	}

	fields := []zap.Field{zap.Any("Arguments", args)}
	logger.lw.Info(fmt.Sprintf(format, args...), fields...)

	entry := logger.lw.Check(zapcore.InfoLevel, format)
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			fmt.Sprintf(format, args...),
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Debugf logs a formatted message with Debug level if the global log level is set to "debug."
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - format: The format of the log message.
//   - args: The log message.
func (logger *loggerWrapper) Debugf(ctx *ContextMap, format string, args ...interface{}) {
	if (LOG_LEVEL == "fatal") || (LOG_LEVEL == "error") || (LOG_LEVEL == "warn") || (LOG_LEVEL == "info") {
		return
	}

	fields := []zap.Field{zap.Any("Arguments", args)}
	logger.lw.Debug(fmt.Sprintf(format, args...), fields...)

	entry := logger.lw.Check(zapcore.DebugLevel, format)
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			fmt.Sprintf(format, args...),
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Tracef logs a formatted message with Trace level if the global log level is set to "trace."
//
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - format: The format of the log message.
//   - args: The log message.
func (logger *loggerWrapper) Tracef(ctx *ContextMap, format string, args ...interface{}) {
	if LOG_LEVEL != "trace" {
		return
	}

	fields := []zap.Field{zap.Any("Arguments", args)}
	logger.lw.Log(TraceLevel, fmt.Sprintf(format, args...), fields...)

	entry := logger.lw.Check(TraceLevel, format)
	if entry != nil {
		asyncSendLogs(
			ctx,
			entry.Level,
			entry.Time,
			fmt.Sprintf(format, args...),
			entry.Caller,
			entry.Stack,
			entry.Caller.Function,
			args...)
	}
}

// Metrics sends a metric event with the specified name and count to Datadog if Datadog metrics are enabled.
//
// Parameters:
//   - name: The name of the metric.
//   - count: The value of the metric.
func (logger *loggerWrapper) Metrics(name string, count float64) {
	if !DATADOG_METRICS {
		return
	}

	go sendMetrics(name, count)
}

///////////////////////////////////
// Datadog logging helper functions
///////////////////////////////////

// asyncSendLogs dispatches sendLogs in a goroutine, tracking it with pendingLogs
// so that Fatal/Fatalf can wait for all pending writes before exiting.
func asyncSendLogs(ctx *ContextMap, level zapcore.Level, t time.Time, message string, caller zapcore.EntryCaller, stack string, function string, arguments ...interface{}) {
	pendingLogs.Add(1)
	go func() {
		defer pendingLogs.Done()
		sendLogs(ctx, level, t, message, caller, stack, function, arguments...)
	}()
}

// sendLogs sends log entries to Datadog or writes them to a local file, depending on the global configuration settings. It formats log entries and prepares them for transmission.

// The function constructs a log entry with the specified parameters and context, and then sends it to Datadog if enabled in the global configuration. It also writes log entries to a local file if local logs are enabled.
// If any errors occur during this process, they are logged and written to the local error log file.
// Parameters:
//   - ctx: A ContextMap containing context information to be included in the log entry.
//   - level: The log entry's severity level (e.g., Debug, Info, Error).
//   - time: The timestamp of the log entry.
//   - message: The log message.
//   - caller: Information about the caller of the log entry.
//   - stack: The stack trace of the log entry.
//   - function: The function where the log entry was created.
//   - arguments: Additional log entry arguments.
func sendLogs(ctx *ContextMap, level zapcore.Level, time time.Time, message string, caller zapcore.EntryCaller, stack string, function string, arguments ...interface{}) {
	defer func() {
		r := recover()
		if r != nil {
			message := "Error occurred during sendLogs:"
			pan := writeStringToFile(ERROR_FILE_LOCATION, message)
			if pan != nil {
				panic(pan)
			}
			pan = writeInterfaceToFile(ERROR_FILE_LOCATION, r)
			if pan != nil {
				panic(pan)
			}
			return
		}
	}()
	// Convert everything to string
	levelString := levelToString(level)
	timeString := timeToString(time)
	callerString := entryCallerToString(caller)

	// Convert arguments to string representations to ensure JSON serializability
	stringArgs := make([]string, len(arguments))
	for i, arg := range arguments {
		stringArgs[i] = fmt.Sprintf("%v", arg)
	}

	// Create rest API call body structure
	body := []map[string]interface{}{
		{
			"ddsource":  DATADOG_SOURCE,
			"ddtags":    "env:" + DATADOG_STAGE + ",version:" + DATADOG_VERSION,
			"message":   message,
			"time":      timeString,
			"service":   DATADOG_SERVICE_NAME,
			"caller":    callerString,
			"stack":     stack,
			"function":  function,
			"status":    levelString,
			"arguments": stringArgs,
		},
	}

	// Append body with context
	ctx.data.Range(func(key, value interface{}) bool {
		if key.(ContextKey) == FlowkitStream {
			return true
		}
		body[0][string(key.(ContextKey))] = value
		return true
	})

	// Convert body to JSON
	bodyJSON, err := mapsToJSONBytes(body)
	if err != nil {
		message := "Error occurred during mapsToJSONBytes in sendLogs: %v"
		pan := writeStringToFile(ERROR_FILE_LOCATION, fmt.Sprintf(message, body))
		if pan != nil {
			panic(pan)
		}
		pan2 := writeInterfaceToFile(ERROR_FILE_LOCATION, err)
		if pan2 != nil {
			panic(pan2)
		}
	}

	if LOCAL_LOGS {

		// Write logs to local file in human-readable columnar format
		err := writeFormattedLogToFile(dailyLogPath(LOCAL_LOGS_LOCATION), APP_NAME, timeString, levelString, function, callerString, message, stack, stringArgs, ctx)
		if err != nil {
			message := "Error occurred in writeFormattedLogToFile:"
			pan := writeStringToFile(ERROR_FILE_LOCATION, message)
			if pan != nil {
				panic(pan)
			}
			pan2 := writeInterfaceToFile(ERROR_FILE_LOCATION, err)
			if pan2 != nil {
				panic(pan2)
			}
		}

	}

	if DATADOG_LOGS {
		if DATADOG_API_KEY == "" || DATADOG_LOGS_URL == "" {
			message := "'DATADOG_LOGS' set to 'true' in 'config.yaml' file but 'DATADOG_API_KEY' and/or 'DATADOG_LOGS_URL' were not defined"
			pan := writeStringToFile(ERROR_FILE_LOCATION, message)
			if pan != nil {
				panic(pan)
			}
			panic(message)
		}
		// Send POST call to datadog
		_, err2 := sendPostRequestToDatadog(DATADOG_LOGS_URL, bodyJSON, DATADOG_API_KEY)
		if err2 != nil {
			message := "Error occurred during sendPostRequestToDatadog in sendLogs:"
			pan := writeStringToFile(ERROR_FILE_LOCATION, message)
			if pan != nil {
				panic(pan)
			}
			pan2 := writeInterfaceToFile(ERROR_FILE_LOCATION, err2)
			if pan2 != nil {
				panic(pan2)
			}
		}
	}
}

// sendMetrics sends a metric to Datadog using the specified name and count. The function creates a metric object, converts it to JSON, and sends a POST request to Datadog's metrics endpoint.
//
// The function constructs a Metrics object containing the metric data and associated resource information. It then converts the object to JSON and sends it as a POST request to Datadog for metrics reporting. Any errors encountered during this process are logged and written to the local error log file.
//
// Parameters:
//   - name: The name of the metric.
//   - count: The value of the metric.
func sendMetrics(name string, count float64) {
	defer func() {
		r := recover()
		if r != nil {
			message := "Error occurred during sendMetrics:"
			err := writeStringToFile(ERROR_FILE_LOCATION, message)
			if err != nil {
				fmt.Println(err)
			}
			err = writeInterfaceToFile(ERROR_FILE_LOCATION, r)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}()

	// Create metrics object
	metrics := Metrics{
		Series: []Metric{
			{
				Metric: name,
				Type:   0,
				Points: []Point{
					{
						Timestamp: time.Now().Unix(),
						Value:     count,
					},
				},
				Resources: []Resource{
					{
						Type: "host",
					},
				},
			},
		},
	}

	// Convert to json
	jsonBody, err := json.Marshal(metrics)
	if err != nil {
		fmt.Println("Error creating JSON:", err)
		return
	}

	// Send POST call to datadog
	_, err2 := sendPostRequestToDatadog(DATADOG_METRICS_URL, jsonBody, DATADOG_API_KEY)
	if err2 != nil {
		message := "Error occurred during sendPostRequestToDatadog in sendMetrics:"
		pan := writeStringToFile(ERROR_FILE_LOCATION, message)
		if pan != nil {
			panic(pan)
		}
		pan2 := writeInterfaceToFile(ERROR_FILE_LOCATION, err2)
		if pan2 != nil {
			panic(pan2)
		}
	}
}

// sendPostRequestToDatadog sends the metric or logs post request to Datadog.
//
// Parameters:
//   - url: The URL to which the request is sent.
//   - requestBody: The request body.
//   - apiKey: The Datadog API key.
//
// Returns:
//   - *http.Response: The response from the POST request.
//   - error: An error if the POST request fails.
func sendPostRequestToDatadog(url string, requestBody []byte, apiKey string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("DD-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 202 {
		message := "Response of sendPostRequestToDatadog is unequal to 202 Acceptepted:"
		err := writeInterfaceToFile(ERROR_FILE_LOCATION, message)
		if err != nil {
			fmt.Println(err)
		}
		err = writeInterfaceToFile(ERROR_FILE_LOCATION, resp.Status)
		if err != nil {
			fmt.Println(err)
		}
	}

	return resp, nil
}

// mapsToJSONBytes converts a slice of maps to a JSON-encoded byte slice. It takes an array of maps, marshals it to JSON format, and returns the resulting byte slice.
//
// Parameters:
//   - maps: The slice of maps to be converted.
//
// Returns:
//   - []byte: The JSON-encoded byte slice.
//   - error: An error if the conversion fails.
func mapsToJSONBytes(maps []map[string]interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(maps)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

// levelToString converts a zapcore.Level to its string representation.
//
// Parameters:
//   - level: The zapcore.Level to be converted.
//
// Returns:
//   - string: The string representation of the zapcore.Level.
func levelToString(level zapcore.Level) string {
	if level == TraceLevel {
		return "trace"
	}
	return level.String()
}

// timeToString converts a time.Time value to a string representation using the "2006-01-02 15:04:05.000" layout.
//
// Parameters:
//   - t: The time.Time value to be converted.
//
// Returns:
//   - string: The string representation of the time.Time value.
func timeToString(t time.Time) string {
	layout := "2006-01-02 15:04:05.000"
	return t.Format(layout)
}

// entryCallerToString converts a zapcore.EntryCaller to a string representation.
//
// Parameters:
//   - ec: The zapcore.EntryCaller to be converted.
//
// Returns:
//   - string: The string representation of the zapcore.EntryCaller.
func entryCallerToString(ec zapcore.EntryCaller) string {
	return ec.String()
}

///////////////////////////////////
// Local log file writer (columnar)
///////////////////////////////////

// dailyLogPath takes the configured log file location and returns a path with
// today's date (YYYY-MM-DD) prepended to the filename.
// It handles:
//   - absolute paths (Windows or Linux)
//   - relative paths
//   - bare filenames
//   - empty string (defaults to "local.log")
//
// Examples:
//
//	"app.log"                     -> "2026-07-16_app.log"
//	"logs/app.log"                -> "logs/2026-07-16_app.log"
//	"/var/log/app.log"            -> "/var/log/2026-07-16_app.log"
//	"C:\\logs\\app.log"            -> "C:\\logs\\2026-07-16_app.log"
//	""                            -> "2026-07-16_local.log"
func dailyLogPath(location string) string {
	if location == "" {
		location = "local.log"
	}

	dir := filepath.Dir(location)
	base := filepath.Base(location)
	datePrefix := time.Now().Format("2006-01-02")

	return filepath.Join(dir, datePrefix+"_"+base)
}

// shortenFunction shortens a fully-qualified Go function name to its last two dot-separated segments.
// e.g. "github.com/ansys/snps-ai-agent/pkg/workflows/workflowstore.loadPredefinedWorkflows" -> "workflowstore.loadPredefinedWorkflows"
func shortenFunction(fn string) string {
	if fn == "" {
		return ""
	}
	// Find the last '/' to get "package.Function"
	if idx := strings.LastIndex(fn, "/"); idx >= 0 {
		return fn[idx+1:]
	}
	return fn
}

// shortenCaller shortens a full caller path to just "filename:line".
// e.g. "C:/Users/fkuhn/Documents/GitHub/snps-ai-agent/pkg/clients/flowkit/flowkit.go:67" -> "flowkit.go:67"
func shortenCaller(caller string) string {
	if caller == "" {
		return ""
	}
	// Caller format is "path/to/file.go:line"
	// Find the last '/' or '\' before the ':'
	for i := len(caller) - 1; i >= 0; i-- {
		if caller[i] == '/' || caller[i] == '\\' {
			return caller[i+1:]
		}
	}
	return caller
}

// Column widths for the local log file format.
const (
	colWidthTimestamp = 23
	colWidthApp       = 15
	colWidthLevel     = 5
	colWidthMessage   = 60
	colWidthFunction  = 50
	colWidthCaller    = 30
	colWidthStack     = 60
	colWidthContext   = 40
)

// localLogHeader is the title row written when a new log file is created.
var localLogHeader = fmt.Sprintf("%-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s\n",
	colWidthTimestamp, "TIMESTAMP",
	colWidthApp, "APP",
	colWidthLevel, "LEVEL",
	colWidthMessage, "MESSAGE",
	colWidthFunction, "FUNCTION",
	colWidthCaller, "CALLER",
	colWidthStack, "STACK",
	colWidthContext, "CONTEXT") +
	fmt.Sprintf("%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s\n",
		strings.Repeat("-", colWidthTimestamp),
		strings.Repeat("-", colWidthApp),
		strings.Repeat("-", colWidthLevel),
		strings.Repeat("-", colWidthMessage),
		strings.Repeat("-", colWidthFunction),
		strings.Repeat("-", colWidthCaller),
		strings.Repeat("-", colWidthStack),
		strings.Repeat("-", colWidthContext))

// wrapText splits a string into lines of at most width characters, breaking at hard character boundaries.
// Continuation lines are indented with 2 spaces (reducing effective width by 2).
func wrapText(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	var lines []string
	for len(s) > width {
		lines = append(lines, s[:width])
		s = "  " + s[width:]
	}
	lines = append(lines, s)
	return lines
}

// wrapTextDot splits a string into lines of at most width characters, preferring to break after '.' boundaries.
// This is intended for qualified function names (e.g. "package.FunctionName") so that breaks look natural.
// Falls back to hard character break when no suitable dot is found.
func wrapTextDot(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	var lines []string
	for len(s) > width {
		// Prefer breaking after the last '.' within the allowed width
		breakAt := strings.LastIndex(s[:width], ".")
		if breakAt < 1 || breakAt < width/4 {
			// No good dot break found — hard break
			breakAt = width
		} else {
			breakAt++ // include the dot on the first line
		}
		lines = append(lines, s[:breakAt])
		s = "  " + s[breakAt:]
	}
	lines = append(lines, s)
	return lines
}

// wrapTextWords splits a string into lines of at most width characters, preferring to break at word boundaries.
// Continuation lines are indented with 2 spaces.
func wrapTextWords(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	var lines []string
	for len(s) > width {
		// Find the last space within the allowed width, but not in the leading indent
		breakAt := strings.LastIndex(s[:width], " ")
		if breakAt < width/4 {
			// Space only found very early (e.g. in the indent) — hard break to ensure progress
			breakAt = width
		}
		lines = append(lines, s[:breakAt])
		s = "  " + strings.TrimLeft(s[breakAt:], " ")
	}
	lines = append(lines, s)
	return lines
}

// writeFormattedLogToFile writes a log entry to a file in a human-readable columnar format.
// Content that exceeds the column width is wrapped onto continuation lines to keep columns aligned.
// The entire entry is built as a single string and written atomically to avoid interleaving from concurrent goroutines.
func writeFormattedLogToFile(filename, app, timeStr, level, function, caller, message, stack string, args []string, ctx *ContextMap) error {
	// Open file atomically: create if missing, always append. Avoids TOCTOU race with concurrent goroutines.
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header if the file is empty (freshly created)
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		_, err = file.WriteString(localLogHeader)
		if err != nil {
			return err
		}
	}

	shortFunc := shortenFunction(function)
	shortCaller := shortenCaller(caller)
	upperLevel := strings.ToUpper(level)

	// Replace embedded newlines with spaces to preserve table alignment,
	// then strip trailing whitespace.
	message = strings.ReplaceAll(message, "\r\n", " ")
	message = strings.ReplaceAll(message, "\n", " ")
	message = strings.ReplaceAll(message, "\r", " ")
	message = strings.TrimRight(message, " \t")

	// Build stack column: join all frames with " > "
	stackCol := ""
	if stack != "" {
		frames := strings.Split(stack, "\n")
		var trimmed []string
		for _, f := range frames {
			f = strings.TrimSpace(f)
			if f != "" {
				trimmed = append(trimmed, f)
			}
		}
		stackCol = strings.Join(trimmed, " > ")
	}

	// Build context column: key=value pairs joined with ", "
	ctxCol := ""
	if ctx != nil {
		var parts []string
		ctx.data.Range(func(key, value interface{}) bool {
			if key.(ContextKey) == FlowkitStream {
				return true
			}
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
			return true
		})
		ctxCol = strings.Join(parts, ", ")
	}

	// Wrap variable-width columns (word-aware for message, dot-aware for function, hard wrap for others)
	// Context is last column so no wrapping needed
	appLines := wrapText(app, colWidthApp)
	funcLines := wrapTextDot(shortFunc, colWidthFunction)
	callerLines := wrapText(shortCaller, colWidthCaller)
	msgLines := wrapTextWords(message, colWidthMessage)
	stackLines := wrapText(stackCol, colWidthStack)

	// Determine how many output lines we need
	maxLines := len(msgLines)
	if len(stackLines) > maxLines {
		maxLines = len(stackLines)
	}
	if len(funcLines) > maxLines {
		maxLines = len(funcLines)
	}
	if len(callerLines) > maxLines {
		maxLines = len(callerLines)
	}
	if len(appLines) > maxLines {
		maxLines = len(appLines)
	}

	// Blank padding for fixed columns on continuation lines
	blankTimestamp := strings.Repeat(" ", colWidthTimestamp)
	blankLevel := strings.Repeat(" ", colWidthLevel)

	// Build complete entry as a single string for atomic write
	var buf strings.Builder
	for i := 0; i < maxLines; i++ {
		ts := blankTimestamp
		lv := blankLevel
		if i == 0 {
			ts = timeStr
			lv = upperLevel
		}

		ap := ""
		if i < len(appLines) {
			ap = appLines[i]
		}
		fn := ""
		if i < len(funcLines) {
			fn = funcLines[i]
		}
		cl := ""
		if i < len(callerLines) {
			cl = callerLines[i]
		}

		msg := ""
		if i < len(msgLines) {
			msg = msgLines[i]
		}
		stk := ""
		if i < len(stackLines) {
			stk = stackLines[i]
		}
		ctxV := ""
		if i == 0 {
			ctxV = ctxCol
		}

		buf.WriteString(fmt.Sprintf("%-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %s\n",
			colWidthTimestamp, ts,
			colWidthApp, ap,
			colWidthLevel, lv,
			colWidthMessage, msg,
			colWidthFunction, fn,
			colWidthCaller, cl,
			colWidthStack, stk,
			ctxV))
	}

	// Single atomic write
	_, err = file.WriteString(buf.String())
	return err
}

///////////////////////////////////
// Log Error file creator
///////////////////////////////////

// writeInterfaceToFile writes data, which is an interface{} representing structured data, to a file in JSON format. It adds a timestamp to each entry.
//
// Parameters:
//   - filename: The name of the file to write to.
//   - data: The structured data to be written to the file.
//
// Returns:
//   - error: An error if writing to the file fails.
func writeInterfaceToFile(filename string, data interface{}) error {
	var file *os.File
	var err error

	// create file
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// If the file does not exist, create a new file.
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
	} else {
		// If the file already exists, open it in append mode.
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// write to file
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// add time
	timestamp := timeToString(time.Now())

	// write to file
	line := fmt.Sprintf("%s: %s\n", timestamp, string(jsonData))
	_, err = file.Write([]byte(line))
	if err != nil {
		return err
	}

	return nil
}

// writeStringToFile appends a string message to a file, including a timestamp.
//
// Parameters:
//   - filename: The name of the file to write to.
//   - data: The string message to be written to the file.
//
// Returns:
//   - error: An error if writing to the file fails.
func writeStringToFile(filename string, data string) error {
	var file *os.File
	var err error

	// add time
	timestamp := timeToString(time.Now())

	// change string
	data = timestamp + ": " + data

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		// file does not exist, create a new file
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
	} else {
		// file exists, open it for appending
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// append data to file with a new line
	_, err = fmt.Fprintln(file, data)
	return err
}

///////////////////////////////////
// Log Context metadata functions
///////////////////////////////////

// CreateMetaDataFromCtx creates gRPC metadata from the given ContextMap and attaches it to the provided context.
//
// Parameters:
//   - ctx: the logging context map containing metadata values
//   - ctxWithCancel: the gRPC context to which the metadata will be attached
//
// Returns:
//   - ctxWithMetaData: the new gRPC context with the attached metadata
//   - err: an error if the metadata creation or attachment fails
func CreateMetaDataFromCtx(ctx *ContextMap, ctxWithCancel context.Context) (ctxWithMetaData context.Context, err error) {
	// Append body with context
	body := []map[string]interface{}{
		{},
	}
	ctx.data.Range(func(key, value interface{}) bool {
		body[0][string(key.(ContextKey))] = value
		return true
	})

	// Serialize struct to JSON
	jsonData, err := json.Marshal(&body)
	if err != nil {
		return nil, fmt.Errorf("error serializing metadata struct to JSON: %v", err)
	}

	// Attach metadata to gRPC context
	md := metadata.Pairs(
		"snps-ai-logging-context", string(jsonData),
	)
	return metadata.NewOutgoingContext(ctxWithCancel, md), nil
}

// CreateCtxFromMetaData creates a ContextMap from gRPC metadata in the provided context.
//
// Parameters:
//   - ctxWithMetaData: the gRPC context containing the metadata
//
// Returns:
//   - ctx: the logging context map created from the metadata
//   - err: an error if the metadata extraction or deserialization fails
func CreateCtxFromMetaData(ctxWithMetaData context.Context) (ctx *ContextMap, err error) {
	// Create new ContextMap
	ctx = &ContextMap{}

	// Extract metadata from incoming context
	md, ok := metadata.FromIncomingContext(ctxWithMetaData)
	if !ok {
		return ctx, nil
	}

	// Get the snps-ai-logging-context value
	metadataValues := md.Get("snps-ai-logging-context")
	if len(metadataValues) == 0 {
		return ctx, nil
	}

	// Take the first value (there should only be one)
	jsonData := metadataValues[0]

	// Deserialize JSON to body
	var body []map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &body)
	if err != nil {
		return nil, fmt.Errorf("error deserializing JSON to metadata: %v", err)
	}

	// Populate the ContextMap with data from body
	if len(body) > 0 && body[0] != nil {
		for key, value := range body[0] {
			ctx.data.Store(ContextKey(key), value)
		}
	}

	return ctx, nil
}

// CreateDialOptionsFromCtx creates websocket dial options from the given ContextMap.
//
// Parameters:
//   - ctx: the logging context map containing metadata values
//
// Returns:
//   - opts: the websocket dial options with the attached metadata
//   - err: an error if the metadata creation fails
func CreateDialOptionsFromCtx(ctx *ContextMap) (opts *websocket.DialOptions, err error) {
	// Append body with context
	body := []map[string]interface{}{
		{},
	}
	ctx.data.Range(func(key, value interface{}) bool {
		body[0][string(key.(ContextKey))] = value
		return true
	})

	// Serialize struct to JSON
	jsonData, err := json.Marshal(&body)
	if err != nil {
		return nil, fmt.Errorf("error serializing metadata struct to JSON: %v", err)
	}
	opts = &websocket.DialOptions{
		HTTPHeader: http.Header{
			"snps-ai-logging-context": []string{string(jsonData)},
		},
	}
	return opts, nil
}

// CreateCtxFromHeader creates a ContextMap from HTTP request headers.
//
// Parameters:
//   - request: the HTTP request containing the headers
//
// Returns:
//   - ctx: the logging context map created from the headers
//   - err: an error if the header extraction or deserialization fails
func CreateCtxFromHeader(request *http.Request) (ctx *ContextMap, err error) {
	// Create new ContextMap
	ctx = &ContextMap{}

	// Get the snps-ai-logging-context value
	meta := request.Header.Get("snps-ai-logging-context")
	if meta == "" {
		return ctx, nil
	}

	// Deserialize JSON to body
	var body []map[string]interface{}
	err = json.Unmarshal([]byte(meta), &body)
	if err != nil {
		return nil, fmt.Errorf("error deserializing JSON to metadata: %v", err)
	}

	// Populate the ContextMap with data from body
	if len(body) > 0 && body[0] != nil {
		for key, value := range body[0] {
			ctx.data.Store(ContextKey(key), value)
		}
	}
	return ctx, nil
}
