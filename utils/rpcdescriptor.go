/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"reflect"
	"strings"
	"time"
)

// MethodDescriptor describes one RPC method's argument and reply fields.
type MethodDescriptor struct {
	Method string
	Args   []FieldDescriptor
	Result []FieldDescriptor
}

// FieldDescriptor describes one JSON field. Type is a display kind:
// a scalar (string, int, bool, float, duration, time), a struct's qualified name,
// "object" for an anonymous struct, or a []T/map[K]V of those.
type FieldDescriptor struct {
	Name   string
	Type   string
	Fields []FieldDescriptor
}

var (
	durationType = reflect.TypeFor[time.Duration]()
	timeType     = reflect.TypeFor[time.Time]()
)

// DescribeType returns t's fields, following encoding/json's rules.
func DescribeType(t reflect.Type) []FieldDescriptor {
	return describeStruct(t, map[reflect.Type]bool{})
}

func describeStruct(t reflect.Type, seen map[reflect.Type]bool) []FieldDescriptor {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	// Path scoped: defer delete lets a sibling of the same type still expand,
	// only an ancestor cycle is cut.
	if seen[t] {
		return nil
	}
	seen[t] = true
	defer delete(seen, t)

	// Field selection follows encoding/json's typeFields.
	var fields []FieldDescriptor
	for sf := range t.Fields() {
		if sf.Anonymous {
			ft := sf.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			// Keep embedded unexported structs: they may have exported fields.
			if !sf.IsExported() && ft.Kind() != reflect.Struct {
				continue
			}
		} else if !sf.IsExported() {
			continue
		}
		tag := sf.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if sf.Anonymous && name == "" {
			et := sf.Type
			for et.Kind() == reflect.Pointer {
				et = et.Elem()
			}
			if et.Kind() == reflect.Struct && et != timeType {
				fields = append(fields, describeStruct(et, seen)...)
				continue
			}
		}
		if name == "" {
			name = sf.Name
		}
		fields = append(fields, describeField(name, sf.Type, seen))
	}
	return fields
}

func describeField(name string, t reflect.Type, seen map[reflect.Type]bool) FieldDescriptor {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	fd := FieldDescriptor{Name: name}
	switch {
	case t == durationType:
		fd.Type = "duration"
	case t == timeType:
		fd.Type = "time"
	case t.Kind() == reflect.Struct:
		fd.Type = structName(t)
		fd.Fields = describeStruct(t, seen)
	default:
		fd.Type = typeName(t)
		if leaf := leafStruct(t); leaf != nil {
			fd.Fields = describeStruct(leaf, seen)
		}
	}
	return fd
}

func structName(t reflect.Type) string {
	if t.Name() == "" {
		return "object"
	}
	return t.String()
}

func leafStruct(t reflect.Type) reflect.Type {
	for {
		switch t.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Pointer:
			t = t.Elem()
		case reflect.Struct:
			if t == timeType {
				return nil
			}
			return t
		default:
			return nil
		}
	}
}

func typeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Pointer:
		return typeName(t.Elem())
	case reflect.Bool:
		return "bool"
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if t == durationType {
			return "duration"
		}
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Slice, reflect.Array:
		return "[]" + typeName(t.Elem())
	case reflect.Map:
		return "map[" + typeName(t.Key()) + "]" + typeName(t.Elem())
	case reflect.Interface:
		return "any"
	case reflect.Struct:
		if t == timeType {
			return "time"
		}
		return structName(t)
	default:
		return t.Kind().String()
	}
}
