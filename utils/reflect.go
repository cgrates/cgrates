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
	case int:
		strVal = strconv.Itoa(fld.(int))
		converted = true
	case int64:
		strVal = strconv.FormatInt(fld.(int64), 10)
		converted = true
	case bool:
		strVal = strconv.FormatBool(fld.(bool))
		converted = true
	case float64:
		strVal = strconv.FormatFloat(fld.(float64), 'f', -1, 64)
		converted = true
	case []uint8:
		var byteVal []byte
		if byteVal, converted = fld.([]byte); converted {
			strVal = string(byteVal)
		}
	case time.Duration:
		strVal = fld.(time.Duration).String()
		converted = true
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
	// time.Time
	if t, err := ParseTimeDetectLayout(s, "Local"); err == nil {
		return t
	}
	// time.Duration
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	// string
	return s
}

// ReflectFieldInterface parses intf attepting to return the field value or error otherwise
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

func IfaceAsTime(itm interface{}, timezone string) (t time.Time, err error) {
	switch itm.(type) {
	case time.Time:
		return itm.(time.Time), nil
	case string:
		return ParseTimeDetectLayout(itm.(string), timezone)
	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Time", itm)
	}
	return
}

func IfaceAsDuration(itm interface{}) (d time.Duration, err error) {
	switch itm.(type) {
	case time.Duration:
		return itm.(time.Duration), nil
	case float64:
		return time.Duration(int64(itm.(float64) * float64(time.Second))), nil
	case int64:
		return time.Duration(itm.(int64)), nil
	case string:
		return ParseDurationWithNanosecs(itm.(string))

	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Time", itm)
	}
	return
}

func IfaceAsFloat64(itm interface{}) (f float64, err error) {
	switch itm.(type) {
	case float64:
		return itm.(float64), nil
	case time.Duration:
		return float64(itm.(time.Duration).Nanoseconds()), nil
	case int64:
		return float64(itm.(int64)), nil
	case string:
		return strconv.ParseFloat(itm.(string), 64)
	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Time", itm)
	}
	return
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
	valItm := reflect.ValueOf(item)
	valOtItm := reflect.ValueOf(oItem)
	// convert to wider type so we can be compatible with StringToInterface function
	switch valItm.Kind() {
	case reflect.Float32:
		item = valItm.Float()
		valItm = reflect.ValueOf(item)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		item = valItm.Int()
		valItm = reflect.ValueOf(item)
	}
	switch valOtItm.Kind() {
	case reflect.Float32:
		oItem = valOtItm.Float()
		valOtItm = reflect.ValueOf(oItem)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		oItem = valOtItm.Int()
		valOtItm = reflect.ValueOf(oItem)
	}
	typItem := reflect.TypeOf(item)
	typOItem := reflect.TypeOf(oItem)
	if !typItem.Comparable() ||
		!typOItem.Comparable() ||
		typItem != typOItem {
		return false, errors.New("incomparable")
	}
	switch item.(type) {
	case float64:
		if orEqual {
			gte = valItm.Float() >= valOtItm.Float()
		} else {
			gte = valItm.Float() > valOtItm.Float()
		}
	case int64:
		if orEqual {
			gte = valItm.Int() >= valOtItm.Int()
		} else {
			gte = valItm.Int() > valOtItm.Int()
		}
	case time.Time:
		tVal := item.(time.Time)
		tOVal := oItem.(time.Time)
		if orEqual {
			gte = tVal == tOVal
		}
		if !gte {
			gte = tVal.After(tOVal)
		}
	case time.Duration:
		tVal := item.(time.Duration)
		tOVal := oItem.(time.Duration)
		if orEqual {
			gte = tVal == tOVal
		}
		if !gte {
			gte = tVal > tOVal
		}

	default: // unsupported comparison
		err = fmt.Errorf("unsupported comparison type: %v, kind: %v", typItem, typItem.Kind())
	}
	return
}
