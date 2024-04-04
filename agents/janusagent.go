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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

// NewJanusAgent will construct a JanusAgent
func NewJanusAgent(connMgr *engine.ConnManager,
	sessionConns []string, filterS *engine.FilterS,
	reqProcessors []*config.RequestProcessor) *JanusAgent {
	return &JanusAgent{
		connMgr:       connMgr,
		filterS:       filterS,
		reqProcessors: reqProcessors,
		sessionConns:  sessionConns,
	}
}

// JanusAgent is a gateway between HTTP and Janus Server over Websocket
type JanusAgent struct {
	connMgr       *engine.ConnManager
	filterS       *engine.FilterS
	reqProcessors []*config.RequestProcessor
	sessionConns  []string
}
