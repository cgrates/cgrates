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
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// UpdateFromCSV will update LoaderData with data received from fileName,
// contained in record and processed with cfgTpl
func newRecord(ctx *context.Context, req utils.DataProvider, tmpls []*config.FCTemplate, tnt string, filterS *engine.FilterS,
	cfg *config.CGRConfig, cache *ltcache.Cache) (_ utils.MapStorage, err error) {
	r := &record{
		data:  make(utils.MapStorage),
		tmp:   make(utils.MapStorage),
		req:   req,
		cfg:   cfg.GetDataProvider(),
		cache: cache,
	}
	if err = r.parseTemplates(ctx, tmpls, tnt, filterS, cfg.GeneralCfg().RoundingDecimals, cfg.GeneralCfg().DefaultTimezone, cfg.GeneralCfg().RSRSep); err != nil {
		return
	}
	return r.data, nil
}

type record struct {
	data  utils.MapStorage
	tmp   utils.MapStorage
	req   utils.DataProvider
	cfg   utils.DataProvider
	cache *ltcache.Cache
	tntID *string
}

func (r *record) parseTemplates(ctx *context.Context, tmpls []*config.FCTemplate, tnt string, filterS *engine.FilterS, rndDec int, dftTmz, rsrSep string) (err error) {
	for _, fld := range tmpls {
		// Make sure filters are matching
		if len(fld.Filters) != 0 {
			if pass, err := filterS.Pass(ctx, tnt,
				fld.Filters, r); err != nil {
				return err
			} else if !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		var out interface{}
		if out, err = engine.ParseAttribute(r, utils.FirstNonEmpty(fld.Type, utils.MetaVariable), utils.DynamicDataPrefix+fld.Path, fld.Value,
			rndDec, utils.FirstNonEmpty(fld.Timezone, dftTmz), fld.Layout, rsrSep); err != nil {
			return
		}
		ps := fld.GetPathSlice()
		if fld.Type == utils.MetaComposed {
			if val, err := r.FieldAsString(ps); err == nil {
				out = utils.IfaceAsString(out) + val
			}
		}
		if err = r.Set(ps, out); err != nil {
			return
		}
	}
	return
}

func (r *record) String() string { return r.req.String() }

func (r *record) FieldAsInterface(path []string) (val interface{}, err error) {
	switch path[0] {
	case utils.MetaUCH:
		cp := strings.Join(path[1:], utils.NestingSep)
		if path[1] == utils.MetaTntID {
			cp = r.tenatID() + strings.TrimPrefix(cp, utils.MetaTntID)
		}
		var ok bool
		if val, ok = r.cache.Get(cp); !ok || val == nil {
			err = utils.ErrNotFound
		}
		return
	case utils.MetaCfg:
		return r.cfg.FieldAsInterface(path[1:])
	case utils.MetaReq:
		return r.req.FieldAsInterface(path[1:])
	case utils.MetaTmp:
		return r.tmp.FieldAsInterface(path[1:])
	default:
		return r.data.FieldAsInterface(path)
	}
}

func (r *record) FieldAsString(path []string) (str string, err error) {
	var val interface{}
	if val, err = r.FieldAsInterface(path); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}

func (r *record) Set(path []string, val interface{}) (err error) {
	switch path[0] {
	case utils.MetaUCH:
		cp := strings.Join(path[1:], utils.NestingSep)
		if path[1] == utils.MetaTntID {
			cp = r.tenatID() + strings.TrimPrefix(cp, utils.MetaTntID)
		}
		r.cache.Set(cp, val, nil)
		return
	case utils.MetaTmp:
		return r.tmp.Set(path[1:], val)
	default:
		return r.data.Set(path, val)
	}
}

func (r *record) tenatID() string {
	if r.tntID == nil {
		r.tntID = utils.StringPointer(TenantIDFromMap(r.data).TenantID())
	}
	return *r.tntID
}

func TenantIDFromMap(data utils.MapStorage) *utils.TenantID {
	return &utils.TenantID{
		Tenant: utils.IfaceAsString(data[utils.Tenant]),
		ID:     utils.IfaceAsString(data[utils.ID]),
	}
}

func RateIDsFromMap(data utils.MapStorage) ([]string, error) {
	val, has := data[utils.RateIDs]
	if !has {
		return nil, fmt.Errorf("cannot find RateIDs in map")
	}
	return utils.IfaceAsStringSlice(val)
}
