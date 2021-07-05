/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package engine

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewExportRequest returns a new EventRequest
func NewExportRequest(inData map[string]utils.MapStorage,
	tnt string,
	filterS *FilterS, oNM map[string]*utils.OrderedNavigableMap) (eeR *ExportRequest) {
	eeR = &ExportRequest{
		inData:  inData,
		filterS: filterS,
		tnt:     tnt,
		ExpData: oNM,
	}
	return
}

// ExportRequest represents data related to one request towards agent
// implements utils.DataProvider so we can pass it to filters
type ExportRequest struct {
	inData  map[string]utils.MapStorage           // request
	ExpData map[string]*utils.OrderedNavigableMap // *exp:OrderNavMp *trl:OrderNavMp *cdr:OrderNavMp
	tnt     string
	filterS *FilterS
}

// String implements utils.DataProvider
func (eeR *ExportRequest) String() string {
	return utils.ToIJSON(eeR)
}

// FieldAsInterface implements utils.DataProvider
func (eeR *ExportRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		var dp utils.DataProvider
		var has bool
		if dp, has = eeR.ExpData[fldPath[0]]; !has {
			if dp, has = eeR.inData[fldPath[0]]; !has {
				return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
			}
		}
		val, err = dp.FieldAsInterface(fldPath[1:])
	case utils.MetaUCH:
		var ok bool
		if val, ok = Cache.Get(utils.CacheUCH, strings.Join(fldPath[1:], utils.NestingSep)); !ok {
			return nil, utils.ErrNotFound
		}
	case utils.MetaTenant:
		return eeR.tnt, nil
	}
	if err != nil {
		return
	}
	if nmItems, isNMItems := val.(*utils.DataNode); isNMItems && nmItems.Type == utils.NMSliceType { // special handling of NMItems, take the last value out of it
		el := nmItems.Slice[len(nmItems.Slice)-1]
		if el.Type == utils.NMDataType {
			val = el.Value.Data
		}
	}
	return
}

// FieldAsString implements utils.DataProvider
func (eeR *ExportRequest) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = eeR.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

//SetFields will populate fields of AgentRequest out of templates
func (eeR *ExportRequest) SetFields(tplFlds []*config.FCTemplate) (err error) {
	for _, tplFld := range tplFlds {
		if pass, err := eeR.filterS.Pass(eeR.tnt,
			tplFld.Filters, eeR); err != nil {
			return err
		} else if !pass {
			continue
		}

		var out interface{}
		out, err = eeR.ParseField(tplFld)
		if err != nil {
			if err == utils.ErrNotFound {
				if !tplFld.Mandatory {
					err = nil
					continue
				}
				err = utils.ErrPrefixNotFound(tplFld.Tag)
			}
			return
		}
		var fullPath *utils.FullPath
		if fullPath, err = utils.GetFullFieldPath(tplFld.Path, eeR); err != nil {
			return
		} else if fullPath == nil { // no dynamic path
			fullPath = &utils.FullPath{
				PathSlice: utils.CloneStringSlice(tplFld.GetPathSlice()), // need to clone so me do not modify the template
				Path:      tplFld.Path,
			}
		}
		nMItm := &utils.DataLeaf{Data: out, NewBranch: tplFld.NewBranch, AttributeID: tplFld.AttributeID}
		switch tplFld.Type {
		case utils.MetaComposed:
			err = eeR.Compose(fullPath, nMItm)
		case utils.MetaGroup: // in case of *group type simply append to valSet
			err = eeR.Append(fullPath, nMItm)
		default:
			err = eeR.SetAsSlice(fullPath, nMItm)
		}
		if err != nil {
			return
		}

		if tplFld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// Set implements utils.NMInterface
func (eeR *ExportRequest) SetAsSlice(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch prfx := fullPath.PathSlice[0]; prfx {
	case utils.MetaUCH:
		return Cache.Set(utils.CacheUCH, fullPath.Path[5:], val.Data, nil, true, utils.NonTransactional)
	case utils.MetaOpts:
		return eeR.inData[utils.MetaOpts].Set(fullPath.PathSlice[1:], val.Data)
	default:
		oNM, has := eeR.ExpData[prfx]
		if !has {
			return fmt.Errorf("unsupported field prefix: <%s> when set field", prfx)
		}
		return oNM.SetAsSlice(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[len(prfx):],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: val}})
	}
}

// ParseField outputs the value based on the template item
func (eeR *ExportRequest) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {
	tmpType := cfgFld.Type
	switch tmpType {
	case utils.MetaMaskedDestination:
		//check if we have destination in the event
		var dst string
		if dst, err = eeR.inData[utils.MetaReq].FieldAsString([]string{utils.Destination}); err != nil {
			err = fmt.Errorf("error <%s> getting destination for %s",
				err, utils.ToJSON(cfgFld))
			return
		}
		if cfgFld.MaskLen != -1 && len(cfgFld.MaskDestID) != 0 &&
			CachedDestHasPrefix(cfgFld.MaskDestID, dst) {
			out = utils.MaskSuffix(dst, cfgFld.MaskLen)
		}
		return
	case utils.MetaFiller:
		cfgFld.Padding = utils.MetaRight
		tmpType = utils.MetaConstant
	case utils.MetaGroup:
		tmpType = utils.MetaVariable
	}
	out, err = ParseAttribute(eeR, tmpType, cfgFld.Path, cfgFld.Value, config.CgrConfig().GeneralCfg().RoundingDecimals, utils.FirstNonEmpty(cfgFld.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone), cfgFld.Layout, config.CgrConfig().GeneralCfg().RSRSep)

	if err != nil &&
		!strings.HasPrefix(err.Error(), "Could not find") {
		return
	}
	if utils.StringTmplType.Has(tmpType) { // format the string additionally with fmtFieldWidth
		out, err = utils.FmtFieldWidth(cfgFld.Tag, out.(string), cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)
	}
	return
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (eeR *ExportRequest) Append(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch prfx := fullPath.PathSlice[0]; prfx {
	case utils.MetaUCH:
		return Cache.Set(utils.CacheUCH, fullPath.Path[5:], val.Data, nil, true, utils.NonTransactional)
	case utils.MetaOpts:
		return eeR.inData[utils.MetaOpts].Set(fullPath.PathSlice[1:], val.Data)
	default:
		oNM, has := eeR.ExpData[prfx]
		if !has {
			return fmt.Errorf("unsupported field prefix: <%s> when set field", prfx)
		}
		return oNM.Append(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[len(prfx):],
		}, val)
	}
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (eeR *ExportRequest) Compose(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch prfx := fullPath.PathSlice[0]; prfx {
	case utils.MetaUCH:
		path := fullPath.Path[5:]
		var prv interface{}
		if prvI, ok := Cache.Get(utils.CacheUCH, path); !ok {
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prvI) + utils.IfaceAsString(val.Data)
		}
		return Cache.Set(utils.CacheUCH, path, prv, nil, true, utils.NonTransactional)
	case utils.MetaOpts:
		var prv interface{}
		if prv, err = eeR.inData[utils.MetaOpts].FieldAsInterface(fullPath.PathSlice[1:]); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prv) + utils.IfaceAsString(val.Data)
		}
		return eeR.inData[utils.MetaOpts].Set(fullPath.PathSlice[1:], prv)
	default:
		oNM, has := eeR.ExpData[prfx]
		if !has {
			return fmt.Errorf("unsupported field prefix: <%s> when set field", prfx)
		}
		return oNM.Compose(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[len(prfx):],
		}, val)
	}
}
