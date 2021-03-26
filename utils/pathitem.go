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

// NewFullPath is a constructor for FullPath out of string
func NewFullPath(path string) *FullPath {
	return &FullPath{
		Path:      path,
		PathSlice: CompilePath(path),
	}
}

// FullPath is the path to the item with all the needed fields
type FullPath struct {
	PathSlice []string
	Path      string
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
