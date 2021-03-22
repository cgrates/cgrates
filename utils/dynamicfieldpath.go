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
	return NewFullPath(newPath), nil
}

// replaces the dynamic path between <>
func processFieldPath(fldPath string, dP DataProvider) (newPath string, err error) {
	startIdx := strings.IndexByte(fldPath, RSRDynStartChar)
	if startIdx == -1 {
		return // no proccessing requred
	}
	endIdx := strings.IndexByte(fldPath, RSRDynEndChar)
	if endIdx == -1 {
		err = ErrWrongPath
		newPath = EmptyString
		return
	}
	newPath = fldPath[:startIdx]
	for _, path := range strings.Split(fldPath[startIdx+1:endIdx], InfieldSep) { // proccess the found path
		var val string
		if val, err = DPDynamicString(path, dP); err != nil {
			newPath = EmptyString
			return
		}
		newPath += val
	}
	newPath += fldPath[endIdx+1:]
	return
}
