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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
	"github.com/cgrates/rpcclient"
)

const (
	MetaRadReqType   = "*radReqType"
	MetaRadAuth      = "*radAuth"
	MetaRadAcctStart = "*radAcctStart"
	MetaRadReplyCode = "*radReplyCode"
)

func NewRadiusAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	sessionS rpcclient.RpcClientConnection) (ra *RadiusAgent, err error) {
	dts := make(map[string]*radigo.Dictionary, len(cgrCfg.RadiusAgentCfg().ClientDictionaries))
	for clntID, dictPath := range cgrCfg.RadiusAgentCfg().ClientDictionaries {
		utils.Logger.Info(
			fmt.Sprintf("<%s> loading dictionary for clientID: <%s> out of path <%s>",
				utils.RadiusAgent, clntID, dictPath))
		if dts[clntID], err = radigo.NewDictionaryFromFolderWithRFC2865(dictPath); err != nil {
			return
		}
	}
	dicts := radigo.NewDictionaries(dts)
	ra = &RadiusAgent{cgrCfg: cgrCfg, filterS: filterS, sessionS: sessionS}
	secrets := radigo.NewSecrets(cgrCfg.RadiusAgentCfg().ClientSecrets)
	ra.rsAuth = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet,
		cgrCfg.RadiusAgentCfg().ListenAuth, secrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.AccessRequest: ra.handleAuth}, nil)
	ra.rsAcct = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet,
		cgrCfg.RadiusAgentCfg().ListenAcct, secrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.AccountingRequest: ra.handleAcct}, nil)
	return
}

type RadiusAgent struct {
	cgrCfg   *config.CGRConfig             // reference for future config reloads
	sessionS rpcclient.RpcClientConnection // Connection towards CGR-SessionS component
	filterS  *engine.FilterS
	rsAuth   *radigo.Server
	rsAcct   *radigo.Server
}

// handleAuth handles RADIUS Authorization request
func (ra *RadiusAgent) handleAuth(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues()             // populate string values in AVPs
	dcdr := newRADataProvider(req) // dcdr will provide information from request
	rpl = req.Reply()
	rpl.Code = radigo.AccessAccept
	var processed bool
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		agReq := newAgentRequest(dcdr, nil, nil,
			reqProcessor.Tenant, ra.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.filterS)
		agReq.Vars.Set([]string{MetaRadReqType}, utils.StringToInterface(MetaRadAuth), false, true)
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(reqProcessor, agReq, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s> ignoring request: %s",
			utils.RadiusAgent, err.Error(), utils.ToJSON(req)))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s",
			utils.RadiusAgent, utils.ToJSON(req)))
		return nil, nil
	}
	return
}

// handleAcct handles RADIUS Accounting request
// supports: Acct-Status-Type = Start, Interim-Update, Stop
func (ra *RadiusAgent) handleAcct(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues()             // populate string values in AVPs
	dcdr := newRADataProvider(req) // dcdr will provide information from request
	rpl = req.Reply()
	rpl.Code = radigo.AccountingResponse
	var processed bool
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		agReq := newAgentRequest(dcdr, nil, nil,
			reqProcessor.Tenant, ra.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.filterS)
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(reqProcessor, agReq, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s> ignoring request: %s, ",
			utils.RadiusAgent, err.Error(), utils.ToJSON(req)))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s",
			utils.RadiusAgent, utils.ToJSON(req)))
		return nil, nil
	}
	return
}

// processRequest represents one processor processing the request
func (ra *RadiusAgent) processRequest(reqProcessor *config.RARequestProcessor,
	agReq *AgentRequest, rply *radigo.Packet) (processed bool, err error) {
	if pass, err := ra.filterS.Pass(agReq.tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if agReq.CGRRequest, err = agReq.AsNavigableMap(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := agReq.CGRRequest.AsCGREvent(agReq.tenant, utils.NestingSep)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuth,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaEvent,
		utils.MetaCDRs} {
		if reqProcessor.Flags.HasKey(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, radius message: %s",
				utils.RadiusAgent, reqProcessor.Id, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.RadiusAgent, reqProcessor.Id, utils.ToJSON(cgrEv)))
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
		err = ra.sessionS.Call(utils.SessionSv1AuthorizeEvent,
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
		err = ra.sessionS.Call(utils.SessionSv1InitiateSession,
			initArgs, &initReply)
		if agReq.CGRReply, err = NewCGRReply(&initReply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaAccounts), *cgrEv)
		var updateReply sessions.V1UpdateSessionReply
		err = ra.sessionS.Call(utils.SessionSv1UpdateSession,
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
		err = ra.sessionS.Call(utils.SessionSv1TerminateSession,
			terminateArgs, &tRply)
		if agReq.CGRReply, err = NewCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := sessions.NewV1ProcessEventArgs(
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			*cgrEv)
		var eventRply sessions.V1ProcessEventReply
		err = ra.sessionS.Call(utils.SessionSv1ProcessEvent,
			evArgs, &eventRply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if eventRply.MaxUsage != nil {
			cgrEv.Event[utils.Usage] = *eventRply.MaxUsage // make sure the CDR reflects the debit
		}
		if agReq.CGRReply, err = NewCGRReply(&eventRply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow this method
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.HasKey(utils.MetaCDRs) {
		var rplyCDRs string
		if err = ra.sessionS.Call(utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Set([]string{utils.Error}, err.Error(), false, false)
		}
	}
	if nM, err := agReq.AsNavigableMap(reqProcessor.ReplyFields); err != nil {
		return false, err
	} else {
		agReq.Reply.Merge(nM)
	}
	if err := radReplyAppendAttributes(rply, agReq, reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, Radius reply: %s",
				utils.RadiusAgent, utils.ToJSON(rply)))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, Radius reply: %s",
				utils.RadiusAgent, utils.ToJSON(rply)))
	}
	return true, nil
}

func (ra *RadiusAgent) ListenAndServe() (err error) {
	var errListen chan error
	go func() {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> Start listening for auth requests on <%s>", ra.cgrCfg.RadiusAgentCfg().ListenAuth))
		if err := ra.rsAuth.ListenAndServe(); err != nil {
			errListen <- err
		}
	}()
	go func() {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> Start listening for acct req on <%s>", ra.cgrCfg.RadiusAgentCfg().ListenAcct))
		if err := ra.rsAcct.ListenAndServe(); err != nil {
			errListen <- err
		}
	}()
	err = <-errListen
	return
}
