/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"reflect"
	"strconv"
)

// Detects missing field values based on mandatory field names, s should be a pointer to a struct
func MissingStructFields(s interface{}, mandatories []string) []string {
	missing := []string{}
	for _, fieldName := range mandatories {
		fld := reflect.ValueOf(s).Elem().FieldByName(fieldName)
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
func StrucToMap(s interface{}) map[string]interface{} {
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
}

// Update struct with map fields, returns not matching map keys, s is a struct to be updated
func UpdateStructWithStrMap(s interface{}, m map[string]string) []string {
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
