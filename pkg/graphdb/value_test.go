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
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func valueTest1[V Value](t *testing.T, value V, expected any) {
	require := require.New(t)
	assert := assert.New(t)

	var actual []byte

	t.Run("marshal", func(t *testing.T) {
		actualBytes, err := json.Marshal(value)
		require.NoError(err)
		var actualJson any
		require.NoError(json.Unmarshal(actualBytes, &actualJson))
		assert.Equal(expected, actualJson, "marshalling was not as expected")
		actual = actualBytes
	})
	t.Run("unmarshal", func(t *testing.T) {
		var recreated V
		require.NoError(json.Unmarshal(actual, &recreated))
		assert.Equal(value, recreated, "unmarshalling did not recreate original")
	})
	t.Run("generic unmarshal", func(t *testing.T) {
		var unmarshaledVal valueUnmarshalHelper
		require.NoError(json.Unmarshal(actual, &unmarshaledVal))
		assert.Equal(reflect.TypeOf(value), reflect.TypeOf(unmarshaledVal.Value))
	})
}

func valueTestN[V Value](t *testing.T, value V, expecteds []any) {
	require := require.New(t)
	assert := assert.New(t)

	var actual []byte

	t.Run("marshal", func(t *testing.T) {
		actualBytes, err := json.Marshal(value)
		require.NoError(err)
		var actualJson any
		require.NoError(json.Unmarshal(actualBytes, &actualJson))
		assert.Contains(expecteds, actualJson, "marshalling was not as expected")
		actual = actualBytes
	})
	t.Run("unmarshal", func(t *testing.T) {
		var recreated V
		require.NoError(json.Unmarshal(actual, &recreated))
		assert.Equal(value, recreated, "unmarshalling did not recreate original")
	})
	t.Run("generic unmarshal", func(t *testing.T) {
		var unmarshaledVal valueUnmarshalHelper
		require.NoError(json.Unmarshal(actual, &unmarshaledVal))
		assert.Equal(reflect.TypeOf(value), reflect.TypeOf(unmarshaledVal.Value))
	})
}

func TestValueNullAny(t *testing.T) {
	valueTest1(
		t,
		NullValue{AnyLogicalType{}},
		map[string]any{"Null": "Any"},
	)
}
func TestValueNullListFloat(t *testing.T) {
	valueTest1(
		t,
		NullValue{ListLogicalType{FloatLogicalType{}}},
		map[string]any{"Null": map[string]any{"List": map[string]any{"child_type": "Float"}}},
	)
}
func TestValueBool(t *testing.T) {
	valueTest1(
		t,
		BoolValue(true),
		map[string]any{"Bool": true},
	)
}
func TestValueInt64(t *testing.T) {
	valueTest1(
		t,
		Int64Value(82),
		map[string]any{"Int64": float64(82)},
	)
}
func TestValueInt32(t *testing.T) {
	valueTest1(
		t,
		Int32Value(1),
		map[string]any{"Int32": float64(1)},
	)
}
func TestValueInt16(t *testing.T) {
	valueTest1(
		t,
		Int16Value(100),
		map[string]any{"Int16": float64(100)},
	)
}
func TestValueInt8(t *testing.T) {
	valueTest1(
		t,
		Int8Value(-6),
		map[string]any{"Int8": float64(-6)},
	)
}
func TestValueUInt64(t *testing.T) {
	valueTest1(
		t,
		UInt64Value(0),
		map[string]any{"UInt64": float64(0)},
	)
}
func TestValueUInt32(t *testing.T) {
	valueTest1(
		t,
		UInt32Value(1001),
		map[string]any{"UInt32": float64(1001)},
	)
}
func TestValueUInt16(t *testing.T) {
	valueTest1(
		t,
		UInt16Value(212),
		map[string]any{"UInt16": float64(212)},
	)
}
func TestValueUInt8(t *testing.T) {
	valueTest1(
		t,
		UInt8Value(50),
		map[string]any{"UInt8": float64(50)},
	)
}
func TestValueInt128(t *testing.T) {
	valueTest1(
		t,
		Int128Value(9009),
		map[string]any{"Int128": float64(9009)},
	)
}
func TestValueDouble(t *testing.T) {
	valueTest1(
		t,
		DoubleValue(-56.1234),
		map[string]any{"Double": -56.1234},
	)
}
func TestValueFloat(t *testing.T) {
	valueTest1(
		t,
		FloatValue(90.0),
		map[string]any{"Float": 90.0},
	)
}
func TestValueInternalID(t *testing.T) {
	valueTest1(
		t,
		InternalIDValue{0, 0},
		map[string]any{"InternalID": map[string]any{"table_id": float64(0), "offset": float64(0)}},
	)
}
func TestValueString(t *testing.T) {
	valueTest1(
		t,
		StringValue("Hello"),
		map[string]any{"String": "Hello"},
	)
}
func TestValueBlob(t *testing.T) {
	valueTest1(
		t,
		BlobValue([]uint8{0, 1, 2, 3, 4}),
		map[string]any{"Blob": []any{float64(0), float64(1), float64(2), float64(3), float64(4)}},
	)
}
func TestValueList(t *testing.T) {
	valueTest1(
		t,
		ListValue{UInt64LogicalType{}, []Value{UInt64Value(0), UInt64Value(12)}},
		map[string]any{
			"List": []any{
				"UInt64",
				[]any{
					map[string]any{"UInt64": float64(0)},
					map[string]any{"UInt64": float64(12)},
				},
			},
		},
	)
}
func TestValueArray(t *testing.T) {
	valueTest1(
		t,
		ArrayValue{BoolLogicalType{}, []Value{BoolValue(true), BoolValue(false)}},
		map[string]any{
			"Array": []any{
				"Bool",
				[]any{
					map[string]any{"Bool": true},
					map[string]any{"Bool": false},
				},
			},
		},
	)
}
func TestValueStruct(t *testing.T) {
	valueTestN(
		t,
		StructValue(map[string]Value{"a": BoolValue(false), "name": StringValue("Joe")}),
		[]any{
			map[string]any{
				"Struct": []any{
					[]any{"a", map[string]any{"Bool": false}},
					[]any{"name", map[string]any{"String": "Joe"}},
				},
			},
			map[string]any{
				"Struct": []any{
					[]any{"name", map[string]any{"String": "Joe"}},
					[]any{"a", map[string]any{"Bool": false}},
				},
			},
		},
	)
}
func TestValueNode(t *testing.T) {
	valueTest1(
		t,
		NodeValue{InternalID{1, 10}, "my-label", map[string]Value{}},
		map[string]any{
			"Node": map[string]any{
				"id":         map[string]any{"table_id": float64(1), "offset": float64(10)},
				"label":      "my-label",
				"properties": []any{},
			},
		},
	)
}
func TestValueRel(t *testing.T) {
	valueTest1(
		t,
		RelValue{InternalID{4, 1}, InternalID{6, 0}, "lab", map[string]Value{}},
		map[string]any{
			"Rel": map[string]any{
				"src_node":   map[string]any{"table_id": float64(4), "offset": float64(1)},
				"dst_node":   map[string]any{"table_id": float64(6), "offset": float64(0)},
				"label":      "lab",
				"properties": []any{},
			},
		},
	)
}
func TestValueMap(t *testing.T) {
	valueTest1(
		t,
		MapValue{UInt64LogicalType{}, BoolLogicalType{}, map[Value]Value{UInt64Value(4): BoolValue(false)}},
		map[string]any{
			"Map": []any{
				[]any{"UInt64", "Bool"},
				[]any{
					[]any{
						map[string]any{"UInt64": float64(4)},
						map[string]any{"Bool": false},
					},
				},
			},
		},
	)
}
func TestValueUnion(t *testing.T) {
	valueTestN(
		t,
		UnionValue{map[string]LogicalType{"num": Int64LogicalType{}, "str": StringLogicalType{}}, Int64Value(1)},
		[]any{
			map[string]any{
				"Union": map[string]any{
					"types": []any{[]any{"num", "Int64"}, []any{"str", "String"}},
					"value": map[string]any{"Int64": float64(1)},
				},
			},
			map[string]any{
				"Union": map[string]any{
					"types": []any{[]any{"str", "String"}, []any{"num", "Int64"}},
					"value": map[string]any{"Int64": float64(1)},
				},
			},
		},
	)
}
func TestValueUUIDZeros(t *testing.T) {
	valueTest1(
		t,
		UUIDValue(uuid.MustParse("00000000-0000-0000-0000-ffff00000000")),
		map[string]any{"UUID": "00000000-0000-0000-0000-ffff00000000"},
	)
}
func TestValueUUID(t *testing.T) {
	valueTest1(
		t,
		UUIDValue(uuid.MustParse("8f914bce-df4e-4244-9cd4-ea96bf0c58d4")),
		map[string]any{"UUID": "8f914bce-df4e-4244-9cd4-ea96bf0c58d4"},
	)
}
func TestValueDecimalSmall(t *testing.T) {
	valueTest1(
		t,
		DecimalValue(decimal.RequireFromString("12.34")),
		map[string]any{"Decimal": "12.34"},
	)
}
func TestValueDecimalBig(t *testing.T) {
	valueTest1(
		t,
		DecimalValue(decimal.RequireFromString("12.3456789")),
		map[string]any{"Decimal": "12.3456789"},
	)
}
func TestValueDate(t *testing.T) {
	valueTest1(
		t,
		DateValue(civil.Date{Year: 2025, Month: time.April, Day: 23}),
		map[string]any{"Date": "2025-04-23"},
	)
}
func TestValueTimestamp(t *testing.T) {
	valueTest1(
		t,
		TimestampValue(time.Date(2025, time.April, 23, 13, 26, 21, 123450000, time.UTC)),
		map[string]any{"Timestamp": "2025-04-23T13:26:21.12345Z"},
	)
}
func TestValueTimestampTz(t *testing.T) {
	valueTest1(
		t,
		TimestampTzValue(time.Date(2025, time.April, 23, 13, 26, 21, 123450000, time.UTC)),
		map[string]any{"TimestampTz": "2025-04-23T13:26:21.12345Z"},
	)
}
func TestValueTimestampNs(t *testing.T) {
	valueTest1(
		t,
		TimestampNsValue(time.Date(2025, time.April, 23, 13, 26, 21, 123450000, time.UTC)),
		map[string]any{"TimestampNs": "2025-04-23T13:26:21.12345Z"},
	)
}
func TestValueTimestampMs(t *testing.T) {
	valueTest1(
		t,
		TimestampMsValue(time.Date(2025, time.April, 23, 13, 26, 21, 123450000, time.UTC)),
		map[string]any{"TimestampMs": "2025-04-23T13:26:21.12345Z"},
	)
}
func TestValueTimestampSec(t *testing.T) {
	valueTest1(
		t,
		TimestampSecValue(time.Date(2025, time.April, 23, 13, 26, 21, 123450000, time.UTC)),
		map[string]any{"TimestampSec": "2025-04-23T13:26:21.12345Z"},
	)
}
func TestValueInterval(t *testing.T) {
	valueTest1(
		t,
		IntervalValue(23*24*time.Hour),
		map[string]any{"Interval": []any{float64(1987200), float64(0)}},
	)
}
func TestValueIntervalNs(t *testing.T) {
	valueTest1(
		t,
		IntervalValue(23*24*time.Hour+456*time.Nanosecond),
		map[string]any{"Interval": []any{float64(1987200), float64(456)}},
	)
}

// func TestGenericValueUnmarshal(t *testing.T) {
// 	vals := []struct {
// 		name  string
// 		json  string
// 		value reflect.Type
// 	}{}

// 	for _, test := range vals {
// 		t.Run(test.name, func(t *testing.T) {
// 			require := require.New(t)
// 			assert := assert.New(t)

// 			var unmarshaledVal Value
// 			require.NoError(json.Unmarshal([]byte(test.json), &unmarshaledVal))
// 			assert.Equal(test.value, reflect.TypeOf(unmarshaledVal))
// 		})
// 	}
// }
