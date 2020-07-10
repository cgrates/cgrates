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
	"strings"
)

// GetFullFieldPath returns the full path for the
func GetFullFieldPath(fldPath string, dP DataProvider) (fpath *FullPath, err error) {
	var newPath string
	if newPath, err = processFieldPath(fldPath, dP); err != nil || newPath == EmptyString {
		return
	}
	fpath = &FullPath{
		PathItems: NewPathItems(strings.Split(newPath, NestingSep)),
		Path:      newPath,
	}

	return
}

// replaces the dynamic path between <>
func processFieldPath(fldPath string, dP DataProvider) (newPath string, err error) {
	idx := strings.IndexByte(fldPath, RSRDynStartChar)
	if idx == -1 {
		return // no proccessing requred
	}
	newPath = fldPath[:idx] // add the first path of the path without the "<"
	for idx != -1 {         // stop when we do not find any "<"
		fldPath = fldPath[idx+1:]                                   // move the path to the begining of the index
		nextBeginIdx := strings.IndexByte(fldPath, RSRDynStartChar) // get the next "<" if any
		nextEndIdx := strings.IndexByte(fldPath, RSRDynEndChar)     // get the next ">" if any
		if nextEndIdx == -1 {                                       // no end index found so return error
			err = ErrWrongPath
			newPath = EmptyString
			return
		}

		// parse the rest of the field path until we match the [ ]
		bIdx, eIdx := nextBeginIdx, nextEndIdx
		for nextBeginIdx != -1 && nextBeginIdx < nextEndIdx { // do this until no new [ is found or the next begining [ is after the end ]
			nextBeginIdx = strings.IndexByte(fldPath[bIdx+1:], RSRDynStartChar) // get the next "<" if any
			nextEndIdx = strings.IndexByte(fldPath[eIdx+1:], RSRDynEndChar)     // get the next ">" if any
			if nextEndIdx == -1 {                                               // no end index found so return error
				err = ErrWrongPath
				newPath = EmptyString
				return
			}
			if nextBeginIdx == -1 { // if no index found do not increment but replace it
				bIdx = -1
			} else {
				bIdx += nextBeginIdx + 1
			}
			// increment the indexes
			eIdx += nextEndIdx + 1
		}
		var val string
		for _, path := range strings.Split(fldPath[:eIdx], PipeSep) { // proccess the found path
			if val, err = DPDynamicString(path, dP); err != nil {
				newPath = EmptyString
				return
			}
		}

		if bIdx == -1 { // if is the last ocurence add the rest of the path and exit
			newPath += val + fldPath[eIdx+1:]
		} else {
			// else just add until the next [
			newPath += val + fldPath[eIdx+1:bIdx]
		}
		idx = bIdx
	}
	return
}
