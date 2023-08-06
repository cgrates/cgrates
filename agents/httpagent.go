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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewHttpAgent will construct a HTTPAgent
func NewHTTPAgent(connMgr *engine.ConnManager, sessionConns []string,
	filterS *engine.FilterS, dfltTenant, reqPayload, rplyPayload string,
	reqProcessors []*config.RequestProcessor) *HTTPAgent {
	return &HTTPAgent{
		connMgr:       connMgr,
		filterS:       filterS,
		dfltTenant:    dfltTenant,
		reqPayload:    reqPayload,
		rplyPayload:   rplyPayload,
		reqProcessors: reqProcessors,
		sessionConns:  sessionConns,
	}
}

// HTTPAgent is a handler for HTTP requests
type HTTPAgent struct {
	connMgr       *engine.ConnManager
	filterS       *engine.FilterS
	dfltTenant    string
	reqPayload    string
	rplyPayload   string
	reqProcessors []*config.RequestProcessor
	sessionConns  []string
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
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.MapStorage{}
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.RemoteHost: utils.NewLeafNode(req.RemoteAddr)}}
	for _, reqProcessor := range ha.reqProcessors {
		agReq := NewAgentRequest(dcdr, reqVars, cgrRplyNM, rplyNM,
			opts, reqProcessor.Tenant, ha.dfltTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ha.filterS, nil)
		lclProcessed, err := processRequest(context.TODO(), reqProcessor, agReq,
			utils.HTTPAgent, ha.connMgr, ha.sessionConns,
			agReq.filterS)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing request: %s",
					utils.HTTPAgent, err.Error(), utils.ToJSON(agReq)))
			return // FixMe with returning some error on HTTP level
		}
		if !lclProcessed {
			continue
		}
		if lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue) {
			break
		}
	}
	encdr, err := newHAReplyEncoder(ha.rplyPayload, w)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error creating reply encoder: %s",
				utils.HTTPAgent, err.Error()))
		return
	}
	if err = encdr.Encode(rplyNM); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s encoding out %s",
				utils.HTTPAgent, err.Error(), utils.ToJSON(rplyNM)))
		return
	}
}
