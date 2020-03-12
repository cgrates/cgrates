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
	"strings"
)

// DataProvider is a data source from multiple formats
type DataProvider interface {
	String() string // printable version of data
	FieldAsInterface(fldPath []string) (interface{}, error)
	FieldAsString(fldPath []string) (string, error) // remove this
	RemoteHost() net.Addr
}

// NavigableMapper is the interface supported by replies convertible to CGRReply
type NavigableMapper interface {
	AsNavigableMap() NavigableMap2
}

// DPDynamicInterface returns the value of the field if the path is dynamic
func DPDynamicInterface(dnVal string, dP DataProvider) (interface{}, error) {
	if strings.HasPrefix(dnVal, DynamicDataPrefix) {
		dnVal = strings.TrimPrefix(dnVal, DynamicDataPrefix)
		return dP.FieldAsInterface(strings.Split(dnVal, NestingSep))
	}
	return StringToInterface(dnVal), nil
}

// DPDynamicString returns the string value of the field if the path is dynamic
func DPDynamicString(dnVal string, dP DataProvider) (string, error) {
	if strings.HasPrefix(dnVal, DynamicDataPrefix) {
		dnVal = strings.TrimPrefix(dnVal, DynamicDataPrefix)
		return dP.FieldAsString(strings.Split(dnVal, NestingSep))
	}
	return dnVal, nil
}

// NMType the type used for navigable Map
type NMType byte

// posible NMType
const (
	NMInterfaceType NMType = iota
	NMMapType
	NMSliceType
)

// NM the basic interface
type NM interface {
	String() string
	Interface() interface{}
	Field(path PathItems) (val NM, err error)
	GetField(path *PathItem) (val NM, err error)
	SetField(path *PathItem, val NM) (err error)
	Set(path PathItems, val NM) (err error)
	Remove(path PathItems) (err error)
	Type() NMType
	Empty() bool
	Len() int
}

// navMap subset of function for NM interface
type navMap interface {
	Field(path PathItems) (val NM, err error)
	Set(path PathItems, val NM) (err error)
}

// AppendNavMapVal appends value to the map
func AppendNavMapVal(nm navMap, fldPath PathItems, val NM) (err error) {
	var prevItm NM
	var indx int
	if prevItm, err = nm.Field(fldPath); err != nil {
		if err != ErrNotFound {
			return
		}
	} else {
		indx = prevItm.Len()
	}
	fldPath[len(fldPath)-1].Index = &indx
	return nm.Set(fldPath, val)
}

// ComposeNavMapVal compose adds value to prevision item
func ComposeNavMapVal(nm navMap, fldPath PathItems, val NM) (err error) {
	var prevItmSlice NM
	var indx int
	if prevItmSlice, err = nm.Field(fldPath); err != nil {
		if err != ErrNotFound {
			return
		}
	} else {
		indx = prevItmSlice.Len() - 1
		var prevItm NM
		if prevItm, err = prevItmSlice.GetField(&PathItem{Index: &indx}); err != nil {
			if err != ErrNotFound {
				return
			}
		} else if err = val.Set(nil, NewNMInterface(IfaceAsString(prevItm.Interface())+IfaceAsString(val.Interface()))); err != nil {
			return
		}
	}
	fldPath[len(fldPath)-1].Index = &indx
	return nm.Set(fldPath, val)
}
