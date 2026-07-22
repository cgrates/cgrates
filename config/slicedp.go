// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

// NewSliceDP constructs a utils.DataProvider
func NewSliceDP(record []string, indxAls map[string]int) (dP utils.DataProvider) {
	return &SliceDP{
		req:    record,
		cache:  utils.MapStorage{utils.Length: len(record)},
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
	var cfgFieldIdx int
	if cfgFieldIdx, err = cP.getIndex(idx); err != nil {
		return nil, fmt.Errorf("Ignoring record: %v with error : %+v", cP.req, err)
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
	var valIface any
	valIface, err = cP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
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
