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
	cfg *config.CGRConfig, cache *ltcache.Cache) (_ *utils.OrderedNavigableMap, err error) {
	r := &record{
		data:   utils.NewOrderedNavigableMap(),
		tmp:    &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)},
		req:    req,
		cfg:    cfg.GetDataProvider(),
		cache:  cache,
		tenant: tnt,
	}
	if err = r.SetFields(ctx, tmpls, filterS, cfg.GeneralCfg().RoundingDecimals, cfg.GeneralCfg().DefaultTimezone, cfg.GeneralCfg().RSRSep); err != nil {
		return
	}
	return r.data, nil
}

type record struct {
	tenant string
	data   *utils.OrderedNavigableMap
	tmp    *utils.DataNode
	req    utils.DataProvider
	cfg    utils.DataProvider
	cache  *ltcache.Cache
}

func (r *record) String() string { return r.req.String() }

func (r *record) FieldAsString(path []string) (str string, err error) {
	var val interface{}
	if val, err = r.FieldAsInterface(path); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}

func TenantIDFromDataProvider(data utils.DataProvider) *utils.TenantID {
	tnt, _ := data.FieldAsString([]string{utils.Tenant})
	id, _ := data.FieldAsString([]string{utils.ID})
	return &utils.TenantID{
		Tenant: tnt,
		ID:     id,
	}
}

func RateIDsFromDataProvider(data utils.DataProvider) ([]string, error) {
	val, err := data.FieldAsInterface([]string{utils.RateIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot find RateIDs in map")
	}
	return utils.IfaceAsStringSlice(val)
}

// FieldAsInterface implements utils.DataProvider
func (ar *record) FieldAsInterface(fldPath []string) (val interface{}, err error) {
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

//SetFields will populate fields of record out of templates

func (ar *record) SetFields(ctx *context.Context, tmpls []*config.FCTemplate, filterS *engine.FilterS, rndDec int, dftTmz, rsrSep string) (err error) {
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
			if err = ar.RemoveAll(fld.GetPathSlice()[0]); err != nil {
				return
			}
		default:
			var out interface{}
			if out, err = engine.ParseAttribute(ar, fld.Type, fld.Path, fld.Value, rndDec,
				utils.FirstNonEmpty(fld.Timezone, dftTmz), fld.Layout, rsrSep); err != nil {
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
					PathSlice: utils.CloneStringSlice(fld.GetPathSlice()), // need to clone so me do not modify the template
					Path:      fld.Path,
				}
			}

			nMItm := &utils.DataLeaf{Data: out, NewBranch: fld.NewBranch, AttributeID: fld.AttributeID}
			switch fld.Type {
			case utils.MetaComposed:
				err = ar.Compose(fullPath, nMItm)
			case utils.MetaGroup: // in case of *group type simply append to valSet
				err = ar.Append(fullPath, nMItm)
			default:
				err = ar.SetAsSlice(fullPath, nMItm)
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

// Set implements utils.NMInterface
func (ar *record) SetAsSlice(fullPath *utils.FullPath, nm *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	default:
		return ar.data.SetAsSlice(fullPath, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaTmp:
		_, err = ar.tmp.Set(fullPath.PathSlice[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaUCH:
		ar.cache.Set(fullPath.Path[5:], nm.Data, nil)
		return
	}
}

// RemoveAll deletes all fields at given prefix
func (ar *record) RemoveAll(prefix string) error {
	switch prefix {
	default:
		ar.data = utils.NewOrderedNavigableMap()
	case utils.MetaTmp:
		ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaUCH:
		ar.cache.Clear()
	}
	return nil
}

// Remove deletes the fields found at path with the given prefix
func (ar *record) Remove(fullPath *utils.FullPath) error {
	switch fullPath.PathSlice[0] {
	default:
		return ar.data.Remove(&utils.FullPath{
			PathSlice: fullPath.PathSlice,
			Path:      fullPath.Path,
		})
	case utils.MetaTmp:
		return ar.tmp.Remove(utils.CloneStringSlice(fullPath.PathSlice[1:]))
	case utils.MetaUCH:
		ar.cache.Remove(fullPath.Path[5:])
		return nil
	}
}

// ParseField outputs the value based on the template item
func (ar *record) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {

	if err != nil &&
		!strings.HasPrefix(err.Error(), "Could not find") {
		return
	}
	if utils.StringTmplType.Has(cfgFld.Type) { // format the string additionally with fmtFieldWidth
		out, err = utils.FmtFieldWidth(cfgFld.Tag, out.(string), cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)
	}
	return
}

// Append sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *record) Append(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	case utils.MetaTmp:
		_, err = ar.tmp.Append(fullPath.PathSlice[1:], val)
		return
	case utils.MetaUCH:
		ar.cache.Set(fullPath.Path[5:], val.Data, nil)
		return
	default:
		return ar.data.Append(fullPath, val)
	}
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *record) Compose(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	case utils.MetaTmp:
		return ar.tmp.Compose(fullPath.PathSlice[1:], val)
	case utils.MetaUCH:
		path := fullPath.Path[5:]
		var prv interface{}
		if prvI, ok := ar.cache.Get(path); !ok {
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prvI) + utils.IfaceAsString(val.Data)
		}
		ar.cache.Set(path, prv, nil)
		return
	default:
		return ar.data.Compose(fullPath, val)
	}
}

type profile interface {
	Set([]string, interface{}, bool, string) error
}

func prepareData(prf profile, lData []*utils.OrderedNavigableMap, rsrSep string) (err error) {
	for _, mp := range lData {
		newRow := true
		for el := mp.GetFirstElement(); el != nil; el = el.Next() {
			path := el.Value
			nmIt, _ := mp.Field(path)
			if nmIt == nil {
				continue // all attributes, not writable to diameter packet
			}
			// path = path[:len(path)-1] // remove the last index
			if err = prf.Set(path, nmIt.Data, nmIt.NewBranch || newRow, rsrSep); err != nil {
				return
			}
			newRow = false
		}
	}
	return
}
