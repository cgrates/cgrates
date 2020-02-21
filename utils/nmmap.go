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

type NavigableMap2 map[string]NM

func (nmm NavigableMap2) String() (out string) {
	for k, v := range nmm {
		out += ",\"" + k + "\":" + v.String()
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
func (nmm NavigableMap2) Set(path []string, val NM) (err error) {
	if len(path) == 0 {
		return fmt.Errorf("Wrong path")
	}
	opath, indx := getPathIndex(path[0])
	el, has := nmm[opath]
	if len(path) == 1 {
		if !has {
			if indx != "" {
				nel := &NMSlice{}
				if err = nel.Set([]string{indx}, val); err != nil {
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
			return el.Set([]string{indx}, val)
		}
		// if el.Type() == NMSliceType {
		// 	el = &NMSlice{} //only works if the last element from path is part of a slice
		// 	nmm[opath] = el
		// 	return el.Set([]string{"0"}, val)
		// }
		nmm[opath] = val
		return
	}
	if !has {
		if indx != "" {
			nel := &NMSlice{}
			path[0] = indx
			if err = nel.Set(path, val); err != nil {
				return
			}
			nmm[opath] = nel
			return
		}
		nel := NavigableMap2{}
		if err = nel.Set(path[1:], val); err != nil {
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
		return el.Set(path, val)
	}
	if el.Type() != NMMapType { // do not try to overwrite an interface
		return
	}
	return el.Set(path[1:], val)
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
func (nmm NavigableMap2) Empty() bool  { return nmm == nil || len(nmm) == 0 }

func (nmm NavigableMap2) GetField(path string) (val NM, err error) {
	opath, indx := getPathIndex(path)
	el, has := nmm[opath]
	if !has {
		return nil, ErrNotFound
	}
	switch el.Type() {
	case NMInterfaceType:
		if indx != "" {
			return nil, ErrNotFound
		}
		return el, nil
	case NMMapType:
		if indx != "" {
			return nil, ErrNotFound
		}
		return el, nil
	case NMSliceType:
		if indx == "" {
			return el, nil
		}
		return el.GetField(indx)
	}
	panic("BUG")
}

func (nmm NavigableMap2) SetField(path string, val NM) (err error) {
	opath, indx := getPathIndex(path)
	el, has := nmm[opath]
	if !has {
		if indx != "" {
			nel := &NMSlice{}
			if err = nel.Set([]string{indx}, val); err != nil {
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
		return el.Set([]string{indx}, val)
	}
	// if el.Type() == NMSliceType {
	// 	el = &NMSlice{} //only works if the last element from path is part of a slice
	// 	nmm[opath] = el
	// 	return el.Set([]string{"0"}, val)
	// }
	nmm[opath] = val
	return

}
func (nmm NavigableMap2) Len() int { return len(nmm) }

// FieldAsString returns thevalue from path as string
func (nmm NavigableMap2) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	val, err = nmm.Field(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(val), nil
}

func (nmm NavigableMap2) FieldAsInterface(fldPath []string) (str interface{}, err error) {
	return nmm.Field(fldPath)
}
