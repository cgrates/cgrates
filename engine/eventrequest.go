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
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewEventRequest returns a new EventRequest
func NewEventRequest(req utils.DataProvider, dc, opts utils.MapStorage,
	tntTpl config.RSRParsers, dfltTenant, timezone string,
	filterS *FilterS, oNM map[string]*utils.OrderedNavigableMap) (eeR *EventRequest) {
	eeR = &EventRequest{
		req:      req,
		Tenant:   dfltTenant,
		Timezone: timezone,
		filterS:  filterS,
		dc:       dc,
		opts:     opts,
		OrdNavMP: oNM,
		Cfg:      config.CgrConfig().GetDataProvider(),
	}
	if tnt, err := tntTpl.ParseDataProvider(eeR); err == nil && tnt != utils.EmptyString {
		eeR.Tenant = tnt
	}
	return
}

// EventRequest represents data related to one request towards agent
// implements utils.DataProvider so we can pass it to filters
type EventRequest struct {
	req      utils.DataProvider // request
	Tenant   string
	Timezone string
	OrdNavMP map[string]*utils.OrderedNavigableMap // *exp:OrderNavMp *trl:OrderNavMp *cdr:OrderNavMp
	dc       utils.MapStorage
	opts     utils.MapStorage
	Cfg      utils.DataProvider

	filterS *FilterS
}

// String implements utils.DataProvider
func (eeR *EventRequest) String() string {
	return utils.ToIJSON(eeR)
}

// RemoteHost implements utils.DataProvider
func (eeR *EventRequest) RemoteHost() net.Addr {
	return eeR.req.RemoteHost()
}

// FieldAsInterface implements utils.DataProvider
func (eeR *EventRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		dp, has := eeR.OrdNavMP[fldPath[0]]
		if !has {
			return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
		}
		val, err = dp.FieldAsInterface(fldPath[1:])
	case utils.MetaReq:
		val, err = eeR.req.FieldAsInterface(fldPath[1:])
	case utils.MetaUCH:
		var ok bool
		if val, ok = Cache.Get(utils.CacheUCH, strings.Join(fldPath[1:], utils.NestingSep)); !ok {
			return nil, utils.ErrNotFound
		}
	case utils.MetaDC:
		val, err = eeR.dc.FieldAsInterface(fldPath[1:])
	case utils.MetaOpts:
		val, err = eeR.opts.FieldAsInterface(fldPath[1:])
	case utils.MetaCfg:
		val, err = eeR.Cfg.FieldAsInterface(fldPath[1:])
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
func (eeR *EventRequest) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = eeR.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

//SetFields will populate fields of AgentRequest out of templates
func (eeR *EventRequest) SetFields(tplFlds []*config.FCTemplate) (err error) {
	for _, tplFld := range tplFlds {
		if pass, err := eeR.filterS.Pass(eeR.Tenant,
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
		var itmPath []string
		if fullPath, err = utils.GetFullFieldPath(tplFld.Path, eeR); err != nil {
			return
		} else if fullPath == nil { // no dynamic path
			fullPath = &utils.FullPath{
				PathItems: utils.CloneStringSlice(tplFld.GetPathItems()), // need to clone so me do not modify the template
				Path:      tplFld.Path,
			}
			itmPath = tplFld.GetPathSlice()[1:]
		} else {
			itmPath = fullPath.PathItems
		}
		nMItm := &utils.DataLeaf{Data: out, Path: itmPath, NewBranch: tplFld.NewBranch, AttributeID: tplFld.AttributeID}
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
func (eeR *EventRequest) SetAsSlice(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch prfx := fullPath.PathItems[0]; prfx {
	case utils.MetaUCH:
		return Cache.Set(utils.CacheUCH, fullPath.Path[5:], val.Data, nil, true, utils.NonTransactional)
	case utils.MetaOpts:
		return eeR.opts.Set(fullPath.PathItems[1:], val.Data)
	default:
		oNM, has := eeR.OrdNavMP[prfx]
		if !has {
			return fmt.Errorf("unsupported field prefix: <%s> when set field", prfx)
		}
		return oNM.SetAsSlice(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[len(prfx):],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: val}})
	}
}

// ParseField outputs the value based on the template item
func (eeR *EventRequest) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {
	var isString bool
	switch cfgFld.Type {
	default:
		return utils.EmptyString, fmt.Errorf("unsupported type: <%s>", cfgFld.Type)
	case utils.MetaNone:
		return
	case utils.MetaFiller:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		cfgFld.Padding = utils.MetaRight
		isString = true
	case utils.MetaConstant:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		isString = true
	case utils.MetaRemoteHost:
		out = eeR.RemoteHost().String()
		isString = true
	case utils.MetaVariable, utils.MetaComposed, utils.MetaGroup:
		out, err = cfgFld.Value.ParseDataProvider(eeR)
		isString = true
	case utils.MetaUsageDifference:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaUsageDifference)
		}
		var strVal1 string
		if strVal1, err = cfgFld.Value[0].ParseDataProvider(eeR); err != nil {
			return
		}
		var strVal2 string
		if strVal2, err = cfgFld.Value[1].ParseDataProvider(eeR); err != nil {
			return
		}
		var tEnd time.Time
		if tEnd, err = utils.ParseTimeDetectLayout(strVal1, eeR.Timezone); err != nil {
			return
		}
		var tStart time.Time
		if tStart, err = utils.ParseTimeDetectLayout(strVal2, eeR.Timezone); err != nil {
			return
		}
		out = tEnd.Sub(tStart).String()
		isString = true
	case utils.MetaCCUsage:
		if len(cfgFld.Value) != 3 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaCCUsage)
		}
		var strVal1 string
		if strVal1, err = cfgFld.Value[0].ParseDataProvider(eeR); err != nil {
			return
		}
		var reqNr int64
		if reqNr, err = strconv.ParseInt(strVal1, 10, 64); err != nil {
			err = fmt.Errorf("invalid requestNumber <%s> to %s",
				strVal1, utils.MetaCCUsage)
			return
		}
		var strVal2 string
		if strVal2, err = cfgFld.Value[1].ParseDataProvider(eeR); err != nil {
			return
		}
		var usedCCTime time.Duration
		if usedCCTime, err = utils.ParseDurationWithNanosecs(strVal2); err != nil {
			err = fmt.Errorf("invalid usedCCTime <%s> to %s",
				strVal2, utils.MetaCCUsage)
			return
		}
		var strVal3 string
		if strVal3, err = cfgFld.Value[2].ParseDataProvider(eeR); err != nil {
			return
		}
		var debitItvl time.Duration
		if debitItvl, err = utils.ParseDurationWithNanosecs(strVal3); err != nil {
			err = fmt.Errorf("invalid debitInterval <%s> to %s",
				strVal3, utils.MetaCCUsage)
			return
		}

		if reqNr--; reqNr < 0 { // terminate will be ignored (init request should always be 0)
			reqNr = 0
		}
		return usedCCTime + time.Duration(debitItvl.Nanoseconds()*reqNr), nil
	case utils.MetaSum:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(eeR); err != nil {
			return
		}
		out, err = utils.Sum(iFaceVals...)
	case utils.MetaDifference:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(eeR); err != nil {
			return
		}
		out, err = utils.Difference(iFaceVals...)
	case utils.MetaMultiply:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(eeR); err != nil {
			return
		}
		out, err = utils.Multiply(iFaceVals...)
	case utils.MetaDivide:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(eeR); err != nil {
			return
		}
		out, err = utils.Divide(iFaceVals...)
	case utils.MetaValueExponent:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaValueExponent)
		}
		var strVal1 string
		if strVal1, err = cfgFld.Value[0].ParseDataProvider(eeR); err != nil {
			return
		}
		var val float64
		if val, err = strconv.ParseFloat(strVal1, 64); err != nil {
			err = fmt.Errorf("invalid value <%s> to %s",
				strVal1, utils.MetaValueExponent)
			return
		}
		var strVal2 string
		if strVal2, err = cfgFld.Value[1].ParseDataProvider(eeR); err != nil {
			return
		}
		var exp int
		if exp, err = strconv.Atoi(strVal2); err != nil {
			return
		}
		out = strconv.FormatFloat(utils.Round(val*math.Pow10(exp),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.MetaRoundingMiddle), 'f', -1, 64)
	case utils.MetaUnixTimestamp:
		var val string
		if val, err = cfgFld.Value.ParseDataProvider(eeR); err != nil {
			return
		}
		var t1 time.Time
		if t1, err = utils.ParseTimeDetectLayout(val, cfgFld.Timezone); err != nil {
			return
		}
		out = strconv.Itoa(int(t1.Unix()))
	case utils.MetaMaskedDestination:
		//check if we have destination in the event
		var dst string
		if dst, err = eeR.req.FieldAsString([]string{utils.Destination}); err != nil {
			err = fmt.Errorf("error <%s> getting destination for %s",
				err, utils.ToJSON(cfgFld))
			return
		}
		if cfgFld.MaskLen != -1 && len(cfgFld.MaskDestID) != 0 &&
			CachedDestHasPrefix(cfgFld.MaskDestID, dst) {
			out = utils.MaskSuffix(dst, cfgFld.MaskLen)
		}

	}

	if err != nil &&
		!strings.HasPrefix(err.Error(), "Could not find") {
		return
	}
	if isString { // format the string additionally with fmtFieldWidth
		out, err = utils.FmtFieldWidth(cfgFld.Tag, out.(string), cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)
	}
	return
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (eeR *EventRequest) Append(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch prfx := fullPath.PathItems[0]; prfx {
	case utils.MetaUCH:
		return Cache.Set(utils.CacheUCH, fullPath.Path[5:], val.Data, nil, true, utils.NonTransactional)
	case utils.MetaOpts:
		return eeR.opts.Set(fullPath.PathItems[1:], val.Data)
	default:
		oNM, has := eeR.OrdNavMP[prfx]
		if !has {
			return fmt.Errorf("unsupported field prefix: <%s> when set field", prfx)
		}
		return oNM.Append(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[len(prfx):],
		}, val)
	}
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (eeR *EventRequest) Compose(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch prfx := fullPath.PathItems[0]; prfx {
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
		if prv, err = eeR.opts.FieldAsInterface(fullPath.PathItems[1:]); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prv) + utils.IfaceAsString(val.Data)
		}
		return eeR.opts.Set(fullPath.PathItems[1:], prv)
	default:
		oNM, has := eeR.OrdNavMP[prfx]
		if !has {
			return fmt.Errorf("unsupported field prefix: <%s> when set field", prfx)
		}
		return oNM.Compose(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[len(prfx):],
		}, val)
	}
}
