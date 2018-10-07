/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Detects missing field values based on mandatory field names, s should be a pointer to a struct
func MissingStructFields(s interface{}, mandatories []string) []string {
	missing := []string{}
	for _, fieldName := range mandatories {
		fld := reflect.ValueOf(s).Elem().FieldByName(fieldName)
		// sanitize the string fields before checking
		if fld.Kind() == reflect.String && fld.CanSet() {
			fld.SetString(strings.TrimSpace(fld.String()))
		}
		if (fld.Kind() == reflect.String && fld.String() == "") ||
			((fld.Kind() == reflect.Slice || fld.Kind() == reflect.Map) && fld.Len() == 0) ||
			(fld.Kind() == reflect.Int && fld.Int() == 0) {
			missing = append(missing, fieldName)
		}
	}
	return missing
}

// Detects nonempty struct fields, s should be a pointer to a struct
// Useful to not overwrite db fields with non defined params in api
func NonemptyStructFields(s interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < reflect.ValueOf(s).Elem().NumField(); i++ {
		fld := reflect.ValueOf(s).Elem().Field(i)
		switch fld.Kind() {
		case reflect.Bool:
			fields[reflect.TypeOf(s).Elem().Field(i).Name] = fld.Bool()
		case reflect.Int:
			fieldVal := fld.Int()
			if fieldVal != 0 {
				fields[reflect.TypeOf(s).Elem().Field(i).Name] = fieldVal
			}
		case reflect.String:
			fieldVal := fld.String()
			if fieldVal != "" {
				fields[reflect.TypeOf(s).Elem().Field(i).Name] = fieldVal
			}
		}
	}
	return fields
}

// Converts a struct to map
/*func StrucToMap(s interface{}) map[string]interface{} {
	mp := make(map[string]interface{})
	for i := 0; i < reflect.ValueOf(s).Elem().NumField(); i++ {
		fld := reflect.ValueOf(s).Elem().Field(i)
		switch fld.Kind() {
		case reflect.Bool:
			mp[reflect.TypeOf(s).Elem().Field(i).Name] = fld.Bool()
		case reflect.Int:
			mp[reflect.TypeOf(s).Elem().Field(i).Name] = fld.Int()
		case reflect.String:
			mp[reflect.TypeOf(s).Elem().Field(i).Name] = fld.String()
		}
	}
	return mp
}*/

// Converts a struct to map[string]interface{}
func ToMapMapStringInterface(in interface{}) map[string]interface{} { // Got error and it is not used anywhere
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	typ := reflect.TypeOf(in)
	for i := 0; i < v.NumField(); i++ {
		out[typ.Field(i).Name] = v.Field(i).Interface()
	}
	return out
}

// Converts a struct to map[string]string
func ToMapStringString(in interface{}) map[string]string {
	out := make(map[string]string)

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		in = v.Interface()
	}
	typ := reflect.TypeOf(in)
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		typField := typ.Field(i)
		field := v.Field(i)
		if field.Kind() == reflect.String {
			out[typField.Name] = field.String()
		}
	}
	return out
}

func GetMapExtraFields(in interface{}, extraFields string) map[string]string {
	out := make(map[string]string)
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(extraFields)
	if field.Kind() == reflect.Map {
		keys := field.MapKeys()
		for _, key := range keys {
			out[key.String()] = field.MapIndex(key).String()
		}
	}
	return out
}

func SetMapExtraFields(in interface{}, values map[string]string, extraFields string) {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	efField := v.FieldByName(extraFields)
	if efField.IsValid() && efField.Kind() == reflect.Map {
		keys := efField.MapKeys()
		for _, key := range keys {
			if efField.MapIndex(key).String() != "" {
				if val, found := values[key.String()]; found {
					efField.SetMapIndex(key, reflect.ValueOf(val))
				}
			}
		}
	}
	return
}

func FromMapStringString(m map[string]string, in interface{}) {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for fieldName, fieldValue := range m {
		field := v.FieldByName(fieldName)
		if field.IsValid() {
			if field.Kind() == reflect.String {
				if field.String() != "" && field.CanSet() {
					field.SetString(fieldValue)
				}
			}
		}
	}
	return
}

func FromMapStringInterface(m map[string]interface{}, in interface{}) error {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for fieldName, fieldValue := range m {
		field := v.FieldByName(fieldName)
		if field.IsValid() {
			if !field.IsValid() || !field.CanSet() {
				continue
			}
			structFieldType := field.Type()
			val := reflect.ValueOf(fieldValue)
			if structFieldType != val.Type() {
				return errors.New("Provided value type didn't match obj field type")
			}
			field.Set(val)
		}
	}
	return nil
}

// initial intent was to use it with *cgr_rpc but does not handle slice and structure fields
func FromMapStringInterfaceValue(m map[string]interface{}, v reflect.Value) (interface{}, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for fieldName, fieldValue := range m {
		field := v.FieldByName(fieldName)
		if field.IsValid() {
			if !field.IsValid() || !field.CanSet() {
				continue
			}
			val := reflect.ValueOf(fieldValue)
			structFieldType := field.Type()
			if structFieldType.Kind() == reflect.Ptr {
				field.Set(reflect.New(field.Type().Elem()))
				field = field.Elem()
			}
			structFieldType = field.Type()
			if structFieldType != val.Type() {
				return nil, fmt.Errorf("provided value type didn't match obj field type: %v vs %v (%v vs %v)", structFieldType, val.Type(), structFieldType.Kind(), val.Type().Kind())
			}
			field.Set(val)
		}
	}
	return v.Interface(), nil
}

// Update struct with map fields, returns not matching map keys, s is a struct to be updated
func UpdateStructWithStrMap(s interface{}, m map[string]string) []string { // Not tested and declared and used only here
	notMatched := []string{}
	for key, val := range m {
		fld := reflect.ValueOf(s).Elem().FieldByName(key)
		if fld.IsValid() {
			switch fld.Kind() {
			case reflect.Bool:
				if valBool, err := strconv.ParseBool(val); err != nil {
					notMatched = append(notMatched, key)
				} else {
					fld.SetBool(valBool)
				}
			case reflect.Int:
				if valInt, err := strconv.ParseInt(val, 10, 64); err != nil {
					notMatched = append(notMatched, key)
				} else {
					fld.SetInt(valInt)
				}
			case reflect.String:
				fld.SetString(val)
			}
		} else {
			notMatched = append(notMatched, key)
		}
	}
	return notMatched
}

// UpdateStructWithIfaceMap will update struct fields with values coming from map
// if map values are not matching the ones in struct convertion is being attempted
// ToDo: add here more fields
func UpdateStructWithIfaceMap(s interface{}, mp map[string]interface{}) (err error) {
	for key, val := range mp {
		fld := reflect.ValueOf(s).Elem().FieldByName(key)
		if fld.IsValid() {
			switch fld.Kind() {
			case reflect.Bool:
				if val == "" { // auto-populate defaults so we can parse empty strings
					val = false
				}
				if valBool, err := IfaceAsBool(val); err != nil {
					return err
				} else {
					fld.SetBool(valBool)
				}
			case reflect.Int, reflect.Int64:
				if val == "" {
					val = 0
				}
				if valInt, err := IfaceAsInt64(val); err != nil {
					return err
				} else {
					fld.SetInt(valInt)
				}
			case reflect.Float64:
				if val == "" {
					val = 0.0
				}
				if valFlt, err := IfaceAsFloat64(val); err != nil {
					return err
				} else {
					fld.SetFloat(valFlt)
				}
			case reflect.String:
				if valStr, err := IfaceAsString(val); err != nil {
					return fmt.Errorf("cannot convert field: %+v to string", val)
				} else {
					fld.SetString(valStr)
				}
			default: // improper use of function
				return fmt.Errorf("cannot update unsupported struct field: %+v", fld)
			}
		}
	}
	return
}
