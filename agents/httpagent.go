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
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewHttpAgent will construct a HTTPAgent
func NewHTTPAgent(sessionS rpcclient.RpcClientConnection,
	timezone, reqPayload, rplyPayload string,
	reqProcessors []*config.HttpAgntProcCfg) *HTTPAgent {
	return &HTTPAgent{sessionS: sessionS, timezone: timezone,
		reqPayload: reqPayload, rplyPayload: rplyPayload,
		reqProcessors: reqProcessors}
}

// HTTPAgent is a handler for HTTP requests
type HTTPAgent struct {
	sessionS rpcclient.RpcClientConnection
	timezone,
	reqPayload,
	rplyPayload string
	reqProcessors []*config.HttpAgntProcCfg
}

// ServeHTTP implements http.Handler interface
func (ha *HTTPAgent) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, err := newHAReqDecoder(ha.reqPayload, req) // dcdr
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error creating decoder: %s",
				utils.HTTPAgent, err.Error()))
	}
	var processed bool
	procVars := make(processorVars)
	rpl := newHTTPReplyFields()
	for _, reqProcessor := range ha.reqProcessors {
		var lclProcessed bool
		if lclProcessed, err = ha.processRequest(reqProcessor, req,
			procVars, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s processing request: %s, process vars: %+v",
			utils.HTTPAgent, err.Error(), utils.ToJSON(req), procVars))
		return // FixMe with returning some error on HTTP level
	} else if !processed {
		utils.Logger.Warning(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s, process vars: %+v",
			utils.RadiusAgent, utils.ToJSON(req), procVars))
		return // FixMe with returning some error on HTTP level
	}
}

// processRequest represents one processor processing the request
func (ha *HTTPAgent) processRequest(reqProc *config.HttpAgntProcCfg,
	req *http.Request, procVars processorVars,
	reply *httpReplyFields) (processed bool, err error) {
	return
}
