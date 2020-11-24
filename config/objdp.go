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

package config

import (
	"fmt"
	"net"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// NewObjectDP constructs a utils.DataProvider
func NewObjectDP(obj interface{}) (dP utils.DataProvider) {
	dP = &ObjectDP{
		obj:   obj,
		cache: make(map[string]interface{}),
	}
	return
}

// ObjectDP implements the DataProvider for any interface{}
type ObjectDP struct {
	obj   interface{}
	cache map[string]interface{}
}

func (objDP *ObjectDP) setCache(path string, val interface{}) {
	objDP.cache[path] = val
}

func (objDP *ObjectDP) getCache(path string) (val interface{}, has bool) {
	val, has = objDP.cache[path]
	return
}

// String is part of engine.utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (objDP *ObjectDP) String() string {
	return utils.ToJSON(objDP.obj)
}

// FieldAsInterface is part of engine.utils.DataProvider interface
func (objDP *ObjectDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	obj := objDP.obj
	// []string{ BalanceMap *monetary[0] Value }
	var has bool
	if data, has = objDP.getCache(strings.Join(fldPath, utils.NestingSep)); has {
		return
	}
	data = obj // in case the fldPath is empty we need to return the whole object
	var prevFld string
	for _, fld := range fldPath {
		var slctrStr string
		if splt := strings.Split(fld, utils.IdxStart); len(splt) != 1 { // check if we have selector
			fld = splt[0]
			if splt[1][len(splt[1])-1:] != utils.IdxEnd {
				return nil, fmt.Errorf("filter rule <%s> needs to end in ]", splt[1])
			}
			slctrStr = splt[1][:len(splt[1])-1] // also strip the last ]
		}
		if prevFld == utils.EmptyString {
			prevFld += fld
		} else {
			prevFld += utils.NestingSep + fld
		}
		// check if we take the current path from cache
		if data, has = objDP.getCache(prevFld); !has {
			if data, err = utils.ReflectFieldMethodInterface(obj, fld); err != nil { // take the object the field for current path
				// in case of error set nil for the current path and return err
				objDP.setCache(prevFld, nil)
				return nil, err
			}
			// add the current field in prevFld so we can set in cache the full path with it's data
			objDP.setCache(prevFld, data)
		}
		// change the obj to be the current data and continue the processing
		obj = data
		if slctrStr != utils.EmptyString { //we have selector so we need to do an aditional get
			prevFld += utils.IdxStart + slctrStr + utils.IdxEnd
			// check if we take the current path from cache
			if data, has = objDP.getCache(prevFld); !has {
				if data, err = utils.ReflectFieldMethodInterface(obj, slctrStr); err != nil { // take the object the field for current path
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
	objDP.setCache(strings.Join(fldPath, utils.NestingSep), data)
	return
}

// FieldAsString is part of engine.utils.DataProvider interface
func (objDP *ObjectDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = objDP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of engine.utils.DataProvider interface
func (objDP *ObjectDP) RemoteHost() net.Addr {
	return utils.LocalAddr()
}
