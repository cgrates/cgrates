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
	vars, cgrRply *utils.DataNode,
	rply *utils.OrderedNavigableMap,
	opts utils.MapStorage,
	tntTpl config.RSRParsers,
	dfltTenant, timezone string,
	filterS *engine.FilterS,
	header, trailer utils.DataProvider) (ar *AgentRequest) {
	if cgrRply == nil {
		cgrRply = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	}
	if vars == nil {
		vars = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	}
	if rply == nil {
		rply = utils.NewOrderedNavigableMap()
	}
	if opts == nil {
		opts = make(utils.MapStorage)
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
		Cfg:        config.CgrConfig().GetDataProvider(),
	}
	if tnt, err := tntTpl.ParseDataProvider(ar); err == nil && tnt != utils.EmptyString {
		ar.Tenant = tnt
	}
	ar.Vars.Set([]string{utils.NodeID}, config.CgrConfig().GeneralCfg().NodeID)
	return
}

// AgentRequest represents data related to one request towards agent
// implements utils.DataProvider so we can pass it to filters
type AgentRequest struct {
	Request    utils.DataProvider         // request
	Vars       *utils.DataNode            // shared data
	CGRRequest *utils.OrderedNavigableMap // Used in reply to access the request that was send
	CGRReply   *utils.DataNode
	Reply      *utils.OrderedNavigableMap
	Tenant     string
	Timezone   string
	filterS    *engine.FilterS
	Header     utils.DataProvider
	Trailer    utils.DataProvider
	diamreq    *utils.OrderedNavigableMap // used in case of building requests (ie. DisconnectSession)
	tmp        *utils.DataNode            // used in case you want to store temporary items and access them later
	Opts       utils.MapStorage
	Cfg        utils.DataProvider
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
	case utils.MetaCfg:
		val, err = ar.Cfg.FieldAsInterface(fldPath[1:])
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
	ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
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
					PathItems: utils.CloneSlice(tplFld.GetPathItems()), // need to clone so me do not modify the template
					Path:      tplFld.Path,
				}
				itmPath = tplFld.GetPathSlice()[1:]
			} else {
				itmPath = fullPath.PathItems[1:]
			}

			nMItm := &utils.DataLeaf{Data: out, Path: itmPath, NewBranch: tplFld.NewBranch, AttributeID: tplFld.AttributeID}
			switch tplFld.Type {
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
		if tplFld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// Set implements utils.NMInterface
func (ar *AgentRequest) SetAsSlice(fullPath *utils.FullPath, nm *utils.DataLeaf) (err error) {
	switch fullPath.PathItems[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathItems[0])
	case utils.MetaVars:
		_, err = ar.Vars.Set(fullPath.PathItems[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaCgreq:
		return ar.CGRRequest.SetAsSlice(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[7:],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaCgrep:
		_, err = ar.CGRReply.Set(fullPath.PathItems[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaRep:
		return ar.Reply.SetAsSlice(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[5:],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaDiamreq:
		return ar.diamreq.SetAsSlice(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[9:],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaTmp:
		_, err = ar.tmp.Set(fullPath.PathItems[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaOpts:
		return ar.Opts.Set(fullPath.PathItems[1:], nm.Data)
	case utils.MetaUCH:
		return engine.Cache.Set(utils.CacheUCH, fullPath.Path[5:], nm.Data, nil, true, utils.NonTransactional)
	}
}

// RemoveAll deletes all fields at given prefix
func (ar *AgentRequest) RemoveAll(prefix string) error {
	switch prefix {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set fields", prefix)
	case utils.MetaVars:
		ar.Vars = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaCgreq:
		ar.CGRRequest.RemoveAll()
	case utils.MetaCgrep:
		ar.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaRep:
		ar.Reply.RemoveAll()
	case utils.MetaDiamreq:
		ar.diamreq.RemoveAll()
	case utils.MetaTmp:
		ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaUCH:
		engine.Cache.Clear([]string{utils.CacheUCH})
	case utils.MetaOpts:
		ar.Opts = make(utils.MapStorage)
	}
	return nil
}

// Remove deletes the fields found at path with the given prefix
func (ar *AgentRequest) Remove(fullPath *utils.FullPath) error {
	switch fullPath.PathItems[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set fields", fullPath.PathItems[0])
	case utils.MetaVars:
		return ar.Vars.Remove(utils.CloneSlice(fullPath.PathItems[1:]))
	case utils.MetaCgreq:
		return ar.CGRRequest.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[7:],
		})
	case utils.MetaCgrep:
		return ar.CGRReply.Remove(utils.CloneSlice(fullPath.PathItems[1:]))
	case utils.MetaRep:
		return ar.Reply.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[5:],
		})
	case utils.MetaDiamreq:
		return ar.diamreq.Remove(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[9:],
		})
	case utils.MetaTmp:
		return ar.tmp.Remove(utils.CloneSlice(fullPath.PathItems[1:]))
	case utils.MetaOpts:
		return ar.Opts.Remove(fullPath.PathItems[1:])
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
func (ar *AgentRequest) setCGRReply(rply utils.NavigableMapper, err error) {
	ar.CGRReply = &utils.DataNode{
		Type: utils.NMMapType,
		Map:  make(map[string]*utils.DataNode),
	}
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else if rply != nil {
		ar.CGRReply.Map = rply.AsNavigableMap()
	}
	ar.CGRReply.Map[utils.Error] = utils.NewLeafNode(errMsg)
}

func needsMaxUsage(ralsFlags utils.FlagParams) bool {
	return len(ralsFlags) != 0 &&
		(ralsFlags.Has(utils.MetaAuthorize) ||
			ralsFlags.Has(utils.MetaInitiate) ||
			ralsFlags.Has(utils.MetaUpdate))
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *AgentRequest) Append(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathItems[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathItems[0])
	case utils.MetaVars:
		_, err = ar.Vars.Append(fullPath.PathItems[1:], val)
		return
	case utils.MetaCgreq:
		return ar.CGRRequest.Append(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[7:],
		}, val)
	case utils.MetaCgrep:
		_, err = ar.CGRReply.Append(fullPath.PathItems[1:], val)
		return
	case utils.MetaRep:
		return ar.Reply.Append(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[5:],
		}, val)
	case utils.MetaDiamreq:
		return ar.diamreq.Append(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[9:],
		}, val)
	case utils.MetaTmp:
		_, err = ar.tmp.Append(fullPath.PathItems[1:], val)
		return
	case utils.MetaOpts:
		return ar.Opts.Set(fullPath.PathItems[1:], val.Data)
	case utils.MetaUCH:
		return engine.Cache.Set(utils.CacheUCH, fullPath.Path[5:], val.Data, nil, true, utils.NonTransactional)
	}
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *AgentRequest) Compose(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathItems[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathItems[0])
	case utils.MetaVars:
		return ar.Vars.Compose(fullPath.PathItems[1:], val)
	case utils.MetaCgreq:
		return ar.CGRRequest.Compose(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[7:],
		}, val)
	case utils.MetaCgrep:
		return ar.CGRReply.Compose(fullPath.PathItems[1:], val)
	case utils.MetaRep:
		return ar.Reply.Compose(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[5:],
		}, val)
	case utils.MetaDiamreq:
		return ar.diamreq.Compose(&utils.FullPath{
			PathItems: fullPath.PathItems[1:],
			Path:      fullPath.Path[9:],
		}, val)
	case utils.MetaTmp:
		return ar.tmp.Compose(fullPath.PathItems[1:], val)
	case utils.MetaOpts:
		var prv interface{}
		if prv, err = ar.Opts.FieldAsInterface(fullPath.PathItems[1:]); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prv) + utils.IfaceAsString(val.Data)
		}
		return ar.Opts.Set(fullPath.PathItems[1:], prv)

	case utils.MetaUCH:
		path := fullPath.Path[5:]
		var prv interface{}
		if prvI, ok := engine.Cache.Get(utils.CacheUCH, path); !ok {
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prvI) + utils.IfaceAsString(val.Data)
		}
		return engine.Cache.Set(utils.CacheUCH, path, prv, nil, true, utils.NonTransactional)
	}
}
