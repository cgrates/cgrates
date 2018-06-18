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

func newAgentRequest(req engine.DataProvider, tntTpl utils.RSRFields,
	dfltTenant string) (ar *AgentRequest) {
	ar = &AgentRequest{
		Request:  req,
		Vars:     engine.NewNavigableMap(nil),
		CGRReply: engine.NewNavigableMap(nil),
		Reply:    engine.NewNavigableMap(nil),
	}
	// populate tenant
	if tntIf, err := ar.ParseField(
		&config.CfgCdrField{Type: utils.META_COMPOSED,
			Value: tntTpl}); err == nil && tntIf.(string) != "" {
		ar.Tenant = tntIf.(string)
	} else {
		ar.Tenant = dfltTenant
	}
	return
}

// AgentRequest represents data related to one request towards agent
// implements engine.DataProvider so we can pass it to filters
type AgentRequest struct {
	Tenant   string
	Request  engine.DataProvider  // request
	Vars     *engine.NavigableMap // shared data
	CGRReply *engine.NavigableMap
	Reply    *engine.NavigableMap
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
	case utils.MetaRequest:
		return ar.Request.FieldAsInterface(fldPath[1:])
	case utils.MetaVars:
		return ar.Vars.FieldAsInterface(fldPath[1:])
	case utils.MetaCGRReply:
		return ar.CGRReply.FieldAsInterface(fldPath[1:])
	case utils.MetaReply:
		return ar.Reply.FieldAsInterface(fldPath[1:])
	}
}

// FieldAsString implements engine.DataProvider
func (ar *AgentRequest) FieldAsString(fldPath []string) (val string, err error) {
	switch fldPath[0] {
	default:
		return "", fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.MetaRequest:
		return ar.Request.FieldAsString(fldPath[1:])
	case utils.MetaVars:
		return ar.Vars.FieldAsString(fldPath[1:])
	case utils.MetaCGRReply:
		return ar.CGRReply.FieldAsString(fldPath[1:])
	case utils.MetaReply:
		return ar.Reply.FieldAsString(fldPath[1:])
	}
}

// AsNavigableMap implements engine.DataProvider
func (ar *AgentRequest) AsNavigableMap([]*config.CfgCdrField) (
	nM *engine.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// parseField outputs the value based on the template item
func (aReq *AgentRequest) ParseField(
	cfgFld *config.CfgCdrField) (out interface{}, err error) {
	var isString bool
	switch cfgFld.Type {
	default:
		return "", fmt.Errorf("unsupported type: <%s>", cfgFld.Type)
	case utils.META_FILLER:
		out = cfgFld.Value.Id()
		cfgFld.Padding = "right"
		isString = true
	case utils.META_CONSTANT:
		out = cfgFld.Value.Id()
		isString = true
	case utils.META_COMPOSED:
		out = aReq.composedField(cfgFld.Value)
		isString = true
	}
	if isString { // format the string additionally with fmtFieldWidth
		out, err = utils.FmtFieldWidth(cfgFld.Tag, out.(string), cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)
	}
	return
}

// composedField is a subset of ParseField
func (ar *AgentRequest) composedField(outTpl utils.RSRFields) (outVal string) {
	for _, rsrTpl := range outTpl {
		if rsrTpl.IsStatic() {
			if parsed, err := rsrTpl.Parse(""); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> %s",
						utils.HTTPAgent, err.Error()))
			} else {
				outVal += parsed
			}
			continue
		}
		valStr, err := ar.FieldAsString(strings.Split(rsrTpl.Id, utils.CONCATENATED_KEY_SEP))
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> %s",
					utils.HTTPAgent, err.Error()))
			continue
		}
		if parsed, err := rsrTpl.Parse(valStr); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> %s",
					utils.RadiusAgent, err.Error()))
		} else {
			outVal += parsed
		}
	}
	return outVal
}
