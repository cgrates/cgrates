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
)

const (
	MetaRadReqType     = "*radReqType"
	MetaRadAuth        = "*radAuth"
	MetaRadReplyCode   = "*radReplyCode"
	UserPasswordAVP    = "User-Password"
	CHAPPasswordAVP    = "CHAP-Password"
	MSCHAPChallengeAVP = "MS-CHAP-Challenge"
	MSCHAPResponseAVP  = "MS-CHAP-Response"
	MicrosoftVendor    = "Microsoft"
	MSCHAP2SuccessAVP  = "MS-CHAP2-Success"
)

func NewRadiusAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (ra *RadiusAgent, err error) {
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
	ra = &RadiusAgent{cgrCfg: cgrCfg, filterS: filterS, connMgr: connMgr}
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
	cgrCfg  *config.CGRConfig // reference for future config reloads
	connMgr *engine.ConnManager
	filterS *engine.FilterS
	rsAuth  *radigo.Server
	rsAcct  *radigo.Server
}

// handleAuth handles RADIUS Authorization request
func (ra *RadiusAgent) handleAuth(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues()             // populate string values in AVPs
	dcdr := newRADataProvider(req) // dcdr will provide information from request
	rpl = req.Reply()
	rpl.Code = radigo.AccessAccept
	cgrRplyNM := utils.NavigableMap{}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.NewOrderedNavigableMap()
	var processed bool
	reqVars := utils.NavigableMap{utils.RemoteHost: utils.NewNMData(req.RemoteAddr().String())}
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(dcdr, reqVars, &cgrRplyNM, rplyNM, opts,
			reqProcessor.Tenant, ra.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.filterS, nil, nil)
		agReq.Vars.Set(utils.PathItems{{Field: MetaRadReqType}}, utils.NewNMData(MetaRadAuth))
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(req, reqProcessor, agReq, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
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
	if err := radReplyAppendAttributes(rpl, rplyNM); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> err: %s, replying to message: %+v",
			utils.RadiusAgent, err.Error(), utils.ToIJSON(req)))
		return nil, err
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
	cgrRplyNM := utils.NavigableMap{}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.NewOrderedNavigableMap()
	var processed bool
	reqVars := utils.NavigableMap{utils.RemoteHost: utils.NewNMData(req.RemoteAddr().String())}
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(dcdr, reqVars, &cgrRplyNM, rplyNM, opts,
			reqProcessor.Tenant, ra.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.filterS, nil, nil)
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(req, reqProcessor, agReq, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
			break
		}
	}
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s> ignoring request: %s, ",
			utils.RadiusAgent, err.Error(), utils.ToIJSON(req)))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s",
			utils.RadiusAgent, utils.ToIJSON(req)))
		return nil, nil
	}
	if err := radReplyAppendAttributes(rpl, rplyNM); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> err: %s, replying to message: %+v",
			utils.RadiusAgent, err.Error(), utils.ToIJSON(req)))
		return nil, err
	}
	return
}

// processRequest represents one processor processing the request
func (ra *RadiusAgent) processRequest(req *radigo.Packet, reqProcessor *config.RequestProcessor,
	agReq *AgentRequest, rpl *radigo.Packet) (processed bool, err error) {
	if pass, err := ra.filterS.Pass(agReq.Tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if err = agReq.SetFields(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuthorize,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaMessage,
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone, utils.MetaRadauth} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	var cgrArgs utils.Paginator
	if reqType == utils.MetaAuthorize ||
		reqType == utils.MetaMessage ||
		reqType == utils.MetaEvent {
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(cgrEv.Opts); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> args extraction failed because <%s>",
				utils.RadiusAgent, err.Error()))
			err = nil // reset the error and continue the processing
		}
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, radius message: %s",
				utils.RadiusAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.RadiusAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
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
			cgrEv, cgrArgs, reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1AuthorizeReply)
		err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
		rply.SetMaxUsageNeeded(authArgs.GetMaxUsage)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1InitSessionReply)
		err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1InitiateSession,
			initArgs, rply)
		rply.SetMaxUsageNeeded(initArgs.InitSession)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1UpdateSessionReply)
		err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1UpdateSession,
			updateArgs, rply)
		rply.SetMaxUsageNeeded(updateArgs.UpdateSession)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		var rply string
		err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1TerminateSession,
			terminateArgs, &rply)
		if err = agReq.setCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaMessage:
		evArgs := sessions.NewV1ProcessMessageArgs(
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
			cgrEv, cgrArgs, reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1ProcessMessageReply)
		err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1ProcessMessage, evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if evArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(evArgs.Debit)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags:     reqProcessor.Flags.SliceFlags(),
			CGREvent:  cgrEv,
			Paginator: cgrArgs,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1ProcessEvent,
			evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow this method
	case utils.MetaRadauth:
		if pass, err := radauthReq(reqProcessor.Flags, req, agReq, rpl); err != nil {
			agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(err.Error()))
		} else if !pass {
			agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(utils.RadauthFailed))
		}
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.GetBool(utils.MetaCDRs) {
		var rplyCDRs string
		if err = ra.connMgr.Call(ra.cgrCfg.RadiusAgentCfg().SessionSConns, nil, utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(err.Error()))
		}
	}

	if err := agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return false, err
	}

	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, Radius reply: %s",
				utils.RadiusAgent, utils.ToIJSON(agReq.Reply)))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, Radius reply: %s",
				utils.RadiusAgent, utils.ToJSON(agReq.Reply)))
	}
	return true, nil
}

func (ra *RadiusAgent) ListenAndServe(stopChan <-chan struct{}) (err error) {
	errListen := make(chan error, 2)
	go func() {
		utils.Logger.Info(fmt.Sprintf("<%s> Start listening for auth requests on <%s>", utils.RadiusAgent, ra.cgrCfg.RadiusAgentCfg().ListenAuth))
		if err := ra.rsAuth.ListenAndServe(stopChan); err != nil {
			errListen <- err
		}
	}()
	go func() {
		utils.Logger.Info(fmt.Sprintf("<%s> Start listening for acct req on <%s>", utils.RadiusAgent, ra.cgrCfg.RadiusAgentCfg().ListenAcct))
		if err := ra.rsAcct.ListenAndServe(stopChan); err != nil {
			errListen <- err
		}
	}()
	err = <-errListen
	return
}
