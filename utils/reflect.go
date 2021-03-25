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
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

// StringToInterface will parse string into supported types
// if no other conversion possible, original string will be returned
func StringToInterface(s string) interface{} {
	if s == EmptyString {
		return s
	}
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
	default:
		return "", fmt.Errorf("Cannot convert to string field type: %s", vOf.Kind().String())
	}
}

func IfaceAsTime(itm interface{}, timezone string) (t time.Time, err error) {
	switch v := itm.(type) {
	case time.Time:
		return v, nil
	case string:
		return ParseTimeDetectLayout(v, timezone)
	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Time", itm)
	}
	return
}

func IfaceAsBig(itm interface{}) (b *decimal.Big, err error) {
	switch it := itm.(type) {
	case time.Duration:
		return decimal.New(int64(it), 0), nil
	case int: // check every int type
		return decimal.New(int64(it), 0), nil
	case int8:
		return decimal.New(int64(it), 0), nil
	case int16:
		return decimal.New(int64(it), 0), nil
	case int32:
		return decimal.New(int64(it), 0), nil
	case int64:
		return decimal.New(it, 0), nil
	case uint:
		return new(decimal.Big).SetUint64(uint64(it)), nil
	case uint8:
		return new(decimal.Big).SetUint64(uint64(it)), nil
	case uint16:
		return new(decimal.Big).SetUint64(uint64(it)), nil
	case uint32:
		return new(decimal.Big).SetUint64(uint64(it)), nil
	case uint64:
		return new(decimal.Big).SetUint64(it), nil
	case float32: // automatically hitting here also ints
		return new(decimal.Big).SetFloat64(float64(it)), nil
	case float64: // automatically hitting here also ints
		return new(decimal.Big).SetFloat64(it), nil
	case string:
		if strings.HasSuffix(it, NsSuffix) ||
			strings.HasSuffix(it, UsSuffix) ||
			strings.HasSuffix(it, µSuffix) ||
			strings.HasSuffix(it, MsSuffix) ||
			strings.HasSuffix(it, SSuffix) ||
			strings.HasSuffix(it, MSuffix) ||
			strings.HasSuffix(it, HSuffix) {
			var tm time.Duration
			if tm, err = time.ParseDuration(it); err != nil {
				return
			}
			return decimal.New(int64(tm), 0), nil
		}
		z, ok := new(decimal.Big).SetString(it)
		// verify ok and check if the value was converted successfuly
		// and the big is a valid number
		if !ok || z.IsNaN(0) {
			return nil, fmt.Errorf("can't convert <%+v> to decimal", it)
		}
		return z, nil
	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Duration", it)
	}
	return
}

func IfaceAsDuration(itm interface{}) (d time.Duration, err error) {
	switch it := itm.(type) {
	case time.Duration:
		return it, nil
	case int: // check every int type
		return time.Duration(int64(it)), nil
	case int8:
		return time.Duration(int64(it)), nil
	case int16:
		return time.Duration(int64(it)), nil
	case int32:
		return time.Duration(int64(it)), nil
	case int64:
		return time.Duration(int64(it)), nil
	case uint:
		return time.Duration(int64(it)), nil
	case uint8:
		return time.Duration(int64(it)), nil
	case uint16:
		return time.Duration(int64(it)), nil
	case uint32:
		return time.Duration(int64(it)), nil
	case uint64:
		return time.Duration(int64(it)), nil
	case float32: // automatically hitting here also ints
		return time.Duration(int64(it)), nil
	case float64: // automatically hitting here also ints
		return time.Duration(int64(it)), nil
	case string:
		return ParseDurationWithNanosecs(it)
	default:
		err = fmt.Errorf("cannot convert field: %+v to time.Duration", it)
	}
	return
}

func IfaceAsInt64(itm interface{}) (i int64, err error) {
	switch it := itm.(type) {
	case int:
		return int64(it), nil
	case time.Duration:
		return it.Nanoseconds(), nil
	case int32:
		return int64(it), nil
	case int64:
		return it, nil
	case string:
		return strconv.ParseInt(it, 10, 64)
	default:
		err = fmt.Errorf("cannot convert field: %+v to int", it)
	}
	return
}

// same function as IfaceAsInt64 but if the value is float round it to int64 instead of returning error
func IfaceAsTInt64(itm interface{}) (i int64, err error) {
	switch it := itm.(type) {
	case int:
		return int64(it), nil
	case time.Duration:
		return it.Nanoseconds(), nil
	case int32:
		return int64(it), nil
	case int64:
		return it, nil
	case float32:
		return int64(it), nil
	case float64:
		return int64(it), nil
	case string:
		return strconv.ParseInt(it, 10, 64)
	default:
		err = fmt.Errorf("cannot convert field<%T>: %+v to int", it, it)
	}
	return
}

func IfaceAsFloat64(itm interface{}) (f float64, err error) {
	switch it := itm.(type) {
	case float64:
		return it, nil
	case time.Duration:
		return float64(it.Nanoseconds()), nil
	case int:
		return float64(it), nil
	case int64:
		return float64(it), nil
	case string:
		return strconv.ParseFloat(it, 64)
	default:
		err = fmt.Errorf("cannot convert field: %+v to float64", it)
	}
	return
}

func IfaceAsBool(itm interface{}) (b bool, err error) {
	switch v := itm.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v > 0, nil
	case int64:
		return v > 0, nil
	case float64:
		return v > 0, nil
	default:
		err = fmt.Errorf("cannot convert field: %+v to bool", itm)
	}
	return
}

func IfaceAsString(fld interface{}) (out string) {
	switch value := fld.(type) {
	case nil:
		return
	case int:
		return strconv.Itoa(value)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case bool:
		return strconv.FormatBool(value)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case []uint8:
		return string(value) // byte is an alias for uint8 conversions implicit
	case time.Duration:
		return value.String()
	case time.Time:
		return value.Format(time.RFC3339)
	case net.IP:
		return value.String()
	case string:
		return value
	default: // Maybe we are lucky and the value converts to string
		return ToJSON(fld)
	}
}

// IfaceAsSliceString is trying to convert the interface to a slice of strings
func IfaceAsSliceString(fld interface{}) (out []string, err error) {
	switch value := fld.(type) {
	case nil:
		return
	case []int:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.Itoa(val)
		}
	case []int32:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatInt(int64(val), 10)
		}
	case []int64:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatInt(val, 10)
		}
	case []uint:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatUint(uint64(val), 10)
		}
	case []uint32:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatUint(uint64(val), 10)
		}
	case []uint64:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatUint(val, 10)
		}
	case []bool:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatBool(val)
		}
	case []float32:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatFloat(float64(val), 'f', -1, 64)
		}
	case []float64:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = strconv.FormatFloat(val, 'f', -1, 64)
		}
	case [][]uint8:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = string(val) // byte is an alias for uint8 conversions implicit
		}
	case []time.Duration:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = val.String()
		}
	case []time.Time:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = val.Format(time.RFC3339)
		}
	case []net.IP:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = val.String()
		}
	case []string:
		out = value
	case []interface{}:
		out = make([]string, len(value))
		for i, val := range value {
			out[i] = IfaceAsString(val)
		}
	default:
		err = fmt.Errorf("cannot convert field: %T to []string", value)
	}
	return
}

func GetUniformType(item interface{}) (interface{}, error) {
	valItm := reflect.ValueOf(item)
	switch valItm.Kind() { // convert evreting to float64
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(valItm.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(valItm.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return valItm.Float(), nil
	case reflect.Struct: // used only for time
		return valItm.Interface(), nil
	default:
		return nil, errors.New("incomparable")
	}
}

func GetBasicType(item interface{}) interface{} {
	valItm := reflect.ValueOf(item)
	switch valItm.Kind() { // convert evreting to float64
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return valItm.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return valItm.Uint()
	case reflect.Float32, reflect.Float64:
		return valItm.Float()
	default:
		return item
	}
}

// GreaterThan attempts to compare two items
// returns the result or error if not comparable
func GreaterThan(item, oItem interface{}, orEqual bool) (gte bool, err error) {
	item = GetBasicType(item)
	oItem = GetBasicType(oItem)
	typItem := reflect.TypeOf(item)
	typOItem := reflect.TypeOf(oItem)
	if typItem != typOItem {
		if item, err = GetUniformType(item); err == nil { // overwrite type only if possible
			typItem = reflect.TypeOf(item)
		}
		if oItem, err = GetUniformType(oItem); err == nil {
			typOItem = reflect.TypeOf(oItem)
		}
	}
	if !typItem.Comparable() ||
		!typOItem.Comparable() ||
		typItem != typOItem {
		return false, fmt.Errorf("incomparable: <%+v> with <%+v>", item, oItem)
	}
	switch tVal := item.(type) {
	case float64:
		tOVal := oItem.(float64)
		if orEqual {
			gte = tVal >= tOVal
		} else {
			gte = tVal > tOVal
		}
	case uint64:
		tOVal := oItem.(uint64)
		if orEqual {
			gte = tVal >= tOVal
		} else {
			gte = tVal > tOVal
		}
	case int64:
		tOVal := oItem.(int64)
		if orEqual {
			gte = tVal >= tOVal
		} else {
			gte = tVal > tOVal
		}
	case time.Time:
		tOVal := oItem.(time.Time)
		if orEqual {
			gte = tVal == tOVal
		}
		if !gte {
			gte = tVal.After(tOVal)
		}
	default: // unsupported comparison
		err = fmt.Errorf("unsupported comparison type: %v, kind: %v", typItem, typItem.Kind())
	}
	return
}

func EqualTo(item, oItem interface{}) (eq bool, err error) {
	item = GetBasicType(item)
	oItem = GetBasicType(oItem)
	typItem := reflect.TypeOf(item)
	typOItem := reflect.TypeOf(oItem)
	if typItem != typOItem {
		if item, err = GetUniformType(item); err == nil { // overwrite type only if possible
			typItem = reflect.TypeOf(item)
		}
		if oItem, err = GetUniformType(oItem); err == nil {
			typOItem = reflect.TypeOf(oItem)
		}
	}
	if !typItem.Comparable() ||
		!typOItem.Comparable() ||
		typItem != typOItem {
		return false, fmt.Errorf("incomparable: <%+v> with <%+v>", item, oItem)
	}
	switch tVal := item.(type) {
	case float64:
		tOVal := oItem.(float64)
		eq = tVal == tOVal
	case uint64:
		tOVal := oItem.(uint64)
		eq = tVal == tOVal
	case int64:
		tOVal := oItem.(int64)
		eq = tVal == tOVal
	case time.Time:
		tOVal := oItem.(time.Time)
		eq = tVal == tOVal
	case string:
		tOVal := oItem.(string)
		eq = tVal == tOVal
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

	switch dt := items[0].(type) {
	case time.Duration:
		sum = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsDuration(item); err != nil {
				return nil, err
			} else {
				sum = sum.(time.Duration) + itmVal
			}
		}
	case time.Time:
		sum = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsDuration(item); err != nil {
				return nil, err
			} else {
				sum = sum.(time.Time).Add(itmVal)
			}
		}
	case float64:
		sum = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsFloat64(item); err != nil {
				return nil, err
			} else {
				sum = sum.(float64) + itmVal
			}
		}
	case int64:
		sum = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				sum = sum.(int64) + itmVal
			}
		}
	case int:
		sum = int64(dt)
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				sum = sum.(int64) + itmVal
			}
		}
	}
	return
}

// Difference attempts to sum multiple items
// returns the result or error if not comparable
func Difference(items ...interface{}) (diff interface{}, err error) {
	//we need at least 2 items to diff them
	if len(items) < 2 {
		return nil, ErrNotEnoughParameters
	}
	switch dt := items[0].(type) {
	case time.Duration:
		diff = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsDuration(item); err != nil {
				return nil, err
			} else {
				diff = diff.(time.Duration) - itmVal
			}
		}
	case time.Time:
		diff = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsDuration(item); err != nil {
				return nil, err
			} else {
				diff = diff.(time.Time).Add(-itmVal)
			}
		}
	case float64:
		diff = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsFloat64(item); err != nil {
				return nil, err
			} else {
				diff = diff.(float64) - itmVal
			}
		}
	case int64:
		diff = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				diff = diff.(int64) - itmVal
			}
		}
	case int:
		diff = int64(dt)
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				diff = diff.(int64) - itmVal
			}
		}
	default: // unsupported comparison
		return nil, fmt.Errorf("unsupported type")
	}
	return
}

// Multiply attempts to multiply multiple items
// returns the result or error if not comparable
func Multiply(items ...interface{}) (mlt interface{}, err error) {
	//we need at least 2 items to diff them
	if len(items) < 2 {
		return nil, ErrNotEnoughParameters
	}
	switch dt := items[0].(type) {
	case float64:
		mlt = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsFloat64(item); err != nil {
				return nil, err
			} else {
				mlt = mlt.(float64) * itmVal
			}
		}
	case int64:
		mlt = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				mlt = mlt.(int64) * itmVal
			}
		}
	case int:
		mlt = int64(dt)
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				mlt = mlt.(int64) * itmVal
			}
		}
	default: // unsupported comparison
		return nil, fmt.Errorf("unsupported type")
	}
	return
}

// Divide attempts to divide multiple items
// returns the result or error if not comparable
func Divide(items ...interface{}) (mlt interface{}, err error) {
	//we need at least 2 items to diff them
	if len(items) < 2 {
		return nil, ErrNotEnoughParameters
	}
	switch dt := items[0].(type) {
	case float64:
		mlt = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsFloat64(item); err != nil {
				return nil, err
			} else {
				mlt = mlt.(float64) / itmVal
			}
		}
	case int64:
		mlt = dt
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				mlt = mlt.(int64) / itmVal
			}
		}
	case int:
		mlt = int64(dt)
		for _, item := range items[1:] {
			if itmVal, err := IfaceAsInt64(item); err != nil {
				return nil, err
			} else {
				mlt = mlt.(int64) / itmVal
			}
		}
	default: // unsupported comparison
		return nil, fmt.Errorf("unsupported type")
	}
	return
}

// ReflectFieldMethodInterface parses intf attepting to return the field value or error otherwise
// Supports "ExtraFields" where additional fields are dynamically inserted in map with field name: extraFieldsLabel
func ReflectFieldMethodInterface(obj interface{}, fldName string) (retIf interface{}, err error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var field reflect.Value
	switch v.Kind() {
	case reflect.Struct:
		field = v.FieldByName(fldName)
	case reflect.Map:
		field = v.MapIndex(reflect.ValueOf(fldName))
	case reflect.Slice, reflect.Array:
		//convert fldName to int
		idx, err := strconv.Atoi(fldName)
		if err != nil {
			return nil, err
		}
		if idx >= v.Len() {
			return nil, fmt.Errorf("index out of range")
		}
		field = v.Index(idx)
	default:
		return nil, fmt.Errorf("unsupported field kind: %v", v.Kind())
	}
	if !field.IsValid() {
		// handle function with pointer
		v = reflect.ValueOf(obj)
		field = v.MethodByName(fldName)
		if !field.IsValid() {
			return nil, ErrNotFound
		} else {
			if field.Type().NumIn() != 0 {
				return nil, fmt.Errorf("invalid function called")
			}
			if field.Type().NumOut() > 2 {
				return nil, fmt.Errorf("invalid function called")
			}
			// the function have two parameters in return and check if the second is of type error
			if field.Type().NumOut() == 2 {
				errorInterface := reflect.TypeOf((*error)(nil)).Elem()
				if !field.Type().Out(1).Implements(errorInterface) {
					return nil, fmt.Errorf("invalid function called")
				}
			}
			fields := field.Call([]reflect.Value{})
			if len(fields) == 2 && !fields[1].IsNil() {
				return fields[0].Interface(), fields[1].Interface().(error)
			}
			return fields[0].Interface(), nil
		}
	}
	return field.Interface(), nil
}
