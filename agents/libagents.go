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
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func processRequest(ctx *context.Context, reqProcessor *config.RequestProcessor,
	agReq *AgentRequest, agentName string, connMgr *engine.ConnManager,
	sessionsConns, statsConns, thConns []string,
	filterS *engine.FilterS) (_ bool, err error) {
	startTime := time.Now()
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
			fmt.Sprintf("<%s> LOG, processorID: %s, %s message: %s",
				agentName, reqProcessor.ID, strings.ToLower(agentName[:len(agentName)-5]), agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, %sMessage: %s",
				agentName, reqProcessor.ID, agentName[:len(agentName)-5], agReq.Request.String()))
	case utils.MetaAuthorize:
		rply := new(sessions.V1AuthorizeReply)
		sessions.ApplyFlags(reqType, reqProcessor.Flags, cgrEv.APIOpts)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1AuthorizeEvent,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.MetaAccounts))
		agReq.setCGRReply(rply, err)
	case utils.MetaInitiate:
		rply := new(sessions.V1InitSessionReply)
		sessions.ApplyFlags(reqType, reqProcessor.Flags, cgrEv.APIOpts)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1InitiateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.MetaInitiate))
		agReq.setCGRReply(rply, err)
	case utils.MetaUpdate:
		rply := new(sessions.V1UpdateSessionReply)
		sessions.ApplyFlags(reqType, reqProcessor.Flags, cgrEv.APIOpts)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1UpdateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.MetaUpdate))
		agReq.setCGRReply(rply, err)
	case utils.MetaTerminate:
		var rply string
		sessions.ApplyFlags(reqType, reqProcessor.Flags, cgrEv.APIOpts)
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
	endTime := time.Now()
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, %s reply: %s",
				agentName, agentName[:len(agentName)-5], agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, %s reply: %s",
				agentName, agentName[:len(agentName)-5], agReq.Reply))
	}
	if reqProcessor.Flags.Has(utils.MetaDryRun) {
		return true, nil
	}

	var rawStatIDs, rawThIDs string
	switch agentName {
	case utils.DiameterAgent:
		rawStatIDs = reqProcessor.Flags.ParamValue(utils.MetaDAStats)
		rawThIDs = reqProcessor.Flags.ParamValue(utils.MetaDAThresholds)
	case utils.HTTPAgent:
		rawStatIDs = reqProcessor.Flags.ParamValue(utils.MetaHAStats)
		rawThIDs = reqProcessor.Flags.ParamValue(utils.MetaHAThresholds)
	case utils.DNSAgent:
		rawStatIDs = reqProcessor.Flags.ParamValue(utils.MetaDNSStats)
		rawThIDs = reqProcessor.Flags.ParamValue(utils.MetaDNSThresholds)
	}

	// Return early if nothing to process.
	if rawStatIDs == "" && rawThIDs == "" {
		return true, nil
	}

	// Clone is needed to prevent data races if requests are sent
	// asynchronously.
	ev := cgrEv.Clone()

	ev.Event[utils.StartTime] = startTime
	ev.Event[utils.EndTime] = endTime
	ev.Event[utils.ProcessingTime] = endTime.Sub(startTime)
	ev.Event[utils.Source] = agentName
	ev.APIOpts[utils.MetaEventType] = utils.ProcessTime

	if rawStatIDs != "" {
		statIDs := strings.Split(rawStatIDs, utils.ANDSep)
		ev.APIOpts[utils.OptsStatsProfileIDs] = statIDs
		var reply []string
		if err := connMgr.Call(ctx, statsConns, utils.StatSv1ProcessEvent,
			ev, &reply); err != nil {
			return false, fmt.Errorf("failed to process %s event in %s: %v",
				agentName, utils.StatS, err)
		}
		// NOTE: ProfileIDs APIOpts key persists for the ThresholdS request,
		// although it would be ignored. Might want to delete it.
	}
	if rawThIDs != "" {
		thIDs := strings.Split(rawThIDs, utils.ANDSep)
		ev.APIOpts[utils.OptsThresholdsProfileIDs] = thIDs
		var reply []string
		if err := connMgr.Call(ctx, thConns, utils.ThresholdSv1ProcessEvent,
			ev, &reply); err != nil {
			return false, fmt.Errorf("failed to process %s event in %s: %v",
				agentName, utils.ThresholdS, err)
		}
	}
	return true, nil
}
