/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	if pass, err := filterS.Pass(agReq.Tenant,
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
	var cgrArgs utils.Paginator
	if reqType == utils.MetaAuthorize || reqType == utils.MetaMessage || reqType == utils.MetaEvent {
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(cgrEv.APIOpts); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> args extraction failed because <%s>",
				agentName, err.Error()))
			err = nil // reset the error and continue the processing
		}
	}

	if reqProcessor.Flags.Has(utils.MetaLog) || reqType == utils.MetaDryRun {
		logPrefix := "LOG"
		if reqType == utils.MetaDryRun {
			logPrefix = "DRY_RUN"
		}
		utils.Logger.Info(
			fmt.Sprintf("<%s> %s, processorID: %s, %s message: %s",
				agentName, logPrefix, reqProcessor.ID, strings.ToLower(agentName[:len(agentName)-5]), agReq.Request.String()))
		utils.Logger.Info(
			fmt.Sprintf("<%s> %s, processorID: %s, CGREvent: %s",
				agentName, logPrefix, reqProcessor.ID, utils.ToIJSON(cgrEv)))
	}

	replyState := utils.OK
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun: // do nothing on CGRateS side, logging handled above
	case utils.MetaAuthorize:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaIPs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaRoutes),
			reqProcessor.Flags.Has(utils.MetaRoutesIgnoreErrors),
			reqProcessor.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs,
			reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1AuthorizeReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
		if err != nil {
			replyState = utils.ErrReplyStateAuthorize
		}
		rply.SetMaxUsageNeeded(authArgs.GetMaxUsage)
		agReq.setCGRReply(rply, err)
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.GetBool(utils.MetaIPs),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1InitSessionReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1InitiateSession,
			initArgs, rply)
		if err != nil {
			replyState = utils.ErrReplyStateInitiate
		}
		rply.SetMaxUsageNeeded(initArgs.InitSession)
		agReq.setCGRReply(rply, err)
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1UpdateSessionReply)
		rply.SetMaxUsageNeeded(updateArgs.UpdateSession)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1UpdateSession,
			updateArgs, rply)
		if err != nil {
			replyState = utils.ErrReplyStateUpdate
		}
		agReq.setCGRReply(rply, err)
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.GetBool(utils.MetaIPs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		var rply string
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1TerminateSession,
			terminateArgs, &rply)
		if err != nil {
			replyState = utils.ErrReplyStateTerminate
		}
		agReq.setCGRReply(nil, err)
	case utils.MetaMessage:
		msgArgs := sessions.NewV1ProcessMessageArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.GetBool(utils.MetaIPs),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaRoutes),
			reqProcessor.Flags.Has(utils.MetaRoutesIgnoreErrors),
			reqProcessor.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs,
			reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1ProcessMessageReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1ProcessMessage,
			msgArgs, rply)
		if err != nil {
			replyState = utils.ErrReplyStateMessage
		}
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if msgArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(msgArgs.Debit)
		agReq.setCGRReply(rply, err)
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags:     reqProcessor.Flags.SliceFlags(),
			Paginator: cgrArgs,
			CGREvent:  cgrEv,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = connMgr.Call(ctx, sessionsConns, utils.SessionSv1ProcessEvent,
			evArgs, rply)
		if err != nil {
			replyState = utils.ErrReplyStateEvent
		}
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
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
			if replyState == utils.OK {
				replyState = utils.ErrReplyStateCDRs
			} else {
				replyState += ";" + utils.ErrReplyStateCDRs
			}
		}
	}
	if err = agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return
	}
	endTime := time.Now()
	if reqProcessor.Flags.Has(utils.MetaLog) || reqType == utils.MetaDryRun {
		logPrefix := "LOG"
		if reqType == utils.MetaDryRun {
			logPrefix = "DRY_RUN"
		}
		utils.Logger.Info(
			fmt.Sprintf("<%s> %s, %s reply: %s",
				agentName, logPrefix, agentName[:len(agentName)-5], agReq.Reply))
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

	ev := &utils.CGREvent{
		Tenant: cgrEv.Tenant,
		ID:     utils.GenUUID(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]any{
			utils.ReplyState:         replyState,
			utils.StartTime:          startTime,
			utils.EndTime:            endTime,
			utils.ProcessingTime:     endTime.Sub(startTime),
			utils.Source:             agentName,
			utils.RequestProcessorID: reqProcessor.ID,
		},
		APIOpts: map[string]any{
			utils.MetaEventType: utils.EventPerformanceReport,
		},
	}

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
