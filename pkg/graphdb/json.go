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
	"fmt"
	"time"
)

type InternalID struct {
	TableID uint64 `json:"table_id"`
	Offset  uint64 `json:"offset"`
}

func (i InternalID) tag() string { return string(internalidValTag) }

// Json marshal helper for externally-tagged types
//
// Mimics the externally taggged serde format in rust: https://serde.rs/enum-representations.html
//
// examples:
// `"Any"`
// `{"Type": "content"}`
type externallyTagged[T tagged] struct {
	value *T
}

func (exTag externallyTagged[T]) MarshalJSON() ([]byte, error) {
	var t T
	if exTag.value == nil {
		return json.Marshal(t.tag())
	} else {
		return json.Marshal(map[string]any{t.tag(): exTag.value})
	}
}

func (exTag *externallyTagged[T]) UnmarshalJSON(data []byte) error {
	var t T

	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		// try again with string type
		var s string
		err := json.Unmarshal(data, &s)
		if err != nil {
			return err
		}
		if s != t.tag() {
			return fmt.Errorf("found string %q but expected tag %q", s, t.tag())
		}
		exTag.value = nil
		return nil
	}

	if len(raw) != 1 {
		return fmt.Errorf("expected 1 key, got %v", len(raw))
	}
	for key, val := range raw {

		var typedVal T
		err = json.Unmarshal(val, &typedVal)
		if err != nil {
			return err
		}
		if key != t.tag() {
			return fmt.Errorf("expected tag %q but got %q", t.tag(), key)
		}
		exTag.value = &typedVal
		return nil
	}
	return nil
}

type tagged interface {
	tag() string
}

// helper type for serializing 2-tuples
type Twople[A any, B any] struct {
	A A
	B B
}

func NewTwople[A any, B any](a A, b B) Twople[A, B] {
	return Twople[A, B]{a, b}
}

func (tup Twople[A, B]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{tup.A, tup.B})
}

func (tup *Twople[A, B]) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}
	if len(arr) == 2 {
		var a A
		err = json.Unmarshal(arr[0], &a)
		if err != nil {
			return err
		}

		var b B
		err = json.Unmarshal(arr[1], &b)
		if err != nil {
			return err
		}

		tup.A = a
		tup.B = b
		return nil
	} else {
		return fmt.Errorf("cannot unmarshal array of length %d to Twople, must be length 2", len(arr))
	}
}

// typed values converter
// typed values are a "twople" where the first field is a logical type and the second is an array of values
type typedValuesTwople Twople[logicalTypeUnmarshalHelper, valueArrayJson]

func typedValuesTwopleFromList(l listValue) typedValuesTwople {
	return typedValuesTwople{
		A: l.LogicalType,
		B: l.Values,
	}
}

func (tup typedValuesTwople) toList() listValue {
	return listValue{
		LogicalType: tup.A,
		Values:      tup.B,
	}
}

func typedValuesTwopleFromArray(a arrayValue) typedValuesTwople {
	return typedValuesTwople{
		A: a.LogicalType,
		B: a.Values,
	}
}

func (tup typedValuesTwople) toArray() arrayValue {
	return arrayValue{
		LogicalType: tup.A,
		Values:      tup.B,
	}
}

func (t typedValuesTwople) MarshalJSON() ([]byte, error) {
	return json.Marshal(Twople[logicalTypeUnmarshalHelper, valueArrayJson](t))
}
func (t *typedValuesTwople) UnmarshalJSON(data []byte) error {
	var intermediate Twople[logicalTypeUnmarshalHelper, valueArrayJson]
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return err
	}
	*t = typedValuesTwople(intermediate)
	return nil
}

// named fields converter
// named fields are twoples of string, value
type namedFieldsTwoples []Twople[string, valueUnmarshalHelper]

func namedFieldsTwoplesFromMap(m map[string]Value) namedFieldsTwoples {

	tups := make([]Twople[string, valueUnmarshalHelper], len(m))
	i := 0
	for k, v := range m {
		tups[i] = Twople[string, valueUnmarshalHelper]{k, valueUnmarshalHelper{v}}
		i++
	}
	return tups
}

func (t namedFieldsTwoples) toMap() map[string]Value {
	m := make(map[string]Value, len(t))
	for _, tup := range t {
		m[tup.A] = tup.B.Value
	}
	return m
}

func (t namedFieldsTwoples) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Twople[string, valueUnmarshalHelper](t))
}
func (t *namedFieldsTwoples) UnmarshalJSON(data []byte) error {
	var intermediate []Twople[string, valueUnmarshalHelper]
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return err
	}
	*t = namedFieldsTwoples(intermediate)
	return nil
}

// named type twoples converter
// named types are twoples of string, logical type
type namedTypesTwoples []Twople[string, logicalTypeUnmarshalHelper]

func namedTypesTwoplesFromMap(m map[string]LogicalType) namedTypesTwoples {
	tups := make([]Twople[string, logicalTypeUnmarshalHelper], len(m))
	i := 0
	for k, v := range m {
		tups[i] = Twople[string, logicalTypeUnmarshalHelper]{k, logicalTypeUnmarshalHelper{v}}
		i++
	}
	return tups
}

func (t namedTypesTwoples) toMap() map[string]LogicalType {
	m := make(map[string]LogicalType, len(t))
	for _, tup := range t {
		m[tup.A] = tup.B.LogicalType
	}
	return m
}

// null value converter
func (v nullValue) getInnerLogicalType() logicalTypeUnmarshalHelper {
	return v.LogicalType
}
func nullValueLogType(l logicalTypeUnmarshalHelper) nullValue {
	return nullValue{l}
}

// blob converter
type blobValueJson []uint8

func (v blobValueJson) MarshalJSON() ([]byte, error) {
	inner := make([]uint16, len(v))
	for i, element := range v {
		inner[i] = uint16(element)
	}
	return json.Marshal(inner)
}

// timestamps the marshal to RFC3339 format
type rfc3339NanoTime time.Time

func (v rfc3339NanoTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(v).Format(time.RFC3339Nano))
}

func (v *rfc3339NanoTime) UnmarshalJSON(data []byte) error {
	var timeStr string
	err := json.Unmarshal(data, &timeStr)
	if err != nil {
		return err
	}
	val, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return err
	}
	*v = rfc3339NanoTime(val)
	return nil
}

// interval converter
type intervalValueJson time.Duration

func (v intervalValueJson) MarshalJSON() ([]byte, error) {
	// should return an array of [<seconds>, <nanoseconds>]
	val := time.Duration(v)

	roundMultiple, err := time.ParseDuration("1s")
	if err != nil {
		return nil, err
	}
	rounded := val.Round(roundMultiple)
	wholeSecs := int64(rounded.Seconds())
	leftover := val - rounded
	nanos := leftover.Nanoseconds()
	return json.Marshal([]int64{wholeSecs, nanos})
}

func (v *intervalValueJson) UnmarshalJSON(data []byte) error {
	var secsNanos Twople[int64, int64]
	err := json.Unmarshal(data, &secsNanos)
	if err != nil {
		return err
	}
	allNanos := secsNanos.A*1e9 + secsNanos.B
	*v = intervalValueJson(time.Duration(allNanos))
	return nil
}

// value array converter
type valueArrayJson []Value

func (v valueArrayJson) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Value(v))
}
func (v *valueArrayJson) UnmarshalJSON(data []byte) error {
	var intermediate []valueUnmarshalHelper
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return err
	}

	vals := make([]Value, len(intermediate))
	for i, val := range intermediate {
		vals[i] = val.Value
	}
	*v = valueArrayJson(vals)
	return nil
}

// value map converter
type valueMapJson map[Value]Value

func (v valueMapJson) MarshalJSON() ([]byte, error) {
	twoples := make([]Twople[Value, Value], len(v))
	i := 0
	for k, v := range v {
		twoples[i] = Twople[Value, Value]{k, v}
		i++
	}
	return json.Marshal(twoples)
}
func (v *valueMapJson) UnmarshalJSON(data []byte) error {
	var intermediate []Twople[valueUnmarshalHelper, valueUnmarshalHelper]
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return err
	}

	vals := make(map[Value]Value, len(intermediate))
	for _, val := range intermediate {
		vals[val.A.Value] = val.B.Value
	}
	*v = valueMapJson(vals)
	return nil
}

// kuzu MapValue converter
type mapValueJson mapValue

func (v mapValueJson) MarshalJSON() ([]byte, error) {
	types := Twople[LogicalType, LogicalType]{v.KeyType, v.ValueType}
	combined := Twople[Twople[LogicalType, LogicalType], valueMapJson]{types, v.Pairs}
	return json.Marshal(combined)
}
func (v *mapValueJson) UnmarshalJSON(data []byte) error {
	var intermediate Twople[Twople[logicalTypeUnmarshalHelper, logicalTypeUnmarshalHelper], valueMapJson]
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return err
	}
	*v = mapValueJson(mapValue{
		intermediate.A.A,
		intermediate.A.B,
		intermediate.B,
	})
	return nil
}
