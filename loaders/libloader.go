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

package loaders

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type LoaderData map[string]interface{}

func (ld LoaderData) TenantID() string {
	tnt := ld[utils.Tenant].(string)
	prflID := ld[utils.ID].(string)
	return utils.ConcatenatedKey(tnt, prflID)
}

// UpdateFromCSV will update LoaderData with data received from fileName,
// contained in record and processed with cfgTpl
func (ld LoaderData) UpdateFromCSV(fileName string, record []string,
	cfgTpl []*config.FCTemplate, tnt config.RSRParsers, filterS *engine.FilterS) (err error) {
	csvProvider := newCsvProvider(record, fileName)
	for _, cfgFld := range cfgTpl {
		// Make sure filters are matching
		if len(cfgFld.Filters) != 0 {
			tenant, err := tnt.ParseValue("")
			if err != nil {
				return err
			}
			if pass, err := filterS.Pass(tenant,
				cfgFld.Filters, csvProvider); err != nil || !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		out, err := cfgFld.Value.ParseDataProvider(csvProvider, utils.InInFieldSep)
		if err != nil {
			return err
		}
		switch cfgFld.Type {
		case utils.META_COMPOSED:
			if _, has := ld[cfgFld.FieldId]; !has {
				ld[cfgFld.FieldId] = out
			} else if valOrig, canCast := ld[cfgFld.FieldId].(string); canCast {
				valOrig += out
				ld[cfgFld.FieldId] = valOrig
			}
		case utils.MetaString:
			if _, has := ld[cfgFld.FieldId]; !has {
				ld[cfgFld.FieldId] = out
			}
		}
	}
	return
}

// newCsvProvider constructs a DataProvider
func newCsvProvider(record []string, fileName string) (dP config.DataProvider) {
	dP = &csvProvider{req: record, fileName: fileName, cache: config.NewNavigableMap(nil)}
	return
}

// csvProvider implements engine.DataProvider so we can pass it to filters
type csvProvider struct {
	req      []string
	fileName string
	cache    *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cP *csvProvider) String() string {
	return utils.ToJSON(cP)
}

// FieldAsInterface is part of engine.DataProvider interface
func (cP *csvProvider) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if data, err = cP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	idx := fldPath[0]
	var fileName string
	if len(fldPath) == 2 {
		fileName = fldPath[0]
		idx = fldPath[1]
	}
	if fileName != "" && cP.fileName != fileName {
		cP.cache.Set(fldPath, nil, false, false)
		return
	}
	if cfgFieldIdx, err := strconv.Atoi(idx); err != nil || len(cP.req) <= cfgFieldIdx {
		return nil, fmt.Errorf("Ignoring record: %v with error : %+v", cP.req, err)
	} else {
		data = cP.req[cfgFieldIdx]
	}

	cP.cache.Set(fldPath, data, false, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (cP *csvProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = cP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (cP *csvProvider) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// RemoteHost is part of engine.DataProvider interface
func (cP *csvProvider) RemoteHost() net.Addr {
	return utils.LocalAddr()
}
