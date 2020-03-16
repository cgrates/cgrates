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
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

// NewSliceDP constructs a utils.DataProvider
func NewSliceDP(record []string) (dP utils.DataProvider) {
	dP = &SliceDP{req: record, cache: utils.NavigableMap{}}
	return
}

// SliceDP implements utils.DataProvider so we can pass it to filters
type SliceDP struct {
	req   []string
	cache utils.NavigableMap
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cP *SliceDP) String() string {
	return utils.ToJSON(cP)
}

// FieldAsInterface is part of utils.DataProvider interface
func (cP *SliceDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) == 0 {
		return
	}
	if len(fldPath) != 1 {
		return nil, fmt.Errorf("Invalid fieldPath %+v", fldPath)
	}
	idx := fldPath[0]
	if data, err = cP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	if cfgFieldIdx, err := strconv.Atoi(idx); err != nil {
		return nil, fmt.Errorf("Ignoring record: %v with error : %+v", cP.req, err)
	} else if len(cP.req) <= cfgFieldIdx {
		return nil, utils.ErrNotFound
	} else {
		data = cP.req[cfgFieldIdx]
	}
	cP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of utils.DataProvider interface
func (cP *SliceDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = cP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of utils.DataProvider interface
func (cP *SliceDP) RemoteHost() net.Addr {
	return utils.LocalAddr()
}
