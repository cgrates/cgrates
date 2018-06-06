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
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/rpcclient"
)

// NewHttpAgent will construct a HttpAgent
func NewHttpAgent(sessionS rpcclient.RpcClientConnection,
	timezone, reqPayload, rplyPayload string,
	reqProcessors []*config.HttpAgntProcCfg) *HttpAgent {
	return &HttpAgent{sessionS: sessionS, timezone: timezone,
		reqPayload: reqPayload, rplyPayload: rplyPayload,
		reqProcessors: reqProcessors}
}

type HttpAgent struct {
	sessionS rpcclient.RpcClientConnection
	timezone,
	reqPayload,
	rplyPayload string
	reqProcessors []*config.HttpAgntProcCfg
}

// ServeHTTP implements http.Handler interface
func (ha *HttpAgent) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}
