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
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func NewNMInterface(val interface{}) *NMInterface { return &NMInterface{data: val} }

type NMInterface struct{ data interface{} }

func (nmi *NMInterface) String() string         { return IfaceAsString(nmi.data) }
func (nmi *NMInterface) Interface() interface{} { return nmi.data }
func (nmi *NMInterface) Field(path []string) (val NM, err error) {
	return nil, ErrNotImplemented
}
func (nmi *NMInterface) Set(path []string, val NM, overwrite bool) (err error) {
	return ErrNotImplemented
}
func (nmi *NMInterface) Remove(path []string) (err error) {
	return ErrNotImplemented
}
func (nmi *NMInterface) Type() NMType { return NMInterfaceType }
func (nmi *NMInterface) Empty() bool  { return nmi != nil && nmi.data != nil }

type NavigableMap2 map[string]NM

func (nmm NavigableMap2) String() (out string) {
	for k, v := range nmm {
		out = "," + k + ":" + v.String()
	}
	if len(out) == 0 {
		return "{}"
	}
	out = out[1:]
	return "{" + out + "}"
}
func (nmm NavigableMap2) Interface() interface{} { return nmm }
func (nmm NavigableMap2) Field(path []string) (val NM, err error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("Wrong path")
	}
	opath, indx := getPathIndex(path[0])
	el, has := nmm[opath]
	if !has {
		return nil, ErrNotFound
	}
	switch el.Type() {
	case NMInterfaceType:
		if len(path) != 1 || indx != "" {
			return nil, ErrNotFound
		}
		return el, nil
	case NMMapType:
		if indx != "" {
			return nil, ErrNotFound
		}
		if len(path) == 1 {
			return el, nil
		}
		return el.Field(path[1:])
	case NMSliceType:
		if len(path) == 1 && indx == "" {
			return el, nil
		}
		if indx == "" {
			return nil, ErrNotFound
		}
		path[0] = indx
		return el.Field(path)
	}
	panic("BUG")
}
func (nmm NavigableMap2) Set(path []string, val NM, overwrite bool) (err error) {
	if len(path) == 0 {
		return fmt.Errorf("Wrong path")
	}
	opath, indx := getPathIndex(path[0])
	el, has := nmm[opath]
	if len(path) == 1 {
		if !has {
			if indx != "" {
				nel := &NMSlice{}
				if err = nel.Set([]string{indx}, val, overwrite); err != nil {
					return
				}
				nmm[opath] = nel
				return
			}
			nmm[opath] = val
			return
		}
		if indx != "" {
			if el.Type() != NMSliceType {
				return fmt.Errorf("Wrong path")
			}
			return el.Set([]string{indx}, val, overwrite)
		}
		if !overwrite { // do not try to overwrite
			return
		}
		nmm[opath] = val
		return
	}
	if !has {
		if indx != "" {
			nel := &NMSlice{}
			path[0] = indx
			if err = nel.Set(path, val, overwrite); err != nil {
				return
			}
			nmm[opath] = nel
			return
		}
		nel := NavigableMap2{}
		if err = nel.Set(path[1:], val, overwrite); err != nil {
			return
		}
		nmm[opath] = nel
		return
	}
	if indx != "" {
		if el.Type() != NMSliceType {
			return fmt.Errorf("Wrong path")
		}
		path[0] = indx
		return el.Set(path, val, overwrite)
	}
	if !overwrite && el.Type() != NMMapType { // do not try to overwrite an interface
		return
	}
	return el.Set(path[1:], val, overwrite)
}
func (nmm NavigableMap2) Remove(path []string) (err error) {
	if len(path) == 0 {
		return fmt.Errorf("Wrong path")
	}
	opath, indx := getPathIndex(path[0])
	el, has := nmm[opath]
	if !has {
		return // already removed
	}
	if len(path) == 1 {
		if indx != "" {
			if el.Type() != NMSliceType {
				return fmt.Errorf("Wrong path")
			}
			return el.Remove([]string{indx})
		}
		delete(nmm, opath)
		return
	}
	if indx != "" {
		if el.Type() != NMSliceType {
			return fmt.Errorf("Wrong path")
		}
		path[0] = indx
		if err = el.Remove(path); err != nil {
			return
		}
		if el.Empty() {
			delete(nmm, opath)
		}
		return
	}
	if el.Type() != NMMapType {
		return fmt.Errorf("Wrong path")
	}
	if err = el.Remove(path[1:]); err != nil {
		return
	}
	if el.Empty() {
		delete(nmm, opath)
	}
	return
}
func (nmm NavigableMap2) Type() NMType { return NMMapType }
func (nmm NavigableMap2) Empty() bool  { return nmm != nil && len(nmm) != 0 }

type NMSlice []NM

func (nms *NMSlice) String() (out string) {
	for _, v := range *nms {
		out = "," + v.String()
	}
	if len(out) == 0 {
		return "[]"
	}
	out = out[1:]
	return "[" + out + "]"
}
func (nms *NMSlice) Interface() interface{} { return *nms }
func (nms *NMSlice) Field(path []string) (val NM, err error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("Wrong path")
	}
	if nms.Empty() {
		return nil, ErrNotFound
	}
	var idx int
	if idx, err = strconv.Atoi(path[0]); err != nil {
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) {
		return nil, ErrNotFound
	}
	if len(path) == 1 {
		return (*nms)[idx], nil
	}
	return (*nms)[idx].Field(path[1:])
}
func (nms *NMSlice) Set(path []string, val NM, overwrite bool) (err error) {
	if len(path) == 0 {
		return fmt.Errorf("Wrong path")
	}
	if path[0] == MetaAppend {
		if len(path) == 1 {
			*nms = append(*nms, val)
		}
		nel := NavigableMap2{}
		if err = nel.Set(path[1:], val, overwrite); err != nil {
			return
		}
		*nms = append(*nms, nel)
		return
	}
	var idx int
	if idx, err = strconv.Atoi(path[0]); err != nil {
		return
	}
	if idx == 0 && nms.Empty() { // add first element
		if len(path) == 1 {
			*nms = append(*nms, val)
		}
		nel := NavigableMap2{}
		if err = nel.Set(path[1:], val, overwrite); err != nil {
			return
		}
		*nms = append(*nms, nel)
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) {
		return fmt.Errorf("Wrong path")
	}
	if (*nms)[idx].Type() == NMSliceType {
		return fmt.Errorf("Wrong path")
	}
	if !overwrite && (*nms)[idx].Type() == NMInterfaceType {
		return
	}
	return (*nms)[idx].Set(path[1:], val, overwrite)
}
func (nms *NMSlice) Remove(path []string) (err error) {
	if len(path) == 0 {
		return fmt.Errorf("Wrong path")
	}
	var idx int
	if idx, err = strconv.Atoi(path[0]); err != nil {
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) { // already removed
		return
	}
	switch (*nms)[idx].Type() {
	case NMSliceType:
		return fmt.Errorf("Wrong path")
	case NMInterfaceType:
		if len(path) != 1 {
			return fmt.Errorf("Wrong path")
		}
		*nms = append((*nms)[:idx], (*nms)[idx+1:]...)
		return
	case NMMapType:
		if len(path) == 1 {
			*nms = append((*nms)[:idx], (*nms)[idx+1:]...)
			return
		}
		if err = (*nms)[idx].Remove(path[1:]); err != nil {
			return
		}
		if (*nms)[idx].Empty() {
			*nms = append((*nms)[:idx], (*nms)[idx+1:]...)
		}
		return
	}
	panic("BUG")
}
func (nms NMSlice) Type() NMType { return NMSliceType }
func (nms NMSlice) Empty() bool  { return nms != nil && len(nms) != 0 }

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap2 with a NavigableMap2
func NewOrderedNavigableMap() *OrderedNavigableMap {
	return &OrderedNavigableMap{
		nm:       NavigableMap2{},
		order:    [][]string{},
		orderSet: StringSet{},
	}
}

// OrderedNavigableMap is the same as NavigableMap2 but keeps the order of fields
type OrderedNavigableMap struct {
	nm       NM
	order    [][]string
	orderSet StringSet // to prevent duplicate values
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string { return ToJSON(onm.nm) }

// FieldAsInterface returns the value from the path
func (onm *OrderedNavigableMap) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	return onm.nm.Field(fldPath)
}

// FieldAsString returns thevalue from path as string
func (onm *OrderedNavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	val, err = onm.nm.Field(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// Set sets the value at the given path
func (onm *OrderedNavigableMap) Set(fldPath []string, val NM) (err error) {
	if err = onm.nm.Set(fldPath, val, true); err != nil {
		return
	}
	// path := strings.Join(fldPath, NestingSep)
	// if !onm.orderSet.Has(path) {
	onm.order = append(onm.order, fldPath)
	// onm.orderSet.Add(path)
	// }
	return
}

// GetKeys returns all the keys from map
func (onm *OrderedNavigableMap) GetKeys(nesteed bool) (keys []string) {
	return onm.orderSet.AsSlice()
}

// Remove removes the item at path
// this function is not needed for now
// ToDo: remove this function
func (onm *OrderedNavigableMap) Remove(fldPath []string) (err error) {
	return ErrNotImplemented // this should handle the order corectly
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
// ToDo: use GetOrder instead of this as it is not memmory efficient
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

// GetOrder returns the order the fields were set in NavigableMap2
func (onm *OrderedNavigableMap) GetOrder() [][]string { return onm.order }

// ToDo: remove the following functions
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

// used by NavigableMap2 GetKeys to return all values
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
