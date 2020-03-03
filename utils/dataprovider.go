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

const (
	NMInterfaceType NMType = iota
	NMMapType
	NMSliceType
)

// PathItem used by the NM interface to store the path information
type PathItem struct {
	Field string
	Index *int
}

// NM the basic interface
type NM interface {
	String() string
	Interface() interface{}
	Field(path []*PathItem) (val NM, err error)
	GetField(path *PathItem) (val NM, err error)
	SetField(path *PathItem, val NM) (err error)
	Set(path []*PathItem, val NM) (err error)
	Remove(path []*PathItem) (err error)
	Type() NMType
	Empty() bool
	Len() int
}

func NewPathToItemFromSlice(path []string) (pItms []*PathItem) {
	pItms = make([]*PathItem, len(path))
	for i, v := range path {
		field, indx := GetPathIndex(v)
		pItms[i] = &PathItem{
			Field: field,
			Index: indx,
		}
	}
	return
}

func NewPathToItem(path string) []*PathItem {
	return NewPathToItemFromSlice(strings.Split(path, NestingSep))
}

func (p *PathItem) String() (out string) {
	out = p.Field
	if p.Index != nil {
		out += IdxStart + strconv.Itoa(*p.Index) + IdxEnd
	}
	return
}

func (p *PathItem) Equal(p2 *PathItem) bool {
	if p.Field != p2.Field {
		return false
	}
	if p.Index == nil && p2.Index == nil {
		return true
	}
	if p.Index != nil && p2.Index != nil {
		return *p.Index == *p2.Index
	}
	return false
}

func PathItemsToString(path []*PathItem) (out string) {
	for _, v := range path {
		out += NestingSep + v.String()
	}
	if out == "" {
		return
	}
	return out[1:]

}
