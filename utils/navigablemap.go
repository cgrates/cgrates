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
	"strings"
)

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap2 with a NavigableMap2
func NewOrderedNavigableMap() *OrderedNavigableMap {
	return &OrderedNavigableMap{
		nm:    NavigableMap2{},
		order: [][]string{},
	}
}

// OrderedNavigableMap is the same as NavigableMap2 but keeps the order of fields
type OrderedNavigableMap struct {
	nm    NM
	order [][]string
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string { return onm.nm.String() }

// GetOrder returns the order the fields were set in NavigableMap2
func (onm *OrderedNavigableMap) GetOrder() [][]string   { return onm.order }
func (onm *OrderedNavigableMap) Interface() interface{} { return onm.nm }
func (onm *OrderedNavigableMap) Field(fldPath []string) (val NM, err error) {
	path := make([]string, len(fldPath))
	copy(path, fldPath)
	return onm.nm.Field(path)
}
func (onm *OrderedNavigableMap) Type() NMType { return onm.nm.Type() }
func (onm *OrderedNavigableMap) Empty() bool  { return onm.nm.Empty() }

func (onm *OrderedNavigableMap) Remove(path []string) (err error) {
	if err = onm.nm.Remove(path); err != nil {
		return
	}
	onm.removePath(path)
	return
}

// Set sets the value at the given path
func (onm *OrderedNavigableMap) Set(fldPath []string, val NM) (err error) {
	path := make([]string, len(fldPath))
	copy(path, fldPath)

	switch val.Type() {
	case NMInterfaceType:
		var dataMap NM = onm.nm
		for i, spath := range path {
			var newData NM
			newData, err = dataMap.GetField(spath)
			if err == ErrNotFound {
				if err = dataMap.Set(path[i:], val); err != nil {
					return
				}
				onm.order = append(onm.order, fldPath)
				return
			}
			if err != nil {
				return
			}
			if i == len(path)-1 { // last path
				onm.removePath(fldPath)
				if err = dataMap.SetField(spath, val); err != nil {
					return
				}
				onm.order = append(onm.order, fldPath)
				return
			}
			dataMap = newData
		}
	case NMSliceType:
		var dataMap NM = onm.nm
		for i, spath := range path {
			var newData NM
			newData, err = dataMap.GetField(spath)
			if err == ErrNotFound {
				if err = dataMap.Set(path[i:], val); err != nil {
					return
				}
				l := val.Len()
				lastEl := fldPath[len(fldPath)-1]
				for j := 0; j < l; j++ {
					newpath := make([]string, len(fldPath))
					copy(newpath, fldPath)
					newpath[len(newpath)-1] = fmt.Sprintf("%s[%v]", lastEl, j)
					onm.order = append(onm.order, newpath)
				}
				return
			}
			if err != nil {
				return
			}
			if i == len(path)-1 { // last path
				onm.removePath(fldPath)
				if err = dataMap.SetField(spath, val); err != nil {
					return
				}
				l := val.Len()
				lastEl := fldPath[len(fldPath)-1]
				for j := 0; j < l; j++ {
					newpath := make([]string, len(fldPath))
					copy(newpath, fldPath)
					newpath[len(newpath)-1] = fmt.Sprintf("%s[%v]", lastEl, j)
					onm.order = append(onm.order, newpath)
				}
				return
			}
			dataMap = newData
		}
	case NMMapType:
		return ErrNotImplemented
	}
	panic("BUG")
}

// removePath removes any reference to the given path from order
// extremly slow method
func (onm *OrderedNavigableMap) removePath(path []string) {
	lenpath := len(path)
	// cnmorder := make([][]string, 0, len(onm.order))
	for i := 0; i < len(onm.order); {
		p := onm.order[i]
		if len(p) < lenpath {
			i++
			continue
		}
		match := true
		for j, field := range path {
			if lenpath-1 == j {
				match = strings.HasPrefix(p[j], field)
				break
			}
			if match = field == p[j]; !match {
				break
			}
		}
		if !match {
			i++
			continue
			// cnmorder = append(cnmorder, p)
		}
		copy(onm.order[i:], onm.order[i+1:]) // Shift a[i+1:] left one index.
		onm.order[len(onm.order)-1] = nil    // Erase last element (write zero value).
		onm.order = onm.order[:len(onm.order)-1]
	}
	// onm.order = cnmorder
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

func (onm *OrderedNavigableMap) FieldAsInterface(fldPath []string) (str interface{}, err error) {
	return onm.Field(fldPath)
}
