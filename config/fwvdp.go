// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
func (fP *FWVProvider) FieldAsInterface(fldPath []string) (data any, err error) {
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
		return "", fmt.Errorf("Invalid format for index : %+v", fldPath)
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
	var valIface any
	valIface, err = fP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}
