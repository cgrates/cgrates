/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
	"strings"
)

// NewObjectDP constructs a DataProvider
func NewObjectDP(obj any) DataProvider {
	return &ObjectDP{
		obj:   obj,
		cache: make(map[string]any),
	}
}

// ObjectDP implements the DataProvider for any any
type ObjectDP struct {
	obj   any
	cache map[string]any
}

func (objDP *ObjectDP) setCache(path string, val any) {
	objDP.cache[path] = val
}

func (objDP *ObjectDP) getCache(path string) (val any, has bool) {
	val, has = objDP.cache[path]
	return
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (objDP *ObjectDP) String() string {
	return ToJSON(objDP.obj)
}

// FieldAsInterface is part of engine.DataProvider interface
func (objDP *ObjectDP) FieldAsInterface(fldPath []string) (data any, err error) {
	obj := objDP.obj
	// []string{ BalanceMap *monetary[0] Value }
	var has bool
	if data, has = objDP.getCache(strings.Join(fldPath, NestingSep)); has {
		if data == nil {
			err = ErrNotFound
		}
		return
	}
	data = obj // in case the fldPath is empty we need to return the whole object
	var prevFld string
	for _, fld := range fldPath {
		var slctrStr string
		if splt := strings.Split(fld, IdxStart); len(splt) != 1 { // check if we have selector
			fld = splt[0]
			if splt[1][len(splt[1])-1:] != IdxEnd {
				return nil, fmt.Errorf("filter rule <%s> needs to end in ]", splt[1])
			}
			slctrStr = splt[1][:len(splt[1])-1] // also strip the last ]
		}
		if prevFld == EmptyString {
			prevFld += fld
		} else {
			prevFld += NestingSep + fld
		}
		// check if we take the current path from cache
		if data, has = objDP.getCache(prevFld); !has {
			if data, err = ReflectFieldMethodInterface(obj, fld); err != nil { // take the object the field for current path
				// in case of error set nil for the current path and return err
				objDP.setCache(prevFld, nil)
				return nil, err
			}
			// add the current field in prevFld so we can set in cache the full path with it's data
			objDP.setCache(prevFld, data)
		}
		// change the obj to be the current data and continue the processing
		obj = data
		if slctrStr != EmptyString { //we have selector so we need to do an aditional get
			prevFld += IdxStart + slctrStr + IdxEnd
			// check if we take the current path from cache
			if data, has = objDP.getCache(prevFld); !has {
				if data, err = ReflectFieldMethodInterface(obj, slctrStr); err != nil { // take the object the field for current path
					// in case of error set nil for the current path and return err
					objDP.setCache(prevFld, nil)
					return nil, err
				}
				// add the current field in prevFld so we can set in cache the full path with it's data
				objDP.setCache(prevFld, data)
			}
			// change the obj to be the current data and continue the processing
			obj = data
		}

	}
	//add in cache the initial path
	objDP.setCache(strings.Join(fldPath, NestingSep), data)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (objDP *ObjectDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface any
	if valIface, err = objDP.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(valIface), nil
}
