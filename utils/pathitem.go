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

// NewFullPath is a constructor for FullPath out of string and separator
func NewFullPath(pth string, sep string) *FullPath {
	return &FullPath{
		Path:      pth,
		PathItems: NewPathItems(strings.Split(pth, sep)),
	}
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
		field, indx := GetPathIndexSlice(v)
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

// Slice returns the path as string slice
func (path PathItems) Slice() (out []string) {
	out = make([]string, len(path))
	for i, v := range path {
		out[i] = v.String()
	}
	return out
}

// PathItem used by the NM interface to store the path information
type PathItem struct {
	Field string
	Index []string
}

func (p PathItem) String() (out string) {
	out = p.Field
	for _, indx := range p.Index {
		out += IdxStart + indx + IdxEnd
	}
	return
}

// Clone creates a copy
func (p PathItem) Clone() (c PathItem) {
	c.Field = p.Field
	if p.Index != nil {
		c.Index = make([]string, len(p.Index))
		for i, indx := range p.Index {
			c.Index[i] = indx
		}
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

// GetPathIndexString returns the path and index as string if index present
// path[index]=>path,index
// path=>path,nil
func GetPathIndexString(spath string) (opath string, idx *string) {
	idxStart := strings.Index(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath, nil
	}
	idxVal := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	return opath, &idxVal
}

// GetPathIndexSlice returns the path and index as string if index present
// path[index]=>path,[index1,index2]
// path=>path,nil
func GetPathIndexSlice(spath string) (opath string, idx []string) {
	idxStart := strings.Index(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath, nil
	}
	idxVal := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	return opath, strings.Split(idxVal, IdxCombination)
}
