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

import "strconv"

// NMSlice is the basic slice of NM interface
type NMSlice []NMInterface

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
func (nms *NMSlice) Field(path PathItems) (val NMInterface, err error) {
	if len(path) == 0 {
		return nil, ErrWrongPath
	}
	if nms.Empty() || path[0].Index == nil {
		return nil, ErrNotFound
	}
	var idx int
	if idx, err = strconv.Atoi(*path[0].Index); err != nil {
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

// Set sets the value for the given index
func (nms *NMSlice) Set(path PathItems, val NMInterface) (addedNew bool, err error) {
	if len(path) == 0 || path[0].Index == nil {
		return false, ErrWrongPath
	}
	var idx int
	if idx, err = strconv.Atoi(*path[0].Index); err != nil {
		return
	}
	if idx == len(*nms) { // append element
		addedNew = true
		if len(path) == 1 {
			*nms = append(*nms, val)
			return
		}
		nel := NavigableMap2{}
		if _, err = nel.Set(path[1:], val); err != nil {
			return
		}
		*nms = append(*nms, nel)
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) {
		return false, ErrWrongPath
	}
	path[0].Index = StringPointer(strconv.Itoa(idx))
	if len(path) == 1 {
		(*nms)[idx] = val
		return
	}
	if (*nms)[idx].Type() == NMSliceType {
		return false, ErrWrongPath
	}
	return (*nms)[idx].Set(path[1:], val)
}

// Remove removes the item for the given index
func (nms *NMSlice) Remove(path PathItems) (err error) {
	if len(path) == 0 || path[0].Index == nil {
		return ErrWrongPath
	}
	var idx int
	if idx, err = strconv.Atoi(*path[0].Index); err != nil {
		return
	}
	if idx < 0 {
		idx = len(*nms) + idx
	}
	if idx < 0 || idx >= len(*nms) { // already removed
		return
	}
	path[0].Index = StringPointer(strconv.Itoa(idx))
	if len(path) == 1 {
		*nms = append((*nms)[:idx], (*nms)[idx+1:]...)
		return
	}
	if (*nms)[idx].Type() != NMMapType {
		return ErrWrongPath
	}
	if err = (*nms)[idx].Remove(path[1:]); err != nil {
		return
	}
	if (*nms)[idx].Empty() {
		*nms = append((*nms)[:idx], (*nms)[idx+1:]...)
	}
	return
}

// Type returns the type of the NM slice
func (nms NMSlice) Type() NMType {
	return NMSliceType
}

// Empty returns true if the NM is empty(no data)
func (nms NMSlice) Empty() bool {
	return nms == nil || len(nms) == 0
}

// Len returns the lenght of the slice
func (nms *NMSlice) Len() int {
	return len(*nms)
}
