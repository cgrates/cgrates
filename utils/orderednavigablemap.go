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
func (onm *OrderedNavigableMap) Remove(path FullPath) (err error) {
	if err = onm.nm.Remove(path.PathItems); err != nil {
		return
	}
	onm.removePath(path.Path, path.PathItems[len(path.PathItems)-1].Index)
	if path.PathItems[len(path.PathItems)-1].Index != nil {
		return ErrNotImplemented // for the momment we can't remove only a specific element
		// if idx := *path[len(path)-1].Index; idx >= 0 {
		// onm.updateOrderBasedOnIndex(path, idx)
		// }
	}
	return
}

// Set sets the value at the given path
// this is the old to be capable of  building the code without updating all the code
// will be replaced with Set2 after we decide that is the optimal solution
func (onm *OrderedNavigableMap) Set(fldPath PathItems, val NMInterface) (addedNew bool, err error) {
	return onm.Set2(&FullPath{PathItems: fldPath, Path: fldPath.String()}, val)
}

// Set2 sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (onm *OrderedNavigableMap) Set2(fullPath *FullPath, val NMInterface) (addedNew bool, err error) {
	fldPath := fullPath.PathItems
	path := fullPath.Path
	lpath := len(fldPath)
	if lpath == 0 {
		return false, ErrWrongPath
	}
	switch val.Type() {
	case NMDataType:
		if addedNew, err = onm.nm.Set(fullPath.PathItems, val); err != nil {
			return
		}
		// because we only add at back optimize to not modify the slice
		if !addedNew {
			onm.orderIdx.MoveToBack(onm.orderRef[fullPath.Path][len(onm.orderRef[fullPath.Path])-1])
		} else {
			onm.orderRef[path] = append(onm.orderRef[path], onm.orderIdx.PushBack(fldPath))
		}
		// onm.removePath(fullPath.Path, fullPath.PathItems[len(fullPath.PathItems)-1].Index)
		// onm.appendPath(fullPath.Path, fullPath.PathItems)
		return
		// this is the old code. Keep this here to not rewrite it if this is the wanted behavior
		/*
			var dataMap NMInterface = onm.nm
			for i, spath := range fldPath {
				var newData NMInterface
				newData, err = dataMap.GetField(spath)
				if err == ErrNotFound {
					if err = dataMap.Set(fldPath[i:], val); err != nil {
						return
					}
					onm.appendPath(path, fldPath)
					return
				}
				if err != nil {
					return
				}
				if i == lpath-1 { // last path
					if err = dataMap.SetField(spath, val); err == nil {
						for _, el := range onm.orderRef[path] {
							if *el.Value[len(el.Value)-1].Index == *spath.Index {
								onm.orderIdx.Remove(el)
							}
						}
						onm.appendPath(path, fldPath)
					}
					return
				}
				dataMap = newData
			}
		*/
	case NMSliceType:
		if addedNew, err = onm.nm.Set(fullPath.PathItems, val); err != nil {
			return
		}
		l := val.Len()
		for _, el := range onm.orderRef[path] {
			onm.orderIdx.Remove(el)
		}
		onm.orderRef[path] = make([]*PathItemElement, l)
		if l == 1 { // do not coppy the path if we have only one element in slice
			fldPath[len(fldPath)-1] = PathItem{
				Field: fldPath[len(fldPath)-1].Field,
				Index: IntPointer(0),
			}
			onm.orderRef[path][0] = onm.orderIdx.PushBack(fldPath)
			return
		}
		for j := 0; j < l; j++ {
			newpath := make(PathItems, lpath)
			copy(newpath, fldPath)
			newpath[len(newpath)-1] = PathItem{
				Field: newpath[len(newpath)-1].Field,
				Index: IntPointer(j),
			}
			onm.orderRef[path][j] = onm.orderIdx.PushBack(newpath)
		}
		return
		// this is the old code. Keep this here to not rewrite it if this is the wanted behavior
		/*
			var dataMap NMInterface = onm.nm
			for i, spath := range fldPath {
				var newData NMInterface
				newData, err = dataMap.GetField(spath)
				if err == ErrNotFound {
					if err = dataMap.Set(fldPath[i:], val); err != nil {
						return
					}
					l := val.Len()
					onm.orderRef[path] = make([]*PathItemElement, l)
					for j := 0; j < l; j++ {
						newpath := make(PathItems, lpath)
						copy(newpath, fldPath)
						newpath[len(newpath)-1] = PathItem{
							Field: newpath[len(newpath)-1].Field,
							Index: IntPointer(j),
						}
						onm.orderRef[path][j] = onm.orderIdx.PushBack(fldPath)
					}
					return
				}
				if err != nil {
					return
				}
				if i == lpath-1 { // last path
					if err = dataMap.SetField(spath, val); err == nil {
						onm.removePath(path)
						l := val.Len()
						onm.orderRef[path] = make([]*PathItemElement, l)
						for j := 0; j < l; j++ {
							newpath := make(PathItems, lpath)
							copy(newpath, fldPath)
							newpath[len(newpath)-1] = PathItem{
								Field: newpath[len(newpath)-1].Field,
								Index: IntPointer(j),
							}
							onm.orderRef[path][j] = onm.orderIdx.PushBack(fldPath)
						}
					}
					return
				}
				dataMap = newData
			}
		*/
	default:
	}
	return false, ErrNotImplemented
}

func (onm *OrderedNavigableMap) appendPath(path string, fldPath PathItems) {
	onm.orderRef[path] = append(onm.orderRef[path], onm.orderIdx.PushBack(fldPath))
}

// removePath removes any reference to the given path from order
// extremly slow method
func (onm *OrderedNavigableMap) removePath(path string, indx *int) {
	if indx == nil {
		for _, el := range onm.orderRef[path] {
			onm.orderIdx.Remove(el)
		}
		// faster to only overwrite the value than deleting it
		onm.orderRef[path] = nil
		// delete(onm.orderRef, path)
		return
	}
	i := 0
	for ; i < len(onm.orderRef[path]); i++ {
		path := onm.orderRef[path][i].Value
		if *path[len(path)-1].Index == *indx {
			break
		}
	}
	if i < len(onm.orderRef[path]) {
		onm.orderIdx.Remove(onm.orderRef[path][i])
		onm.orderRef[path][i] = nil
		onm.orderRef[path] = onm.orderRef[path][:i+copy(onm.orderRef[path][:i], onm.orderRef[path][i+1:])]
	}
}

// GetField the same as Field but for one level deep
// used to implement NM interface
func (onm *OrderedNavigableMap) GetField(path PathItem) (val NMInterface, err error) {
	return onm.nm.GetField(path)
}

/*
// This functione are not needed for the curent implementation
// will decomment them after all are done
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
				onm.orderIdx.PushBack(PathItems{path})
			}
			return
		}
		onm.removePath(PathItems{path})
		if err = onm.nm.SetField(path, val); err == nil {
			onm.orderIdx.PushBack(PathItems{path})
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
					onm.orderIdx.PushBack(PathItems{path})
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
				onm.orderIdx.PushBack(newpath)
			}
		}
		return
	default:
		return ErrNotImplemented
	}
}
*/

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

/*
// updateOrderBasedOnIndex updates the index of the slice elements that are bigger that the removed element
func (onm *OrderedNavigableMap) updateOrderBasedOnIndex(path PathItems, idx int) {
	lenpath := len(path)
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		p := el.Value
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
//*/

// RemoteHost is part of dataStorage interface
func (OrderedNavigableMap) RemoteHost() net.Addr {
	return LocalAddr()
}

// GetOrder returns the elements order as a slice
// use this only for testing
func (onm *OrderedNavigableMap) GetOrder() (order []PathItems) {
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		order = append(order, el.Value)
	}
	return
}
