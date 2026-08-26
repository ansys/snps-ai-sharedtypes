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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/semver"
)

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type StdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}

var imageName string
var apiKey string

func init() {
	flag.StringVar(&imageName, "imagename", "ghcr.io/ansys/aali-graphdb:edge", "Name of the snps-ai-graphdb image to run the tests against")
	flag.StringVar(&apiKey, "apikey", "", "Set the tests to use an API key")
}

// if you are using Colima see https://golang.testcontainers.org/system_requirements/using_colima/
func getTestClient(t *testing.T) *Client {
	ctx := context.Background()

	fmt.Printf("Running test against image: %q\n", imageName)

	env := map[string]string{"RUST_LOG": "debug"}
	if apiKey != "" {
		env["SNPS_AI_GRAPHDB_API_KEY"] = apiKey
	}

	req := testcontainers.ContainerRequest{
		Image:        imageName,
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(30 * time.Second),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{}},
		},
		Env: env,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      t.TempDir(),
				ContainerFilePath: "/data",
				FileMode:          0o700,
			},
		},
	}
	dbCont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	defer testcontainers.CleanupContainer(t, dbCont)
	require.NoError(t, err)

	port, err := dbCont.MappedPort(ctx, "8080/tcp")
	require.NoError(t, err)
	host, err := dbCont.Host(ctx)
	require.NoError(t, err)

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	httpClient := http.Client{
		Transport: tr,
	}

	address := fmt.Sprintf("http://%s:%s", host, port.Port())
	client, err := NewClient(address, apiKey, &httpClient)
	require.NoError(t, err)
	return client
}

func TestGetHealth(t *testing.T) {
	client := getTestClient(t)
	dbs, err := client.GetHealth()
	require.NoError(t, err)
	assert.Equal(t, true, dbs)
}

func TestGetVersion(t *testing.T) {
	client := getTestClient(t)
	version, err := client.GetVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, version.Version)
	if semver.Compare("v"+version.Version, "v1.0.8") >= 0 {
		// these were introduced in v1.0.8, so cannot assert on them if testing against earlier server version
		assert.NotEmpty(t, version.KuzuVersion)
		assert.NotEmpty(t, version.KuzuStorageVersion)
	}
}

func TestGetDatabases(t *testing.T) {
	client := getTestClient(t)
	dbs, err := client.GetDatabases()
	require.NoError(t, err)
	assert.Equal(t, []string{}, dbs)
}

func TestCreateDatabase(t *testing.T) {
	var warnLogs strings.Builder
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&warnLogs),
		zapcore.WarnLevel,
	))
	defer func() { require.NoError(t, logger.Sync()) }()

	client := getTestClient(t)
	client.logger = logger

	// make sure no dbs
	dbs, err := client.GetDatabases()
	require.NoError(t, err)
	assert.Equal(t, []string{}, dbs)

	// now insert 1
	const NEWDBNAME = "test-create-db"
	err = client.CreateDatabase(NEWDBNAME)
	require.NoError(t, err)

	// check warnings, if appropriate
	ver, err := client.GetVersion()
	require.NoError(t, err)
	if semverComp(ver.Version, createDbPutMinVer) < 0 {
		logs := strings.Split(strings.TrimSpace(warnLogs.String()), "\n")
		assert.Len(t, logs, 1)
		assert.True(t, slices.ContainsFunc(logs, func(log string) bool {
			return strings.Contains(log, "The `POST /databases` method for creating a new DB is deprecated. Upgrade your snps-ai-graphdb server to use the newer `PUT /databases/{name}` method")
		}))
	}

	// now check the dbs again
	dbs, err = client.GetDatabases()
	require.NoError(t, err)
	assert.Equal(t, []string{NEWDBNAME}, dbs)
}

func TestDeleteDatabase(t *testing.T) {
	client := getTestClient(t)

	// insert db
	const NEWDBNAME = "test-delete-db"
	err := client.CreateDatabase(NEWDBNAME)
	require.NoError(t, err)

	// check its there
	dbs, err := client.GetDatabases()
	require.NoError(t, err)
	assert.Equal(t, []string{NEWDBNAME}, dbs)

	// delete it
	err = client.DeleteDatabase(NEWDBNAME)
	require.NoError(t, err)

	// now check the dbs again
	dbs, err = client.GetDatabases()
	require.NoError(t, err)
	assert.Equal(t, []string{}, dbs)
}

func TestGetSchema(t *testing.T) {
	client := getTestClient(t)

	const DBNAME = "test-db"
	err := client.CreateDatabase(DBNAME)
	require.NoError(t, err)

	for _, query := range []string{
		"CREATE NODE TABLE User(name STRING, age INT64, PRIMARY KEY (name))",
		"CREATE NODE TABLE City(name STRING, population INT64, PRIMARY KEY (name))",
		"CREATE REL TABLE LivesIn(FROM User TO City)",
	} {
		_, err := client.CypherQueryWrite(DBNAME, query, nil)
		require.NoError(t, err)
	}

	ver, err := client.GetVersion()
	require.NoError(t, err)
	schema, err := client.GetSchema(DBNAME)
	if semverComp(ver.Version, getSchemaMinVer) < 0 {
		assert.ErrorContains(t, err, "the `GET /databases/{name}/schema` endpoint was introduced in ")
		assert.Empty(t, schema)
	} else {
		assert.NoError(t, err)
		schemaLines := strings.Split(strings.TrimSpace(schema), "\n")
		expectedLines := []string{
			"CREATE NODE TABLE `User` (`name` STRING,`age` INT64, PRIMARY KEY(`name`));",
			"CREATE NODE TABLE `City` (`name` STRING,`population` INT64, PRIMARY KEY(`name`));",
			"CREATE REL TABLE `LivesIn` (FROM `User` TO `City`, MANY_MANY);",
		}
		assert.Len(t, schemaLines, len(expectedLines))
		for _, line := range schemaLines {
			assert.Contains(t, expectedLines, line)
		}
	}
}

func TestReadWriteData(t *testing.T) {
	client := getTestClient(t)
	const DBNAME = "my-db"

	// create db
	err := client.CreateDatabase(DBNAME)
	require.NoError(t, err)

	// write some data in there
	queries := []string{
		// setup schema
		"CREATE NODE TABLE User(name STRING, age INT64, PRIMARY KEY (name))",
		"CREATE NODE TABLE City(name STRING, population INT64, PRIMARY KEY (name))",
		"CREATE REL TABLE Follows(FROM User TO User, since INT64)",
		"CREATE REL TABLE LivesIn(FROM User TO City)",

		// add a few users
		"CREATE (:User {name: 'Adam', age: 30});",
		"CREATE (:User {name: 'Karissa', age: 40});",
		"CREATE (:User {name: 'Zhang', age: 50});",
		"CREATE (:User {name: 'Noura', age: 25});",

		// create a few cities
		"CREATE (:City {name: 'Waterloo', population: 150000});",
		"CREATE (:City {name: 'Kitchener', population: 200000});",
		"CREATE (:City {name: 'Guelph', population: 75000});",

		// add a few follows relationships
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Adam' AND u2.name = 'Karissa' CREATE (u1)-[:Follows {since: 2020}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Adam' AND u2.name = 'Zhang' CREATE (u1)-[:Follows {since: 2020}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Karissa' AND u2.name = 'Zhang' CREATE (u1)-[:Follows {since: 2021}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Zhang' AND u2.name = 'Noura' CREATE (u1)-[:Follows {since: 2022}]->(u2);",

		// add a few lives-in relationships
		"MATCH (u:User), (c:City) WHERE u.name = 'Adam' AND c.name = 'Waterloo' CREATE (u)-[:LivesIn {}]->(c);",
		"MATCH (u:User), (c:City) WHERE u.name = 'Karissa' AND c.name = 'Waterloo' CREATE (u)-[:LivesIn {}]->(c);",
		"MATCH (u:User), (c:City) WHERE u.name = 'Zhang' AND c.name = 'Kitchener' CREATE (u)-[:LivesIn {}]->(c);",
		"MATCH (u:User), (c:City) WHERE u.name = 'Noura' AND c.name = 'Guelph' CREATE (u)-[:LivesIn {}]->(c);",
	}
	for _, query := range queries {
		_, err := client.CypherQueryWrite(DBNAME, query, nil)
		require.NoError(t, err)
	}

	// read it back
	res, err := client.CypherQueryRead(DBNAME, "MATCH (a:User)-[e:Follows]->(b:User) RETURN a.name, e.since, b.name ORDER BY a.name ASC, b.name ASC", nil)
	require.NoError(t, err)
	expected := []map[string]any{
		{"a.name": "Adam", "b.name": "Karissa", "e.since": float64(2020)},
		{"a.name": "Adam", "b.name": "Zhang", "e.since": float64(2020)},
		{"a.name": "Karissa", "b.name": "Zhang", "e.since": float64(2021)},
		{"a.name": "Zhang", "b.name": "Noura", "e.since": float64(2022)},
	}
	assert.Equal(t, expected, res)
}

func TestReadWriteGeneric(t *testing.T) {
	client := getTestClient(t)
	const DBNAME = "my-db-generics"

	// create db
	err := client.CreateDatabase(DBNAME)
	require.NoError(t, err)

	// write some data in there
	type Result struct {
		Result string `json:"result"`
	}
	queries := []string{
		// setup schema
		"CREATE NODE TABLE User(name STRING, age INT64, PRIMARY KEY (name))",
		"CREATE NODE TABLE City(name STRING, population INT64, PRIMARY KEY (name))",
		"CREATE REL TABLE Follows(FROM User TO User, since INT64)",
		"CREATE REL TABLE LivesIn(FROM User TO City)",

		// add a few users
		"CREATE (:User {name: 'Adam', age: 30});",
		"CREATE (:User {name: 'Karissa', age: 40});",
		"CREATE (:User {name: 'Zhang', age: 50});",
		"CREATE (:User {name: 'Noura', age: 25});",

		// create a few cities
		"CREATE (:City {name: 'Waterloo', population: 150000});",
		"CREATE (:City {name: 'Kitchener', population: 200000});",
		"CREATE (:City {name: 'Guelph', population: 75000});",

		// add a few follows relationships
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Adam' AND u2.name = 'Karissa' CREATE (u1)-[:Follows {since: 2020}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Adam' AND u2.name = 'Zhang' CREATE (u1)-[:Follows {since: 2020}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Karissa' AND u2.name = 'Zhang' CREATE (u1)-[:Follows {since: 2021}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Zhang' AND u2.name = 'Noura' CREATE (u1)-[:Follows {since: 2022}]->(u2);",

		// add a few lives-in relationships
		"MATCH (u:User), (c:City) WHERE u.name = 'Adam' AND c.name = 'Waterloo' CREATE (u)-[:LivesIn {}]->(c);",
		"MATCH (u:User), (c:City) WHERE u.name = 'Karissa' AND c.name = 'Waterloo' CREATE (u)-[:LivesIn {}]->(c);",
		"MATCH (u:User), (c:City) WHERE u.name = 'Zhang' AND c.name = 'Kitchener' CREATE (u)-[:LivesIn {}]->(c);",
		"MATCH (u:User), (c:City) WHERE u.name = 'Noura' AND c.name = 'Guelph' CREATE (u)-[:LivesIn {}]->(c);",
	}
	for _, query := range queries {
		_, err := CypherQueryWriteGeneric[Result](client, DBNAME, query, nil)
		require.NoError(t, err)
	}

	type Person struct {
		A     string `json:"a.name"`
		B     string `json:"b.name"`
		Since int64  `json:"e.since"`
	}

	// read it back
	res, err := CypherQueryReadGeneric[Person](client, DBNAME, "MATCH (a:User)-[e:Follows]->(b:User) RETURN a.name, e.since, b.name ORDER BY a.name ASC, b.name ASC", nil)
	require.NoError(t, err)
	expected := []Person{
		{"Adam", "Karissa", 2020},
		{"Adam", "Zhang", 2020},
		{"Karissa", "Zhang", 2021},
		{"Zhang", "Noura", 2022},
	}
	assert.Equal(t, expected, res)
}

func TestReadWriteDataWithParameters(t *testing.T) {
	client := getTestClient(t)
	const DBNAME = "my-db-with-params"

	// create db
	err := client.CreateDatabase(DBNAME)
	require.NoError(t, err)

	// create schema
	queries := []string{
		// setup schema
		"CREATE NODE TABLE User(name STRING, age INT64, PRIMARY KEY (name))",
		"CREATE REL TABLE Follows(FROM User TO User, since INT64)",

		// add a few follows relationships
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Adam' AND u2.name = 'Karissa' CREATE (u1)-[:Follows {since: 2020}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Adam' AND u2.name = 'Zhang' CREATE (u1)-[:Follows {since: 2020}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Karissa' AND u2.name = 'Zhang' CREATE (u1)-[:Follows {since: 2021}]->(u2);",
		"MATCH (u1:User), (u2:User) WHERE u1.name = 'Zhang' AND u2.name = 'Noura' CREATE (u1)-[:Follows {since: 2022}]->(u2);",
	}
	for _, query := range queries {
		_, err := client.CypherQueryWrite(DBNAME, query, nil)
		require.NoError(t, err)
	}

	// insert user data w/ parameters
	userQuery := "CREATE (:User {name: $name, age: $age});"
	users := []ParameterMap{
		{"name": StringValue("Adam"), "age": Int64Value(30)},
		{"name": StringValue("Karissa"), "age": Int64Value(40)},
		{"name": StringValue("Zhang"), "age": Int64Value(50)},
		{"name": StringValue("Noura"), "age": Int64Value(25)},
	}
	for _, user := range users {
		_, err := client.CypherQueryWrite(DBNAME, userQuery, user)
		require.NoError(t, err)
	}

	// insert relationships data w/ parameters
	followsQuery := "MATCH (u1:User{name:$from}), (u2:User{name:$to}) CREATE (u1)-[:Follows {since: $since}]->(u2);"
	follows := []ParameterMap{
		{"from": StringValue("Adam"), "to": StringValue("Karissa"), "since": Int64Value(2020)},
		{"from": StringValue("Adam"), "to": StringValue("Zhang"), "since": Int64Value(2020)},
		{"from": StringValue("Karissa"), "to": StringValue("Zhang"), "since": Int64Value(2021)},
		{"from": StringValue("Zhang"), "to": StringValue("Noura"), "since": Int64Value(2022)},
	}
	for _, follow := range follows {
		_, err := client.CypherQueryWrite(DBNAME, followsQuery, follow)
		require.NoError(t, err)
	}

	// now read it back w/ parameters
	query := "MATCH (a:User)-[e:Follows]->(b:User) WHERE a.name = $from OR e.since > $after RETURN a.name, e.since, b.name ORDER BY a.name ASC, b.name ASC"
	params := ParameterMap{
		"from": StringValue("Adam"), "after": Int64Value(2021),
	}
	res, err := client.CypherQueryRead(DBNAME, query, params)
	require.NoError(t, err)
	expected := []map[string]any{
		{"a.name": "Adam", "b.name": "Karissa", "e.since": float64(2020)},
		{"a.name": "Adam", "b.name": "Zhang", "e.since": float64(2020)},
		{"a.name": "Zhang", "b.name": "Noura", "e.since": float64(2022)},
	}
	assert.Equal(t, expected, res)
}

type ParamsStruct struct {
	name    string
	boolean bool
	date    civil.Date
	age     uint64
}

func (ps ParamsStruct) AsParameters() (map[string]Value, error) {
	return map[string]Value{
		"name":    StringValue(ps.name),
		"boolean": BoolValue(ps.boolean),
		"date":    DateValue(ps.date),
		"age":     UInt64Value(ps.age),
	}, nil
}

func TestParametersStruct(t *testing.T) {
	client := getTestClient(t)
	const DBNAME = "my-db-from-paramstruct"

	// create db
	err := client.CreateDatabase(DBNAME)
	require.NoError(t, err)

	// create schema
	_, err = client.CypherQueryWrite(DBNAME, "CREATE NODE TABLE User(name STRING, boolean BOOL, date DATE, age INT64, PRIMARY KEY (name))", nil)
	require.NoError(t, err)

	// insert user data w/ parameters
	userQuery := "CREATE (:User {name: $name, age: $age, boolean: $boolean, date: $date});"
	users := []ParamsStruct{
		{"Adam", true, civil.Date{Year: 2024, Month: time.August, Day: 22}, 30},
		{"Karissa", true, civil.Date{Year: 2022, Month: time.January, Day: 7}, 40},
		{"Zhang", false, civil.Date{Year: 2025, Month: time.July, Day: 3}, 50},
		{"Noura", true, civil.Date{Year: 2023, Month: time.October, Day: 15}, 25},
	}
	for _, user := range users {
		_, err := client.CypherQueryWrite(DBNAME, userQuery, user)
		require.NoError(t, err)
	}

	// now read it back w/ parameters
	query := "MATCH (u:User) RETURN u.* ORDER BY u.name ASC"
	res, err := client.CypherQueryRead(DBNAME, query, nil)
	require.NoError(t, err)
	expected := []map[string]any{
		{"u.name": "Adam", "u.boolean": true, "u.date": "2024-08-22", "u.age": float64(30)},
		{"u.name": "Karissa", "u.boolean": true, "u.date": "2022-01-07", "u.age": float64(40)},
		{"u.name": "Noura", "u.boolean": true, "u.date": "2023-10-15", "u.age": float64(25)},
		{"u.name": "Zhang", "u.boolean": false, "u.date": "2025-07-03", "u.age": float64(50)},
	}
	assert.Equal(t, expected, res)
}

func TestErrorsReturned(t *testing.T) {
	client := getTestClient(t)
	const DBNAME = "test-errors"

	err := client.CreateDatabase(DBNAME)
	require.NoError(t, err)

	query := "not a real cypher query"
	pat := regexp.MustCompile(`Query execution failed:[\s\S]*` + query)

	t.Run("Read", func(t *testing.T) {
		_, err = client.CypherQueryRead(DBNAME, query, nil)
		require.Error(t, err)
		assert.True(t, pat.MatchString(fmt.Sprint(err)))
	})
	t.Run("Write", func(t *testing.T) {
		_, err = client.CypherQueryWrite(DBNAME, query, nil)
		require.Error(t, err)
		assert.True(t, pat.MatchString(fmt.Sprint(err)))
	})
}

func TestParameterMapJson(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	pMap := ParameterMap{"myStr": StringValue("hi"), "myBool": BoolValue(true), "myInt": Int16Value(23)}
	jsonParamMap, err := json.Marshal(pMap)
	require.NoError(err)

	expected := map[string]any{
		"myStr":  map[string]any{"String": "hi"},
		"myBool": map[string]any{"Bool": true},
		"myInt":  map[string]any{"Int16": float64(23)},
	}
	var unmarshalledMap map[string]any
	require.NoError(json.Unmarshal(jsonParamMap, &unmarshalledMap))
	assert.Equal(expected, unmarshalledMap)

	var unmarshalledParamMap ParameterMap
	require.NoError(json.Unmarshal(jsonParamMap, &unmarshalledParamMap))
	assert.Equal(pMap, unmarshalledParamMap)
}

func TestRequiresApiKey(t *testing.T) {
	if apiKey == "" {
		t.Skip("this test is only relevant for tests with API keys configured")
	}

	require := require.New(t)
	assert := assert.New(t)

	client := getTestClient(t)
	version, err := client.GetVersion()
	require.NoError(err)

	if semver.Compare("v"+version.Version, "v1.2.2") < 0 {
		t.Skip("API key auth was not released prior to server version v1.2.2")
	}

	// try to make a call without api key and make sure you get an error
	_, err = client.WithApiKey("").GetDatabases()
	assert.EqualError(err, "unexpected status code: 401")
}

func TestExport(t *testing.T) {
	client := getTestClient(t)

	const Db = "test-db"
	require.NoError(t, client.CreateDatabase(Db))

	// put some data in there
	for _, query := range []string{
		"CREATE NODE TABLE User(name STRING, age INT64, PRIMARY KEY (name))",
		"CREATE (:User {name: 'Adam', age: 30});",
		"CREATE (:User {name: 'Karissa', age: 40});",
		"CREATE (:User {name: 'Zhang', age: 50});",
		"CREATE (:User {name: 'Noura', age: 25});",
	} {
		_, err := client.CypherQueryWrite(Db, query, nil)
		require.NoError(t, err)
	}

	testCases := map[string]struct {
		opts []GraphDbExportOpt
		ext  string
	}{
		"NoOpts":             {[]GraphDbExportOpt{}, "parquet"},
		"Parquet":            {[]GraphDbExportOpt{WithFormatParquet{}}, "parquet"},
		"Csv":                {[]GraphDbExportOpt{WithFormatCsv{}}, "csv"},
		"DefaultCompression": {[]GraphDbExportOpt{WithCompressionDefault{}}, "parquet"},
		"BestCompression":    {[]GraphDbExportOpt{WithCompressionBest{}}, "parquet"},
		"FastCompression":    {[]GraphDbExportOpt{WithCompressionFast{}}, "parquet"},
		"ParquetBest":        {[]GraphDbExportOpt{WithFormatParquet{}, WithCompressionBest{}}, "parquet"},
		"CsvFast":            {[]GraphDbExportOpt{WithCompressionFast{}, WithFormatCsv{}}, "csv"},
	}

	for mode, exportBuilder := range map[string]func(*testing.T) io.ReadWriter{
		"File": func(t *testing.T) io.ReadWriter {
			tmpdir := t.TempDir()
			exportFile, err := os.CreateTemp(tmpdir, "*.tar.gz")
			require.NoError(t, err)
			return exportFile
		},
		"Buffer": func(t *testing.T) io.ReadWriter {
			return &bytes.Buffer{}
		},
	} {
		t.Run(mode, func(t *testing.T) {
			for name, tc := range testCases {
				t.Run(name, func(t *testing.T) {
					export := exportBuilder(t)

					// close file-like export locations
					if closer, ok := export.(io.Closer); ok {
						defer func() { require.NoError(t, closer.Close()) }()
					}
					require.NoError(t, client.ExportDatabase(Db, export, tc.opts...))

					// seek back to the start for file-like export locations
					if seeker, ok := export.(io.Seeker); ok {
						_, err := seeker.Seek(0, io.SeekStart)
						require.NoError(t, err)
					}

					tmpdir := t.TempDir()
					unarchivedPath := path.Join(tmpdir, "unarchived")
					require.NoError(t, os.Mkdir(unarchivedPath, os.ModePerm))
					require.NoError(t, extractTarGz(export, unarchivedPath))

					exportDir := path.Join(unarchivedPath, "snps-ai-graphdb-export")
					assert.DirExists(t, exportDir)
					assert.FileExists(t, path.Join(exportDir, "copy.cypher"))
					assert.FileExists(t, path.Join(exportDir, "schema.cypher"))
					assert.FileExists(t, path.Join(exportDir, "index.cypher"))
					assert.FileExists(t, path.Join(exportDir, fmt.Sprintf("User.%s", tc.ext)))
				})
			}
		})
	}

}

func extractTarGz(r io.Reader, dir string) error {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		dstPath := path.Join(dir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dstPath, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeReg:
			dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer func() { _ = dst.Close() }()

			_, err = io.Copy(dst, tarReader)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TestImport(t *testing.T) {
	client := getTestClient(t)

	// make a seed DB that an export can be made from
	const SeedDb = "seeder"
	require.NoError(t, client.CreateDatabase(SeedDb))
	for _, query := range []string{
		"CREATE NODE TABLE User(name STRING, age INT64, PRIMARY KEY (name))",
		"CREATE (:User {name: 'Adam', age: 30});",
		"CREATE (:User {name: 'Karissa', age: 40});",
		"CREATE (:User {name: 'Zhang', age: 50});",
		"CREATE (:User {name: 'Noura', age: 25});",
	} {
		_, err := client.CypherQueryWrite(SeedDb, query, nil)
		require.NoError(t, err)
	}

	for mode, exportBuilder := range map[string]func(*testing.T) io.ReadWriter{
		"File": func(t *testing.T) io.ReadWriter {
			tmpdir := t.TempDir()
			exportFile, err := os.CreateTemp(tmpdir, "*.tar.gz")
			require.NoError(t, err)
			return exportFile
		},
		"Buffer": func(t *testing.T) io.ReadWriter {
			return &bytes.Buffer{}
		},
	} {
		t.Run(mode, func(t *testing.T) {
			for name, format := range map[string]GraphDbExportOpt{
				"Parquet": WithFormatParquet{},
				"Csv":     WithFormatCsv{},
			} {
				t.Run(name, func(t *testing.T) {
					db := fmt.Sprintf("%s-%s", mode, name)

					export := exportBuilder(t)
					if closer, ok := export.(io.Closer); ok {
						// make sure file-like exports are closed at the end
						defer func() { require.NoError(t, closer.Close()) }()
					}
					require.NoError(t, client.ExportDatabase(SeedDb, export, format))
					if seeker, ok := export.(io.Seeker); ok {
						// make sure file-like exports are rewinded back to start
						_, err := seeker.Seek(0, io.SeekStart)
						require.NoError(t, err)
					}

					require.NoError(t, client.CreateDatabase(db))
					resBefore, err := client.CypherQueryRead(db, "MATCH (u) RETURN COUNT(*) AS count;", nil)
					require.NoError(t, err)
					require.Equal(t, 0., resBefore[0]["count"])

					require.NoError(t, client.ImportDatabase(db, export))
					resAfter, err := client.CypherQueryRead(db, "MATCH (u) RETURN COUNT(*) AS count;", nil)
					require.NoError(t, err)
					assert.Equal(t, 4., resAfter[0]["count"])
				})
			}
		})
	}
}
