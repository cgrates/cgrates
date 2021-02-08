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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type LoaderData map[string]interface{}

func (ld LoaderData) TenantID() string {
	return utils.ConcatenatedKey(utils.IfaceAsString(ld[utils.Tenant]),
		utils.IfaceAsString(ld[utils.ID]))
}

func (ld LoaderData) GetRateIDs() ([]string, error) {
	if _, has := ld[utils.RateIDs]; !has {
		return nil, fmt.Errorf("cannot find RateIDs in <%+v>", ld)
	}
	if rateIDs := ld[utils.RateIDs].(string); len(rateIDs) != 0 {
		return strings.Split(rateIDs, utils.InfieldSep), nil
	}
	return []string{}, nil
}

// UpdateFromCSV will update LoaderData with data received from fileName,
// contained in record and processed with cfgTpl
func (ld LoaderData) UpdateFromCSV(fileName string, record []string,
	cfgTpl []*config.FCTemplate, tnt config.RSRParsers, filterS *engine.FilterS) (err error) {
	csvProvider := newCsvProvider(record, fileName)
	tenant, err := tnt.ParseValue("")
	if err != nil {
		return err
	}
	for _, cfgFld := range cfgTpl {
		// Make sure filters are matching
		if len(cfgFld.Filters) != 0 {
			if pass, err := filterS.Pass(tenant,
				cfgFld.Filters, csvProvider); err != nil {
				return err
			} else if !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		out, err := cfgFld.Value.ParseDataProvider(csvProvider)
		if err != nil {
			return err
		}
		switch cfgFld.Type {
		case utils.MetaComposed:
			if _, has := ld[cfgFld.Path]; !has {
				ld[cfgFld.Path] = out
			} else if valOrig, canCast := ld[cfgFld.Path].(string); canCast {
				valOrig += out
				ld[cfgFld.Path] = valOrig
			}
		case utils.MetaVariable:
			ld[cfgFld.Path] = out
		case utils.MetaString:
			if _, has := ld[cfgFld.Path]; !has {
				ld[cfgFld.Path] = out
			}
		}
	}
	return
}

// newCsvProvider constructs a DataProvider
func newCsvProvider(record []string, fileName string) (dP utils.DataProvider) {
	return &csvProvider{
		req:      record,
		fileName: fileName,
		cache:    utils.MapStorage{},
		cfg:      config.CgrConfig().GetDataProvider(),
	}
}

// csvProvider implements utils.DataProvider so we can pass it to filters
type csvProvider struct {
	req      []string
	fileName string
	cache    utils.MapStorage
	cfg      utils.DataProvider
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cP *csvProvider) String() string {
	return utils.ToJSON(cP)
}

// FieldAsInterface is part of utils.DataProvider interface
func (cP *csvProvider) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if data, err = cP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err

	switch {
	case strings.HasPrefix(fldPath[0], utils.MetaFile+utils.FilterValStart):
		fileName := strings.TrimPrefix(fldPath[0], utils.MetaFile+utils.FilterValStart)
		hasSelEnd := false
		for _, val := range fldPath[1:] {
			if hasSelEnd = strings.HasSuffix(val, utils.FilterValEnd); hasSelEnd {
				fileName = fileName + utils.NestingSep + val[:len(val)-1]
				break
			}
			fileName = fileName + utils.NestingSep + val
		}
		if !hasSelEnd {
			return nil, fmt.Errorf("filter rule <%s> needs to end in )", fldPath)
		}
		if cP.fileName != fileName {
			cP.cache.Set(fldPath, nil)
			return
		}
	case fldPath[0] == utils.MetaReq:
	case fldPath[0] == utils.MetaCfg:
		data, err = cP.cfg.FieldAsInterface(fldPath[1:])
		if err != nil {
			return
		}
		cP.cache.Set(fldPath, data)
		return
	default:
		return nil, fmt.Errorf("invalid prefix for : %s", fldPath)
	}
	var cfgFieldIdx int
	if cfgFieldIdx, err = strconv.Atoi(fldPath[len(fldPath)-1]); err != nil || len(cP.req) <= cfgFieldIdx {
		return nil, fmt.Errorf("Ignoring record: %q with error : %+v", cP.req, err)
	}
	data = cP.req[cfgFieldIdx]

	cP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of utils.DataProvider interface
func (cP *csvProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = cP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of utils.DataProvider interface
func (cP *csvProvider) RemoteHost() net.Addr {
	return utils.LocalAddr()
}
