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
	"net"
	"strconv"
	"strings"
)

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap with a NavigableMap2
func NewOrderedNavigableMap() *OrderedNavigableMap {
	return &OrderedNavigableMap{
		nm:       NavigableMap2{},
		orderIdx: NewPathItemList(),
		orderRef: make(map[string][]*PathItemElement),
	}
}

// OrderedNavigableMap is the same as NavigableMap2 but keeps the order of fields
type OrderedNavigableMap struct {
	nm       NMInterface
	orderIdx *PathItemList
	orderRef map[string][]*PathItemElement
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string {
	return onm.nm.String()
}

// GetFirstElement returns the first element from the order
func (onm *OrderedNavigableMap) GetFirstElement() *PathItemElement {
	return onm.orderIdx.Front()
}

// Interface returns navigble map that's inside
func (onm *OrderedNavigableMap) Interface() interface{} {
	return onm.nm
}

// Field returns the item on the given path
func (onm *OrderedNavigableMap) Field(fldPath PathItems) (val NMInterface, err error) {
	return onm.nm.Field(fldPath)
}

// Type returns the type of the NM map
func (onm *OrderedNavigableMap) Type() NMType {
	return onm.nm.Type()
}

// Empty returns true if the NM is empty(no data)
func (onm *OrderedNavigableMap) Empty() bool {
	return onm.nm.Empty()
}

// Remove removes the item for the given path and updates the order
func (onm *OrderedNavigableMap) Remove(fullPath *FullPath) (err error) {
	path := stripIdxFromLastPathElm(fullPath.Path)
	if path == EmptyString || fullPath.PathItems[len(fullPath.PathItems)-1].Index != nil {
		return ErrWrongPath
	}
	if err = onm.nm.Remove(fullPath.PathItems); err != nil {
		return
	}

	for idxPath, slcIdx := range onm.orderRef {
		if !strings.HasPrefix(idxPath, path) {
			continue
		}
		for _, el := range slcIdx {
			onm.orderIdx.Remove(el)
		}
		delete(onm.orderRef, idxPath)
	}
	return
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (onm *OrderedNavigableMap) Set(fullPath *FullPath, val NMInterface) (addedNew bool, err error) {
	if fullPath == nil || len(fullPath.PathItems) == 0 {
		return false, ErrWrongPath
	}
	if addedNew, err = onm.nm.Set(fullPath.PathItems, val); err != nil {
		return
	}

	var pathItmsSet []PathItems // can be multiples if we need to inflate due to missing Index in slice set
	var nonIndexedSlcPath bool
	if val.Type() == NMSliceType && fullPath.PathItems[len(fullPath.PathItems)-1].Index == nil { // special case when we overwrite with a slice without specifying indexes
		nonIndexedSlcPath = true
		pathItmsSet = make([]PathItems, len(*val.(*NMSlice)))
		for i := 0; i < val.Len(); i++ {
			pathItms := fullPath.PathItems.Clone()
			pathItms[len(pathItms)-1].Index = StringPointer(strconv.Itoa(i))
			pathItmsSet[i] = pathItms
		}
	} else {
		pathItmsSet = []PathItems{fullPath.PathItems}
	}
	path := stripIdxFromLastPathElm(fullPath.Path)
	if !addedNew && nonIndexedSlcPath { // cleanup old references since the value is being overwritten
		for idxPath, slcIdx := range onm.orderRef {
			if !strings.HasPrefix(idxPath, path) {
				continue
			}
			for _, el := range slcIdx {
				onm.orderIdx.Remove(el)
			}
			delete(onm.orderRef, idxPath)
		}
	}
	_, hasRef := onm.orderRef[path]
	for _, pathItms := range pathItmsSet {
		if addedNew || !hasRef {
			onm.orderRef[path] = append(onm.orderRef[path], onm.orderIdx.PushBack(pathItms))
		} else {
			onm.orderIdx.MoveToBack(onm.orderRef[path][len(onm.orderRef[path])-1])
		}
	}
	return
}

// Len returns the lenght of the map
func (onm OrderedNavigableMap) Len() int {
	return onm.nm.Len()
}

// FieldAsString returns the value from path as string
func (onm *OrderedNavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	var val NMInterface
	val, err = onm.nm.Field(NewPathItems(fldPath))
	if err != nil {
		return
	}
	return IfaceAsString(val.Interface()), nil
}

// FieldAsInterface returns the interface at the path
func (onm *OrderedNavigableMap) FieldAsInterface(fldPath []string) (iface interface{}, err error) {
	var val NMInterface
	val, err = onm.nm.Field(NewPathItems(fldPath))
	if err != nil {
		return
	}
	return val.Interface(), nil
}

// RemoteHost is part of dataStorage interface
func (OrderedNavigableMap) RemoteHost() net.Addr {
	return LocalAddr()
}

// GetOrder returns the elements order as a slice
func (onm *OrderedNavigableMap) GetOrder() (order []PathItems) {
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		order = append(order, el.Value)
	}
	return
}

// OrderedFields returns the elements in order they were inserted
func (onm *OrderedNavigableMap) OrderedFields() (flds []interface{}) {
	flds = make([]interface{}, 0, onm.Len())
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		fld, _ := onm.Field(el.Value)
		flds = append(flds, fld.Interface())
	}
	return
}

// RemoveAll will clean the data and the odrder from OrderedNavigableMap
func (onm *OrderedNavigableMap) RemoveAll() {
	onm.nm = NavigableMap2{}
	onm.orderIdx = NewPathItemList()
	onm.orderRef = make(map[string][]*PathItemElement)
}

// OrderedFieldsAsStrings returns the elements as strings in order they were inserted
func (onm *OrderedNavigableMap) OrderedFieldsAsStrings() (flds []string) {
	flds = make([]string, 0, onm.Len())
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		fld, _ := onm.Field(el.Value)
		flds = append(flds, IfaceAsString(fld.Interface()))
	}
	return
}
