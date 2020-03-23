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
	"strconv"
)

// FullPath is the path to the item with all the needed fields
type FullPath struct {
	PathItems PathItems
	Path      string
	// PathSlice []string
}

// NewPathToItem returns the prelucrated path to the item
func NewPathToItem(path []string) (pItms PathItems) {
	pItms = make(PathItems, len(path))
	for i, v := range path {
		field, indx := GetPathIndex(v)
		pItms[i] = PathItem{
			Field: field,
			Index: indx,
		}
	}
	return
}

// PathItem used by the NM interface to store the path information
type PathItem struct {
	Field string
	Index *int
}

// PathItems a list of PathItem used to describe the path to an item from a NavigableMap
type PathItems []PathItem

func (p PathItem) String() (out string) {
	out = p.Field
	if p.Index != nil {
		out += IdxStart + strconv.Itoa(*p.Index) + IdxEnd
	}
	return
}

// Equal returns true if p==p2
func (p PathItem) Equal(p2 PathItem) bool {
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

// Clone creates a copy
func (p PathItem) Clone() (c PathItem) {
	// if p == nil {
	// 	return
	// }
	c.Field = p.Field
	if p.Index != nil {
		c.Index = IntPointer(*p.Index)
	}
	return
}

func (path PathItems) String() (out string) {
	for _, v := range path {
		out += NestingSep + v.String()
	}
	if out == "" {
		return
	}
	return out[1:]

}

// Clone creates a copy
func (path PathItems) Clone() (c PathItems) {
	if path == nil {
		return
	}
	c = make(PathItems, len(path))
	for i, v := range path {
		c[i] = v.Clone()
	}
	return
}
