// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// NewObjectDP constructs a utils.DataProvider
func NewObjectDP(obj any) utils.DataProvider {
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

// String is part of engine.utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (objDP *ObjectDP) String() string {
	return utils.ToJSON(objDP.obj)
}

// FieldAsInterface is part of engine.utils.DataProvider interface
func (objDP *ObjectDP) FieldAsInterface(fldPath []string) (data any, err error) {
	obj := objDP.obj
	// []string{ BalanceMap *monetary[0] Value }
	var has bool
	if data, has = objDP.getCache(strings.Join(fldPath, utils.NestingSep)); has {
		if data == nil { // field doesn't exist
			err = utils.ErrNotFound
		}
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
	var valIface any
	if valIface, err = objDP.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}
