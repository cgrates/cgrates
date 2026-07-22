// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
func NewHTTPAgent(connMgr *engine.ConnManager,
	sessionConns, statsConns, thresholdsConns []string,
	filterS *engine.FilterS, dfltTenant, reqPayload, rplyPayload string,
	reqProcessors []*config.RequestProcessor, caps *engine.Caps) *HTTPAgent {
	return &HTTPAgent{
		connMgr:         connMgr,
		filterS:         filterS,
		dfltTenant:      dfltTenant,
		reqPayload:      reqPayload,
		rplyPayload:     rplyPayload,
		reqProcessors:   reqProcessors,
		sessionConns:    sessionConns,
		statsConns:      statsConns,
		thresholdsConns: thresholdsConns,
		caps:            caps,
	}
}

// HTTPAgent is a handler for HTTP requests
type HTTPAgent struct {
	connMgr         *engine.ConnManager
	filterS         *engine.FilterS
	dfltTenant      string
	reqPayload      string
	rplyPayload     string
	reqProcessors   []*config.RequestProcessor
	sessionConns    []string
	statsConns      []string
	thresholdsConns []string
	caps            *engine.Caps
}

// ServeHTTP implements http.Handler interface
func (ha *HTTPAgent) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if ha.caps.IsLimited() {
		if err := ha.caps.Allocate(); err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		defer ha.caps.Deallocate()
	}
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
			utils.HTTPAgent, ha.connMgr, ha.sessionConns, ha.statsConns, ha.thresholdsConns,
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
