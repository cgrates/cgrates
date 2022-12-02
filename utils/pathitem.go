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
	"strings"
)

// stripIdxFromLastPathElm will remove the index from the last path element
func stripIdxFromLastPathElm(path string) string {
	lastDotIdx := strings.LastIndexByte(path, '.')
	lastIdxStart := strings.LastIndexByte(path, '[')
	if lastIdxStart == -1 ||
		(lastDotIdx != -1 && lastDotIdx > lastIdxStart) {
		return path
	}
	return path[:lastIdxStart]
}

// FullPath is the path to the item with all the needed fields
type FullPath struct {
	PathItems PathItems
	Path      string
}

// NewPathItems returns the computed PathItems out of slice one
func NewPathItems(path []string) (pItms PathItems) {
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

// PathItems a list of PathItem used to describe the path to an item from a NavigableMap
type PathItems []PathItem

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

func (path PathItems) String() (out string) {
	for _, v := range path {
		out += NestingSep + v.String()
	}
	if out == "" {
		return
	}
	return out[1:]

}

// PathItem used by the NM interface to store the path information
type PathItem struct {
	Field string
	Index *int
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

func (p PathItem) String() (out string) {
	out = p.Field
	if p.Index != nil {
		out += IdxStart + strconv.Itoa(*p.Index) + IdxEnd
	}
	return
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

// GetPathIndex returns the path and index if index present
// path[index]=>path,index
// path=>path,nil
func GetPathIndex(spath string) (opath string, idx *int) {
	idxStart := strings.Index(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath, nil
	}
	slctr := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	// if strings.HasPrefix(slctr, DynamicDataPrefix) {
	// 	return
	// }
	idxVal, err := strconv.Atoi(slctr)
	if err != nil {
		return spath, nil
	}
	return opath, &idxVal
}

func GetPathWithoutIndex(spath string) (opath string) {
	idxStart := strings.LastIndex(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath
	}
	opath = spath[:idxStart]
	return
}
