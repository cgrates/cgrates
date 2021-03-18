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
)

// NavigableMap is the basic map of NM interface
type NavigableMap map[string]NMInterface

func (nm NavigableMap) String() (out string) {
	for k, v := range nm {
		out += ",\"" + k + "\":" + v.String()
	}
	if len(out) == 0 {
		return "{}"
	}
	out = out[1:]
	return "{" + out + "}"
}

// Interface returns itself
func (nm NavigableMap) Interface() interface{} {
	return nm
}

// Field returns the item on the given path
func (nm NavigableMap) Field(path PathItems) (val NMInterface, err error) {
	if len(path) == 0 {
		return nil, ErrWrongPath
	}
	el, has := nm[path[0].Field]
	if !has {
		return nil, ErrNotFound
	}
	if len(path) == 1 &&
		len(path[0].Index) == 0 {
		return el, nil
	}
	switch el.Type() {
	default:
		return nil, ErrNotFound
	case NMMapType:
		if len(path[0].Index) != 0 { // in case we have multiple indexes move the cursor and send the path to next element
			path[0].Field = path[0].Index[0]
			path[0].Index = path[0].Index[1:] // this should not panic as the length is not 0
			return el.Field(path)
		}
		return el.Field(path[1:])
	case NMSliceType:
		return el.Field(path)
	}
}

// Set sets the value for the given path
func (nm NavigableMap) Set(path PathItems, val NMInterface) (added bool, err error) {
	if len(path) == 0 {
		return false, ErrWrongPath
	}
	nmItm, has := nm[path[0].Field]

	if len(path[0].Index) > 0 { // has indexes, should be a slice which is kinda part of our map, hence separate handling
		if !has { // the NMInterface doesn't exit so create it based on indexes and the rest of the path
			if nmItm, err = createFromIndexes(path[0].Index, path[1:], val); err != nil {
				return
			}
			added = true
			nm[path[0].Field] = nmItm
			return
		}
		switch nmItm.Type() { // based on type we handle the indexes
		default: // NMDataType
			return false, ErrWrongPath
		case NMSliceType: // let the slice handle the indexes
			return nmItm.Set(path, val)
		case NMMapType: // recreate the path list in order to not update the one above(it is needed for OrderNavigableMap indexing)
			return nmItm.Set(append(PathItems{{Field: path[0].Index[0],
				Index: path[0].Index[1:]}}, path[1:]...), val)
		}
	}

	// standard handling
	if len(path) == 1 { // always overwrite for single path
		nm[path[0].Field] = val
		if !has {
			added = true
		}
		return
	}
	// from here we should deal only with navmaps due to multiple path
	if !has {
		nmItm = NavigableMap{}
		nm[path[0].Field] = nmItm
	}
	if nmItm.Type() != NMMapType { // do not try to overwrite an interface
		return false, ErrWrongPath
	}
	return nmItm.Set(path[1:], val)
}

// Remove removes the item for the given path
func (nm NavigableMap) Remove(path PathItems) (err error) {
	if len(path) == 0 {
		return ErrWrongPath
	}
	el, has := nm[path[0].Field]
	if !has {
		return // already removed
	}
	if len(path[0].Index) > 0 {
		switch el.Type() { // based on type we handle the indexes
		default: // NMDataType
			return ErrWrongPath
		case NMSliceType: // let the slice handle the indexes
			err = el.Remove(path)
		case NMMapType: // recreate the path list in order to not update the one above(it is needed for OrderNavigableMap indexing)
			err = el.Remove(append(PathItems{{Field: path[0].Index[0],
				Index: path[0].Index[1:]}}, path[1:]...))
		}
		if el.Empty() {
			delete(nm, path[0].Field)
		}
		return
	}
	if len(path) == 1 {
		delete(nm, path[0].Field)
		return
	}
	if el.Type() != NMMapType {
		return ErrWrongPath
	}
	if err = el.Remove(path[1:]); err != nil {
		return
	}
	if el.Empty() {
		delete(nm, path[0].Field)
	}
	return
}

// Type returns the type of the NM map
func (nm NavigableMap) Type() NMType {
	return NMMapType
}

// Empty returns true if the NM is empty(no data)
func (nm NavigableMap) Empty() bool {
	return nm == nil || len(nm) == 0
}

// Len returns the lenght of the map
func (nm NavigableMap) Len() int {
	return len(nm)
}

// FieldAsInterface returns the interface at the path
// Is used by AgentRequest FieldAsInterface
func (nm NavigableMap) FieldAsInterface(fldPath []string) (str interface{}, err error) {
	var nmi NMInterface
	if nmi, err = nm.Field(NewPathItems(fldPath)); err != nil {
		return
	}
	return nmi.Interface(), nil
}

// FieldAsString returns the string at the path
// Used only to implement the DataProvider interface
func (nm NavigableMap) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	val, err = nm.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// RemoteHost is part of dataStorage interface
func (NavigableMap) RemoteHost() net.Addr {
	return LocalAddr()
}

func createFromIndexes(indx []string, path PathItems, nm NMInterface) (nmItm NMInterface, err error) {
	// safe to assume that len(indx) is greater than 0
	if len(indx) != 1 { // not the last index
		var nmEl NMInterface
		if nmEl, err = createFromIndexes(indx[1:], path, nm); err != nil {
			return
		}
		nm = nmEl
	} else if len(path) != 0 { // last element in the indexes but it has extra path
		nmEl := NavigableMap{}                       // we have path the next item is a map
		if _, err = nmEl.Set(path, nm); err != nil { // set the element in map
			return
		}
		nm = nmEl // rewrite the nm with the new map
	}

	// nm will be:
	// 	- a new created NMInterface if we have more indexes
	//  - a NavigableMap if we do not have any more indexes but we have extra path
	//	- the original nm passed to the function if do not have indexes and extra path
	// create the nmItem based on the index type
	val, idxErr := strconv.Atoi(indx[0]) // ignore this error as if we can not convert this we asume that we want to create a map
	if idxErr != nil {                   // is a map
		return NavigableMap{indx[0]: nm}, nil
	}
	// is a slice
	if val != 0 { // only create if index is 0
		return nil, ErrWrongPath
	}
	return &NMSlice{nm}, nil
}
