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
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// UpdateFromCSV will update LoaderData with data received from fileName,
// contained in record and processed with cfgTpl
func newRecord(req utils.DataProvider, data profile, tnt string, cfg *config.CGRConfig, cache *ltcache.Cache) *record {
	return &record{
		data:   data,
		tmp:    &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)},
		req:    req,
		cfg:    cfg.GetDataProvider(),
		cache:  cache,
		tenant: tnt,
	}
}

type record struct {
	tenant string
	data   profile
	tmp    *utils.DataNode
	req    utils.DataProvider
	cfg    utils.DataProvider
	cache  *ltcache.Cache
}

func (r *record) String() string { return r.req.String() }

func (r *record) FieldAsString(path []string) (str string, err error) {
	var val any
	if val, err = r.FieldAsInterface(path); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}

func TenantIDFromOrderedNavigableMap(data *utils.OrderedNavigableMap) *utils.TenantID {
	tnt, _ := data.FieldAsString([]string{utils.Tenant, "0"})
	id, _ := data.FieldAsString([]string{utils.ID, "0"})
	return &utils.TenantID{
		Tenant: tnt,
		ID:     id,
	}
}

func RateIDsFromOrderedNavigableMap(data *utils.OrderedNavigableMap) ([]string, error) {
	val, err := data.FieldAsInterface([]string{utils.RateIDs, "0"})
	if err != nil {
		return nil, fmt.Errorf("cannot find RateIDs in map")
	}
	return utils.IfaceAsStringSlice(val)
}

// FieldAsInterface implements utils.DataProvider
func (ar *record) FieldAsInterface(fldPath []string) (val any, err error) {
	switch fldPath[0] {
	default:
		val, err = ar.data.FieldAsInterface(fldPath)
	case utils.MetaReq:
		if len(fldPath) != 1 {
			val, err = ar.req.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.req
		}
	case utils.MetaTmp:
		if len(fldPath) != 1 {
			val, err = ar.tmp.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.tmp
		}
	case utils.MetaUCH:
		if cacheVal, ok := ar.cache.Get(strings.Join(fldPath[1:], utils.NestingSep)); !ok {
			err = utils.ErrNotFound
		} else {
			val = cacheVal
		}
	case utils.MetaCfg:
		if len(fldPath) != 1 {
			val, err = ar.cfg.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.cfg
		}
	case utils.MetaTenant:
		return ar.tenant, nil
	}
	if err != nil {
		return
	}
	if nmItems, isNMItems := val.([]*utils.DataNode); isNMItems { // special handling of NMItems, take the last value out of it
		el := nmItems[len(nmItems)-1]
		if el.Type == utils.NMDataType {
			val = el.Value.Data
		}
	}
	return
}

// SetFields will populate fields of record out of templates
func (ar *record) SetFields(ctx *context.Context, tmpls []*config.FCTemplate, filterS *engine.FilterS, rndDec int, dftTmz string) (err error) {
	ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	for _, fld := range tmpls {
		if pass, err := filterS.Pass(context.TODO(), ar.tenant,
			fld.Filters, ar); err != nil {
			return err
		} else if !pass {
			continue
		}
		switch fld.Type {
		case utils.MetaNone:
		case utils.MetaRemove:
			if err = ar.Remove(&utils.FullPath{
				PathSlice: fld.GetPathSlice(),
				Path:      fld.Path,
			}); err != nil {
				return
			}
		case utils.MetaRemoveAll:
			ar.RemoveAll(fld.GetPathSlice()[0])
		default:
			var out any
			if out, err = engine.ParseAttribute(ar, fld.Type, fld.Path, fld.Value, rndDec,
				utils.FirstNonEmpty(fld.Timezone, dftTmz), fld.Layout); err != nil {
				if err == utils.ErrNotFound {
					if !fld.Mandatory {
						err = nil
						continue
					}
					err = utils.ErrPrefixNotFound(fld.Tag)
				}
				return
			}
			var fullPath *utils.FullPath
			if fullPath, err = utils.GetFullFieldPath(fld.Path, ar); err != nil {
				return
			} else if fullPath == nil { // no dynamic path
				fullPath = &utils.FullPath{
					PathSlice: slices.Clone(fld.GetPathSlice()), // need to clone so me do not modify the template
					Path:      fld.Path,
				}
			}

			nMItm := &utils.DataLeaf{Data: out, NewBranch: fld.NewBranch, AttributeID: fld.AttributeID}
			switch fld.Type {
			case utils.MetaComposed:
				err = ar.Compose(fullPath, nMItm)
			default:
				err = ar.Set(fullPath, nMItm)
			}
			if err != nil {
				return
			}
		}
		if fld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// RemoveAll deletes all fields at given prefix
func (ar *record) RemoveAll(prefix string) {
	switch prefix {
	default:
		// ar.data = utils.NewOrderedNavigableMap()
	case utils.MetaTmp:
		ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaUCH:
		ar.cache.Clear()
	}
}

// Remove deletes the fields found at path with the given prefix
func (ar *record) Remove(fullPath *utils.FullPath) error {
	switch fullPath.PathSlice[0] {
	default:
		/* ar.data.Remove(&utils.FullPath{
			PathSlice: fullPath.PathSlice,
			Path:      fullPath.Path,
		})*/
	case utils.MetaTmp:
		return ar.tmp.Remove(slices.Clone(fullPath.PathSlice[1:]))
	case utils.MetaUCH:
		ar.cache.Remove(fullPath.Path[5:])
	}
	return nil
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *record) Compose(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	case utils.MetaTmp:
		return ar.tmp.Compose(fullPath.PathSlice[1:], val)
	case utils.MetaUCH:
		path := fullPath.Path[5:]
		var prv any
		if prvI, ok := ar.cache.Get(path); !ok {
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prvI) + utils.IfaceAsString(val.Data)
		}
		ar.cache.Set(path, prv, nil)
		return
	default:
		var valStr string
		if valStr, err = ar.FieldAsString(fullPath.PathSlice); err != nil && err != utils.ErrNotFound {
			return
		}
		return ar.data.Set(fullPath.PathSlice, valStr+utils.IfaceAsString(val.Data), val.NewBranch)
	}
}

// Set implements utils.NMInterface
func (ar *record) Set(fullPath *utils.FullPath, nm *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	default:
		return ar.data.Set(fullPath.PathSlice, nm.Data, nm.NewBranch)
	case utils.MetaTmp:
		_, err = ar.tmp.Set(fullPath.PathSlice[1:], nm.Data)
		return
	case utils.MetaUCH:
		ar.cache.Set(fullPath.Path[5:], nm.Data, nil)
		return
	}
}

type profile interface {
	utils.DataProvider
	Set([]string, any, bool) error
	Merge(any)
	TenantID() string
}

func newProfileFunc(lType string) func() profile {
	switch lType {
	case utils.MetaAttributes:
		return func() profile {
			return new(engine.AttributeProfile)
		}
	case utils.MetaResources:
		return func() profile {
			return new(engine.ResourceProfile)
		}
	case utils.MetaFilters:
		return func() profile {
			return new(engine.Filter)
		}
	case utils.MetaStats:
		return func() profile {
			return new(engine.StatQueueProfile)
		}
	case utils.MetaThresholds:
		return func() profile {
			return new(engine.ThresholdProfile)
		}
	case utils.MetaRoutes:
		return func() profile {
			return new(engine.RouteProfile)
		}
	case utils.MetaChargers:
		return func() profile {
			return new(engine.ChargerProfile)
		}
	case utils.MetaRateProfiles:
		return func() profile {
			return &utils.RateProfile{
				Rates:   make(map[string]*utils.Rate),
				MinCost: utils.NewDecimal(0, 0),
				MaxCost: utils.NewDecimal(0, 0),
			}
		}
	case utils.MetaActionProfiles:
		return func() profile {
			return &engine.ActionProfile{
				Targets: make(map[string]utils.StringSet),
			}
		}
	case utils.MetaAccounts:
		return func() profile {
			return &utils.Account{
				Opts:     make(map[string]any),
				Balances: make(map[string]*utils.Balance),
			}
		}
	case utils.MetaTrends:
		return func() profile {
			return new(engine.TrendProfile)
		}
	case utils.MetaRankings:
		return func() profile {
			return new(engine.RankingProfile)
		}
	default:
		return func() profile { return nil }
	}
}
