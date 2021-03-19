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
	"strings"
	"time"
)

// dataStorage is the DataProvider that can be updated
type dataStorage interface {
	DataProvider

	Set(fldPath []string, val interface{}) error
	Remove(fldPath []string) error
	GetKeys(nesteed bool) []string
}

// MapStorage is the basic dataStorage
type MapStorage map[string]interface{}

// String returns the map as json string
func (ms MapStorage) String() string { return ToJSON(ms) }

// FieldAsInterface returns the value from the path
func (ms MapStorage) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) == 0 {
		err = errors.New("empty field path")
		return

	}
	opath, indx := GetPathIndex(fldPath[0])
	var has bool
	if val, has = ms[opath]; !has {
		err = ErrNotFound
		return
	}
	if len(fldPath) == 1 {
		if indx == nil {
			return
		}
		switch rv := val.(type) {
		case []string:
			if len(rv) <= *indx {
				return nil, ErrNotFound
			}
			val = rv[*indx]
			return
		case []interface{}:
			if len(rv) <= *indx {
				return nil, ErrNotFound
			}
			val = rv[*indx]
			return
		default:
		}
		// only if all above fails use reflect:
		vr := reflect.ValueOf(val)
		if vr.Kind() == reflect.Ptr {
			vr = vr.Elem()
		}
		if vr.Kind() != reflect.Slice && vr.Kind() != reflect.Array {
			return nil, ErrNotFound

		}
		if *indx >= vr.Len() {
			return nil, ErrNotFound
		}
		return vr.Index(*indx).Interface(), nil

	}
	if indx == nil {
		switch dp := ms[fldPath[0]].(type) {
		case DataProvider:
			return dp.FieldAsInterface(fldPath[1:])
		case map[string]interface{}:
			return MapStorage(dp).FieldAsInterface(fldPath[1:])
		default:
			err = ErrWrongPath

			return
		}
	}
	switch dp := ms[opath].(type) {
	case []DataProvider:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		return dp[*indx].FieldAsInterface(fldPath[1:])
	case []MapStorage:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		return dp[*indx].FieldAsInterface(fldPath[1:])
	case []map[string]interface{}:
		if len(dp) <= *indx {
			return nil, ErrNotFound

		}
		return MapStorage(dp[*indx]).FieldAsInterface(fldPath[1:])
	case []interface{}:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		switch ds := dp[*indx].(type) {
		case DataProvider:
			return ds.FieldAsInterface(fldPath[1:])
		case map[string]interface{}:
			return MapStorage(ds).FieldAsInterface(fldPath[1:])
		default:
		}
	default:

	}
	err = ErrNotFound // xml compatible
	val = nil
	return
}

// FieldAsString returns the value from path as string
func (ms MapStorage) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	if val, err = ms.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// Set sets the value at the given path
func (ms MapStorage) Set(fldPath []string, val interface{}) (err error) {
	if len(fldPath) == 0 {
		return ErrWrongPath
	}
	if len(fldPath) == 1 {
		ms[fldPath[0]] = val

		return
	}

	if _, has := ms[fldPath[0]]; !has {
		nMap := MapStorage{}
		ms[fldPath[0]] = nMap
		return nMap.Set(fldPath[1:], val)
	}
	switch dp := ms[fldPath[0]].(type) {
	case dataStorage:
		return dp.Set(fldPath[1:], val)
	case map[string]interface{}:
		return MapStorage(dp).Set(fldPath[1:], val)
	default:
		return ErrWrongPath
	}

}

// GetKeys returns all the keys from map
func (ms MapStorage) GetKeys(nesteed bool) (keys []string) {
	if !nesteed {
		keys = make([]string, len(ms))
		i := 0
		for k := range ms {
			keys[i] = k
			i++

		}
		return
	}
	for k, v := range ms {
		// keys = append(keys, k)
		switch rv := v.(type) {
		case dataStorage:
			for _, dsKey := range rv.GetKeys(nesteed) {
				keys = append(keys, k+NestingSep+dsKey)
			}
		case map[string]interface{}:
			for _, dsKey := range MapStorage(rv).GetKeys(nesteed) {
				keys = append(keys, k+NestingSep+dsKey)
			}
		case []MapStorage:
			for i, dp := range rv {
				pref := k + fmt.Sprintf("[%v]", i)
				// keys = append(keys, pref)
				for _, dsKey := range dp.GetKeys(nesteed) {
					keys = append(keys, pref+NestingSep+dsKey)
				}
			}
		case []dataStorage:
			for i, dp := range rv {
				pref := k + fmt.Sprintf("[%v]", i)
				// keys = append(keys, pref)
				for _, dsKey := range dp.GetKeys(nesteed) {
					keys = append(keys, pref+NestingSep+dsKey)
				}
			}
		case []map[string]interface{}:
			for i, dp := range rv {
				pref := k + fmt.Sprintf("[%v]", i)
				// keys = append(keys, pref)
				for _, dsKey := range MapStorage(dp).GetKeys(nesteed) {
					keys = append(keys, pref+NestingSep+dsKey)
				}
			}
		case []interface{}:
			for i := range rv {
				keys = append(keys, k+fmt.Sprintf("[%v]", i))
			}
		case []string:
			for i := range rv {
				keys = append(keys, k+fmt.Sprintf("[%v]", i))
			}
		case nil, int, int32, int64, uint32, uint64, bool, float32, float64, []uint8, time.Duration, time.Time, string: //no path
			keys = append(keys, k)
		default:
			// ToDo:should not be called
			keys = append(keys, getPathFromValue(reflect.ValueOf(v), k+NestingSep)...)
		}

	}
	return

}

// Remove removes the item at path
func (ms MapStorage) Remove(fldPath []string) (err error) {
	if len(fldPath) == 0 {
		return ErrWrongPath
	}
	var val interface{}
	var has bool
	if val, has = ms[fldPath[0]]; !has {
		return // ignore (already removed)
	}
	if len(fldPath) == 1 {
		delete(ms, fldPath[0])

		return
	}
	switch dp := val.(type) {
	case dataStorage:
		return dp.Remove(fldPath[1:])
	case map[string]interface{}:
		return MapStorage(dp).Remove(fldPath[1:])
	default:
		return ErrWrongPath

	}

}

// RemoteHost is part of dataStorage interface
func (ms MapStorage) RemoteHost() net.Addr {
	return LocalAddr()
}

// ToDo: remove the following functions
func getPathFromValue(in reflect.Value, prefix string) (out []string) {
	switch in.Kind() {
	case reflect.Ptr:
		return getPathFromValue(in.Elem(), prefix)
	case reflect.Array, reflect.Slice:
		prefix = strings.TrimSuffix(prefix, NestingSep)
		for i := 0; i < in.Len(); i++ {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			// out = append(out, pref)
			out = append(out, getPathFromValue(in.Index(i), pref+NestingSep)...)
		}
	case reflect.Map:
		iter := reflect.ValueOf(in).MapRange()
		for iter.Next() {
			pref := prefix + iter.Key().String()
			// out = append(out, pref)
			out = append(out, getPathFromValue(iter.Value(), pref+NestingSep)...)
		}
	case reflect.Struct:
		inType := in.Type()
		for i := 0; i < in.NumField(); i++ {
			pref := prefix + inType.Field(i).Name
			// out = append(out, pref)
			out = append(out, getPathFromValue(in.Field(i), pref+NestingSep)...)
		}
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Interface:
		out = append(out, strings.TrimSuffix(prefix, NestingSep))
	default:
	}
	return
}
