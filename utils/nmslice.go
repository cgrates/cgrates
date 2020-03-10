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
)

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
func (nms *NMSlice) Interface() interface{} { return nms }
func (nms *NMSlice) Field(path PathItems) (val NM, err error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("Wrong path")
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
func (nms *NMSlice) Set(path PathItems, val NM) (err error) {
	if len(path) == 0 || path[0].Index == nil {
		return fmt.Errorf("Wrong path")
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
		return fmt.Errorf("Wrong path")
	}
	if len(path) == 1 {
		(*nms)[idx] = val
		return
	}
	if (*nms)[idx].Type() == NMSliceType {
		return fmt.Errorf("Wrong path")
	}
	return (*nms)[idx].Set(path[1:], val)
}
func (nms *NMSlice) Remove(path PathItems) (err error) {
	if len(path) == 0 || path[0].Index == nil {
		return fmt.Errorf("Wrong path")
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
func (nms NMSlice) Empty() bool  { return nms == nil || len(nms) == 0 }

func (nms *NMSlice) GetField(path *PathItem) (val NM, err error) {
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

func (nms *NMSlice) SetField(path *PathItem, val NM) (err error) {
	if path.Index == nil {
		return fmt.Errorf("Wrong path")
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
		return fmt.Errorf("Wrong path")
	}
	(*nms)[idx] = val
	return
}
func (nms *NMSlice) Len() int { return len(*nms) }
