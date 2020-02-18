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
func NewAgentRequest(req config.DataProvider,
	vars map[string]interface{},
	cgrRply *config.NavigableMap,
	rply *config.NavigableMap,
	tntTpl config.RSRParsers,
	dfltTenant, timezone string,
	filterS *engine.FilterS,
	header, trailer config.DataProvider) (ar *AgentRequest) {
	if cgrRply == nil {
		cgrRply = config.NewNavigableMap(nil)
	}
	if rply == nil {
		rply = config.NewNavigableMap(nil)
	}
	ar = &AgentRequest{
		Request:    req,
		Vars:       config.NewNavigableMap(vars),
		CGRRequest: config.NewNavigableMap(nil),
		diamreq:    config.NewNavigableMap(nil), // special case when CGRateS is building the request
		CGRReply:   cgrRply,
		Reply:      rply,
		Timezone:   timezone,
		filterS:    filterS,
		Header:     header,
		Trailer:    trailer,
	}
	// populate tenant
	if tntIf, err := ar.ParseField(
		&config.FCTemplate{Type: utils.META_COMPOSED,
			Value: tntTpl}); err == nil && tntIf.(string) != "" {
		ar.Tenant = tntIf.(string)
	} else {
		ar.Tenant = dfltTenant
	}
	ar.Vars.Set([]string{utils.NodeID}, config.CgrConfig().GeneralCfg().NodeID, false, true)
	return
}

// AgentRequest represents data related to one request towards agent
// implements engine.DataProvider so we can pass it to filters
type AgentRequest struct {
	Request    config.DataProvider  // request
	Vars       *config.NavigableMap // shared data
	CGRRequest *config.NavigableMap // Used in reply to access the request that was send
	CGRReply   *config.NavigableMap
	Reply      *config.NavigableMap
	Tenant,
	Timezone string
	filterS *engine.FilterS
	Header  config.DataProvider
	Trailer config.DataProvider
	diamreq *config.NavigableMap // used in case of building requests (ie. DisconnectSession)
}

// String implements engine.DataProvider
func (ar *AgentRequest) String() string {
	return utils.ToIJSON(ar)
}

// RemoteHost implements engine.DataProvider
func (ar *AgentRequest) RemoteHost() net.Addr {
	return ar.Request.RemoteHost()
}

// FieldAsInterface implements engine.DataProvider
func (ar *AgentRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaReq:
		val, err = ar.Request.FieldAsInterface(fldPath[1:])
	case utils.MetaVars:
		val, err = ar.Vars.GetField(fldPath[1:])
	case utils.MetaCgreq:
		val, err = ar.CGRRequest.GetField(fldPath[1:])
	case utils.MetaCgrep:
		val, err = ar.CGRReply.GetField(fldPath[1:])
	case utils.MetaDiamreq:
		val, err = ar.diamreq.FieldAsInterface(fldPath[1:])
	case utils.MetaRep:
		val, err = ar.Reply.GetField(fldPath[1:])
	case utils.MetaHdr:
		val, err = ar.Header.FieldAsInterface(fldPath[1:])
	case utils.MetaTrl:
		val, err = ar.Trailer.FieldAsInterface(fldPath[1:])
	}
	return
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
	return utils.IfaceAsString(iface), nil
}

// AsNavigableMap implements engine.DataProvider
func (ar *AgentRequest) AsNavigableMap(tplFlds []*config.FCTemplate) (
	nM *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

//SetFields will populate fields of AgentRequest out of templates
func (ar *AgentRequest) SetFields(tplFlds []*config.FCTemplate) (err error) {
	for _, tplFld := range tplFlds {
		if pass, err := ar.filterS.Pass(ar.Tenant,
			tplFld.Filters, ar); err != nil {
			return err
		} else if !pass {
			continue
		}
		if tplFld.Type != utils.META_NONE {
			out, err := ar.ParseField(tplFld)
			if err != nil {
				if err == utils.ErrNotFound {
					if !tplFld.Mandatory {
						err = nil
						continue
					}
					err = utils.ErrPrefixNotFound(tplFld.Tag)
				}
				return err
			}
			var valSet []*config.NMItem
			fldPath := strings.Split(tplFld.Path, utils.NestingSep)

			nMItm := &config.NMItem{Data: out, Path: fldPath[1:], Config: tplFld}
			if nMFields, err := ar.FieldAsInterface(fldPath); err != nil {
				if err != utils.ErrNotFound {
					return err
				}
			} else {
				valSet = nMFields.([]*config.NMItem) // start from previous stored fields
				if tplFld.Type == utils.META_COMPOSED {
					prevNMItem := valSet[len(valSet)-1] // could be we need nil protection here
					*nMItm = *prevNMItem                // inherit the particularities, ie AttributeName
					nMItm.Data = utils.IfaceAsString(prevNMItem.Data) + utils.IfaceAsString(out)
					valSet = valSet[:len(valSet)-1] // discard the last item since we have captured it in nmItem
				} else {
					valSet = nil
				}
			}
			valSet = append(valSet, nMItm)
			switch fldPath[0] {
			default:
				return fmt.Errorf("unsupported field prefix: <%s> when set fields", fldPath[0])
			case utils.MetaVars:
				ar.Vars.Set(fldPath[1:], valSet, false, true)
			case utils.MetaCgreq:
				ar.CGRRequest.Set(fldPath[1:], valSet, false, true)
			case utils.MetaCgrep:
				ar.CGRReply.Set(fldPath[1:], valSet, false, true)
			case utils.MetaRep:
				ar.Reply.Set(fldPath[1:], valSet, false, true)
			case utils.MetaDiamreq:
				ar.diamreq.Set(fldPath[1:], valSet, false, true)
			}
		}
		if tplFld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// ParseField outputs the value based on the template item
func (ar *AgentRequest) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {
	var isString bool
	switch cfgFld.Type {
	default:
		return "", fmt.Errorf("unsupported type: <%s>", cfgFld.Type)
	case utils.META_NONE:
		return
	case utils.META_FILLER:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		cfgFld.Padding = "right"
		isString = true
	case utils.META_CONSTANT:
		out, err = cfgFld.Value.ParseValue(utils.EmptyString)
		isString = true
	case utils.MetaRemoteHost:
		out = ar.RemoteHost().String()
		isString = true
	case utils.MetaVariable, utils.META_COMPOSED:
		out, err = cfgFld.Value.ParseDataProvider(ar, utils.NestingSep)
		isString = true
	case utils.META_USAGE_DIFFERENCE:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.META_USAGE_DIFFERENCE)
		}
		strVal1, err := cfgFld.Value[0].ParseDataProvider(ar, utils.NestingSep)
		if err != nil {
			return "", err
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(ar, utils.NestingSep)
		if err != nil {
			return "", err
		}
		tEnd, err := utils.ParseTimeDetectLayout(strVal1, ar.Timezone)
		if err != nil {
			return "", err
		}
		tStart, err := utils.ParseTimeDetectLayout(strVal2, ar.Timezone)
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
		strVal1, err := cfgFld.Value[0].ParseDataProvider(ar, utils.NestingSep) // ReqNr
		if err != nil {
			return "", err
		}
		reqNr, err := strconv.ParseInt(strVal1, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid requestNumber <%s> to %s",
				strVal1, utils.MetaCCUsage)
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(ar, utils.NestingSep) // TotalUsage
		if err != nil {
			return "", err
		}
		usedCCTime, err := utils.ParseDurationWithNanosecs(strVal2)
		if err != nil {
			return "", fmt.Errorf("invalid usedCCTime <%s> to %s",
				strVal2, utils.MetaCCUsage)
		}
		strVal3, err := cfgFld.Value[2].ParseDataProvider(ar, utils.NestingSep) // DebitInterval
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
			strVal, err := val.ParseDataProvider(ar, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Sum(iFaceVals...)
	case utils.MetaDifference:
		iFaceVals := make([]interface{}, len(cfgFld.Value))
		for i, val := range cfgFld.Value {
			strVal, err := val.ParseDataProvider(ar, utils.NestingSep)
			if err != nil {
				return "", err
			}
			iFaceVals[i] = utils.StringToInterface(strVal)
		}
		out, err = utils.Difference(iFaceVals...)
	case utils.MetaValueExponent:
		if len(cfgFld.Value) != 2 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(cfgFld.Value), utils.MetaValueExponent)
		}
		strVal1, err := cfgFld.Value[0].ParseDataProvider(ar, utils.NestingSep) // String Value
		if err != nil {
			return "", err
		}
		val, err := strconv.ParseFloat(strVal1, 64)
		if err != nil {
			return "", fmt.Errorf("invalid value <%s> to %s",
				strVal1, utils.MetaValueExponent)
		}
		strVal2, err := cfgFld.Value[1].ParseDataProvider(ar, utils.NestingSep) // String Exponent
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
		val, err := cfgFld.Value.ParseDataProvider(ar, utils.NestingSep)
		if err != nil {
			return nil, err
		}
		t, err := utils.ParseTimeDetectLayout(val, cfgFld.Timezone)
		if err != nil {
			return nil, err
		}
		out = strconv.Itoa(int(t.Unix()))
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
func (ar *AgentRequest) setCGRReply(rply config.NavigableMapper, errRply error) (err error) {
	var nm *config.NavigableMap
	if errRply != nil {
		nm = config.NewNavigableMap(map[string]interface{}{
			utils.Error: errRply.Error()})
	} else {
		nm = config.NewNavigableMap(nil)
		if rply != nil {
			if nm, err = rply.AsNavigableMap(nil); err != nil {
				return
			}
		}
		nm.Set([]string{utils.Error}, "", false, false) // enforce empty error
	}
	*ar.CGRReply = *nm // update value so we can share CGRReply
	return
}
