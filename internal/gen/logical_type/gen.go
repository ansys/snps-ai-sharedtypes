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

type LogicalTypeDef struct {
	Tag         string
	UnitVariant bool
	Fields      []Field
}

func UnitLogicalType(tag string) LogicalTypeDef {
	return LogicalTypeDef{Tag: tag, UnitVariant: true, Fields: nil}
}

func StructLogicalType(tag string, fields []Field) LogicalTypeDef {
	return LogicalTypeDef{Tag: tag, UnitVariant: false, Fields: fields}
}

type Field struct {
	Name             string
	Type             string
	Tag              string
	UnmarshallerType string
	Converter        string
}

func jsonTag(name string) string {
	snakeName := strcase.ToSnake(name)
	return fmt.Sprintf("`json:%q`", snakeName)
}

func BasicField(name string, typ string) Field {
	return Field{Name: name, Type: typ, Tag: jsonTag(name), UnmarshallerType: "", Converter: ""}
}
func LogicalTypeField(name string) Field {
	return Field{Name: name, Type: "LogicalType", Tag: jsonTag(name), UnmarshallerType: "logicalTypeUnmarshalHelper", Converter: "logicalTypeUnmarshalHelper.getInnerLogicalType"}
}

func NamedTypesField(name string) Field {
	return Field{Name: name, Type: "[]Twople[string, LogicalType]", Tag: jsonTag(name), UnmarshallerType: "[]Twople[string, logicalTypeUnmarshalHelper]", Converter: "getTwopleLogicalTypeFromHelper"}
}

func main() {
	logtypes := []LogicalTypeDef{
		UnitLogicalType("Any"),
		UnitLogicalType("Bool"),
		UnitLogicalType("Serial"),
		UnitLogicalType("Int64"),
		UnitLogicalType("Int32"),
		UnitLogicalType("Int16"),
		UnitLogicalType("Int8"),
		UnitLogicalType("UInt64"),
		UnitLogicalType("UInt32"),
		UnitLogicalType("UInt16"),
		UnitLogicalType("UInt8"),
		UnitLogicalType("Int128"),
		UnitLogicalType("Double"),
		UnitLogicalType("Float"),
		UnitLogicalType("Date"),
		UnitLogicalType("Interval"),
		UnitLogicalType("Timestamp"),
		UnitLogicalType("TimestampTz"),
		UnitLogicalType("TimestampNs"),
		UnitLogicalType("TimestampMs"),
		UnitLogicalType("TimestampSec"),
		UnitLogicalType("InternalID"),
		UnitLogicalType("String"),
		UnitLogicalType("Blob"),
		StructLogicalType("List", []Field{
			LogicalTypeField("ChildType"),
		}),
		StructLogicalType("Array", []Field{
			LogicalTypeField("ChildType"),
			BasicField("NumElements", "uint64"),
		}),
		StructLogicalType("Struct", []Field{
			NamedTypesField("Fields"),
		}),
		UnitLogicalType("Node"),
		UnitLogicalType("Rel"),
		UnitLogicalType("RecursiveRel"),
		StructLogicalType("Map", []Field{
			LogicalTypeField("KeyType"),
			LogicalTypeField("ValueType"),
		}),
		StructLogicalType("Union", []Field{
			NamedTypesField("Fields"),
		}),
		UnitLogicalType("UUID"),
		StructLogicalType("Decimal", []Field{
			BasicField("Precision", "uint32"),
			BasicField("Scale", "uint32"),
		}),
	}

	_, thisFile, _, _ := runtime.Caller(0)
	genDir := filepath.Dir(thisFile)
	tmplFile := filepath.Join(genDir, "logical_type.gotmpl")
	outFile := filepath.Join(genDir, "../../../pkg/graphdb/logical_type.go")

	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"toLower": strings.ToLower,
			"toUpper": strings.ToUpper,
		}).ParseFiles(tmplFile))

	// execute template w/ data
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "logical_type.gotmpl", logtypes)
	if err != nil {
		panic(fmt.Sprintf("unable to execute template: %v", err))
	}

	// format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("unable to format generated code: %v", err))
	}

	// write to file
	err = os.WriteFile(outFile, formatted, 0644)
	if err != nil {
		panic(fmt.Sprintf("unable to write generated code to file: %v", err))
	}
}
