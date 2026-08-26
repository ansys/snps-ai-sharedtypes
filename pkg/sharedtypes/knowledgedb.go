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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/civil"
	"github.com/ansys/snps-ai-sharedtypes/pkg/graphdb"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DbFilters represents the filters for the database.
type DbFilters struct {
	// Filters for string fields
	GuidFilter         []string `json:"guid,omitempty"`
	DocumentIdFilter   []string `json:"document_id,omitempty"`
	DocumentNameFilter []string `json:"document_name,omitempty"`
	LevelFilter        []string `json:"level,omitempty"`

	// Filters for array fields
	TagsFilter     DbArrayFilter `json:"tags,omitempty"`
	KeywordsFilter DbArrayFilter `json:"keywords,omitempty"`

	// Filters for JSON fields
	MetadataFilter []DbJsonFilter `json:"metadata,omitempty"`
}

// DbArrayFilter represents the filter for an array field in the database.
type DbArrayFilter struct {
	NeedAll    bool     `json:"needAll"`
	FilterData []string `json:"filterData"`
}

// DbJsonFilter represents the filter for a JSON field in the database.
type DbJsonFilter struct {
	FieldName  string   `json:"fieldName"`
	FieldType  string   `json:"fieldType" description:"Can be either string or array."` // "string" or "array"
	FilterData []string `json:"filterData"`
	NeedAll    bool     `json:"needAll" description:"Only needed if the FieldType is array."` // only needed for array fields
}

// DbData represents the data stored in the database.
type DbData struct {
	Guid              uuid.UUID              `json:"guid"`
	DocumentId        string                 `json:"document_id"`
	DocumentName      string                 `json:"document_name"`
	Text              string                 `json:"text"`
	Keywords          []string               `json:"keywords"`
	Summary           string                 `json:"summary"`
	Embedding         []float32              `json:"embeddings"`
	Tags              []string               `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	ParentId          *uuid.UUID             `json:"parent_id"`
	ChildIds          []uuid.UUID            `json:"child_ids"`
	PreviousSiblingId *uuid.UUID             `json:"previous_sibling_id"`
	NextSiblingId     *uuid.UUID             `json:"next_sibling_id"`
	LastChildId       *uuid.UUID             `json:"last_child_id"`
	FirstChildId      *uuid.UUID             `json:"first_child_id"`
	Level             int                    `json:"level"`
	HasNeo4jEntry     bool                   `json:"has_neo4j_entry"`
}

// DbResponse can accommodate non-conflicting data from:
// - StoreElementsInVectorDatabase (API/Element data)
// - StoreExamplesInVectorDatabase (Example data)
// - StoreUserGuideSectionsInVectorDatabase (User guide data)
// - Standard document data
type DbResponse struct {
	// Standard DbResponse fields
	Guid              uuid.UUID              `json:"guid"`
	DocumentId        string                 `json:"document_id"`
	DocumentName      string                 `json:"document_name"`
	Text              string                 `json:"text"`
	Keywords          []string               `json:"keywords"`
	Summary           string                 `json:"summary"`
	Embedding         []float32              `json:"embeddings"`
	Tags              []string               `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	ParentId          *uuid.UUID             `json:"parent_id"`
	ChildIds          []uuid.UUID            `json:"child_ids"`
	PreviousSiblingId *uuid.UUID             `json:"previous_sibling_id"`
	NextSiblingId     *uuid.UUID             `json:"next_sibling_id"`
	LastChildId       *uuid.UUID             `json:"last_child_id"`
	FirstChildId      *uuid.UUID             `json:"first_child_id"`
	Distance          float64                `json:"distance"`
	Level             int                    `json:"level"`
	HasNeo4jEntry     bool                   `json:"has_neo4j_entry"`

	// Siblings
	Parent    *DbData  `json:"parent,omitempty"`
	Children  []DbData `json:"children,omitempty"`
	LeafNodes []DbData `json:"leaf_nodes,omitempty"`
	Siblings  []DbData `json:"siblings,omitempty"`

	// From StoreElementsInVectorDatabase payload
	Type           string `json:"type,omitempty"`
	NamePseudocode string `json:"name_pseudocode,omitempty"`
	NameFormatted  string `json:"name_formatted,omitempty"`
	Name           string `json:"name,omitempty"`
	ParentClass    string `json:"parent_class,omitempty"`

	// From StoreExamplesInVectorDatabase payload
	Dependencies           []interface{}          `json:"dependencies,omitempty"`
	DependencyEquivalences map[string]interface{} `json:"dependency_equivalences,omitempty"`
	PreviousChunk          string                 `json:"previous_chunk,omitempty"`
	NextChunk              string                 `json:"next_chunk,omitempty"`

	// From StoreUserGuideSectionsInVectorDatabase payload
	SectionName       string `json:"section_name,omitempty"`
	Title             string `json:"title,omitempty"`
	ParentSectionName string `json:"parent_section_name,omitempty"`
}

// DBListCollectionsOutput represents the output of listing collections in the database.
type DBListCollectionsOutput struct {
	Success     bool     `json:"success" description:"Returns true if the collections were listed successfully. Returns false or an error if not."`
	Collections []string `json:"collections" description:"A list of collection names."`
}

// GeneralNeo4jQueryInput represents the input for executing a Neo4j query.
type GeneralNeo4jQueryInput struct {
	Query string `json:"query" description:"Neo4j query to be executed. Required for executing a query." required:"true"`
}

// GeneralNeo4jQueryOutput represents the output of executing a Neo4j query.
type GeneralNeo4jQueryOutput struct {
	Success  bool          `json:"success" description:"Returns true if the query was executed successfully. Returns false or an error if not."`
	Response Neo4jResponse `json:"response" description:"Summary and records of the query execution."`
}

// neo4jResponse represents the response from the Neo4j query.
type Neo4jResponse struct {
	Record          neo4jRecord     `json:"record"`
	SummaryCounters summaryCounters `json:"summaryCounters"`
}

// neo4jRecord represents the record from the Neo4j query.
type neo4jRecord []struct {
	Values []value `json:"Values"`
}

// value represents the value from the Neo4j query.
type value struct {
	Id        int      `json:"Id"`
	NodeTypes []string `json:"Labels"`
	Props     props    `json:"Props"`
}

// props represents the properties from the Neo4j query.
type props struct {
	CollectionName string    `json:"collectionName"`
	DocumentId     string    `json:"documentId"`
	DocumentTypes  []string  `json:"documentTypes,omitempty"`
	Guid           uuid.UUID `json:"guid,omitempty"`
}

// summaryCounters represents the summary counters from the Neo4j query.
type summaryCounters struct {
	NodesCreated         int `json:"nodes_created"`
	NodesDeleted         int `json:"nodes_deleted"`
	RelationshipsCreated int `json:"relationships_created"`
	RelationshipsDeleted int `json:"relationships_deleted"`
	PropertiesSet        int `json:"properties_set"`
	LabelsAdded          int `json:"labels_added"`
	LabelsRemoved        int `json:"labels_removed"`
	IndexesAdded         int `json:"indexes_added"`
	IndexesRemoved       int `json:"indexes_removed"`
	ConstraintsAdded     int `json:"constraints_added"`
	ConstraintsRemoved   int `json:"constraints_removed"`
}

// DbAddDataInput represents the input for adding data to the database.
type DbAddDataInput struct {
	CollectionName string   `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	Data           []DbData `json:"data" description:"Data objects to be added to the DB." required:"true"`
}

// DbAddDataOutput represents the output of adding data to the database.
type DbAddDataOutput struct {
	Success             bool   `json:"success" description:"Returns true if the data was added successfully. Returns false or an error if not."`
	IgnoredObjectsCount int    `json:"ignored_objects_count" description:"Number of ignored documents."`
	Error               string `json:"error" description:"Error message if the collection could not be created."`
}

// DbCreateCollectionInput represents the input for creating a collection in the database.
type DbCreateCollectionInput struct {
	CollectionName string `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
}

// DbCreateCollectionOutput represents the output of creating a collection in the database.
type DbCreateCollectionOutput struct {
	Success bool   `json:"success" description:"Returns true if the collection was created successfully. Returns false or an error if not."`
	Error   string `json:"error" description:"Error message if the collection could not be created."`
}

type GraphDbValueType string

// some types don't really make sense to be parsed in the snps-ai-flowkit context. These are not included here:
// - Null
// - InternalID
// - List/Array (would require somehow parsing/infering the LogicalType as well)
// - Struct (would require knowing type of each value)
// - Node/Rel/RecursiveRel
// - Map
// - Union
const (
	Bool         GraphDbValueType = "bool"
	Int64        GraphDbValueType = "int64"
	Int32        GraphDbValueType = "int32"
	Int16        GraphDbValueType = "int16"
	Int8         GraphDbValueType = "int8"
	UInt64       GraphDbValueType = "uint64"
	UInt32       GraphDbValueType = "uint32"
	UInt16       GraphDbValueType = "uint16"
	UInt8        GraphDbValueType = "uint8"
	Int128       GraphDbValueType = "int128"
	Double       GraphDbValueType = "double"
	Float        GraphDbValueType = "float"
	Date         GraphDbValueType = "date"
	Interval     GraphDbValueType = "interval"
	Timestamp    GraphDbValueType = "timestamp"
	TimestampTz  GraphDbValueType = "timestamptz"
	TimestampNs  GraphDbValueType = "timestampns"
	TimestampMs  GraphDbValueType = "timestampms"
	TimestampSec GraphDbValueType = "timestampsec"
	String       GraphDbValueType = "string"
	Blob         GraphDbValueType = "blob"
	Struct       GraphDbValueType = "struct"
	UUID         GraphDbValueType = "uuid"
	Decimal      GraphDbValueType = "decimal"
)

func (valType GraphDbValueType) Parse(val string) (graphdb.Value, error) {
	switch valType {
	case Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return graphdb.BoolValue(b), nil
	case Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return graphdb.Int64Value(i), nil
	case Int32:
		i, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return nil, err
		}
		return graphdb.Int32Value(i), nil
	case Int16:
		i, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return nil, err
		}
		return graphdb.Int16Value(i), nil
	case Int8:
		i, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return nil, err
		}
		return graphdb.Int8Value(i), nil
	case UInt64:
		i, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return graphdb.UInt64Value(i), nil
	case UInt32:
		i, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		return graphdb.UInt32Value(i), nil
	case UInt16:
		i, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return nil, err
		}
		return graphdb.UInt16Value(i), nil
	case UInt8:
		i, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return nil, err
		}
		return graphdb.UInt8Value(i), nil
	case Int128:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return graphdb.Int128Value(i), nil
	case Double:
		d, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		return graphdb.DoubleValue(d), nil
	case Float:
		f, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return nil, err
		}
		return graphdb.FloatValue(f), nil
	case Date:
		d, err := civil.ParseDate(val)
		if err != nil {
			return nil, err
		}
		return graphdb.DateValue(d), nil
	case Interval:
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, err
		}
		return graphdb.IntervalValue(d), nil
	case TimestampTz:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return graphdb.TimestampTzValue(t), err
	case TimestampNs:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return graphdb.TimestampNsValue(t), err
	case TimestampMs:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return graphdb.TimestampMsValue(t), err
	case TimestampSec:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return graphdb.TimestampSecValue(t), err
	case String:
		return graphdb.StringValue(val), nil
	case Blob:
		var b []uint8
		err := json.Unmarshal([]byte(val), &b)
		if err != nil {
			return nil, err
		}

		return graphdb.BlobValue(b), nil
	case UUID:
		u, err := uuid.Parse(val)
		if err != nil {
			return nil, err
		}
		return graphdb.UUIDValue(u), nil
	case Decimal:
		d, err := decimal.NewFromString(val)
		if err != nil {
			return nil, err
		}
		return graphdb.DecimalValue(d), nil
	default:
		return nil, fmt.Errorf("unknown value type %q", valType)
	}
}
