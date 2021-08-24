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
	"github.com/cgrates/cgrates/sessions"
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
		lclProcessed, err := ha.processRequest(reqProcessor, agReq)
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

// processRequest represents one processor processing the request
func (ha *HTTPAgent) processRequest(reqProcessor *config.RequestProcessor,
	agReq *AgentRequest) (processed bool, err error) {
	if pass, err := ha.filterS.Pass(context.TODO(), agReq.Tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if err = agReq.SetFields(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuthorize,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaMessage,
		utils.MetaCDRs, utils.MetaEvent, utils.MetaEmpty} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}

	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, http message: %s",
				utils.HTTPAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.HTTPAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuthorize:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaRoutes),
			reqProcessor.Flags.Has(utils.MetaRoutesIgnoreErrors),
			reqProcessor.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1AuthorizeReply)
		err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
		rply.SetMaxUsageNeeded(authArgs.GetMaxUsage)
		agReq.setCGRReply(rply, err)
	case utils.MetaInitiate:
		rply := new(sessions.V1InitSessionReply)
		err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1InitiateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesInitiate))
		agReq.setCGRReply(rply, err)
	case utils.MetaUpdate:
		rply := new(sessions.V1UpdateSessionReply)
		err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1UpdateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesUpdate))
		agReq.setCGRReply(rply, err)
	case utils.MetaTerminate:
		var rply string
		err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1TerminateSession,
			cgrEv, &rply)
		agReq.setCGRReply(nil, err)
	case utils.MetaMessage:
		rply := new(sessions.V1ProcessMessageReply)
		err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1ProcessMessage,
			cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else
		messageS := utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesMessage)
		if messageS {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(messageS)
		agReq.setCGRReply(nil, err)
	case utils.MetaEvent:
		rply := new(sessions.V1ProcessEventReply)
		err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1ProcessEvent,
			cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else
		// if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
		// cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		// }
		agReq.setCGRReply(rply, err)
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.GetBool(utils.MetaCDRs) &&
		!reqProcessor.Flags.Has(utils.MetaDryRun) {
		var rplyCDRs string
		if err = ha.connMgr.Call(context.TODO(), ha.sessionConns, utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Map[utils.Error] = utils.NewLeafNode(err.Error())
		}
	}
	if err := agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, HTTP reply: %s",
				utils.HTTPAgent, agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, HTTP reply: %s",
				utils.HTTPAgent, agReq.Reply))
	}
	return true, nil
}
