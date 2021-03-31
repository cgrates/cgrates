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
func NewSliceDP(record []string, indxAls map[string]int) (dP utils.DataProvider) {
	return &SliceDP{
		req:    record,
		cache:  utils.MapStorage{},
		idxAls: indxAls,
	}
}

// SliceDP implements engine.utils.DataProvider so we can pass it to filters
type SliceDP struct {
	req    []string
	cache  utils.MapStorage
	idxAls map[string]int // aliases for indexes
}

// String is part of engine.utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cP *SliceDP) String() string {
	return utils.ToJSON(cP.req)
}

// FieldAsInterface is part of engine.utils.DataProvider interface
func (cP *SliceDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) == 0 {
		return
	}
	if len(fldPath) != 1 {
		return nil, fmt.Errorf("Invalid fieldPath %+v ", fldPath)
	}
	idx := fldPath[0]
	if data, err = cP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	var cfgFieldIdx int
	if cfgFieldIdx, err = cP.getIndex(idx); err != nil {
		return nil, fmt.Errorf("Ignoring record: %v with error : %+v ", cP.req, err)
	}
	if len(cP.req) <= cfgFieldIdx {
		return nil, utils.ErrNotFound
	}
	data = cP.req[cfgFieldIdx]
	cP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of engine.utils.DataProvider interface
func (cP *SliceDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = cP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of engine.utils.DataProvider interface
func (cP *SliceDP) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

// getIndex returns the index from index alias map or if not found try to convert it to int
func (cP *SliceDP) getIndex(idx string) (fieldIdx int, err error) {
	if cP.idxAls != nil {
		var has bool
		if fieldIdx, has = cP.idxAls[idx]; has {
			return
		}
	}
	return strconv.Atoi(idx)
}
