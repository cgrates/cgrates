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
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newAgentRequest(req config.DataProvider,
	vars map[string]interface{},
	rply *config.NavigableMap,
	tntTpl config.RSRParsers,
	dfltTenant, timezone string,
	filterS *engine.FilterS) (ar *AgentRequest) {
	if rply == nil {
		rply = config.NewNavigableMap(nil)
	}
	ar = &AgentRequest{
		Request:    req,
		Vars:       config.NewNavigableMap(vars),
		CGRRequest: config.NewNavigableMap(nil),
		CGRReply:   config.NewNavigableMap(nil),
		Reply:      rply,
		timezone:   timezone,
		filterS:    filterS,
	}
	// populate tenant
	if tntIf, err := ar.ParseField(
		&config.FCTemplate{Type: utils.META_COMPOSED,
			Value: tntTpl}); err == nil && tntIf.(string) != "" {
		ar.tenant = tntIf.(string)
	} else {
		ar.tenant = dfltTenant
	}

	return
}

// AgentRequest represents data related to one request towards agent
// implements engine.DataProvider so we can pass it to filters
type AgentRequest struct {
	Request    config.DataProvider  // request
	Vars       *config.NavigableMap // shared data
	CGRRequest *config.NavigableMap
	CGRReply   *config.NavigableMap
	Reply      *config.NavigableMap
	tenant,
	timezone string
	filterS *engine.FilterS
}

// String implements engine.DataProvider
func (ar *AgentRequest) String() string {
	return utils.ToIJSON(ar)
}

// RemoteHost implements engine.DataProvider
func (aReq *AgentRequest) RemoteHost() net.Addr {
	return aReq.Request.RemoteHost()
}

// FieldAsInterface implements engine.DataProvider
func (ar *AgentRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaReq:
		return ar.Request.FieldAsInterface(fldPath[1:])
	case utils.MetaVars:
		return ar.Vars.FieldAsInterface(fldPath[1:])
	case utils.MetaCgreq:
		return ar.CGRRequest.FieldAsInterface(fldPath[1:])
	case utils.MetaCgrep:
		return ar.CGRReply.FieldAsInterface(fldPath[1:])
	case utils.MetaRep:
		return ar.Reply.FieldAsInterface(fldPath[1:])
	}
}

// FieldAsString implements engine.DataProvider
func (ar *AgentRequest) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = ar.FieldAsInterface(fldPath); err != nil {
		return
	}
	if nmItems, isNMItems := iface.([]*config.NMItem); isNMItems { // special handling of NMItems, take the last value out of it
		iface = nmItems[len(nmItems)-1].Data // could be we need nil protection here
	}
	return utils.IfaceAsString(iface)
}

// AsNavigableMap implements engine.DataProvider
func (ar *AgentRequest) AsNavigableMap(tplFlds []*config.FCTemplate) (
	nM *config.NavigableMap, err error) {
	nM = config.NewNavigableMap(nil)

	for _, tplFld := range tplFlds {
		if pass, err := ar.filterS.Pass(ar.tenant,
			tplFld.Filters, ar); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		out, err := ar.ParseField(tplFld)
		if err != nil {
			if err == utils.ErrNotFound {
				if !tplFld.Mandatory {
					err = nil
					continue
				}
				err = utils.ErrPrefixNotFound(tplFld.Tag)
			}
			return nil, err
		}
		var valSet []*config.NMItem
		fldPath := strings.Split(tplFld.FieldId, utils.NestingSep)
		nMItm := &config.NMItem{Data: out, Path: fldPath, Config: tplFld}
		if nMFields, err := nM.FieldAsInterface(fldPath); err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
		} else {
			valSet = nMFields.([]*config.NMItem) // start from previous stored fields
			if tplFld.Type == utils.META_COMPOSED {
				prevNMItem := valSet[len(valSet)-1] // could be we need nil protection here
				prevDataStr, err := utils.IfaceAsString(prevNMItem.Data)
				if err != nil {
					return nil, err
				}
				outStr, err := utils.IfaceAsString(out)
				if err != nil {
					return nil, err
				}
				*nMItm = *prevNMItem // inherit the particularities, ie AttributeName
				nMItm.Data = prevDataStr + outStr
				valSet = valSet[:len(valSet)-1] // discard the last item
			}
		}
		valSet = append(valSet, nMItm)
		nM.Set(fldPath, valSet, false, true)
		if tplFld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// parseField outputs the value based on the template item
func (aReq *AgentRequest) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {
	var isString bool
	switch cfgFld.Type {
	default:
		return "", fmt.Errorf("unsupported type: <%s>", cfgFld.Type)
	case utils.META_FILLER:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		cfgFld.Padding = "right"
		isString = true
	case utils.META_CONSTANT:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		isString = true
	case utils.MetaRemoteHost:
		out = aReq.RemoteHost().String()
		isString = true
	case utils.MetaVariable, utils.META_COMPOSED:
		out, err = cfgFld.Value.ParseDataProvider(aReq, utils.NestingSep)
		isString = true
	case utils.META_USAGE_DIFFERENCE:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s>", utils.ToJSON(cfgFld.Value))
		}
		strVal1, err := cfgFld.Value[0].ParseDataProvider(aReq, utils.NestingSep)
		if err != nil {
			return "", err
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(aReq, utils.NestingSep)
		if err != nil {
			return "", err
		}
		tEnd, err := utils.ParseTimeDetectLayout(strVal1, aReq.timezone)
		if err != nil {
			return "", err
		}
		tStart, err := utils.ParseTimeDetectLayout(strVal2, aReq.timezone)
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
		strVal1, err := cfgFld.Value[0].ParseDataProvider(aReq, utils.NestingSep) // ReqNr
		if err != nil {
			return "", err
		}
		reqNr, err := strconv.ParseInt(strVal1, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid requestNumber <%s> to %s",
				strVal1, utils.MetaCCUsage)
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(aReq, utils.NestingSep) // TotalUsage
		if err != nil {
			return "", err
		}
		usedCCTime, err := utils.ParseDurationWithNanosecs(strVal2)
		if err != nil {
			return "", fmt.Errorf("invalid usedCCTime <%s> to %s",
				strVal2, utils.MetaCCUsage)
		}
		strVal3, err := cfgFld.Value[2].ParseDataProvider(aReq, utils.NestingSep) // DebitInterval
		if err != nil {
			return "", err
		}
		debitItvl, err := utils.ParseDurationWithNanosecs(strVal3)
		if err != nil {
			return "", fmt.Errorf("invalid debitInterval <%s> to %s",
				strVal3, utils.MetaCCUsage)
		}
		mltpl := reqNr - 2 // init and terminate will be ignored
		if mltpl < 0 {
			mltpl = 0
		}
		return usedCCTime + time.Duration(debitItvl.Nanoseconds()*mltpl), nil
	case utils.MetaSum:
		iFaceVals := make([]interface{}, len(cfgFld.Value))
		for i, val := range cfgFld.Value {
			strVal, err := val.ParseDataProvider(aReq, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Sum(iFaceVals...)
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
