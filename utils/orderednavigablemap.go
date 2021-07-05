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
	"strconv"
	"strings"
)

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap with a DataNode
func NewOrderedNavigableMap() *OrderedNavigableMap {
	return &OrderedNavigableMap{
		nm:       &DataNode{Type: NMMapType, Map: make(map[string]*DataNode)},
		orderIdx: NewPathItemList(),
		orderRef: make(map[string][]*PathItemElement),
	}
}

// OrderedNavigableMap is the same as DataNode but keeps the order of fields
type OrderedNavigableMap struct {
	nm       *DataNode
	orderIdx *PathItemList
	orderRef map[string][]*PathItemElement
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string {
	return ToJSON(onm.nm)
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
func (onm *OrderedNavigableMap) Field(fldPath []string) (val *DataLeaf, err error) {
	return onm.nm.Field(fldPath)
}

// Empty returns true if the NM is empty(no data)
func (onm *OrderedNavigableMap) Empty() bool {
	return onm.nm.IsEmpty()
}

func (onm *OrderedNavigableMap) removeRef(path string) {
	for idxPath, slcIdx := range onm.orderRef { //remove the references from order
		if strings.HasPrefix(idxPath, path) {
			for _, el := range slcIdx {
				onm.orderIdx.Remove(el)
			}
			delete(onm.orderRef, idxPath)
		}
	}
}

// Remove removes the item for the given path and updates the order
func (onm *OrderedNavigableMap) Remove(fullPath *FullPath) (err error) {
	if fullPath.Path == EmptyString {
		return ErrWrongPath
	}
	// fullPath.PathSlice = CloneStringSlice(fullPath.PathSlice) // clone the items to not modify the templates
	if err = onm.nm.Remove(fullPath.PathSlice); err != nil { // remove them from DataNode
		return
	}
	onm.removeRef(fullPath.Path)
	return
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
// used in tests
func (onm *OrderedNavigableMap) Set(fullPath *FullPath, val interface{}) (err error) {
	if fullPath == nil || len(fullPath.PathSlice) == 0 {
		return ErrWrongPath
	}
	var addedNew bool // determine if the node was created or only overwriten
	if addedNew, err = onm.nm.Set(fullPath.PathSlice, val); err != nil {
		return
	}

	// update the  order reference
	path := stripIdxFromLastPathElm(fullPath.Path)
	_, hasRef := onm.orderRef[path]
	if addedNew || !hasRef { // the element is new so append to order
		onm.orderRef[path] = append(onm.orderRef[path], onm.orderIdx.PushBack(fullPath.PathSlice))
	} else { // move element in the back of order list
		onm.orderIdx.MoveToBack(onm.orderRef[path][len(onm.orderRef[path])-1])
	}
	return
}

// SetAsSlice sets the slice of nodes at the given path
// this used with full path and the processed path to not calculate them for every set
// this is used by agent request to overwrite the slice at path
func (onm *OrderedNavigableMap) SetAsSlice(fullPath *FullPath, vals []*DataNode) (err error) {
	if fullPath == nil || len(fullPath.PathSlice) == 0 {
		return ErrWrongPath
	}
	var addedNew bool
	if addedNew, err = onm.nm.Set(fullPath.PathSlice, vals); err != nil {
		return
	}

	pathItmsSet := make([][]string, len(vals)) // prepare the path for order update
	for i := range vals {
		pathItmsSet[i] = append(CloneStringSlice(fullPath.PathSlice), strconv.Itoa(i)) // clone the slice as we will append an index
	}
	path := stripIdxFromLastPathElm(fullPath.Path)
	if !addedNew { // cleanup old references since the value is being overwritten
		onm.removeRef(path)
	}
	for _, pathItms := range pathItmsSet { // add the path in order list
		onm.orderRef[path] = append(onm.orderRef[path], onm.orderIdx.PushBack(pathItms))
	}
	return
}

// FieldAsString returns the value from path as string
func (onm *OrderedNavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	if val, err = onm.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface returns the interface at the path
func (onm *OrderedNavigableMap) FieldAsInterface(fldPath []string) (iface interface{}, err error) {
	return onm.nm.FieldAsInterface(fldPath)
}

// GetOrder returns the elements order as a slice
func (onm *OrderedNavigableMap) GetOrder() (order [][]string) {
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		order = append(order, el.Value)
	}
	return
}

// OrderedFields returns the elements in order they were inserted
func (onm *OrderedNavigableMap) OrderedFields() (flds []interface{}) {
	flds = make([]interface{}, 0, len(onm.nm.Map))
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		fld, _ := onm.Field(el.Value)
		flds = append(flds, fld.Data)
	}
	return
}

// RemoveAll will clean the data and the odrder from OrderedNavigableMap
func (onm *OrderedNavigableMap) RemoveAll() {
	onm.nm = &DataNode{Type: NMMapType, Map: make(map[string]*DataNode)}
	onm.orderIdx = NewPathItemList()
	onm.orderRef = make(map[string][]*PathItemElement)
}

// OrderedFieldsAsStrings returns the elements as strings in order they were inserted
func (onm *OrderedNavigableMap) OrderedFieldsAsStrings() (flds []string) {
	flds = make([]string, 0, len(onm.nm.Map))
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		fld, _ := onm.Field(el.Value)
		flds = append(flds, fld.String())
	}
	return
}

// Append appends the leaf at the given path the end must be a slice
// this used with full path and the processed path to not calculate them for every set
func (onm *OrderedNavigableMap) Append(fullPath *FullPath, val *DataLeaf) (err error) {
	if fullPath == nil || len(fullPath.PathSlice) == 0 {
		return ErrWrongPath
	}
	var idx int
	if idx, err = onm.nm.Append(fullPath.PathSlice, val); err != nil {
		return
	}
	// add the path to order
	onm.orderRef[fullPath.Path] = append(onm.orderRef[fullPath.Path],
		onm.orderIdx.PushBack(
			append(CloneStringSlice(fullPath.PathSlice), // clone the slice as we will append an index
				strconv.Itoa(idx))))
	return
}

// Compose comose the value of the leaf at the given path
// this used with full path and the processed path to not calculate them for every set
func (onm *OrderedNavigableMap) Compose(fullPath *FullPath, val *DataLeaf) (err error) {
	if fullPath == nil || len(fullPath.PathSlice) == 0 {
		return ErrWrongPath
	}
	if err = onm.nm.Compose(fullPath.PathSlice, val); err != nil {
		return
	}

	if _, hasRef := onm.orderRef[fullPath.Path]; !hasRef { // the element is new so append to order
		onm.orderRef[fullPath.Path] = append(onm.orderRef[fullPath.Path],
			onm.orderIdx.PushBack(
				append(CloneStringSlice(fullPath.PathSlice), // clone the slice as we will append an index
					"0")))
	} else { // move element in the back of order list
		onm.orderIdx.MoveToBack(onm.orderRef[fullPath.Path][len(onm.orderRef[fullPath.Path])-1])
	}
	return
}
