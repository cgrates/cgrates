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
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// NewFWVProvider constructs a utils.DataProvider
func NewFWVProvider(record string) (dP utils.DataProvider) {
	dP = &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	return
}

// FWVProvider implements engine.utils.DataProvider so we can pass it to filters
type FWVProvider struct {
	req   string
	cache utils.MapStorage
}

// String is part of engine.utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (fP *FWVProvider) String() string {
	return utils.ToJSON(fP.req)
}

// FieldAsInterface is part of engine.utils.DataProvider interface
func (fP *FWVProvider) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) == 0 {
		return
	}
	fwvIdx := fldPath[0]
	if data, err = fP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	indexes := strings.Split(fwvIdx, "-")
	if len(indexes) != 2 {
		return "", fmt.Errorf("Invalid format for index : %+v ", fldPath)
	}
	startIndex, err := strconv.Atoi(indexes[0])
	if err != nil {
		return nil, err
	}
	if startIndex > len(fP.req) {
		return "", fmt.Errorf("StartIndex : %+v is greater than : %+v", startIndex, len(fP.req))
	}
	finalIndex, err := strconv.Atoi(indexes[1])
	if err != nil {
		return nil, err
	}
	if finalIndex > len(fP.req) {
		return "", fmt.Errorf("FinalIndex : %+v is greater than : %+v", finalIndex, len(fP.req))
	}
	data = fP.req[startIndex:finalIndex]
	fP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of engine.utils.DataProvider interface
func (fP *FWVProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = fP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}
