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
	"time"
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

// StringToInterface will parse string into supported types
// if no other conversion possible, original string will be returned
func StringToInterface(s string) interface{} {
	// int64
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	// bool
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	// float64
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// time.Duration
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	// string
	return s
}

// ReflectFieldInterface parses intf attepting to return the field as string or error otherwise
// Supports "ExtraFields" where additional fields are dynamically inserted in map with field name: extraFieldsLabel
func ReflectFieldInterface(intf interface{}, fldName, extraFieldsLabel string) (retIf interface{}, err error) {
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
			return nil, ErrNotFound
		}
	default:
		return nil, fmt.Errorf("Unsupported field kind: %v", v.Kind())
	}

	if !field.IsValid() {
		if extraFieldsLabel == "" {
			return nil, ErrNotFound
		}
		mpVal := v.FieldByName(extraFieldsLabel)
		if !mpVal.IsValid() || mpVal.Kind() != reflect.Map {
			return nil, ErrNotFound
		}
		field = mpVal.MapIndex(reflect.ValueOf(fldName))
		if !field.IsValid() {
			return nil, ErrNotFound
		}
	}
	return field.Interface(), nil
}

// ReflectFieldAsString parses intf and attepting to return the field as string or error otherwise
// Supports "ExtraFields" where additional fields are dynamically inserted in map with field name: extraFieldsLabel
func ReflectFieldAsString(intf interface{}, fldName, extraFieldsLabel string) (string, error) {
	field, err := ReflectFieldInterface(intf, fldName, extraFieldsLabel)
	if err != nil {
		return "", err
	}
	vOf := reflect.ValueOf(field)
	switch vOf.Kind() {
	case reflect.String:
		return vOf.String(), nil
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(vOf.Int(), 10), nil
	case reflect.Float64:
		return strconv.FormatFloat(vOf.Float(), 'f', -1, 64), nil
	case reflect.Interface:
		strVal, converted := CastFieldIfToString(field)
		if !converted {
			return "", fmt.Errorf("Cannot convert to string field type: %s", vOf.Kind().String())
		} else {
			return strVal, nil
		}
	default:
		return "", fmt.Errorf("Cannot convert to string field type: %s", vOf.Kind().String())
	}
}

// AsMapStringIface converts an item (mostly struct) as map[string]interface{}
func AsMapStringIface(item interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct { // Only structs for now
		return nil, fmt.Errorf("AsMapStringIface only accepts structs; got %T", v)
	}
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		out[typ.Field(i).Name] = v.Field(i).Interface()
	}
	return out, nil
}

// GreaterThan attempts to compare two items
// returns the result or error if not comparable
func GreaterThan(item, oItem interface{}, orEqual bool) (gte bool, err error) {
	typItem := reflect.TypeOf(item)
	typOItem := reflect.TypeOf(oItem)
	fmt.Println(typItem.Comparable(),
		typOItem.Comparable(),
		typItem,
		typOItem,
		typItem == typOItem)
	if !typItem.Comparable() ||
		!typOItem.Comparable() ||
		typItem != typOItem {
		return false, errors.New("incomparable")
	}
	if orEqual && reflect.DeepEqual(item, oItem) {
		return true, nil
	}
	valItm := reflect.ValueOf(item)
	valOItm := reflect.ValueOf(oItem)
	switch typItem.Kind() {
	case reflect.Float32, reflect.Float64:
		gte = valItm.Float() > valOItm.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		gte = valItm.Int() > valOItm.Int()
	default: // unsupported comparison
		err = fmt.Errorf("unsupported type: %v", typItem)
	}
	return
}
