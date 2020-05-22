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

// NewDynamicDataProvider constructs a dynamic data provider
func NewDynamicDataProvider(dp DataProvider) *DynamicDataProvider {
	return &DynamicDataProvider{
		DataProvider: dp,
		cache:        make(map[string]interface{}),
	}
}

// DynamicDataProvider is a data source from multiple formats
type DynamicDataProvider struct {
	DataProvider
	cache map[string]interface{}
}

// FieldAsInterface to overwrite the FieldAsInterface function from the given DataProvider
func (ddp *DynamicDataProvider) FieldAsInterface(fldPath []string) (out interface{}, err error) {
	path := strings.Join(fldPath, NestingSep) // join the path so we can check it in cache and parse it more easy
	if val, has := ddp.cache[path]; has {     // check if we have the path in cache
		return val, nil
	}
	var newPath string
	if newPath, err = ddp.proccesFieldPath(path); err != nil { // proccess the path
		return
	}
	if newPath == EmptyString { // no new path means no dynamic path so just take the value from the data provider
		return ddp.DataProvider.FieldAsInterface(fldPath)
	}
	// split the new path and get that field
	if out, err = ddp.DataProvider.FieldAsInterface(strings.Split(newPath, NestingSep)); err != nil {
		return
	}
	// if no error save in cache the path
	ddp.cache[path] = out
	return
}

func (ddp *DynamicDataProvider) proccesFieldPath(fldPath string) (newPath string, err error) {
	idx := strings.Index(fldPath, IdxStart)
	if idx == -1 {
		return // no proccessing requred
	}
	newPath = fldPath[:idx+1] // add the first path of the path with the "[" included
	for idx != -1 {           // stop when we do not find any "["
		fldPath = fldPath[idx+1:]                        // move the path to the begining of the index
		nextBeginIdx := strings.Index(fldPath, IdxStart) // get the next "[" if any
		nextEndIdx := strings.Index(fldPath, IdxEnd)     // get the next "]" if any
		if nextEndIdx == -1 {                            // no end index found so return error
			err = ErrWrongPath
			newPath = EmptyString
			return
		}

		// parse the rest of the field path until we match the [ ]
		bIdx, eIdx := nextBeginIdx, nextEndIdx
		for nextBeginIdx != -1 && nextBeginIdx < nextEndIdx { // do this until no new [ is found or the next begining [ is after the end ]
			nextBeginIdx = strings.Index(fldPath[bIdx+1:], IdxStart) // get the next "[" if any
			nextEndIdx = strings.Index(fldPath[eIdx+1:], IdxEnd)     // get the next "]" if any
			if nextEndIdx == -1 {                                    // no end index found so return error
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
			var iface interface{}
			if iface, err = DPDynamicInterface(path, ddp); err != nil {
				newPath = EmptyString
				return
			}
			val += IfaceAsString(iface) // compose the value
		}
		if bIdx == -1 { // if is the last ocurence add the rest of the path and exit
			newPath += val + fldPath[eIdx:]
		} else {
			// else just add until the next [
			newPath += val + fldPath[eIdx:bIdx+1]
		}
		idx = bIdx
	}
	return
}

// GetFullFieldPath returns the full path for the
func (ddp *DynamicDataProvider) GetFullFieldPath(fldPath string) (fpath *FullPath, err error) {
	var newPath string
	if newPath, err = ddp.proccesFieldPathForSet(fldPath); err != nil || newPath == EmptyString {
		return
	}
	fpath = &FullPath{
		PathItems: NewPathItems(strings.Split(newPath, NestingSep)),
		Path:      newPath,
	}

	return
}

// FieldAsString returns the value from path as string
func (ddp DynamicDataProvider) FieldAsString(fldPath []string) (str string, err error) {
	var val interface{}
	if val, err = ddp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// does the same thing as ... but replaces [ with . if the value between [] is dynamic
func (ddp *DynamicDataProvider) proccesFieldPathForSet(fldPath string) (newPath string, err error) {
	idx := strings.Index(fldPath, IdxStart)
	if idx == -1 {
		return // no proccessing requred
	}
	var hasDyn bool         // to be able to determine if the path has selector
	newPath = fldPath[:idx] // add the first path of the path without the "["
	for idx != -1 {         // stop when we do not find any "["
		fldPath = fldPath[idx+1:]                        // move the path to the begining of the index
		nextBeginIdx := strings.Index(fldPath, IdxStart) // get the next "[" if any
		nextEndIdx := strings.Index(fldPath, IdxEnd)     // get the next "]" if any
		if nextEndIdx == -1 {                            // no end index found so return error
			err = ErrWrongPath
			newPath = EmptyString
			return
		}

		// parse the rest of the field path until we match the [ ]
		bIdx, eIdx := nextBeginIdx, nextEndIdx
		for nextBeginIdx != -1 && nextBeginIdx < nextEndIdx { // do this until no new [ is found or the next begining [ is after the end ]
			nextBeginIdx = strings.Index(fldPath[bIdx+1:], IdxStart) // get the next "[" if any
			nextEndIdx = strings.Index(fldPath[eIdx+1:], IdxEnd)     // get the next "]" if any
			if nextEndIdx == -1 {                                    // no end index found so return error
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
		var isDyn bool
		for _, path := range strings.Split(fldPath[:eIdx], PipeSep) { // proccess the found path
			var iface interface{}
			if strings.HasPrefix(path, DynamicDataPrefix) {
				isDyn = true
				path2 := strings.TrimPrefix(path, DynamicDataPrefix)
				if iface, err = ddp.FieldAsInterface(strings.Split(path2, NestingSep)); err != nil {
					newPath = EmptyString
					return
				}
				val += IfaceAsString(iface) // compose the value
			} else {

				val += IfaceAsString(path) // compose the value
			}
		}
		if isDyn {
			hasDyn = true
			val = NestingSep + val
		} else {
			val = IdxStart + val + IdxEnd
		}

		if bIdx == -1 { // if is the last ocurence add the rest of the path and exit
			newPath += val + fldPath[eIdx+1:]
		} else {
			// else just add until the next [
			newPath += val + fldPath[eIdx+1:bIdx]
		}
		idx = bIdx
	}
	if !hasDyn { // the path doesn't have dynamic selector
		newPath = EmptyString
	}
	return
}
