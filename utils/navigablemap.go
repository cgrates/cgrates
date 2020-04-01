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

// NavigableMap2 is the basic map of NM interface
type NavigableMap2 map[string]NMInterface

func (nm NavigableMap2) String() (out string) {
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
func (nm NavigableMap2) Interface() interface{} {
	return nm
}

// Field returns the item on the given path
func (nm NavigableMap2) Field(path PathItems) (val NMInterface, err error) {
	if len(path) == 0 {
		return nil, ErrWrongPath
	}
	el, has := nm[path[0].Field]
	if !has {
		return nil, ErrNotFound
	}
	if len(path) == 1 && path[0].Index == nil {
		return el, nil
	}
	switch el.Type() {
	default:
		return nil, ErrNotFound
	case NMMapType:
		if path[0].Index != nil {
			return nil, ErrNotFound
		}
		return el.Field(path[1:])
	case NMSliceType:
		return el.Field(path)
	}
}

// Set sets the value for the given path
func (nm NavigableMap2) Set(path PathItems, val NMInterface) (addedNew bool, err error) {
	if len(path) == 0 {
		return false, ErrWrongPath
	}
	el, has := nm[path[0].Field]
	if len(path) == 1 {
		if !has {
			addedNew = true
			if path[0].Index != nil {
				nel := &NMSlice{}
				if _, err = nel.Set(path, val); err != nil {
					return
				}
				nm[path[0].Field] = nel
				return
			}
			nm[path[0].Field] = val
			return
		}
		if path[0].Index != nil {
			if el.Type() != NMSliceType {
				return false, ErrWrongPath
			}
			return el.Set(path, val)
		}
		nm[path[0].Field] = val
		return
	}
	if !has {
		addedNew = true
		if path[0].Index != nil {
			nel := &NMSlice{}
			if _, err = nel.Set(path, val); err != nil {
				return
			}
			nm[path[0].Field] = nel
			return
		}
		nel := NavigableMap2{}
		if _, err = nel.Set(path[1:], val); err != nil {
			return
		}
		nm[path[0].Field] = nel
		return
	}
	if path[0].Index != nil {
		if el.Type() != NMSliceType {
			return false, ErrWrongPath
		}
		return el.Set(path, val)
	}
	if el.Type() != NMMapType { // do not try to overwrite an interface
		return false, ErrWrongPath
	}
	return el.Set(path[1:], val)
}

// Remove removes the item for the given path
func (nm NavigableMap2) Remove(path PathItems) (err error) {
	if len(path) == 0 {
		return ErrWrongPath
	}
	el, has := nm[path[0].Field]
	if !has {
		return // already removed
	}
	if len(path) == 1 {
		if path[0].Index != nil {
			if el.Type() != NMSliceType {
				return ErrWrongPath
			}
			// this should not return error
			// but in case it does we propagate it further
			err = el.Remove(path)
			if el.Empty() {
				delete(nm, path[0].Field)
			}
			return
		}
		delete(nm, path[0].Field)
		return
	}
	if path[0].Index != nil {
		if el.Type() != NMSliceType {
			return ErrWrongPath
		}
		if err = el.Remove(path); err != nil {
			return
		}
		if el.Empty() {
			delete(nm, path[0].Field)
		}
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
func (nm NavigableMap2) Type() NMType {
	return NMMapType
}

// Empty returns true if the NM is empty(no data)
func (nm NavigableMap2) Empty() bool {
	return nm == nil || len(nm) == 0
}

// GetField is the same as Field but for one level deep
// Is used for OrderedNavigableMap parsing when setting a value
func (nm NavigableMap2) GetField(path PathItem) (val NMInterface, err error) {
	// if path == nil {
	// 	return nil, ErrWrongPath
	// }
	el, has := nm[path.Field]
	if !has {
		return nil, ErrNotFound
	}
	if path.Index == nil {
		return el, nil
	}
	if el.Type() == NMSliceType {
		return el.GetField(path)
	}
	return nil, ErrNotFound
}

// SetField is the same as Set but for one level deep
// Is used for OrderedNavigableMap parsing when setting a value
func (nm NavigableMap2) SetField(path PathItem, val NMInterface) (err error) {
	// if path == nil {
	// 	return ErrWrongPath
	// }
	el, has := nm[path.Field]
	if !has {
		if path.Index != nil {
			nel := &NMSlice{}
			if err = nel.SetField(path, val); err != nil {
				return
			}
			nm[path.Field] = nel
			return
		}
		nm[path.Field] = val
		return
	}
	if path.Index != nil {
		if el.Type() != NMSliceType {
			return ErrWrongPath
		}
		return el.SetField(path, val)
	}
	nm[path.Field] = val
	return

}

// Len returns the lenght of the map
func (nm NavigableMap2) Len() int {
	return len(nm)
}

// FieldAsInterface returns the interface at the path
// Is used by AgentRequest FieldAsInterface
func (nm NavigableMap2) FieldAsInterface(fldPath []string) (str interface{}, err error) {
	var nmi NMInterface
	if nmi, err = nm.Field(NewPathToItem(fldPath)); err != nil {
		return
	}
	return nmi.Interface(), nil
}

// FieldAsString returns the string at the path
// Used only to implement the DataProvider interface
func (nm NavigableMap2) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	val, err = nm.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// RemoteHost is part of dataStorage interface
func (NavigableMap2) RemoteHost() net.Addr {
	return LocalAddr()
}
