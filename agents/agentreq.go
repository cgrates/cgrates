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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newAgentRequest(req engine.DataProvider) *AgentRequest {
	return &AgentRequest{
		Request:  req,
		Vars:     engine.NewNavigableMap(nil),
		CGRReply: engine.NewNavigableMap(nil),
		Reply:    engine.NewNavigableMap(nil),
	}

}

// AgentRequest represents data related to one request towards agent
// implements engine.DataProvider so we can pass it to filters
type AgentRequest struct {
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
