// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"reflect"
	"strings"
)

func fieldByIndexIsEmpty(v reflect.Value, index []int) bool {
	if len(index) == 1 {
		return valueIsEmpty(v.Field(index[0]))
	}
	for i, x := range index {
		if i > 0 {
			if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					return true
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}
	return valueIsEmpty(v)
}

func valueIsEmpty(fld reflect.Value) bool {
	if fld.Kind() == reflect.String && fld.CanSet() {
		fld.SetString(strings.TrimSpace(fld.String()))
	}
	return fld.Kind() == reflect.String && fld.String() == EmptyString ||
		(fld.Kind() == reflect.Slice || fld.Kind() == reflect.Map) && fld.Len() == 0 ||
		fld.Kind() == reflect.Int && fld.Int() == 0
}

// Detects missing field values based on mandatory field names, s should be a pointer to a struct
func MissingStructFields(s any, mandatories []string) []string {
	missing := []string{}
	sValue := reflect.ValueOf(s).Elem()
	sType := sValue.Type()
	for _, fieldName := range mandatories {
		fldStr, ok := sType.FieldByName(fieldName)
		if !ok || fieldByIndexIsEmpty(sValue, fldStr.Index) {
			missing = append(missing, fieldName)
		}
	}
	return missing
}

// UpdateStructWithIfaceMap will update struct fields with values coming from map
// if map values are not matching the ones in struct convertion is being attempted
// ToDo: add here more fields
func UpdateStructWithIfaceMap(s any, mp map[string]any) (err error) {
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
				fld.SetString(IfaceAsString(val))
			default: // improper use of function
				return fmt.Errorf("cannot update unsupported struct field: %+v", fld)
			}
		}
	}
	return
}
