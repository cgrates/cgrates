// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

// NewSliceDP constructs a utils.DataProvider
func NewSliceDP(record []string) (dP utils.DataProvider) {
	dP = &SliceDP{req: record, cache: utils.MapStorage{}}
	return
}

// SliceDP implements engine.utils.DataProvider so we can pass it to filters
type SliceDP struct {
	req   []string
	cache utils.MapStorage
}

// String is part of engine.utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cP *SliceDP) String() string {
	return utils.ToJSON(cP)
}

// FieldAsInterface is part of engine.utils.DataProvider interface
func (cP *SliceDP) FieldAsInterface(fldPath []string) (data any, err error) {
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

// FieldAsString is part of engine.utils.DataProvider interface
func (cP *SliceDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface any
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
