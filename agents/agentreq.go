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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newAgentRequest(req config.DataProvider, tntTpl config.RSRParsers,
	dfltTenant, timezone string, filterS *engine.FilterS) (ar *AgentRequest) {
	ar = &AgentRequest{
		Request:    req,
		Vars:       config.NewNavigableMap(nil),
		CGRRequest: config.NewNavigableMap(nil),
		CGRReply:   config.NewNavigableMap(nil),
		Reply:      config.NewNavigableMap(nil),
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
	return utils.ToJSON(ar)
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
	switch fldPath[0] {
	default:
		return "", fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaReq:
		return ar.Request.FieldAsString(fldPath[1:])
	case utils.MetaVars:
		return ar.Vars.FieldAsString(fldPath[1:])
	case utils.MetaCgreq:
		return ar.CGRRequest.FieldAsString(fldPath[1:])
	case utils.MetaCgrep:
		return ar.CGRReply.FieldAsString(fldPath[1:])
	case utils.MetaRep:
		return ar.Reply.FieldAsString(fldPath[1:])
	}
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
			return nil, err
		}
		var valSet []*config.NMItem
		fldPath := strings.Split(tplFld.FieldId, utils.NestingSep)
		if nMFields, err := nM.FieldAsInterface(fldPath); err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
		} else {
			valSet = nMFields.([]*config.NMItem) // start from previous stored fields
		}
		valSet = append(valSet, &config.NMItem{Data: out, Path: fldPath, Config: tplFld})
		nM.Set(fldPath, valSet, true)
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
	case utils.META_COMPOSED:
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
	}
	if err != nil {
		return
	}
	if isString { // format the string additionally with fmtFieldWidth
		out, err = utils.FmtFieldWidth(cfgFld.Tag, out.(string), cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)
	}
	return
}
