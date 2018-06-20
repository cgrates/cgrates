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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewHttpAgent will construct a HTTPAgent
func NewHTTPAgent(
	sessionS rpcclient.RpcClientConnection,
	filterS *engine.FilterS, tenantCfg utils.RSRFields,
	dfltTenant, timezone, reqPayload, rplyPayload string,
	reqProcessors []*config.HttpAgntProcCfg) *HTTPAgent {
	return &HTTPAgent{sessionS: sessionS,
		dfltTenant: dfltTenant, timezone: timezone,
		reqPayload: reqPayload, rplyPayload: rplyPayload,
		reqProcessors: reqProcessors}
}

// HTTPAgent is a handler for HTTP requests
type HTTPAgent struct {
	sessionS  rpcclient.RpcClientConnection
	filterS   *engine.FilterS
	tenantCfg utils.RSRFields
	dfltTenant,
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
	agReq := newAgentRequest(dcdr, ha.tenantCfg, ha.dfltTenant, ha.filterS)
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
	if pass, err := ha.filterS.Pass(agReq.Tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if agReq.CGRRequest, err = agReq.AsNavigableMap(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := &utils.CGREvent{
		Tenant: agReq.Tenant,
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event:  agReq.CGRRequest.AsMapStringInterface(),
	}
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuth,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaEvent} {
		if reqProcessor.Flags.HasKey(typ) { // request type is identified through flags
			reqType = typ
		}
	}
	switch reqType {
	default:
		return false, errors.New("unknown request type")
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, HTTP request: %s",
				utils.HTTPAgent, reqProcessor.Id, utils.ToJSON(agReq.Request)))
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.HTTPAgent, reqProcessor.Id, utils.ToJSON(cgrEv)))
	case utils.MetaAuth:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaSuppliers),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersEventCost),
			*cgrEv)
		var authReply sessions.V1AuthorizeReply
		err = ha.sessionS.Call(utils.SessionSv1AuthorizeEvent,
			authArgs, &authReply)
		if agReq.CGRReply, err = NewCGRReply(&authReply, err); err != nil {
			return
		}
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats), *cgrEv)
		var initReply sessions.V1InitSessionReply
		err = ha.sessionS.Call(utils.SessionSv1InitiateSession,
			initArgs, &initReply)
		if agReq.CGRReply, err = NewCGRReply(&initReply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaAccounts), *cgrEv)
		var updateReply sessions.V1UpdateSessionReply
		err = ha.sessionS.Call(utils.SessionSv1UpdateSession,
			updateArgs, &updateReply)
		if agReq.CGRReply, err = NewCGRReply(&updateReply, err); err != nil {
			return
		}
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats), *cgrEv)
		var tRply string
		err = ha.sessionS.Call(utils.SessionSv1TerminateSession,
			terminateArgs, &tRply)
		if agReq.CGRReply, err = NewCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := sessions.NewV1ProcessEventArgs(
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaAttributes), *cgrEv)
		var eventRply sessions.V1ProcessEventReply
		err = ha.sessionS.Call(utils.SessionSv1ProcessEvent,
			evArgs, &eventRply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if eventRply.MaxUsage != nil {
			cgrEv.Event[utils.Usage] = *eventRply.MaxUsage // make sure the CDR reflects the debit
		}
		if agReq.CGRReply, err = NewCGRReply(&eventRply, err); err != nil {
			return
		}
	}
	if reqProcessor.Flags.HasKey(utils.MetaCDRs) &&
		utils.IsSliceMember([]string{utils.MetaTerminate, utils.MetaEvent}, reqType) {
		var rplyCDRs string
		if err = ha.sessionS.Call(utils.SessionSv1ProcessCDR,
			*cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Set([]string{utils.Error}, err.Error(), false)
		}
	}
	if nM, err := agReq.AsNavigableMap(reqProcessor.ReplyFields); err != nil {
		return false, err
	} else {
		agReq.Reply.Merge(nM)
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, HTTP reply: %s",
				utils.HTTPAgent, agReq.Reply))
	}
	return
}
