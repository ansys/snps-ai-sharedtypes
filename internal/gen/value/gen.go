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

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

type ValueDef interface {
	ValueType() string
	Tag() string
	ConverterType() string
	ToConverter() string
	FromConverter() string
}

type TypeAliasValue struct {
	tag           string
	Type          string
	converterType string
	toConverter   string
	fromConverter string
}

func (t TypeAliasValue) ValueType() string     { return "alias" }
func (t TypeAliasValue) Tag() string           { return t.tag }
func (t TypeAliasValue) ConverterType() string { return t.converterType }
func (t TypeAliasValue) ToConverter() string   { return t.toConverter }
func (t TypeAliasValue) FromConverter() string { return t.fromConverter }

func BasicTypeAliasValue(tag string, typ string) TypeAliasValue {
	return TypeAliasValue{tag, typ, "", "", ""}
}

type StructValue struct {
	tag           string
	Fields        []Field
	converterType string
	toConverter   string
	fromConverter string
}

func (t StructValue) ValueType() string     { return "struct" }
func (t StructValue) Tag() string           { return t.tag }
func (t StructValue) ConverterType() string { return t.converterType }
func (t StructValue) ToConverter() string   { return t.toConverter }
func (t StructValue) FromConverter() string { return t.fromConverter }

type Field struct {
	Name          string
	Type          string
	Tag           string
	ConverterType string
	ToConverter   string
	FromConverter string
}

func BasicField(name string, typ string) Field {
	snakeName := strcase.ToSnake(name)
	return Field{name, typ, fmt.Sprintf("`json:%q`", snakeName), "", "", ""}
}

func LogicalTypeField(name string) Field {
	snakeName := strcase.ToSnake(name)
	return Field{
		name,
		"LogicalType",
		fmt.Sprintf("`json:%q`", snakeName),
		"logicalTypeUnmarshalHelper",
		"newLogicalTypeHelper",
		"logicalTypeUnmarshalHelper.getInnerLogicalType",
	}
}

func fail(msg string) string {
	panic("error in template: " + msg)
}

func main() {
	valtypes := []ValueDef{
		StructValue{
			"Null",
			[]Field{LogicalTypeField("LogicalType")},
			"logicalTypeUnmarshalHelper",
			"nullValue.getInnerLogicalType",
			"nullValueLogType",
		},
		BasicTypeAliasValue("Bool", "bool"),
		BasicTypeAliasValue("Int64", "int64"),
		BasicTypeAliasValue("Int32", "int32"),
		BasicTypeAliasValue("Int16", "int16"),
		BasicTypeAliasValue("Int8", "int8"),
		BasicTypeAliasValue("UInt64", "uint64"),
		BasicTypeAliasValue("UInt32", "uint32"),
		BasicTypeAliasValue("UInt16", "uint16"),
		BasicTypeAliasValue("UInt8", "uint8"),
		BasicTypeAliasValue("Int128", "int64"),
		BasicTypeAliasValue("Double", "float64"),
		BasicTypeAliasValue("Float", "float32"),
		TypeAliasValue{"Date", "civil.Date", "civil.Date", "civil.Date", "dateValue"},
		TypeAliasValue{"Interval", "time.Duration", "intervalValueJson", "intervalValueJson", "intervalValue"},
		TypeAliasValue{"Timestamp", "time.Time", "rfc3339NanoTime", "rfc3339NanoTime", "timestampValue"},
		TypeAliasValue{"TimestampTz", "time.Time", "rfc3339NanoTime", "rfc3339NanoTime", "timestamptzValue"},
		TypeAliasValue{"TimestampNs", "time.Time", "rfc3339NanoTime", "rfc3339NanoTime", "timestampnsValue"},
		TypeAliasValue{"TimestampMs", "time.Time", "rfc3339NanoTime", "rfc3339NanoTime", "timestampmsValue"},
		TypeAliasValue{"TimestampSec", "time.Time", "rfc3339NanoTime", "rfc3339NanoTime", "timestampsecValue"},
		BasicTypeAliasValue("InternalID", "InternalID"),
		BasicTypeAliasValue("String", "string"),
		TypeAliasValue{
			"Blob",
			"[]uint8",
			"blobValueJson",
			"blobValueJson",
			"[]uint8",
		},
		StructValue{
			"List",
			[]Field{
				LogicalTypeField("LogicalType"),
				{"Values", "[]Value", "`json:\"values\"`", "valueArrayJson", "", ""},
			},
			"typedValuesTwople",
			"typedValuesTwopleFromList",
			"typedValuesTwople.toList",
		},
		StructValue{
			"Array",
			[]Field{
				LogicalTypeField("LogicalType"),
				{"Values", "[]Value", "`json:\"values\"`", "valueArrayJson", "", ""},
			},
			"typedValuesTwople",
			"typedValuesTwopleFromArray",
			"typedValuesTwople.toArray",
		},
		TypeAliasValue{
			"Struct",
			"map[string]Value",
			"namedFieldsTwoples",
			"namedFieldsTwoplesFromMap",
			"namedFieldsTwoples.toMap",
		},
		StructValue{
			"Node",
			[]Field{
				BasicField("ID", "InternalID"),
				BasicField("Label", "string"),
				{"Properties", "map[string]Value", "`json:\"properties\"`", "namedFieldsTwoples", "namedFieldsTwoplesFromMap", "namedFieldsTwoples.toMap"},
			},
			"",
			"",
			"",
		},
		StructValue{
			"Rel",
			[]Field{
				BasicField("SrcNode", "InternalID"),
				BasicField("DstNode", "InternalID"),
				BasicField("Label", "string"),
				{"Properties", "map[string]Value", "`json:\"properties\"`", "namedFieldsTwoples", "namedFieldsTwoplesFromMap", "namedFieldsTwoples.toMap"},
			},
			"",
			"",
			"",
		},
		StructValue{
			"RecursiveRel",
			[]Field{
				BasicField("Nodes", "[]NodeValue"),
				BasicField("Rels", "[]RelValue"),
			},
			"",
			"",
			"",
		},
		StructValue{
			"Map",
			[]Field{
				LogicalTypeField("KeyType"),
				LogicalTypeField("ValueType"),
				{"Pairs", "map[Value]Value", "`json:\"pairs\"`", "valueMapJson", "valueMapJson", "map[Value]Value"},
			},
			"mapValueJson",
			"mapValueJson",
			"mapValue",
		},
		StructValue{
			"Union",
			[]Field{
				{"Types", "map[string]LogicalType", "`json:\"types\"`", "namedTypesTwoples", "namedTypesTwoplesFromMap", "namedTypesTwoples.toMap"},
				{"Value", "Value", "`json:\"value\"`", "valueUnmarshalHelper", "newValueUnmarshalHelper", "valueUnmarshalHelper.getInnerValue"},
			},
			"",
			"",
			"",
		},
		TypeAliasValue{
			"UUID",
			"uuid.UUID",
			"uuid.UUID",
			"uuid.UUID",
			"uuidValue",
		},
		TypeAliasValue{
			"Decimal",
			"decimal.Decimal",
			"decimal.Decimal",
			"decimal.Decimal",
			"decimalValue",
		},
	}

	_, thisFile, _, _ := runtime.Caller(0)
	genDir := filepath.Dir(thisFile)
	tmplFile := filepath.Join(genDir, "value.gotmpl")
	outFile := filepath.Join(genDir, "../../../pkg/graphdb/value.go")

	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"toLower": strings.ToLower,
			"toUpper": strings.ToUpper,
			"fail":    fail,
		}).ParseFiles(tmplFile))

	// execute template w/ data
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "value.gotmpl", valtypes)
	if err != nil {
		panic(fmt.Sprintf("unable to execute template: %v", err))
	}

	// format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("unable to format generated code: %v\n\n%v", err, buf.String()))
	}

	// write to file
	err = os.WriteFile(outFile, formatted, 0644)
	if err != nil {
		panic(fmt.Sprintf("unable to write generated code to file: %v", err))
	}
}
