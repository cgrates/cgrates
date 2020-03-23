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

import "net"

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap2 with a NavigableMap2
func NewOrderedNavigableMap() *OrderedNavigableMap {
	return &OrderedNavigableMap{
		nm:    NavigableMap2{},
		order: []PathItems{},
	}
}

// OrderedNavigableMap is the same as NavigableMap2 but keeps the order of fields
type OrderedNavigableMap struct {
	nm    NMInterface
	order []PathItems
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string {
	return onm.nm.String()
}

// GetOrder returns the order the fields were set in NavigableMap2
func (onm *OrderedNavigableMap) GetOrder() []PathItems {
	return onm.order
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
func (onm *OrderedNavigableMap) Remove(path PathItems) (err error) {
	if err = onm.nm.Remove(path); err != nil {
		return
	}
	onm.removePath(path)
	if path[len(path)-1].Index != nil {
		if idx := *path[len(path)-1].Index; idx >= 0 {
			onm.updateOrderBasedOnIndex(path, idx)
		}
	}
	return
}

// Set sets the value at the given path
func (onm *OrderedNavigableMap) Set(fldPath PathItems, val NMInterface) (err error) {
	lpath := len(fldPath)
	if lpath == 0 {
		return ErrWrongPath
	}
	switch val.Type() {
	case NMDataType:
		var dataMap NMInterface = onm.nm
		for i, spath := range fldPath {
			var newData NMInterface
			newData, err = dataMap.GetField(spath)
			if err == ErrNotFound {
				if err = dataMap.Set(fldPath[i:], val); err != nil {
					return
				}
				onm.order = append(onm.order, fldPath)
				return
			}
			if err != nil {
				return
			}
			if i == lpath-1 { // last path
				if err = dataMap.SetField(spath, val); err == nil {
					onm.removePath(fldPath)
					onm.order = append(onm.order, fldPath)
				}
				return
			}
			dataMap = newData
		}
	case NMSliceType:
		var dataMap NMInterface = onm.nm
		for i, spath := range fldPath {
			var newData NMInterface
			newData, err = dataMap.GetField(spath)
			if err == ErrNotFound {
				if err = dataMap.Set(fldPath[i:], val); err != nil {
					return
				}
				l := val.Len()
				for j := 0; j < l; j++ {
					newpath := make(PathItems, lpath)
					copy(newpath, fldPath)
					newpath[len(newpath)-1] = PathItem{
						Field: newpath[len(newpath)-1].Field,
						Index: IntPointer(j),
					}
					onm.order = append(onm.order, newpath)
				}
				return
			}
			if err != nil {
				return
			}
			if i == lpath-1 { // last path
				if err = dataMap.SetField(spath, val); err == nil {
					onm.removePath(fldPath)
					l := val.Len()
					for j := 0; j < l; j++ {
						newpath := make(PathItems, lpath)
						copy(newpath, fldPath)
						newpath[len(newpath)-1] = PathItem{
							Field: newpath[len(newpath)-1].Field,
							Index: IntPointer(j),
						}
						onm.order = append(onm.order, newpath)
					}
				}
				return
			}
			dataMap = newData
		}
	default:
	}
	return ErrNotImplemented
}

// removePath removes any reference to the given path from order
// extremly slow method
func (onm *OrderedNavigableMap) removePath(path PathItems) {
	lenpath := len(path)
	for i := 0; i < len(onm.order); {
		p := onm.order[i]
		if len(p) < lenpath {
			i++
			continue
		}
		match := true
		for j, field := range path {
			if lenpath-1 == j {
				if match = field.Field == p[j].Field; !match {
					break
				}
				if match = field.Index == nil; match {
					break
				}
				match = p[j].Index != nil && *p[j].Index == *field.Index
				break
			}
			if match = field.Equal(p[j]); !match {
				break
			}
		}
		if !match {
			i++
			continue
		}
		copy(onm.order[i:], onm.order[i+1:]) // Shift a[i+1:] left one index.
		onm.order[len(onm.order)-1] = nil    // Erase last element (write zero value).
		onm.order = onm.order[:len(onm.order)-1]
	}
}

// GetField the same as Field but for one level deep
// used to implement NM interface
func (onm *OrderedNavigableMap) GetField(path PathItem) (val NMInterface, err error) {
	return onm.nm.GetField(path)
}

// SetField the same as Set but for one level deep
// used to implement NM interface
func (onm *OrderedNavigableMap) SetField(path PathItem, val NMInterface) (err error) {
	// if path == nil {
	// 	return ErrWrongPath
	// }
	switch val.Type() {
	case NMDataType:
		_, err = onm.nm.GetField(path)
		if err != nil {
			if err == ErrNotFound {
				if err = onm.nm.SetField(path, val); err != nil {
					return
				}
				onm.order = append(onm.order, PathItems{path})
			}
			return
		}
		onm.removePath(PathItems{path})
		if err = onm.nm.SetField(path, val); err == nil {
			onm.order = append(onm.order, PathItems{path})
		}
		return
	case NMSliceType:
		_, err = onm.nm.GetField(path)
		if err != nil {
			if err == ErrNotFound {
				if err = onm.nm.SetField(path, val); err != nil {
					return
				}
				l := val.Len()
				for j := 0; j < l; j++ {
					newpath := make(PathItems, 1)
					newpath[0] = PathItem{
						Field: path.Field,
						Index: IntPointer(j),
					}
					onm.order = append(onm.order, newpath)
				}
			}
			return
		}
		onm.removePath(PathItems{path})
		if err = onm.nm.SetField(path, val); err == nil {
			l := val.Len()
			for j := 0; j < l; j++ {
				newpath := make(PathItems, 1)
				newpath[0] = PathItem{
					Field: path.Field,
					Index: IntPointer(j),
				}
				onm.order = append(onm.order, newpath)
			}
		}
		return
	default:
		return ErrNotImplemented
	}
}

// Len returns the lenght of the map
func (onm OrderedNavigableMap) Len() int {
	return onm.nm.Len()
}

// FieldAsString returns thevalue from path as string
func (onm *OrderedNavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	var val NMInterface
	val, err = onm.nm.Field(NewPathToItem(fldPath))
	if err != nil {
		return
	}
	return IfaceAsString(val.Interface()), nil
}

// FieldAsInterface returns the interface at the path
func (onm *OrderedNavigableMap) FieldAsInterface(fldPath []string) (str interface{}, err error) {
	var val NMInterface
	val, err = onm.nm.Field(NewPathToItem(fldPath))
	if err != nil {
		return
	}
	return val.Interface(), nil
}

// updateOrderBasedOnIndex updates the index of the slice elements that are bigger that the removed element
func (onm *OrderedNavigableMap) updateOrderBasedOnIndex(path PathItems, idx int) {
	lenpath := len(path)
	for _, p := range onm.order {
		if len(p) < lenpath {
			continue
		}
		for j, field := range path {
			if lenpath-1 == j {
				if field.Field != p[j].Field {
					break
				}
				if p[j].Index != nil && *p[j].Index > idx {
					(*p[j].Index)--
				}
				break
			}
			if !field.Equal(p[j]) {
				break
			}
		}
	}
}

// RemoteHost is part of dataStorage interface
func (OrderedNavigableMap) RemoteHost() net.Addr {
	return LocalAddr()
}
