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
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"gopkg.in/yaml.v2"
)

////////////////////////////////////////////
// Standard Config init for Synopsys AI Tools Go Modules
////////////////////////////////////////////

// InitConfig initializes the configuration for the Synopsys AI Tools service.
//
// Parameters:
//   - requiredProperties: The list of required properties.
//   - optionalDefaultValues: The map of optional properties and their default values.
func InitConfig(requiredProperties []string, optionalDefaultValues map[string]interface{}) {
	// Get config file location
	// 1st option: read from environment variable
	configFile := os.Getenv("SNPS_AI_CONFIG_PATH")
	if configFile == "" {
		// 2nd option: read from default location... root directory
		configFile = "config.yaml"
	}

	// Get config properties from CLI
	err := CreateUpdateConfigFileFromCLI(configFile)
	if err != nil {
		pan := writeStringToFile("error in creating and/or updating configuration file from command line:")
		if pan != nil {
			panic(pan)
		}
		pan = writeInterfaceToFile(err)
		if pan != nil {
			panic(pan)
		}
		panic(err)
	}

	// Initialize config From File
	err = InitGlobalConfigFromFile(configFile, requiredProperties, optionalDefaultValues)
	if err != nil {
		pan := writeStringToFile("error in reading configuration values from configuration file:")
		if pan != nil {
			panic(pan)
		}
		pan = writeInterfaceToFile(err)
		if pan != nil {
			panic(pan)
		}
		panic(err)
	}

	// Optionally retrieve secrets from Azure Key Vault with Managed Identity (only works inside Azure Services)
	if GlobalConfig.EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT {
		// Validate the required properties for Azure Key Vault are set
		err = ValidateConfig(*GlobalConfig, []string{"AZURE_KEY_VAULT_NAME", "AZURE_MANAGED_IDENTITY_ID"})
		if err != nil {
			pan := writeStringToFile("error in validating the mandatory configuration values for extracting configuration from Azure Key Vault:")
			if pan != nil {
				panic(pan)
			}
			pan = writeInterfaceToFile(err)
			if pan != nil {
				panic(pan)
			}
			panic(err)
		}

		// Initialize the config from Azure Key Vault
		err = InitGlobalConfigFromAzureKeyVault()
		if err != nil {
			pan := writeStringToFile("error in retrieving configuration values from Azure Key Vault:")
			if pan != nil {
				panic(pan)
			}
			pan = writeInterfaceToFile(err)
			if pan != nil {
				panic(pan)
			}
			panic(err)
		}
	}

	// Validate mandatory config properties
	err = ValidateConfig(*GlobalConfig, requiredProperties)
	if err != nil {
		pan := writeStringToFile("error in validating configuration variables:")
		if pan != nil {
			panic(pan)
		}
		pan = writeInterfaceToFile(err)
		if pan != nil {
			panic(pan)
		}
		panic(err)
	}
}

//////////////////////////////////////////
// Read Config variables from Config file
//////////////////////////////////////////

// InitGlobalConfigFromFile reads the configuration file and initializes the Config object.
//
// Parameters:
//   - fileName: The name of the configuration file.
//   - requiredProperties: The list of required properties.
//   - optionalDefaultValues: The map of optional properties and their default values.
//
// Returns:
//   - err: An error if there was an issue initializing the configuration.
func InitGlobalConfigFromFile(fileName string, requiredProperties []string, optionalDefaultValues map[string]interface{}) (err error) {
	var config Config
	configResult, err := readYaml(fileName, config)
	if err != nil {
		return err
	}

	// Assign to global config
	GlobalConfig = &configResult

	// Set optional properties if missing
	err = defineOptionalProperties(GlobalConfig, optionalDefaultValues)
	if err != nil {
		return err
	}

	return nil
}

// readYaml reads the yaml specified in `fileName` parameter and saves it to `config_struct`
//
// Parameters:
//   - fileName: The name of the configuration file.
//   - configStruct: Struct with the parameters of the YAML to read.
//
// Returns:
//   - extractedConfigStruct: The extracted configuration struct.
//   - err: An error if there was an issue reading the YAML file.
func readYaml(fileName string, configStruct Config) (extractedConfigStruct Config, err error) {
	// Read the YAML file into a byte slice
	data, err := os.ReadFile(fileName)
	if err != nil {
		message := "config.yaml file is missing from directory or not accessible"
		return Config{}, errors.New(message)
	}

	// Unmarshal the YAML data into the config object
	err = yaml.Unmarshal(data, &configStruct)
	if err != nil {
		// Create a new Config struct with field names and types
		configStruct := reflect.TypeOf(configStruct)
		var fieldList []string
		for i := 0; i < configStruct.NumField(); i++ {
			field := configStruct.Field(i)
			fieldList = append(fieldList, fmt.Sprintf("%q: %s", field.Name, field.Type.String()))
		}

		// Define error message
		message := fileName + " contains incorrect content. The allowed values are as follows: {"
		message += strings.Join(fieldList, ",")
		message = message + "}"

		// Write message and error to error file
		return Config{}, errors.New(message)
	}
	return configStruct, nil
}

// defineOptionalProperties sets optional properties for the configuration.
//
// Parameters:
//   - config: The configuration object to validate.
//   - optionalDefaultValues: The map of optional properties and their default values.
//
// Returns:
//   - err: An error if there was an issue setting the optional properties.
func defineOptionalProperties(config *Config, optionalDefaultValues map[string]interface{}) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			// Write message to error file
			message := fmt.Sprintf("Error in defineOptionalProperties: %v", r)
			err = errors.New(message)
		}
	}()

	// Iterate over the optional default values
	for key, defaultValue := range optionalDefaultValues {
		// Use reflection to check if the field exists and is set to its zero value
		fieldValue := reflect.ValueOf(config).Elem().FieldByName(key)
		if fieldValue.IsValid() && isZeroValue(fieldValue) {
			// Set the default value using reflection
			if reflect.TypeOf(defaultValue) == fieldValue.Type() {
				fieldValue.Set(reflect.ValueOf(defaultValue))
			} else {
				// Handle type mismatch
				message := fmt.Sprintf("Type mismatch for key '%s': expected %v, got %v", key, fieldValue.Type(), reflect.TypeOf(defaultValue))
				return errors.New(message)
			}
		}
	}

	return nil
}

// isZeroValue checks if a reflect.Value is zero for its type.
//
// Parameters:
//   - v: The reflect.Value to check.
//
// Returns:
//   - bool: True if the value is zero, false otherwise.
func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

/////////////////////////////////////////
// Create or update Config file from CLI
/////////////////////////////////////////

// CreateUpdateConfigFileFromCLI reads and updates the configuration file based on command-line arguments.
//
// Parameters:
//   - fileName: The name of the configuration file.
//
// Returns:
//   - err: An error if there was an issue creating or updating the configuration file.
func CreateUpdateConfigFileFromCLI(fileName string) (err error) {
	// Checking for any command-line arguments
	if len(os.Args) == 1 {
		log.Println("No command line options given; full config will be retrieved from existing config.yaml file and/or Azure Key Vault.")
		return
	}

	// Create a new config to store command-line options
	cliConfig := Config{}

	// Use reflection to create flags for each field in Config
	createFlags(reflect.ValueOf(&cliConfig).Elem(), "")

	// Parse the flags
	flag.Parse()

	// Track which flags were actually set
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	// Checking if config.yaml file already exists
	_, err = os.Stat(fileName)

	if os.IsNotExist(err) {
		// If it doesn't exist, create a new one
		log.Println("config.yaml file does not exist. Creating a new one with provided CLI values...")

		file, _ := yaml.Marshal(cliConfig)
		_ = os.WriteFile(fileName, file, 0644)
	} else {
		// If it does exist, open and append to it
		log.Println("config.yaml file exists. Appending command line options with provided CLI values...")

		file, _ := os.ReadFile(fileName)
		config := Config{}
		err := yaml.Unmarshal(file, &config)
		if err != nil {
			message := fmt.Sprintf("Error in yaml.Unmarshal: %v", err)
			return errors.New(message)
		}

		// Use reflection to update fields that were actually set on the command line
		updateConfigWithCLI(&config, &cliConfig, setFlags, "")

		// Write back to the file
		file, _ = yaml.Marshal(config)
		_ = os.WriteFile(fileName, file, 0644)
	}

	return nil
}

// CreateFlags initializes command-line flags for configuration.
//
// Parameters:
//   - val: The value to create flags for.
//   - prefix: The prefix to use for the flags.
func createFlags(val reflect.Value, prefix string) {
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := strings.ToUpper(prefix + field.Name)
		switch field.Type.Kind() {
		case reflect.String:
			flag.StringVar(val.Field(i).Addr().Interface().(*string), name, "", "config option")
		case reflect.Int:
			flag.IntVar(val.Field(i).Addr().Interface().(*int), name, 0, "config option")
		case reflect.Bool:
			flag.BoolVar(val.Field(i).Addr().Interface().(*bool), name, false, "config option")
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				flag.Var((*flagStringSlice)(val.Field(i).Addr().Interface().(*[]string)), name, "config option")
			}
		case reflect.Struct:
			createFlags(val.Field(i), name+"_")
		}
	}
}

// updateConfigWithCLI updates the config with CLI values, but only for flags that were actually set
//
// Parameters:
//   - config: The existing config to update
//   - cliConfig: The CLI config with parsed values
//   - setFlags: Map of flag names that were actually set on command line
//   - prefix: The current prefix for nested structs
func updateConfigWithCLI(config interface{}, cliConfig interface{}, setFlags map[string]bool, prefix string) {
	valConfig := reflect.ValueOf(config).Elem()
	valCli := reflect.ValueOf(cliConfig).Elem()
	t := reflect.TypeOf(cliConfig).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		flagName := strings.ToUpper(prefix + field.Name)

		configField := valConfig.Field(i)
		cliField := valCli.Field(i)

		switch field.Type.Kind() {
		case reflect.Struct:
			// Recursively handle nested structs
			updateConfigWithCLI(configField.Addr().Interface(), cliField.Addr().Interface(), setFlags, flagName+"_")
		default:
			// Only update if this flag was actually set on the command line
			if setFlags[flagName] {
				configField.Set(cliField)
			}
		}
	}
}

/////////////////////////////////////////////////
// Extract Config variables from Azure Key Vault
/////////////////////////////////////////////////

// InitGlobalConfigFromAzureKeyVault extracts the configuration from Azure Key Vault.
// It iterates over all secrets in the key vault and if the secret name matches a field in the Config struct,
// it sets the field to the value of the secret.
//
// Returns:
//   - err: An error if there was an issue extracting the configuration.
func InitGlobalConfigFromAzureKeyVault() (err error) {
	// log
	log.Println("Extracting configuration from Azure Key Vault...")

	// get environment variables
	azureManagedIdentity := os.Getenv(GlobalConfig.AZURE_MANAGED_IDENTITY_ID)
	azureKeyVaultName := os.Getenv(GlobalConfig.AZURE_KEY_VAULT_NAME)

	// check if all required environment variables are set
	if azureManagedIdentity == "" {
		return fmt.Errorf("environment variable for %v is not set", azureManagedIdentity)
	}
	if azureKeyVaultName == "" {
		return fmt.Errorf("environment variable for %v is not set", azureKeyVaultName)
	}

	// create key vault URL
	keyVaultUrl := fmt.Sprintf("https://%s.vault.azure.net/", azureKeyVaultName)

	// create Managed Identity credential
	cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(azureManagedIdentity),
	})
	if err != nil {
		return fmt.Errorf("failed to get Managed Identity credential: %w", err)
	}

	// Test the managed id by getting a token
	scope := "https://vault.azure.net/.default" // Scope for Azure Key Vault
	_, err = cred.GetToken(context.TODO(), policy.TokenRequestOptions{
		Scopes: []string{scope},
	})
	if err != nil {
		return fmt.Errorf("failed to get token from managed ID: %w", err)
	}

	// Reflect on the struct
	GlobalConfigValue := reflect.ValueOf(GlobalConfig).Elem()
	GlobalConfigType := GlobalConfigValue.Type()

	// create azsecrets client
	clientSecrets, err := azsecrets.NewClient(keyVaultUrl, cred, nil)
	if err != nil {
		return err
	}

	// list all secrets
	pagerSecerts := clientSecrets.NewListSecretPropertiesPager(nil)
	// iterate over all secrets
	for pagerSecerts.More() {
		page, err := pagerSecerts.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("error, when iterating over Azure key vault secrets: %v", err)
		}
		for _, secret := range page.Value {
			// iterate over all fields in the struct
			for i := 0; i < GlobalConfigValue.NumField(); i++ {
				// Get the YAML tag
				fieldType := GlobalConfigType.Field(i)
				yamlTag := fieldType.Tag.Get("json")

				// Check if the field name matches the target field name
				if yamlTag == secret.ID.Name() {
					// Get the key value
					resp, err := clientSecrets.GetSecret(context.TODO(), secret.ID.Name(), "", nil)
					if err != nil {
						return fmt.Errorf("error while getting value for secret '%v': %v", yamlTag, err)
					}
					// Apply secret value to config field
					err = applySecretValueToConfig(GlobalConfig, yamlTag, *resp.Value)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// applySecretValueToConfig applies a secret value (identified by its JSON tag) to the corresponding field in the Config struct.
//
// Parameters:
//   - config: The configuration object to update.
//   - secretName: The JSON tag name of the secret.
//   - secretValue: The string value of the secret.
//
// Returns:
//   - err: An error if there was an issue applying the secret value.
func applySecretValueToConfig(config *Config, secretName string, secretValue string) error {
	configValue := reflect.ValueOf(config).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)
		yamlTag := fieldType.Tag.Get("json")

		if yamlTag != secretName {
			continue
		}

		if !field.CanSet() {
			return fmt.Errorf("field '%v' is not settable", secretName)
		}

		switch field.Kind() {
		case reflect.String:
			value := secretValue
			unquoted, err := strconv.Unquote(`"` + strings.ReplaceAll(value, `"`, `\"`) + `"`)
			if err == nil {
				value = unquoted
			}
			field.SetString(value)
		case reflect.Bool:
			if secretValue == "" {
				field.SetBool(false)
				break
			}
			value, err := strconv.ParseBool(secretValue)
			if err != nil {
				return fmt.Errorf("error in strconv.ParseBool for secret '%v' with value '%v': %v", secretName, secretValue, err)
			}
			field.SetBool(value)
		case reflect.Int:
			if secretValue == "" {
				field.SetInt(0)
				break
			}
			value, err := strconv.Atoi(secretValue)
			if err != nil {
				return fmt.Errorf("error in strconv.Atoi for secret '%v' with value '%v': %v", secretName, secretValue, err)
			}
			field.SetInt(int64(value))
		case reflect.Slice:
			if secretValue == "" {
				field.Set(reflect.MakeSlice(field.Type(), 0, 0))
				break
			}
			switch field.Type().Elem().Kind() {
			case reflect.String:
				var value []string
				err := json.Unmarshal([]byte(secretValue), &value)
				if err != nil {
					return fmt.Errorf("error in json.Unmarshal []string for secret '%v' with value '%v': %v", secretName, secretValue, err)
				}
				field.Set(reflect.ValueOf(value))
			case reflect.Int:
				var value []int
				err := json.Unmarshal([]byte(secretValue), &value)
				if err != nil {
					return fmt.Errorf("error in json.Unmarshal []int for secret '%v' with value '%v': %v", secretName, secretValue, err)
				}
				field.Set(reflect.ValueOf(value))
			default:
				return fmt.Errorf("unsupported slice element type '%v' for secret '%v' with value '%v'", field.Type().Elem().Kind(), secretName, secretValue)
			}
		case reflect.Map:
			if secretValue == "" {
				field.Set(reflect.ValueOf(map[string]string{}))
				break
			}
			var value map[string]string
			err := json.Unmarshal([]byte(secretValue), &value)
			if err != nil {
				return fmt.Errorf("error in json.Unmarshal map[string]string for secret '%v' with value '%v': %v", secretName, secretValue, err)
			}
			field.Set(reflect.ValueOf(value))
		default:
			return fmt.Errorf("unsupported field type '%v' for secret '%v' with value '%v'", field.Kind(), secretName, secretValue)
		}

		return nil
	}

	return nil
}

///////////////////////
// Helper Functions
///////////////////////

// ValidateConfig checks for mandatory entries in the configuration and validates chosen models.
//
// Parameters:
//   - config: The configuration object to validate.
//   - requiredProperties: The list of required properties.
//
// Returns:
//   - err: An error if there was an issue validating the configuration.
func ValidateConfig(config Config, requiredProperties []string) (err error) {
	// Check if all mandatory properties are present
	configValue := reflect.ValueOf(config)

	for _, property := range requiredProperties {
		field := configValue.FieldByName(property)

		if !field.IsValid() || field.IsZero() {
			return fmt.Errorf("config.yaml is missing mandatory property '%v': ", property)
		}
	}

	return nil
}

// GetGlobalConfigAsJSON returns the global configuration as a JSON string.
//
// Returns:
//   - string: The global configuration as a JSON string.
func GetGlobalConfigAsJSON() string {
	jsonData, err := json.Marshal(GlobalConfig)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

///////////////////////
// Error file creator
///////////////////////

// writeInterfaceToFile writes interface data to an error log file.
//
// Parameters:
//   - data: The data to write to the file.
//
// Returns:
//   - error: An error if there was an issue writing to the file.
func writeInterfaceToFile(data interface{}) error {
	var file *os.File
	var err error

	// Get file name
	filename := "error.log"

	// Create file
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

	// Write to file
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Add time
	timestamp := timeToString(time.Now())

	// Write to file
	line := fmt.Sprintf("%s: %s\n", timestamp, string(jsonData))
	_, err = file.Write([]byte(line))
	if err != nil {
		return err
	}

	return nil
}

// writeStringToFile writes string data to an error log file.
//
// Parameters:
//   - data: The data to write to the file.
//
// Returns:
//   - error: An error if there was an issue writing to the file.
func writeStringToFile(data string) error {
	var file *os.File
	var err error

	// Get file name
	filename := "error.log"

	// Add time
	timestamp := timeToString(time.Now())

	// Change string
	data = timestamp + ": " + data

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		// File does not exist, create a new file
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
	} else {
		// File exists, open it for appending
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// Append data to file with a new line
	_, err = fmt.Fprintln(file, data)
	return err
}

// timeToString converts a time.Time value to a formatted string.
//
// Parameters:
//   - t: The time.Time value to convert.
//
// Returns:
//   - string: The formatted string.
func timeToString(t time.Time) string {
	layout := "2006-01-02 15:04:05.000"
	return t.Format(layout)
}

////////////////////////////
// Legacy Config Converters
////////////////////////////

// HandleLegacyPortDefinition checks if the address is set, and if not, uses the legacy port to define the web server address.
// If both are empty, it returns an error.
//
// Parameters:
//   - address: The address to use for the web server.
//   - legacyPort: The legacy port to use if the address is not set.
//
// Returns:
//   - webserverAddress: The web server address to use.
//   - err: An error if both address and legacy port are empty.
func HandleLegacyPortDefinition(configAddress string, legacyPort string) (webserverAddress string, err error) {
	if configAddress != "" {
		return configAddress, nil
	}
	if legacyPort != "" {
		return "0.0.0.0:" + legacyPort, nil
	}
	return "", fmt.Errorf("both address and legacy port are empty")
}
