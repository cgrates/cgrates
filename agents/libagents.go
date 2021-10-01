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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func processRequest(ctx *context.Context, reqProcessor *config.RequestProcessor, agReq *AgentRequest,
	agentName string, connMgr *engine.ConnManager, sessionsConns []string,
	filterS *engine.FilterS) (processed bool, err error) {
	if pass, err := filterS.Pass(ctx, agReq.Tenant,
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
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, diameter message: %s",
				agentName, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, DiameterMessage: %s",
				agentName, reqProcessor.ID, agReq.Request.String()))
	case utils.MetaAuthorize:
		rply := new(sessions.V1AuthorizeReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1AuthorizeEvent,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsAccountS))
		agReq.setCGRReply(rply, err)
	case utils.MetaInitiate:
		rply := new(sessions.V1InitSessionReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1InitiateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesInitiate))
		agReq.setCGRReply(rply, err)
	case utils.MetaUpdate:
		rply := new(sessions.V1UpdateSessionReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1UpdateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesUpdate))
		agReq.setCGRReply(rply, err)
	case utils.MetaTerminate:
		var rply string
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1TerminateSession,
			cgrEv, &rply)
		agReq.setCGRReply(nil, err)
	case utils.MetaMessage:
		rply := new(sessions.V1ProcessMessageReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1ProcessMessage,
			cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// 	cgrEv.Event[utils.Usage] = 0 // avoid further debits
		messageS := utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesMessage)
		if messageS {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(messageS)
		agReq.setCGRReply(rply, err)
	case utils.MetaEvent:
		rply := new(sessions.V1ProcessEventReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1ProcessEvent,
			cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
		// cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		// }
		agReq.setCGRReply(rply, err)
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.GetBool(utils.MetaCDRs) &&
		!reqProcessor.Flags.Has(utils.MetaDryRun) {
		var rplyCDRs string
		if err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Map[utils.Error] = utils.NewLeafNode(err.Error())
		}
	}
	if err = agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, Diameter reply: %s",
				agentName, agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, Diameter reply: %s",
				agentName, agReq.Reply))
	}
	return true, nil
}
