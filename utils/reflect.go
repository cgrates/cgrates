/*
Real-time Charging System for Telecom & ISP environments
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
	"fmt"
	"reflect"
	"strconv"
)

func CastFieldIfToString(fld interface{}) (string, bool) {
	var strVal string
	var converted bool
	switch fld.(type) {
	case string:
		strVal = fld.(string)
		converted = true
	case int:
		strVal = strconv.Itoa(fld.(int))
		converted = true
	case int64:
		strVal = strconv.FormatInt(fld.(int64), 10)
		converted = true
	case bool:
		strVal = strconv.FormatBool(fld.(bool))
	case float64:
		strVal = strconv.FormatFloat(fld.(float64), 'f', -1, 64)
		converted = true
	case []uint8:
		var byteVal []byte
		if byteVal, converted = fld.([]byte); converted {
			strVal = string(byteVal)
		}
	default: // Maybe we are lucky and the value converts to string
		strVal, converted = fld.(string)
	}
	return strVal, converted
}

// ReflectFieldAsString parses intf and attepting to return the field as string or error otherwise
// Supports "ExtraFields" where additional fields are dynamically inserted in map with field name: extraFieldsLabel
func ReflectFieldAsString(intf interface{}, fldName, extraFieldsLabel string) (string, error) {
	v := reflect.ValueOf(intf)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var field reflect.Value
	switch v.Kind() {
	case reflect.Struct:
		field = v.FieldByName(fldName)
	case reflect.Map:
		field = v.MapIndex(reflect.ValueOf(fldName))
		if !field.IsValid() { // Not looking in extra fields anymore
			return "", ErrNotFound
		}
	default:
		return "", fmt.Errorf("Unsupported field kind: %v", v.Kind())
	}

	if !field.IsValid() {
		if extraFieldsLabel == "" {
			return "", ErrNotFound
		}
		mpVal := v.FieldByName(extraFieldsLabel)
		if !mpVal.IsValid() || mpVal.Kind() != reflect.Map {
			return "", ErrNotFound
		}
		field = mpVal.MapIndex(reflect.ValueOf(fldName))
		if !field.IsValid() {
			return "", ErrNotFound
		}
	}
	switch field.Kind() {
	case reflect.String:
		return field.String(), nil
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64), nil
	case reflect.Interface:
		strVal, converted := CastFieldIfToString(field.Interface())
		if !converted {
			return "", fmt.Errorf("Cannot convert to string field type: %s", field.Kind().String())
		} else {
			return strVal, nil
		}
	default:
		return "", fmt.Errorf("Cannot convert to string field type: %s", field.Kind().String())
	}
}
