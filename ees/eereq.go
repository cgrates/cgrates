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

package ees

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewEventExporterRequest returns a new EventExporterRequest
func NewEventExporterRequest(req utils.DataProvider, dc utils.MapStorage,
	tnt, timezone string, filterS *engine.FilterS) (eeR *EventExporterRequest) {
	eeR = &EventExporterRequest{
		req:     req,
		tmz:     timezone,
		tnt:     tnt,
		filterS: filterS,
		cnt:     utils.NewOrderedNavigableMap(),
		hdr:     utils.NewOrderedNavigableMap(),
		trl:     utils.NewOrderedNavigableMap(),
		dc:      dc,
	}
	eeR.dynamicProvider = utils.NewDynamicDataProvider(eeR)
	return
}

// EventExporterRequest represents data related to one request towards agent
// implements utils.DataProvider so we can pass it to filters
type EventExporterRequest struct {
	req  utils.DataProvider // request
	eeDP utils.DataProvider // eventExporter DataProvider
	tnt  string
	tmz  string
	cnt  *utils.OrderedNavigableMap // Used in reply to access the request that was send
	hdr  *utils.OrderedNavigableMap // Used in reply to access the request that was send
	trl  *utils.OrderedNavigableMap // Used in reply to access the request that was send
	dc   utils.MapStorage

	filterS         *engine.FilterS
	dynamicProvider *utils.DynamicDataProvider
}

// String implements utils.DataProvider
func (eeR *EventExporterRequest) String() string {
	return utils.ToIJSON(eeR)
}

// RemoteHost implements utils.DataProvider
func (eeR *EventExporterRequest) RemoteHost() net.Addr {
	return eeR.req.RemoteHost()
}

// FieldAsInterface implements utils.DataProvider
func (eeR *EventExporterRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaReq:
		val, err = eeR.req.FieldAsInterface(fldPath[1:])
	case utils.MetaCache:
		if cacheVal, ok := engine.Cache.Get(utils.CacheUCH, strings.Join(fldPath[1:], utils.NestingSep)); !ok {
			err = utils.ErrNotFound
		} else {
			val = cacheVal
		}
	case utils.MetaDC:
		val, err = eeR.dc.FieldAsInterface(fldPath[1:])
	}
	if err != nil {
		return
	}
	if nmItems, isNMItems := val.(*utils.NMSlice); isNMItems { // special handling of NMItems, take the last value out of it
		val = (*nmItems)[len(*nmItems)-1].Interface()
	}
	return
}

// Field implements utils.NMInterface
func (eeR *EventExporterRequest) Field(fldPath utils.PathItems) (val utils.NMInterface, err error) {
	switch fldPath[0].Field {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaExp:
		val, err = eeR.cnt.Field(fldPath[1:])
	case utils.MetaHdr:
		val, err = eeR.hdr.Field(fldPath[1:])
	case utils.MetaTrl:
		val, err = eeR.trl.Field(fldPath[1:])
	}
	return
}

// FieldAsString implements utils.DataProvider
func (eeR *EventExporterRequest) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = eeR.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

//SetFields will populate fields of AgentRequest out of templates
func (eeR *EventExporterRequest) SetFields(tplFlds []*config.FCTemplate) (err error) {
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
		var itmPath []string
		if fullPath, err = eeR.dynamicProvider.GetFullFieldPath(tplFld.Path); err != nil {
			return
		} else if fullPath == nil { // no dynamic path
			fullPath = &utils.FullPath{
				PathItems: tplFld.GetPathItems().Clone(), // need to clone so me do not modify the template
				Path:      tplFld.Path,
			}
			itmPath = tplFld.GetPathSlice()[1:]
		} else {
			itmPath = fullPath.PathItems.Slice()[1:]
		}

		nMItm := &config.NMItem{Data: out, Path: itmPath, Config: tplFld}
		switch tplFld.Type {
		case utils.META_COMPOSED:
			err = utils.ComposeNavMapVal(eeR, fullPath, nMItm)
		default:
			_, err = eeR.Set(fullPath, &utils.NMSlice{nMItm})
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
func (eeR *EventExporterRequest) Set(fullPath *utils.FullPath, nm utils.NMInterface) (added bool, err error) {
	switch fullPath.PathItems[0].Field {
	default:
		return false, fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathItems[0].Field)
	case utils.MetaExp:
		return eeR.cnt.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[4:],
		}, nm)
	case utils.MetaHdr:
		return eeR.hdr.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[4:],
		}, nm)
	case utils.MetaTrl:
		return eeR.trl.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[4:],
		}, nm)
	case utils.MetaCache:
		err = engine.Cache.Set(utils.CacheUCH, fullPath.Path[7:], nm, nil, true, utils.NonTransactional)
	}
	return false, err
}

// ParseField outputs the value based on the template item
func (eeR *EventExporterRequest) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {
	var isString bool
	switch cfgFld.Type {
	default:
		return utils.EmptyString, fmt.Errorf("unsupported type: <%s>", cfgFld.Type)
	case utils.META_NONE:
		return
	case utils.META_FILLER:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		cfgFld.Padding = utils.MetaRight
		isString = true
	case utils.META_CONSTANT:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		isString = true
	case utils.MetaRemoteHost:
		out = eeR.RemoteHost().String()
		isString = true
	case utils.MetaVariable, utils.META_COMPOSED, utils.MetaGroup:
		out, err = cfgFld.Value.ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
		isString = true
	case utils.META_USAGE_DIFFERENCE:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.META_USAGE_DIFFERENCE)
		}
		strVal1, err := cfgFld.Value[0].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
		if err != nil {
			return "", err
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
		if err != nil {
			return "", err
		}
		tEnd, err := utils.ParseTimeDetectLayout(strVal1, eeR.tmz)
		if err != nil {
			return "", err
		}
		tStart, err := utils.ParseTimeDetectLayout(strVal2, eeR.tmz)
		if err != nil {
			return "", err
		}
		out = tEnd.Sub(tStart).String()
		isString = true
	case utils.MetaCCUsage:
		if len(cfgFld.Value) != 3 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaCCUsage)
		}
		strVal1, err := cfgFld.Value[0].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep) // ReqNr
		if err != nil {
			return "", err
		}
		reqNr, err := strconv.ParseInt(strVal1, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid requestNumber <%s> to %s",
				strVal1, utils.MetaCCUsage)
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep) // TotalUsage
		if err != nil {
			return "", err
		}
		usedCCTime, err := utils.ParseDurationWithNanosecs(strVal2)
		if err != nil {
			return "", fmt.Errorf("invalid usedCCTime <%s> to %s",
				strVal2, utils.MetaCCUsage)
		}
		strVal3, err := cfgFld.Value[2].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep) // DebitInterval
		if err != nil {
			return "", err
		}
		debitItvl, err := utils.ParseDurationWithNanosecs(strVal3)
		if err != nil {
			return "", fmt.Errorf("invalid debitInterval <%s> to %s",
				strVal3, utils.MetaCCUsage)
		}
		mltpl := reqNr - 1 // terminate will be ignored (init request should always be 0)
		if mltpl < 0 {
			mltpl = 0
		}
		return usedCCTime + time.Duration(debitItvl.Nanoseconds()*mltpl), nil
	case utils.MetaSum:
		iFaceVals := make([]interface{}, len(cfgFld.Value))
		for i, val := range cfgFld.Value {
			strVal, err := val.ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Sum(iFaceVals...)
	case utils.MetaDifference:
		iFaceVals := make([]interface{}, len(cfgFld.Value))
		for i, val := range cfgFld.Value {
			strVal, err := val.ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Difference(iFaceVals...)
	case utils.MetaMultiply:
		iFaceVals := make([]interface{}, len(cfgFld.Value))
		for i, val := range cfgFld.Value {
			strVal, err := val.ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Multiply(iFaceVals...)
	case utils.MetaDivide:
		iFaceVals := make([]interface{}, len(cfgFld.Value))
		for i, val := range cfgFld.Value {
			strVal, err := val.ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Divide(iFaceVals...)
	case utils.MetaValueExponent:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaValueExponent)
		}
		strVal1, err := cfgFld.Value[0].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep) // String Value
		if err != nil {
			return "", err
		}
		val, err := strconv.ParseFloat(strVal1, 64)
		if err != nil {
			return "", fmt.Errorf("invalid value <%s> to %s",
				strVal1, utils.MetaValueExponent)
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(eeR.dynamicProvider, utils.NestingSep) // String Exponent
		if err != nil {
			return "", err
		}
		exp, err := strconv.Atoi(strVal2)
		if err != nil {
			return "", err
		}
		out = strconv.FormatFloat(utils.Round(val*math.Pow10(exp),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64)
	case utils.MetaUnixTimestamp:
		val, err := cfgFld.Value.ParseDataProvider(eeR.dynamicProvider, utils.NestingSep)
		if err != nil {
			return nil, err
		}
		t, err := utils.ParseTimeDetectLayout(val, cfgFld.Timezone)
		if err != nil {
			return nil, err
		}
		out = strconv.Itoa(int(t.Unix()))
	case utils.MetaMaskedDestination:
		//check if we have destination in the event
		if dst, err := eeR.req.FieldAsString([]string{utils.Destination}); err != nil {
			return nil, fmt.Errorf("error <%s> getting destination for %s",
				err, utils.ToJSON(cfgFld))
		} else if cfgFld.MaskLen != -1 && len(cfgFld.MaskDestID) != 0 &&
			engine.CachedDestHasPrefix(cfgFld.MaskDestID, dst) {
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
