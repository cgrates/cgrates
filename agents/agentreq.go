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

package agents

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

// NewAgentRequest returns a new AgentRequest
func NewAgentRequest(req utils.DataProvider,
	vars utils.NavigableMap2,
	cgrRply *utils.NavigableMap2,
	rply, opts *utils.OrderedNavigableMap,
	tntTpl config.RSRParsers,
	dfltTenant, timezone string,
	filterS *engine.FilterS,
	header, trailer utils.DataProvider) (ar *AgentRequest) {
	if cgrRply == nil {
		cgrRply = &utils.NavigableMap2{}
	}
	if vars == nil {
		vars = make(utils.NavigableMap2)
	}
	if rply == nil {
		rply = utils.NewOrderedNavigableMap()
	}
	if opts == nil {
		opts = utils.NewOrderedNavigableMap()
	}
	ar = &AgentRequest{
		Request:    req,
		Tenant:     dfltTenant,
		Vars:       vars,
		CGRRequest: utils.NewOrderedNavigableMap(),
		diamreq:    utils.NewOrderedNavigableMap(), // special case when CGRateS is building the request
		CGRReply:   cgrRply,
		Reply:      rply,
		Timezone:   timezone,
		filterS:    filterS,
		Header:     header,
		Trailer:    trailer,
		Opts:       opts,
	}
	if tntTpl != nil {
		if tntIf, err := ar.ParseField(
			&config.FCTemplate{Type: utils.MetaComposed,
				Value: tntTpl}); err == nil && tntIf.(string) != "" {
			ar.Tenant = tntIf.(string)
		}
	}
	ar.Vars.Set(utils.PathItems{{Field: utils.NodeID}}, utils.NewNMData(config.CgrConfig().GeneralCfg().NodeID))
	return
}

// AgentRequest represents data related to one request towards agent
// implements utils.DataProvider so we can pass it to filters
type AgentRequest struct {
	Request    utils.DataProvider         // request
	Vars       utils.NavigableMap2        // shared data
	CGRRequest *utils.OrderedNavigableMap // Used in reply to access the request that was send
	CGRReply   *utils.NavigableMap2
	Reply      *utils.OrderedNavigableMap
	Tenant     string
	Timezone   string
	filterS    *engine.FilterS
	Header     utils.DataProvider
	Trailer    utils.DataProvider
	diamreq    *utils.OrderedNavigableMap // used in case of building requests (ie. DisconnectSession)
	tmp        utils.NavigableMap2        // used in case you want to store temporary items and access them later
	Opts       *utils.OrderedNavigableMap
}

// String implements utils.DataProvider
func (ar *AgentRequest) String() string {
	return utils.ToIJSON(ar)
}

// RemoteHost implements utils.DataProvider
func (ar *AgentRequest) RemoteHost() net.Addr {
	return ar.Request.RemoteHost()
}

// FieldAsInterface implements utils.DataProvider
func (ar *AgentRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaReq:
		val, err = ar.Request.FieldAsInterface(fldPath[1:])
	case utils.MetaVars:
		val, err = ar.Vars.FieldAsInterface(fldPath[1:])
	case utils.MetaCgreq:
		val, err = ar.CGRRequest.FieldAsInterface(fldPath[1:])
	case utils.MetaCgrep:
		val, err = ar.CGRReply.FieldAsInterface(fldPath[1:])
	case utils.MetaDiamreq:
		val, err = ar.diamreq.FieldAsInterface(fldPath[1:])
	case utils.MetaRep:
		val, err = ar.Reply.FieldAsInterface(fldPath[1:])
	case utils.MetaHdr:
		val, err = ar.Header.FieldAsInterface(fldPath[1:])
	case utils.MetaTrl:
		val, err = ar.Trailer.FieldAsInterface(fldPath[1:])
	case utils.MetaTmp:
		val, err = ar.tmp.FieldAsInterface(fldPath[1:])
	case utils.MetaUCH:
		if cacheVal, ok := engine.Cache.Get(utils.CacheUCH, strings.Join(fldPath[1:], utils.NestingSep)); !ok {
			err = utils.ErrNotFound
		} else {
			val = cacheVal
		}
	case utils.MetaOpts:
		val, err = ar.Opts.FieldAsInterface(fldPath[1:])
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
func (ar *AgentRequest) Field(fldPath utils.PathItems) (val utils.NMInterface, err error) {
	switch fldPath[0].Field {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaVars:
		val, err = ar.Vars.Field(fldPath[1:])
	case utils.MetaCgreq:
		val, err = ar.CGRRequest.Field(fldPath[1:])
	case utils.MetaCgrep:
		val, err = ar.CGRReply.Field(fldPath[1:])
	case utils.MetaDiamreq:
		val, err = ar.diamreq.Field(fldPath[1:])
	case utils.MetaRep:
		val, err = ar.Reply.Field(fldPath[1:])
	case utils.MetaTmp:
		val, err = ar.tmp.Field(fldPath[1:])
	case utils.MetaOpts:
		val, err = ar.Opts.Field(fldPath[1:])
	}
	return
}

// FieldAsString implements utils.DataProvider
func (ar *AgentRequest) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = ar.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

//SetFields will populate fields of AgentRequest out of templates
func (ar *AgentRequest) SetFields(tplFlds []*config.FCTemplate) (err error) {
	ar.tmp = utils.NavigableMap2{}
	for _, tplFld := range tplFlds {
		if pass, err := ar.filterS.Pass(ar.Tenant,
			tplFld.Filters, ar); err != nil {
			return err
		} else if !pass {
			continue
		}
		switch tplFld.Type {
		case utils.MetaNone:
		case utils.MetaRemove:
			if err = ar.Remove(&utils.FullPath{
				PathItems: tplFld.GetPathItems(),
				Path:      tplFld.Path,
			}); err != nil {
				return
			}
		case utils.MetaRemoveAll:
			if err = ar.RemoveAll(tplFld.GetPathSlice()[0]); err != nil {
				return
			}
		default:
			var out interface{}
			out, err = ar.ParseField(tplFld)
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
			if fullPath, err = utils.GetFullFieldPath(tplFld.Path, ar); err != nil {
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
			case utils.MetaComposed:
				err = utils.ComposeNavMapVal(ar, fullPath, nMItm)
			case utils.MetaGroup: // in case of *group type simply append to valSet
				err = utils.AppendNavMapVal(ar, fullPath, nMItm)
			default:
				_, err = ar.Set(fullPath, &utils.NMSlice{nMItm})
			}
			if err != nil {
				return
			}
		}
		if tplFld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// Set implements utils.NMInterface
func (ar *AgentRequest) Set(fullPath *utils.FullPath, nm utils.NMInterface) (added bool, err error) {
	switch fullPath.PathItems[0].Field {
	default:
		return false, fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathItems[0].Field)
	case utils.MetaVars:
		return ar.Vars.Set(fullPath.PathItems[1:], nm)
	case utils.MetaCgreq:
		return ar.CGRRequest.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[7:],
		}, nm)
	case utils.MetaCgrep:
		return ar.CGRReply.Set(fullPath.PathItems[1:], nm)
	case utils.MetaRep:
		return ar.Reply.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[5:],
		}, nm)
	case utils.MetaDiamreq:
		return ar.diamreq.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[9:],
		}, nm)
	case utils.MetaTmp:
		return ar.tmp.Set(fullPath.PathItems[1:], nm)
	case utils.MetaOpts:
		return ar.Opts.Set(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[6:],
		}, nm)
	case utils.MetaUCH:
		err = engine.Cache.Set(utils.CacheUCH, fullPath.Path[5:], nm, nil, true, utils.NonTransactional)
	}
	return false, err
}

// RemoveAll deletes all fields at given prefix
func (ar *AgentRequest) RemoveAll(prefix string) error {
	switch prefix {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set fields", prefix)
	case utils.MetaVars:
		ar.Vars = utils.NavigableMap2{}
	case utils.MetaCgreq:
		ar.CGRRequest.RemoveAll()
	case utils.MetaCgrep:
		ar.CGRReply = &utils.NavigableMap2{}
	case utils.MetaRep:
		ar.Reply.RemoveAll()
	case utils.MetaDiamreq:
		ar.diamreq.RemoveAll()
	case utils.MetaTmp:
		ar.tmp = utils.NavigableMap2{}
	case utils.MetaUCH:
		engine.Cache.Clear([]string{utils.CacheUCH})
	case utils.MetaOpts:
		ar.Opts.RemoveAll()
	}
	return nil
}

// Remove deletes the fields found at path with the given prefix
func (ar *AgentRequest) Remove(fullPath *utils.FullPath) error {
	switch fullPath.PathItems[0].Field {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set fields", fullPath.PathItems[0].Field)
	case utils.MetaVars:
		return ar.Vars.Remove(fullPath.PathItems[1:])
	case utils.MetaCgreq:
		return ar.CGRRequest.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:].Clone(),
			Path:      fullPath.Path[7:],
		})
	case utils.MetaCgrep:
		return ar.CGRReply.Remove(fullPath.PathItems[1:])
	case utils.MetaRep:
		return ar.Reply.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:].Clone(),
			Path:      fullPath.Path[5:],
		})
	case utils.MetaDiamreq:
		return ar.diamreq.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:].Clone(),
			Path:      fullPath.Path[9:],
		})
	case utils.MetaTmp:
		return ar.tmp.Remove(fullPath.PathItems[1:])
	case utils.MetaOpts:
		return ar.Opts.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:].Clone(),
			Path:      fullPath.Path[6:],
		})
	case utils.MetaUCH:
		return engine.Cache.Remove(utils.CacheUCH, fullPath.Path[5:], true, utils.NonTransactional)
	}
}

// ParseField outputs the value based on the template item
func (ar *AgentRequest) ParseField(
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
		out = ar.RemoteHost().String()
		isString = true
	case utils.MetaVariable, utils.MetaComposed, utils.MetaGroup:
		out, err = cfgFld.Value.ParseDataProvider(ar)
		isString = true
	case utils.MetaUsageDifference:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaUsageDifference)
		}
		var strVal1 string
		if strVal1, err = cfgFld.Value[0].ParseDataProvider(ar); err != nil {
			return
		}
		var strVal2 string
		if strVal2, err = cfgFld.Value[1].ParseDataProvider(ar); err != nil {
			return
		}
		var tEnd time.Time
		if tEnd, err = utils.ParseTimeDetectLayout(strVal1, ar.Timezone); err != nil {
			return
		}
		var tStart time.Time
		if tStart, err = utils.ParseTimeDetectLayout(strVal2, ar.Timezone); err != nil {
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
		if strVal1, err = cfgFld.Value[0].ParseDataProvider(ar); err != nil {
			return
		}
		var reqNr int64
		if reqNr, err = strconv.ParseInt(strVal1, 10, 64); err != nil {
			err = fmt.Errorf("invalid requestNumber <%s> to %s",
				strVal1, utils.MetaCCUsage)
			return
		}
		var strVal2 string
		if strVal2, err = cfgFld.Value[1].ParseDataProvider(ar); err != nil {
			return
		}
		var usedCCTime time.Duration
		if usedCCTime, err = utils.ParseDurationWithNanosecs(strVal2); err != nil {
			err = fmt.Errorf("invalid usedCCTime <%s> to %s",
				strVal2, utils.MetaCCUsage)
			return
		}
		var strVal3 string
		if strVal3, err = cfgFld.Value[2].ParseDataProvider(ar); err != nil {
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
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(ar); err != nil {
			return
		}
		out, err = utils.Sum(iFaceVals...)
	case utils.MetaDifference:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(ar); err != nil {
			return
		}
		out, err = utils.Difference(iFaceVals...)
	case utils.MetaMultiply:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(ar); err != nil {
			return
		}
		out, err = utils.Multiply(iFaceVals...)
	case utils.MetaDivide:
		var iFaceVals []interface{}
		if iFaceVals, err = cfgFld.Value.GetIfaceFromValues(ar); err != nil {
			return
		}
		out, err = utils.Divide(iFaceVals...)
	case utils.MetaValueExponent:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaValueExponent)
		}
		var strVal1 string
		if strVal1, err = cfgFld.Value[0].ParseDataProvider(ar); err != nil {
			return
		}
		var val float64
		if val, err = strconv.ParseFloat(strVal1, 64); err != nil {
			err = fmt.Errorf("invalid value <%s> to %s",
				strVal1, utils.MetaValueExponent)
			return
		}
		var strVal2 string
		if strVal2, err = cfgFld.Value[1].ParseDataProvider(ar); err != nil {
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
		if val, err = cfgFld.Value.ParseDataProvider(ar); err != nil {
			return
		}
		var t1 time.Time
		if t1, err = utils.ParseTimeDetectLayout(val, cfgFld.Timezone); err != nil {
			return
		}
		out = strconv.Itoa(int(t1.Unix()))
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

// setCGRReply will set the aReq.cgrReply based on reply coming from upstream or error
// returns error in case of reply not converting to NavigableMap
func (ar *AgentRequest) setCGRReply(rply utils.NavigableMapper, errRply error) (err error) {
	var nm utils.NavigableMap2
	if errRply != nil {
		nm = utils.NavigableMap2{utils.Error: utils.NewNMData(errRply.Error())}
	} else {
		nm = utils.NavigableMap2{}
		if rply != nil {
			nm = rply.AsNavigableMap()
		}
		nm.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData("")) // enforce empty error
	}
	*ar.CGRReply = nm // update value so we can share CGRReply
	return
}

func needsMaxUsage(ralsFlags utils.FlagParams) bool {
	if len(ralsFlags) == 0 {
		return false
	}
	for _, flag := range []string{utils.MetaAuthorize, utils.MetaInitiate, utils.MetaUpdate} {
		if ralsFlags.Has(flag) {
			return true
		}
	}
	return false
}
