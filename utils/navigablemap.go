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

// NavigableMap is the basic dataStorage
type NavigableMap map[string]interface{}

// String returns the map as json string
func (ms NavigableMap) String() string { return ToJSON(ms) }

// FieldAsInterface returns the value from the path
func (ms NavigableMap) FieldAsInterface(fldPath []string) (val interface{}, err error) {
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
			return NavigableMap(dp).FieldAsInterface(fldPath[1:])
		default:
			err = fmt.Errorf("Wrong path")
			return
		}
	}
	switch dp := ms[opath].(type) {
	case []DataProvider:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		return dp[*indx].FieldAsInterface(fldPath[1:])
	case []NavigableMap:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		return dp[*indx].FieldAsInterface(fldPath[1:])
	case []map[string]interface{}:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		return NavigableMap(dp[*indx]).FieldAsInterface(fldPath[1:])
	case []interface{}:
		if len(dp) <= *indx {
			return nil, ErrNotFound
		}
		switch ds := dp[*indx].(type) {
		case DataProvider:
			return ds.FieldAsInterface(fldPath[1:])
		case map[string]interface{}:
			return NavigableMap(ds).FieldAsInterface(fldPath[1:])
		default:
		}
	default:
	}
	err = ErrNotFound // xml compatible
	val = nil
	return
}

// FieldAsString returns the value from path as string
func (ms NavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	if val, err = ms.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// Set sets the value at the given path
func (ms NavigableMap) Set(fldPath []string, val interface{}) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
	}
	if len(fldPath) == 1 {
		ms[fldPath[0]] = val
		return
	}

	if _, has := ms[fldPath[0]]; !has {
		nMap := NavigableMap{}
		ms[fldPath[0]] = nMap
		return nMap.Set(fldPath[1:], val)
	}
	switch dp := ms[fldPath[0]].(type) {
	case dataStorage:
		return dp.Set(fldPath[1:], val)
	case map[string]interface{}:
		return NavigableMap(dp).Set(fldPath[1:], val)
	default:
		return fmt.Errorf("Wrong path")
	}
}

// GetKeys returns all the keys from map
func (ms NavigableMap) GetKeys(nesteed bool) (keys []string) {
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
		keys = append(keys, k)
		switch rv := v.(type) {
		case dataStorage:
			for _, dsKey := range rv.GetKeys(nesteed) {
				keys = append(keys, k+NestingSep+dsKey)
			}
		case map[string]interface{}:
			for _, dsKey := range NavigableMap(rv).GetKeys(nesteed) {
				keys = append(keys, k+NestingSep+dsKey)
			}
		case []NavigableMap:
			for i, dp := range rv {
				pref := k + fmt.Sprintf("[%v]", i)
				keys = append(keys, pref)
				for _, dsKey := range dp.GetKeys(nesteed) {
					keys = append(keys, pref+NestingSep+dsKey)
				}
			}
		case []dataStorage:
			for i, dp := range rv {
				pref := k + fmt.Sprintf("[%v]", i)
				keys = append(keys, pref)
				for _, dsKey := range dp.GetKeys(nesteed) {
					keys = append(keys, pref+NestingSep+dsKey)
				}
			}
		case []map[string]interface{}:
			for i, dp := range rv {
				pref := k + fmt.Sprintf("[%v]", i)
				keys = append(keys, pref)
				for _, dsKey := range NavigableMap(dp).GetKeys(nesteed) {
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
		default:
			keys = append(keys, getPathFromInterface(v, k+NestingSep)...)
		}
	}
	return
}

// Remove removes the item at path
func (ms NavigableMap) Remove(fldPath []string) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
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
		return NavigableMap(dp).Remove(fldPath[1:])
	default:
		return fmt.Errorf("Wrong path")
	}
}

// RemoteHost is part of dataStorage interface
func (ms NavigableMap) RemoteHost() net.Addr {
	return LocalAddr()
}

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap with a NavigableMap
func NewOrderedNavigableMap(nm dataStorage) *OrderedNavigableMap {
	if nm == nil {
		return &OrderedNavigableMap{
			nm:    NavigableMap{},
			order: [][]string{},
		}
	}
	// Index problem
	keys := nm.GetKeys(true)
	order := make([][]string, len(keys))
	for i, k := range keys {
		order[i] = strings.Split(k, NestingSep)
	}
	return &OrderedNavigableMap{
		nm:    nm,
		order: order,
	}
}

// OrderedNavigableMap is the same as NavigableMap but keeps the order of fields
type OrderedNavigableMap struct {
	nm    dataStorage
	order [][]string
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string { return ToJSON(onm.nm) }

// FieldAsInterface returns the value from the path
func (onm *OrderedNavigableMap) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	return onm.nm.FieldAsInterface(fldPath)
}

// FieldAsString returns thevalue from path as string
func (onm *OrderedNavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	return onm.nm.FieldAsString(fldPath)
}

// Set sets the value at the given path
func (onm *OrderedNavigableMap) Set(fldPath []string, val interface{}) (err error) {
	if err = onm.nm.Set(fldPath, val); err != nil {
		return
	}
	onm.order = append(onm.order, fldPath)
	// if dp, canCast := val.(dataStorage); canCast {
	// 	for _, key := range dp.GetKeys(true) {
	// 		onm.order = append(onm.order, append(fldPath, strings.Split(key, NestingSep)...))
	// 	}
	// }
	return
}

// GetKeys returns all the keys from map
func (onm *OrderedNavigableMap) GetKeys(nesteed bool) (keys []string) {
	keys = make([]string, len(onm.order))
	for i, k := range onm.order {
		keys[i] = strings.Join(k, NestingSep)
	}
	return
}

// Remove removes the item at path
// this function is not needed for now
func (onm *OrderedNavigableMap) Remove(fldPath []string) (err error) {
	return ErrNotImplemented
	/*
		if len(fldPath) == 0 {
			return fmt.Errorf("Wrong path")
		}
		if err = onm.nm.Remove(fldPath); err != nil {
			return
		}
		fld := strings.Join(fldPath, NestingSep)
		for i, order := range onm.order {
			o := strings.Join(order, NestingSep)
			if len(o) == 0 || strings.HasPrefix(o, fld) {
				onm.order = append(onm.order[:i], onm.order[i+1:]...)
			}
		}
		return
	*/
}

// RemoteHost is part of dataStorage interface
func (onm OrderedNavigableMap) RemoteHost() net.Addr {
	return LocalAddr()
}

// Values returns the values in map, ordered by order information
func (onm *OrderedNavigableMap) Values() (vals []interface{}) {
	if len(onm.order) == 0 {
		return
	}
	vals = make([]interface{}, len(onm.order))
	for i, path := range onm.order {
		val, _ := onm.FieldAsInterface(path)
		vals[i] = val
	}
	return
}

// Walk returns the values in map, ordered by order information
func (onm *OrderedNavigableMap) Walk(proccess func(interface{}) error) (err error) {
	for _, path := range onm.order {
		val, _ := onm.FieldAsInterface(path)
		if err = proccess(val); err != nil {
			return
		}
	}
	return
}

// func (onm *OrderedNavigableMap) indexOrder() {
// 	keys := nm.GetKeys(true)
// 	order := make([][]string, len(keys))
// 	for i, k := range keys {
// 		order[i] = strings.Split(k, NestingSep)
// 	}
// }

// // indexMapElements will recursively go through map and index the element paths into elmns
// func indexMapElements(mp map[string]interface{}, path []string, vals *[]interface{}) {
// 	for k, v := range mp {
// 		vPath := append(path, k)
// 		if mpIface, isMap := v.(map[string]interface{}); isMap {
// 			indexMapElements(mpIface, vPath, vals)
// 			continue
// 		}
// 		valsOut := append(*vals, v)
// 		*vals = valsOut
// 	}
// }

// // indexMapPaths parses map returning the parsed branchPath, useful when not having order for NavigableMap
// func indexMapPaths(mp map[string]interface{}, branchPath []string, parsedPaths *[][]string) {
// 	for k, v := range mp {
// 		if mpIface, isMap := v.(map[string]interface{}); isMap {
// 			indexMapPaths(mpIface, append(branchPath, k), parsedPaths)
// 			continue
// 		}
// 		tmpPaths := append(*parsedPaths, append(branchPath, k))
// 		*parsedPaths = tmpPaths
// 	}
// }

// GetOrder returns the order the fields were set in map
func (onm *OrderedNavigableMap) GetOrder() [][]string { return onm.order }

// AsCGREvent builds a CGREvent considering Time as time.Now()
// and Event as linear map[string]interface{} with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
// func (onm *OrderedNavigableMap) AsCGREvent(tnt string, pathSep string) (cgrEv *CGREvent) {
// 	if onm == nil || len(onm.order) == 0 {
// 		return
// 	}
// 	cgrEv = &CGREvent{
// 		Tenant: tnt,
// 		ID:     UUIDSha1Prefix(),
// 		Time:   TimePointer(time.Now()),
// 		Event:  make(map[string]interface{})}
// 	for _, branchPath := range onm.order {
// 		val, _ := onm.FieldAsInterface(branchPath)
// 		if nmItms, isNMItems := val.([]NMItem); isNMItems { // special case when we have added multiple items inside a key, used in agents
// 			for _, nmItm := range nmItms {
// 				if !nmItm.IsAttribute() {
// 					val = nmItm.GetData() // first item which is not an attribute will become the value
// 					break
// 				}
// 			}
// 		} else {
// 		}
// 		cgrEv.Event[strings.Join(branchPath, pathSep)] = val
// 	}
// 	return
// }

func getPathFromValue(in reflect.Value, prefix string) (out []string) {
	switch in.Kind() {
	case reflect.Ptr:
		return getPathFromValue(in.Elem(), prefix)
	case reflect.Array, reflect.Slice:
		prefix = strings.TrimSuffix(prefix, NestingSep)
		for i := 0; i < in.Len(); i++ {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
			out = append(out, getPathFromValue(in.Index(i), pref+NestingSep)...)
		}
	case reflect.Map:
		iter := reflect.ValueOf(in).MapRange()
		for iter.Next() {
			pref := prefix + iter.Key().String()
			out = append(out, pref)
			out = append(out, getPathFromValue(iter.Value(), pref+NestingSep)...)
		}
	case reflect.Struct:
		inType := in.Type()
		for i := 0; i < in.NumField(); i++ {
			pref := prefix + inType.Field(i).Name
			out = append(out, pref)
			out = append(out, getPathFromValue(in.Field(i), pref+NestingSep)...)
		}
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Interface:
	default:
	}
	return
}

func getPathFromInterface(in interface{}, prefix string) (out []string) {
	switch vin := in.(type) {
	case map[string]interface{}:
		for k, val := range vin {
			pref := prefix + k
			out = append(out, pref)
			out = append(out, getPathFromInterface(val, pref+NestingSep)...)
		}
	case []map[string]interface{}:
		prefix = strings.TrimSuffix(prefix, NestingSep)
		for i, val := range vin {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
			out = append(out, getPathFromInterface(val, pref+NestingSep)...)
		}
	case []interface{}:
		prefix = strings.TrimSuffix(prefix, NestingSep)
		for i, val := range vin {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
			out = append(out, getPathFromInterface(val, pref+NestingSep)...)
		}
	case []string:
		prefix = strings.TrimSuffix(prefix, NestingSep)
		for i := range vin {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
		}
	case nil, int, int32, int64, uint32, uint64, bool, float32, float64, []uint8, time.Duration, time.Time, string: //no path
	default: //reflect based
		out = getPathFromValue(reflect.ValueOf(vin), prefix)
	}
	return
}
