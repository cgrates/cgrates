// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		lastDotIdx != -1 && lastDotIdx > lastIdxStart {
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

// StripTrailingIndex removes the last element from path if it is a
// numeric index (from array storage). Scalar paths are returned as-is.
func StripTrailingIndex(path []string) []string {
	if len(path) == 0 {
		return path
	}
	if _, err := strconv.Atoi(path[len(path)-1]); err == nil {
		return path[:len(path)-1]
	}
	return path
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
