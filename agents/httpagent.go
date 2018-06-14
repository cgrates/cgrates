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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewHttpAgent will construct a HTTPAgent
func NewHTTPAgent(sessionS rpcclient.RpcClientConnection,
	filterS *engine.FilterS,
	timezone, reqPayload, rplyPayload string,
	reqProcessors []*config.HttpAgntProcCfg) *HTTPAgent {
	return &HTTPAgent{sessionS: sessionS, timezone: timezone,
		reqPayload: reqPayload, rplyPayload: rplyPayload,
		reqProcessors: reqProcessors}
}

// HTTPAgent is a handler for HTTP requests
type HTTPAgent struct {
	sessionS rpcclient.RpcClientConnection
	filterS  *engine.FilterS
	timezone,
	reqPayload,
	rplyPayload string
	reqProcessors []*config.HttpAgntProcCfg
}

// ServeHTTP implements http.Handler interface
func (ha *HTTPAgent) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	dcdr, err := newHADataProvider(ha.reqPayload, req) // dcdr will provide information from request
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error creating decoder: %s",
				utils.HTTPAgent, err.Error()))
		return
	}
	agReq := newAgentRequest(dcdr)
	var processed bool
	for _, reqProcessor := range ha.reqProcessors {
		var lclProcessed bool
		if lclProcessed, err = ha.processRequest(reqProcessor, agReq); lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing request: %s",
				utils.HTTPAgent, err.Error(), utils.ToJSON(req)))
		return // FixMe with returning some error on HTTP level
	} else if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring request %s",
				utils.HTTPAgent, utils.ToJSON(req)))
		return // FixMe with returning some error on HTTP level
	}
	encdr, err := newHAReplyEncoder(ha.rplyPayload, w)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error creating reply encoder: %s",
				utils.HTTPAgent, err.Error()))
		return
	}
	if err = encdr.encode(agReq.Reply); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s encoding out %s",
				utils.HTTPAgent, err.Error(), utils.ToJSON(agReq.Reply)))
		return
	}
}

// processRequest represents one processor processing the request
func (ha *HTTPAgent) processRequest(reqProcessor *config.HttpAgntProcCfg,
	agReq *AgentRequest) (processed bool, err error) {
	tnt, err := agReq.Request.FieldAsString([]string{utils.Tenant})
	if err != nil {
		return false, err
	}
	if pass, err := ha.filterS.Pass(tnt, reqProcessor.Filters, agReq); err != nil {
		return false, err
	} else if !pass {
		return false, nil
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<%s> DRY_RUN, HTTP request: %s", utils.HTTPAgent, agReq))
	}
	/*
		ev, err := radReqAsCGREvent(req, procVars, reqProcessor.Flags, reqProcessor.RequestFields)
		if err != nil {
			return false, err
		}
		if reqProcessor.DryRun {
			utils.Logger.Info(fmt.Sprintf("<%s> DRY_RUN, CGREvent: %s", utils.RadiusAgent, utils.ToJSON(cgrEv)))
		}
	*/
	return
}
