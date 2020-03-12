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

// NMSlice is the basic slice of NM interface
type NMSlice []NM

func (nms *NMSlice) String() (out string) {
	for _, v := range *nms {
		out += "," + v.String()
	}
	if len(out) == 0 {
		return "[]"
	}
	out = out[1:]
	return "[" + out + "]"
}

// Interface returns itself
func (nms *NMSlice) Interface() interface{} {
	return nms
}

// Field returns the item on the given path
// for NMSlice only the Index field is considered
func (nms *NMSlice) Field(path PathItems) (val NM, err error) {
	if len(path) == 0 {
		return nil, ErrWrongPath
	}
	if nms.Empty() || path[0].Index == nil {
		return nil, ErrNotFound
	}
	idx := *path[0].Index
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

// Set sets the value for the given index
func (nms *NMSlice) Set(path PathItems, val NM) (err error) {
	if len(path) == 0 || path[0].Index == nil {
		return ErrWrongPath
	}
	idx := *path[0].Index
	if idx == len(*nms) { // append element
		if len(path) == 1 {
			*nms = append(*nms, val)
			return
		}
		nel := NavigableMap2{}
		if err = nel.Set(path[1:], val); err != nil {
			return
		}
		*nms = append(*nms, nel)
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) {
		return ErrWrongPath
	}
	if len(path) == 1 {
		(*nms)[idx] = val
		return
	}
	if (*nms)[idx].Type() == NMSliceType {
		return ErrWrongPath
	}
	return (*nms)[idx].Set(path[1:], val)
}

// Remove removes the item for the given index
func (nms *NMSlice) Remove(path PathItems) (err error) {
	if len(path) == 0 || path[0].Index == nil {
		return ErrWrongPath
	}
	idx := *path[0].Index
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) { // already removed
		return
	}
	switch (*nms)[idx].Type() {
	case NMSliceType:
		return ErrWrongPath
	case NMInterfaceType:
		if len(path) != 1 {
			return ErrWrongPath
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

// Type returns the type of the NM slice
func (nms NMSlice) Type() NMType {
	return NMSliceType
}

// Empty returns true if the NM is empty(no data)
func (nms NMSlice) Empty() bool {
	return nms == nil || len(nms) == 0
}

// GetField the same as Field but for one level deep
// used for OrderedNavigableMap parsing
func (nms *NMSlice) GetField(path *PathItem) (val NM, err error) {
	if path == nil {
		return nil, ErrWrongPath
	}
	if nms.Empty() || path.Index == nil {
		return nil, ErrNotFound
	}
	idx := *path.Index
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) {
		return nil, ErrNotFound
	}
	return (*nms)[idx], nil
}

// SetField the same as Set but for one level deep
// used for OrderedNavigableMap parsing
func (nms *NMSlice) SetField(path *PathItem, val NM) (err error) {
	if path.Index == nil {
		return ErrWrongPath
	}
	idx := *path.Index
	if idx == len(*nms) { // append element
		*nms = append(*nms, val)
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) {
		return ErrWrongPath
	}
	(*nms)[idx] = val
	return
}

// Len returns the lenght of the slice
func (nms *NMSlice) Len() int {
	return len(*nms)
}
