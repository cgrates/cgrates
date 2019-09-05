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
	"net"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// DataProvider is a data source from multiple formats
type DataProvider interface {
	String() string // printable version of data
	FieldAsInterface(fldPath []string) (interface{}, error)
	FieldAsString(fldPath []string) (string, error)
	AsNavigableMap([]*FCTemplate) (*NavigableMap, error)
	RemoteHost() net.Addr
}

func GetDynamicInterface(dnVal string, dP DataProvider) (interface{}, error) {
	if strings.HasPrefix(dnVal, utils.DynamicDataPrefix) {
		dnVal = strings.TrimPrefix(dnVal, utils.DynamicDataPrefix)
		return dP.FieldAsInterface(strings.Split(dnVal, utils.NestingSep))
	}
	return utils.StringToInterface(dnVal), nil
}

func GetDynamicString(dnVal string, dP DataProvider) (string, error) {
	if strings.HasPrefix(dnVal, utils.DynamicDataPrefix) {
		dnVal = strings.TrimPrefix(dnVal, utils.DynamicDataPrefix)
		return dP.FieldAsString(strings.Split(dnVal, utils.NestingSep))
	}
	return dnVal, nil
}

//NewObjectDP constructs a DataProvider
func NewObjectDP(obj interface{}) (dP DataProvider) {
	dP = &ObjectDP{obj: obj, cache: NewNavigableMap(nil)}
	return
}

type ObjectDP struct {
	obj   interface{}
	cache *NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (objDP *ObjectDP) String() string {
	return utils.ToJSON(objDP.obj)
}

// FieldAsInterface is part of engine.DataProvider interface
func (objDP *ObjectDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	// []string{ BalanceMap *monetary[0] Value }
	if data, err = objDP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	// for _, fld := range fldPath {

	// 	//process each field
	// }
	objDP.cache.Set(fldPath, data, false, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (objDP *ObjectDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = objDP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// AsNavigableMap is part of engine.DataProvider interface
func (objDP *ObjectDP) AsNavigableMap([]*FCTemplate) (
	nm *NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// RemoteHost is part of engine.DataProvider interface
func (objDP *ObjectDP) RemoteHost() net.Addr {
	return utils.LocalAddr()
}
