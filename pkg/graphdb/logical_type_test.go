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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func logicalTypeTest[T LogicalType](t *testing.T, lt T, expected any) {
	require := require.New(t)
	assert := assert.New(t)

	actualBytes, err := json.Marshal(lt)
	require.NoError(err)
	var actualJson any
	require.NoError(json.Unmarshal(actualBytes, &actualJson))
	assert.Equal(expected, actualJson)

	var recreated T
	require.NoError(json.Unmarshal(actualBytes, &recreated))
	assert.Equal(lt, recreated)
}

func TestLogicalTypeAny(t *testing.T) {
	logicalTypeTest(t, AnyLogicalType{}, "Any")
}
func TestLogicalTypeBool(t *testing.T) {
	logicalTypeTest(t, BoolLogicalType{}, "Bool")
}
func TestLogicalTypeSerial(t *testing.T) {
	logicalTypeTest(t, SerialLogicalType{}, "Serial")
}
func TestLogicalTypeInt64(t *testing.T) {
	logicalTypeTest(t, Int64LogicalType{}, "Int64")
}
func TestLogicalTypeInt32(t *testing.T) {
	logicalTypeTest(t, Int32LogicalType{}, "Int32")
}
func TestLogicalTypeInt16(t *testing.T) {
	logicalTypeTest(t, Int16LogicalType{}, "Int16")
}
func TestLogicalTypeInt8(t *testing.T) {
	logicalTypeTest(t, Int8LogicalType{}, "Int8")
}
func TestLogicalTypeUInt64(t *testing.T) {
	logicalTypeTest(t, UInt64LogicalType{}, "UInt64")
}
func TestLogicalTypeUInt32(t *testing.T) {
	logicalTypeTest(t, UInt32LogicalType{}, "UInt32")
}
func TestLogicalTypeUInt16(t *testing.T) {
	logicalTypeTest(t, UInt16LogicalType{}, "UInt16")
}
func TestLogicalTypeUInt8(t *testing.T) {
	logicalTypeTest(t, UInt8LogicalType{}, "UInt8")
}
func TestLogicalTypeInt128(t *testing.T) {
	logicalTypeTest(t, Int128LogicalType{}, "Int128")
}
func TestLogicalTypeDouble(t *testing.T) {
	logicalTypeTest(t, DoubleLogicalType{}, "Double")
}
func TestLogicalTypeFloat(t *testing.T) {
	logicalTypeTest(t, FloatLogicalType{}, "Float")
}
func TestLogicalTypeDate(t *testing.T) {
	logicalTypeTest(t, DateLogicalType{}, "Date")
}
func TestLogicalTypeInterval(t *testing.T) {
	logicalTypeTest(t, IntervalLogicalType{}, "Interval")
}
func TestLogicalTypeTimestamp(t *testing.T) {
	logicalTypeTest(t, TimestampLogicalType{}, "Timestamp")
}
func TestLogicalTypeTimestampTz(t *testing.T) {
	logicalTypeTest(t, TimestampTzLogicalType{}, "TimestampTz")
}
func TestLogicalTypeTimestampNs(t *testing.T) {
	logicalTypeTest(t, TimestampNsLogicalType{}, "TimestampNs")
}
func TestLogicalTypeTimestampSec(t *testing.T) {
	logicalTypeTest(t, TimestampSecLogicalType{}, "TimestampSec")
}
func TestLogicalTypeInternalIDType(t *testing.T) {
	logicalTypeTest(t, InternalIDLogicalType{}, "InternalID")
}
func TestLogicalTypeString(t *testing.T) {
	logicalTypeTest(t, StringLogicalType{}, "String")
}
func TestLogicalTypeBlob(t *testing.T) {
	logicalTypeTest(t, BlobLogicalType{}, "Blob")
}
func TestLogicalTypeListAny(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{AnyLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Any"}},
	)
}
func TestLogicalTypeListBool(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{BoolLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Bool"}},
	)
}
func TestLogicalTypeListSerial(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{SerialLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Serial"}},
	)
}
func TestLogicalTypeListInt64(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{Int64LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Int64"}},
	)
}
func TestLogicalTypeListInt32(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{Int32LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Int32"}},
	)
}
func TestLogicalTypeListInt16(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{Int16LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Int16"}},
	)
}
func TestLogicalTypeListInt8(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{Int8LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Int8"}},
	)
}
func TestLogicalTypeListUInt64(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{UInt64LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "UInt64"}},
	)
}
func TestLogicalTypeListUInt32(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{UInt32LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "UInt32"}},
	)
}
func TestLogicalTypeListUInt16(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{UInt16LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "UInt16"}},
	)
}
func TestLogicalTypeListUInt8(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{UInt8LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "UInt8"}},
	)
}
func TestLogicalTypeListInt128(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{Int128LogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Int128"}},
	)
}
func TestLogicalTypeListDouble(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{DoubleLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Double"}},
	)
}
func TestLogicalTypeListFloat(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{FloatLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Float"}},
	)
}
func TestLogicalTypeListDate(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{DateLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Date"}},
	)
}
func TestLogicalTypeListInterval(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{IntervalLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Interval"}},
	)
}
func TestLogicalTypeListTimestamp(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{TimestampLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Timestamp"}},
	)
}
func TestLogicalTypeListTimestampTz(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{TimestampTzLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "TimestampTz"}},
	)
}
func TestLogicalTypeListTimestampNs(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{TimestampNsLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "TimestampNs"}},
	)
}
func TestLogicalTypeListTimestampMs(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{TimestampMsLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "TimestampMs"}},
	)
}
func TestLogicalTypeListTimestampSec(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{TimestampSecLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "TimestampSec"}},
	)
}
func TestLogicalTypeListInternalID(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{InternalIDLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "InternalID"}},
	)
}
func TestLogicalTypeListString(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{StringLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "String"}},
	)
}
func TestLogicalTypeListBlob(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{BlobLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Blob"}},
	)
}
func TestLogicalTypeListList(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{ListLogicalType{BoolLogicalType{}}},
		map[string]any{"List": map[string]any{"child_type": map[string]any{"List": map[string]any{"child_type": "Bool"}}}},
	)
}
func TestLogicalTypeListArray(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{ArrayLogicalType{Int16LogicalType{}, 2}},
		map[string]any{"List": map[string]any{"child_type": map[string]any{"Array": map[string]any{"child_type": "Int16", "num_elements": float64(2)}}}},
	)
}
func TestLogicalTypeListStruct(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{StructLogicalType{[]Twople[string, LogicalType]{
			{"name", StringLogicalType{}},
			{"age", UInt16LogicalType{}},
		}}},
		map[string]any{"List": map[string]any{"child_type": map[string]any{"Struct": map[string]any{"fields": []any{[]any{"name", "String"}, []any{"age", "UInt16"}}}}}},
	)
}
func TestLogicalTypeListNode(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{NodeLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Node"}},
	)
}
func TestLogicalTypeListRel(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{RelLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "Rel"}},
	)
}
func TestLogicalTypeListRecursiveRel(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{RecursiveRelLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "RecursiveRel"}},
	)
}
func TestLogicalTypeListMap(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{MapLogicalType{UInt32LogicalType{}, FloatLogicalType{}}},
		map[string]any{"List": map[string]any{"child_type": map[string]any{"Map": map[string]any{"key_type": "UInt32", "value_type": "Float"}}}},
	)
}
func TestLogicalTypeListUnion(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{UnionLogicalType{[]Twople[string, LogicalType]{
			{"hi", StringLogicalType{}},
		}}},
		map[string]any{"List": map[string]any{"child_type": map[string]any{"Union": map[string]any{"fields": []any{
			[]any{"hi", "String"},
		}}}}},
	)
}
func TestLogicalTypeListUUID(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{UUIDLogicalType{}},
		map[string]any{"List": map[string]any{"child_type": "UUID"}},
	)
}
func TestLogicalTypeListDecimal(t *testing.T) {
	logicalTypeTest(
		t,
		ListLogicalType{DecimalLogicalType{1, 7}},
		map[string]any{"List": map[string]any{"child_type": map[string]any{"Decimal": map[string]any{"precision": float64(1), "scale": float64(7)}}}},
	)
}
func TestLogicalTypeArray(t *testing.T) {
	logicalTypeTest(
		t,
		ArrayLogicalType{UInt16LogicalType{}, 12},
		map[string]any{"Array": map[string]any{"child_type": "UInt16", "num_elements": float64(12)}},
	)
}
func TestLogicalTypeStruct(t *testing.T) {
	logicalTypeTest(
		t,
		StructLogicalType{[]Twople[string, LogicalType]{
			{
				"name",
				StringLogicalType{},
			},
			{
				"age",
				UInt32LogicalType{},
			},
			{
				"items",
				ListLogicalType{StringLogicalType{}},
			},
		}},
		map[string]any{"Struct": map[string]any{
			"fields": []any{
				[]any{"name", "String"},
				[]any{"age", "UInt32"},
				[]any{"items", map[string]any{"List": map[string]any{"child_type": "String"}}},
			}}},
	)
}
func TestLogicalTypeNode(t *testing.T) {
	logicalTypeTest(t, NodeLogicalType{}, "Node")
}
func TestLogicalTypeRel(t *testing.T) {
	logicalTypeTest(t, RelLogicalType{}, "Rel")
}
func TestLogicalTypeRecursiveRel(t *testing.T) {
	logicalTypeTest(t, RecursiveRelLogicalType{}, "RecursiveRel")
}
func TestLogicalTypeMap(t *testing.T) {
	logicalTypeTest(
		t,
		MapLogicalType{StringLogicalType{}, Int8LogicalType{}},
		map[string]any{"Map": map[string]any{"key_type": "String", "value_type": "Int8"}},
	)
}
func TestLogicalTypeUnion(t *testing.T) {
	logicalTypeTest(
		t,
		UnionLogicalType{[]Twople[string, LogicalType]{
			{
				"name",
				StringLogicalType{},
			},
			{
				"age",
				UInt32LogicalType{},
			},
			{
				"items",
				ListLogicalType{StringLogicalType{}},
			},
		}},
		map[string]any{"Union": map[string]any{
			"fields": []any{
				[]any{"name", "String"},
				[]any{"age", "UInt32"},
				[]any{"items", map[string]any{"List": map[string]any{"child_type": "String"}}},
			}}},
	)
}
func TestLogicalTypeUUID(t *testing.T) {
	logicalTypeTest(t, UUIDLogicalType{}, "UUID")
}
func TestLogicalTypeDecimal(t *testing.T) {
	logicalTypeTest(
		t,
		DecimalLogicalType{5, 3},
		map[string]any{"Decimal": map[string]any{"precision": float64(5), "scale": float64(3)}},
	)
}
