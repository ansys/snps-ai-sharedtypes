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
	"fmt"
	"strings"
)

// Config contains all the configuration settings for the Synopsys AI Tools service.
type Config struct {

	// Logging
	///////////
	LOG_LEVEL string `yaml:"LOG_LEVEL" json:"LOGLEVEL"`
	// Local Logs
	LOCAL_LOGS          bool   `yaml:"LOCAL_LOGS" json:"LOCALLOGS"`
	LOCAL_LOGS_LOCATION string `yaml:"LOCAL_LOGS_LOCATION" json:"LOCALLOGSLOCATION"`
	// Datadog Logs
	DATADOG_LOGS        bool   `yaml:"DATADOG_LOGS" json:"DATADOGLOGS"`
	STAGE               string `yaml:"STAGE" json:"STAGE"`
	VERSION             string `yaml:"VERSION" json:"VERSION"`
	SERVICE_NAME        string `yaml:"SERVICE_NAME" json:"SERVICENAME"`
	ERROR_FILE_LOCATION string `yaml:"ERROR_FILE_LOCATION" json:"ERRORFILELOCATION"`
	LOGGING_URL         string `yaml:"LOGGING_URL" json:"LOGGINGURL"`
	LOGGING_API_KEY     string `yaml:"LOGGING_API_KEY" json:"LOGGINGAPIKEY"`
	DATADOG_SOURCE      string `yaml:"DATADOG_SOURCE" json:"DATADOGSOURCE"`
	// Datadog Metrics
	DATADOG_METRICS bool   `yaml:"DATADOG_METRICS" json:"DATADOGMETRICS"`
	METRICS_URL     string `yaml:"METRICS_URL" json:"METRICSURL"`

	// SSL Settings
	/////////////////
	USE_SSL                   bool   `yaml:"USE_SSL" json:"USESSL"`
	SSL_CERT_PUBLIC_KEY_FILE  string `yaml:"SSL_CERT_PUBLIC_KEY_FILE" json:"SSLCERTPUBLICKEYFILE"`
	SSL_CERT_PRIVATE_KEY_FILE string `yaml:"SSL_CERT_PRIVATE_KEY_FILE" json:"SSLCERTPRIVATEKEYFILE"`

	// Azure Key Vault Settings
	////////////////////////////
	EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT bool   `yaml:"EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT" json:"EXTRACTCONFIGFROMAZUREKEYVAULT"`
	AZURE_KEY_VAULT_NAME                string `yaml:"AZURE_KEY_VAULT_NAME" json:"AZUREKEYVAULTNAME"`
	AZURE_MANAGED_IDENTITY_ID           string `yaml:"AZURE_MANAGED_IDENTITY_ID" json:"AZUREMANAGEDIDENTITYID"`

	// Chat
	///////
	CHAT_ADDRESS string `yaml:"CHAT_ADDRESS" json:"CHATADDRESS"`

	// Agent
	////////
	AGENT_ADDRESS    string `yaml:"AGENT_ADDRESS" json:"AGENTADDRESS"`
	WORKFLOW_API_KEY string `yaml:"WORKFLOW_API_KEY" json:"WORKFLOWAPIKEY"`
	// Workflow Runs
	NUMBER_OF_WORKFLOW_WORKERS                 int  `yaml:"NUMBER_OF_WORKFLOW_WORKERS" json:"NUMBEROFWORKFLOWWORKERS"`
	PRODUCTION_MODE                            bool `yaml:"PRODUCTION_MODE" json:"PRODUCTIONMODE"` // If true, the agent error messages will be generic and workflow build API is disabled
	DISABLE_CONVERSATION_HISTORY_API           bool `yaml:"DISABLE_CONVERSATION_HISTORY_API" json:"DISABLECONVERSATIONHISTORYAPI"`
	DISABLE_WORKFLOW_RUN_REST_API              bool `yaml:"DISABLE_WORKFLOW_RUN_REST_API" json:"DISABLEWORKFLOWRUNRESTAPI"`
	ENFORCE_WORKFLOW_API_KEY_FOR_WORKFLOW_RUNS bool `yaml:"ENFORCE_WORKFLOW_API_KEY_FOR_WORKFLOW_RUNS" json:"ENFORCEWORKFLOWAPIKEYFORWORKFLOWRUNS"`
	// MCP Server
	ENABLE_MCP_SERVER                bool `yaml:"ENABLE_MCP_SERVER" json:"ENABLEMCPSERVER"` // If true, the agent exposes workflow runs via an MCP (Model Context Protocol) server at /mcp
	MCP_SESSION_IDLE_TIMEOUT_MINUTES int  `yaml:"MCP_SESSION_IDLE_TIMEOUT_MINUTES" json:"MCPSESSIONIDLETIMEOUTMINUTES"`
	MCP_MAX_CONCURRENT_SESSIONS      int  `yaml:"MCP_MAX_CONCURRENT_SESSIONS" json:"MCPMAXCONCURRENTSESSIONS"`
	MCP_TOOL_CALL_TIMEOUT_SECONDS    int  `yaml:"MCP_TOOL_CALL_TIMEOUT_SECONDS" json:"MCPTOOLCALLTIMEOUTSECONDS"`
	// Workflow Files
	WORKFLOW_STORE_PATH       string   `yaml:"WORKFLOW_STORE_PATH" json:"WORKFLOWSTOREPATH"`
	BINARY_STORE_PATH         string   `yaml:"BINARY_STORE_PATH" json:"BINARYSTOREPATH"`
	DISABLE_PUBLIC_WORKFLOWS  bool     `yaml:"DISABLE_PUBLIC_WORKFLOWS" json:"DISABLEPUBLICWORKFLOWS"`
	LOAD_PRIVATE_WORKFLOWS    bool     `yaml:"LOAD_PRIVATE_WORKFLOWS" json:"LOADPRIVATEWORKFLOWS"`
	GITHUB_USER               string   `yaml:"GITHUB_USER" json:"GITHUBUSER"`
	GITHUB_TOKEN              string   `yaml:"GITHUB_TOKEN" json:"GITHUBTOKEN"`
	PRIVATE_WORKFLOWS_FOLDERS []string `yaml:"PRIVATE_WORKFLOWS_FOLDERS" json:"PRIVATEWORKFLOWSFOLDERS"`
	// Flowkit Connection
	FLOWKIT_CONNECTIONS        []FlowkitConnection `yaml:"FLOWKIT_CONNECTIONS" json:"FLOWKITCONNECTIONS"`              // Contains the URL and API key for the FlowKit server
	FLOWKIT_PYTHON_CONNECTIONS []FlowkitConnection `yaml:"FLOWKIT_PYTHON_CONNECTIONS" json:"FLOWKITPYTHONCONNECTIONS"` // Contains the URL and API key for the FlowKit-Python server
	// External Function Endpoints (Legacy)
	EXTERNALFUNCTIONS_ENDPOINT string `yaml:"EXTERNALFUNCTIONS_ENDPOINT" json:"EXTERNALFUNCTIONSENDPOINT"`
	FLOWKIT_PYTHON_ENDPOINT    string `yaml:"FLOWKIT_PYTHON_ENDPOINT" json:"FLOWKITPYTHONENDPOINT"`
	// Exec Settings
	EXEC_ENDPOINT                        string `yaml:"EXEC_ENDPOINT" json:"EXECENDPOINT"`
	EXEC_AGENT_API_KEY                   string `yaml:"EXEC_AGENT_API_KEY" json:"EXECAGENTAPIKEY"`
	MONGO_DB_FOR_MULTI_AGENT             bool   `yaml:"MONGO_DB_FOR_MULTI_AGENT" json:"MONGODBFORMULTIAGENT"`
	MONGO_DB_ENDPOINT                    string `yaml:"MONGO_DB_ENDPOINT" json:"MONGODBENDPOINT"`
	MILLISECONDS_MONGODB_UPDATE_INTERVAL int    `yaml:"MILLISECONDS_MONGODB_UPDATE_INTERVAL" json:"MILLISECONDSMONGODBUPDATEINTERVAL"`
	EXEC_FILE_STORE_PATH                 string `yaml:"EXEC_FILE_STORE_PATH" json:"EXECFILESTOREPATH"`
	// DB Connection
	KVDB_ENDPOINT string `yaml:"KVDB_ENDPOINT" json:"KVDBENDPOINT"`
	// LLM Connection
	LLM_REST_ENDPOINT string `yaml:"LLM_REST_ENDPOINT" json:"LLMRESTENDPOINT"`
	// Authentication & Authorization
	ENABLE_AUTH                            bool     `yaml:"ENABLE_AUTH" json:"ENABLEAUTH"` // If true, the agent will require authentication/authorization for workflows
	AZURE_AD_AUTHENTICATION_URL            string   `yaml:"AZURE_AD_AUTHENTICATION_URL" json:"AZUREADAUTHENTICATIONURL"`
	AZURE_AD_AUTHENTICATION_URLS           []string `yaml:"AZURE_AD_AUTHENTICATION_URLS" json:"AZUREADAUTHENTICATIONURLS"` // List of Azure AD authentication URLs for multi-tenant support; overwrites AZURE_AD_AUTHENTICATION_URL if provided
	ANSYS_AUTHORIZATION_URL                string   `yaml:"ANSYS_AUTHORIZATION_URL" json:"ANSYSAUTHORIZATIONURL"`
	ANSYS_GATING_AND_ENTITLEMENT_URL       string   `yaml:"ANSYS_GATING_AND_ENTITLEMENT_URL" json:"ANSYSGATINGANDENTITLEMENTURL"`
	ANSYS_AUTHORIZATION_CRYPT_KEY          string   `yaml:"ANSYS_AUTHORIZATION_CRYPT_KEY" json:"ANSYSAUTHORIZATIONCRYPTKEY"`
	ANSYS_AUTHORIZATION_SECRET_KEY         string   `yaml:"ANSYS_AUTHORIZATION_SECRET_KEY" json:"ANSYSAUTHORIZATIONSECRETKEY"`
	ANSYS_AUTHORIZATION_SECRET_KEY_2       string   `yaml:"ANSYS_AUTHORIZATION_SECRET_KEY_2" json:"ANSYSAUTHORIZATIONSECRETKEY2"`
	ANSYS_AUTHORIZATION_SECRET_KEY_2_VALUE string   `yaml:"ANSYS_AUTHORIZATION_SECRET_KEY_2_VALUE" json:"ANSYSAUTHORIZATIONSECRETKEY2VALUE"`
	ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS  []int    `yaml:"ANSYS_AUTHORIZATION_ACCEPTED_PERSONAS" json:"ANSYSAUTHORIZATIONACCEPTEDPERSONAS"` // List of Ansys personas that are allowed to access the agent
	ANSYS_DISCO_CRYPT_PRIVAT_KEY           string   `yaml:"ANSYS_DISCO_CRYPT_PRIVAT_KEY" json:"ANSYSDISCOCRYPTPRIVATKEY"`
	ANSYS_DISOC_SIGN_PUBLIC_KEY            string   `yaml:"ANSYS_DISOC_SIGN_PUBLIC_KEY" json:"ANSYSDISOCSIGNPUBLICKEY"`
	// OKTA Authentication URL
	OKTA_AUTHENTICATION_URL string `yaml:"OKTA_AUTHENTICATION_URL" json:"OKTAAUTHENTICATIONURL"`
	// Workflow Store
	WORKFLOW_CONFIG_VARIABLES map[string]string `yaml:"WORKFLOW_CONFIG_VARIABLES" json:"WORKFLOWCONFIGVARIABLES"`

	// LLM
	///////
	LLM_ADDRESS            string `yaml:"LLM_ADDRESS" json:"LLMADDRESS"`
	MODELS_CONFIG_LOCATION string `yaml:"MODELS_CONFIG_LOCATION" json:"MODELSCONFIGLOCATION"`
	LLM_API_KEY            string `yaml:"LLM_API_KEY" json:"LLMAPIKEY"`

	// Exec
	///////
	EXEC_ADDRESS string `yaml:"EXEC_ADDRESS" json:"EXECADDRESS"`
	EXEC_ID      string `yaml:"EXEC_ID" json:"EXECID"`
	EXEC_API_KEY string `yaml:"EXEC_API_KEY" json:"EXECAPIKEY"`
	// Python executable name
	PYTHON_EXECUTABLE string `yaml:"PYTHON_EXECUTABLE" json:"PYTHONEXECUTABLE"`
	BASH_EXECUTABLE   string `yaml:"BASH_EXECUTABLE" json:"BASHEXECUTABLE"`
	// File transfer
	WATCH_FOLDER_PATH              string `yaml:"WATCH_FOLDER_PATH" json:"WATCHFOLDERPATH"`
	MILLISECONDS_SINCE_LAST_CHANGE int    `yaml:"MILLISECONDS_SINCE_LAST_CHANGE" json:"MILLISECONDSSINCELASTCHANGE"`
	// Agent connection
	AGENT_ENDPOINT string `yaml:"AGENT_ENDPOINT" json:"AGENTENDPOINT"`

	// KVDB
	///////
	KVDB_ADDRESS   string `yaml:"KVDB_ADDRESS" json:"KVDBADDRESS"`
	KVDB_API_KEY   string `yaml:"KVDB_API_KEY" json:"KVDBAPIKEY"`
	KVDB_PATH      string `yaml:"KVDB_PATH" json:"KVDBPATH"`
	KVDB_IN_MEMORY bool   `yaml:"KVDB_IN_MEMORY" json:"KVDBINMEMORY"`

	// Flowkit
	//////////
	FLOWKIT_ADDRESS string `yaml:"FLOWKIT_ADDRESS" json:"FLOWKITADDRESS"`
	FLOWKIT_API_KEY string `yaml:"FLOWKIT_API_KEY" json:"FLOWKITAPIKEY"`
	// Connections to other Modules
	LLM_HANDLER_ENDPOINT  string `yaml:"LLM_HANDLER_ENDPOINT" json:"LLMHANDLERENDPOINT"`
	KNOWLEDGE_DB_ENDPOINT string `yaml:"KNOWLEDGE_DB_ENDPOINT" json:"KNOWLEDGEDBENDPOINT"`
	GRAPHDB_ADDRESS       string `yaml:"GRAPHDB_ADDRESS" json:"GRAPHDBADDRESS"`
	GRAPHDB_API_KEY       string `yaml:"GRAPHDB_API_KEY" json:"GRAPHDBAPIKEY"`
	WIKI_ADDRESS          string `yaml:"WIKI_ADDRESS" json:"WIKIADDRESS"`
	WIKI_API_KEY          string `yaml:"WIKI_API_KEY" json:"WIKIAPIKEY"`
	WIKI_TRANSPORT        string `yaml:"WIKI_TRANSPORT" json:"WIKITRANSPORT"`
	WIKI_DB_PATH          string `yaml:"WIKI_DB_PATH" json:"WIKIDBPATH"`
	WIKI_API_PORT         int    `yaml:"WIKI_API_PORT" json:"WIKIAPIPORT"`
	WIKI_LLM_PROVIDER     string `yaml:"WIKI_LLM_PROVIDER" json:"WIKILLMPROVIDER"`
	WIKI_LLM_API_KEY      string `yaml:"WIKI_LLM_API_KEY" json:"WIKILLMAPIKEY"`
	WIKI_LLM_MODEL        string `yaml:"WIKI_LLM_MODEL" json:"WIKILLMMODEL"`
	WIKI_LLM_MAX_TOKENS   int    `yaml:"WIKI_LLM_MAX_TOKENS" json:"WIKILLMMAXTOKENS"`
	WIKI_LLM_ENDPOINT     string `yaml:"WIKI_LLM_ENDPOINT" json:"WIKILLMENDPOINT"`
	QDRANT_HOST           string `yaml:"QDRANT_HOST" json:"QDRANTHOST"`
	QDRANT_PORT           int    `yaml:"QDRANT_PORT" json:"QDRANTPORT"`
	QDRANT_API_KEY        string `yaml:"QDRANT_API_KEY" json:"QDRANTAPIKEY"`
	// Connections to external services
	MONGODB_CS string `yaml:"MONGODB_CS" json:"MONGODBCS"`

	// Flowkit Python
	/////////////////
	FLOWKIT_PYTHON_ADDRESS string `yaml:"FLOWKIT_PYTHON_ADDRESS" json:"FLOWKITPYTHONADDRESS"`
	FLOWKIT_PYTHON_API_KEY string `yaml:"FLOWKIT_PYTHON_API_KEY" json:"FLOWKITPYTHONAPIKEY"`

	// Database Settings
	////////////////////
	GRAPHDB_ADDR      string `yaml:"GRAPHDB_ADDR" json:"GRAPHDBADDR"`
	GRAPHDB_DATA_PATH string `yaml:"GRAPHDB_DATA_PATH" json:"GRAPHDBDATAPATH"`

	// Proxy / data-shield
	//////////////////////
	// datashield uses these to determine which addresses to listen on for incoming proxy requests
	PROXY_ADMIN_ADDRESS   string `yaml:"PROXY_ADMIN_ADDRESS" json:"PROXYADMINADDRESS"`
	PROXY_GRAPHDB_ADDRESS string `yaml:"PROXY_GRAPHDB_ADDRESS" json:"PROXYGRAPHDBADDRESS"`
	PROXY_QDRANT_ADDRESS  string `yaml:"PROXY_QDRANT_ADDRESS" json:"PROXYQDRANTADDRESS"`
	// other services use this to determine what address to call to get encrypted graphdb data from
	GRAPHDB_ADDRESS_ENCRYPTED string `yaml:"GRAPHDB_ADDRESS_ENCRYPTED" json:"GRAPHDBADDRESSENCRYPTED"`
	// other services use these to determine what address to call to get encrypted qdrant data from
	QDRANT_HOST_ENCRYPTED string `yaml:"QDRANT_HOST_ENCRYPTED" json:"QDRANTHOSTENCRYPTED"`
	QDRANT_PORT_ENCRYPTED int    `yaml:"QDRANT_PORT_ENCRYPTED" json:"QDRANTPORTENCRYPTED"`

	// Legacy Port definitions (are overwritten by the new ADDRESS variables)
	/////////////////////////////////////////////////////////////////////////
	AGENT_PORT                  string `yaml:"AGENT_PORT" json:"AGENTPORT"`
	EXTERNALFUNCTIONS_GRPC_PORT string `yaml:"EXTERNALFUNCTIONS_GRPC_PORT" json:"EXTERNALFUNCTIONSGRPCPORT"`
	WEBSERVER_PORT              string `yaml:"WEBSERVER_PORT" json:"WEBSERVERPORT"`
	WEBSERVER_PORT_EXEC         string `yaml:"WEBSERVER_PORT_EXEC" json:"WEBSERVERPORTEXEC"`
}

// FlowkitConnections contains the configuration for connecting to the FlowKit server.
type FlowkitConnection struct {
	URL     string `yaml:"URL" json:"URL"`        // URL of the FlowKit server
	API_KEY string `yaml:"API_KEY" json:"APIKEY"` // API key for the FlowKit server
}

// Initialize conifg dict
var GlobalConfig *Config

// flagStringSlice is a custom flag type for string slices.
type flagStringSlice []string

// String returns a string representation of the flagStringSlice.
//
// Returns:
//   - string: The string representation of the flagStringSlice.
func (fss *flagStringSlice) String() string {
	return fmt.Sprintf("%v", *fss)
}

// Set sets the value of the flagStringSlice.
//
// Parameters:
//   - value: The value to set.
//
// Returns:
//   - error: An error if there was an issue setting the value.
func (fss *flagStringSlice) Set(value string) error {
	*fss = strings.Split(value, ",")
	return nil
}
