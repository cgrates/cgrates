// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"net"
	"strings"
)

// posible NMType
const (
	NMDataType NMType = iota
	NMMapType
	NMSliceType
)

// DataProvider is a data source from multiple formats
type DataProvider interface {
	String() string // printable version of data
	FieldAsInterface(fldPath []string) (any, error)
	FieldAsString(fldPath []string) (string, error) // remove this
	RemoteHost() net.Addr
}

// NavigableMapper is the interface supported by replies convertible to CGRReply
type NavigableMapper interface {
	AsNavigableMap() NavigableMap2
}

// DPDynamicInterface returns the value of the field if the path is dynamic
func DPDynamicInterface(dnVal string, dP DataProvider) (any, error) {
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

// NMInterface the basic interface
type NMInterface interface {
	String() string
	Interface() any
	Field(path PathItems) (val NMInterface, err error)
	Set(path PathItems, val NMInterface) (addedNew bool, err error)
	Remove(path PathItems) (err error)
	Type() NMType
	Empty() bool
	Len() int
}

// navMap subset of function for NM interface
type navMap interface {
	Field(path PathItems) (val NMInterface, err error)
	Set(fullpath *FullPath, val NMInterface) (addedNew bool, err error)
}

// AppendNavMapVal appends value to the map
func AppendNavMapVal(nm navMap, fldPath *FullPath, val NMInterface) (err error) {
	var prevItm NMInterface
	var indx int
	if prevItm, err = nm.Field(fldPath.PathItems); err != nil {
		if err != ErrNotFound {
			return
		}
	} else {
		indx = prevItm.Len()
	}
	fldPath.PathItems[len(fldPath.PathItems)-1].Index = &indx
	_, err = nm.Set(fldPath, val)
	return
}

// ComposeNavMapVal compose adds value to prevision item
func ComposeNavMapVal(nm navMap, fldPath *FullPath, val NMInterface) (err error) {
	var prevItmSlice NMInterface
	var indx int
	if prevItmSlice, err = nm.Field(fldPath.PathItems); err != nil {
		if err != ErrNotFound {
			return
		}
	} else {
		indx = prevItmSlice.Len() - 1
		var prevItm NMInterface
		if prevItm, err = prevItmSlice.Field(PathItems{{Index: &indx}}); err != nil {
			if err != ErrNotFound {
				return
			}
		} else if _, err = val.Set(nil, NewNMData(IfaceAsString(prevItm.Interface())+IfaceAsString(val.Interface()))); err != nil {
			return
		}
	}
	fldPath.PathItems[len(fldPath.PathItems)-1].Index = &indx
	_, err = nm.Set(fldPath, val)
	return
}
