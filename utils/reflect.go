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
	"net"
	"reflect"
	"strconv"
	"time"
)

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
		strVal, err := IfaceAsString(field)
		if err != nil {
			return "", fmt.Errorf("Cannot convert to string field type: %s", vOf.Kind().String())
		}
		return strVal, nil
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
	case float64: // automatically hitting here also ints
		return time.Duration(int64(itm.(float64))), nil
	case int64:
		return time.Duration(itm.(int64)), nil
	case int:
		return time.Duration(itm.(int)), nil
	case string:
		return ParseDurationWithNanosecs(itm.(string))

	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Duration", itm)
	}
	return
}

func IfaceAsInt64(itm interface{}) (i int64, err error) {
	switch itm.(type) {
	case int:
		return int64(itm.(int)), nil
	case time.Duration:
		return itm.(time.Duration).Nanoseconds(), nil
	case int64:
		return itm.(int64), nil
	case string:
		return strconv.ParseInt(itm.(string), 10, 64)
	default:
		err = fmt.Errorf("cannot convert field: %+v to int", itm)
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
		err = fmt.Errorf("cannot convert field: %+v to float64", itm)
	}
	return
}

func IfaceAsBool(itm interface{}) (b bool, err error) {
	switch itm.(type) {
	case bool:
		return itm.(bool), nil
	case string:
		return strconv.ParseBool(itm.(string))
	case int:
		return itm.(int) > 0, nil
	case int64:
		return itm.(int64) > 0, nil
	case float64:
		return itm.(float64) > 0, nil
	default:
		err = fmt.Errorf("cannot convert field: %+v to bool", itm)
	}
	return
}

func IfaceAsString(fld interface{}) (out string, err error) {
	switch fld.(type) {
	case nil:
		return
	case int:
		return strconv.Itoa(fld.(int)), nil
	case int32:
		return strconv.FormatInt(int64(fld.(int32)), 10), nil
	case int64:
		return strconv.FormatInt(fld.(int64), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(fld.(uint32)), 10), nil
	case uint64:
		return strconv.FormatUint(fld.(uint64), 10), nil
	case bool:
		return strconv.FormatBool(fld.(bool)), nil
	case float32:
		return strconv.FormatFloat(float64(fld.(float32)), 'f', -1, 64), nil
	case float64:
		return strconv.FormatFloat(fld.(float64), 'f', -1, 64), nil
	case []uint8:
		if byteVal, canCast := fld.([]byte); !canCast {
			return "", ErrNotConvertibleNoCaps
		} else {
			return string(byteVal), nil
		}
	case time.Duration:
		return fld.(time.Duration).String(), nil
	case time.Time:
		return fld.(time.Time).Format(time.RFC3339), nil
	case net.IP:
		return fld.(net.IP).String(), nil
	case string:
		return fld.(string), nil
	default: // Maybe we are lucky and the value converts to string
		return ToJSON(fld), nil
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

// Sum attempts to sum multiple items
// returns the result or error if not comparable
func Sum(items ...interface{}) (sum interface{}, err error) {
	//we need at least 2 items to sum them
	if len(items) < 2 {
		return nil, ErrNotEnoughParameters
	}

	// convert the type for first item
	valItm := reflect.ValueOf(items[0])
	switch valItm.Kind() {
	case reflect.Float32:
		items[0] = valItm.Float()
		valItm = reflect.ValueOf(items[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		items[0] = valItm.Int()
		valItm = reflect.ValueOf(items[0])
	}
	typItem := reflect.TypeOf(items[0])

	//populate sum with first item so we can add after
	switch items[0].(type) {
	case float64:
		sum = valItm.Float()
	case int64:
		sum = valItm.Int()
	case time.Duration:
		tVal := items[0].(time.Duration)
		sum = tVal
	}

	for _, item := range items[1:] {
		valOtItm := reflect.ValueOf(item)
		// convert to wider type so we can be compatible with StringToInterface function
		switch valOtItm.Kind() {
		case reflect.Float32:
			item = valOtItm.Float()
			valOtItm = reflect.ValueOf(item)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			item = valOtItm.Int()
			valOtItm = reflect.ValueOf(item)
		}
		typOItem := reflect.TypeOf(item)
		// check if the type for rest items is the same as first
		if !typItem.Comparable() ||
			!typOItem.Comparable() ||
			typItem != typOItem {
			return false, errors.New("incomparable")
		}

		switch items[0].(type) {
		case float64:
			sum = reflect.ValueOf(sum).Float() + valOtItm.Float()
		case int64:
			sum = reflect.ValueOf(sum).Int() + valOtItm.Int()
		case time.Duration:
			tOVal := item.(time.Duration)
			sum = sum.(time.Duration) + tOVal
		default: // unsupported comparison
			err = fmt.Errorf("unsupported comparison type: %v, kind: %v", typItem, typItem.Kind())
			break
		}
	}
	return

}
