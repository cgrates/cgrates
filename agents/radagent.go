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
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
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

func NewRadiusAgent(cfg *config.CGRConfig, fltrs *engine.FilterS,
	cm *engine.ConnManager) (*RadiusAgent, error) {
	ra := &RadiusAgent{
		cfg:   cfg,
		fltrs: fltrs,
		cm:    cm,
	}

	raCfg := cfg.RadiusAgentCfg()
	dts := make(map[string]*radigo.Dictionary, len(raCfg.ClientDictionaries))
	for clntID, dictPath := range raCfg.ClientDictionaries {
		utils.Logger.Info(fmt.Sprintf(
			"<%s> loading dictionary for clientID %q out of path %q",
			utils.RadiusAgent, clntID, dictPath))
		dt, err := radigo.NewDictionaryFromFoldersWithRFC2865(dictPath)
		if err != nil {
			return nil, err
		}
		dts[clntID] = dt
	}
	dicts := radigo.NewDictionaries(dts)
	secrets := radigo.NewSecrets(raCfg.ClientSecrets)

	ra.rsAuth = make(map[string]*radigo.Server, len(raCfg.Listeners))
	ra.rsAcct = make(map[string]*radigo.Server, len(raCfg.Listeners))
	for i := range raCfg.Listeners {
		net := raCfg.Listeners[i].Network
		authAddr := raCfg.Listeners[i].AuthAddr
		ra.rsAuth[net+"://"+authAddr] = radigo.NewServer(net, authAddr, secrets, dicts,
			map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
				radigo.AccessRequest: ra.handleAuth,
			}, nil, utils.Logger)
		acctAddr := raCfg.Listeners[i].AcctAddr
		ra.rsAcct[net+"://"+acctAddr] = radigo.NewServer(net, acctAddr, secrets, dicts,
			map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
				radigo.AccountingRequest: ra.handleAcct,
			}, nil, utils.Logger)
	}

	return ra, nil
}

type RadiusAgent struct {
	mu     sync.RWMutex
	cfg    *config.CGRConfig // reference for future config reloads
	cm     *engine.ConnManager
	fltrs  *engine.FilterS
	rsAuth map[string]*radigo.Server
	rsAcct map[string]*radigo.Server
	wg     sync.WaitGroup
}

// handleAuth handles RADIUS Authorization request
func (ra *RadiusAgent) handleAuth(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues()             // populate string values in AVPs
	dcdr := newRADataProvider(req) // dcdr will provide information from request
	rpl = req.Reply()
	rpl.Code = radigo.AccessAccept
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.MapStorage{}
	var processed bool
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.RemoteHost: utils.NewLeafNode(req.RemoteAddr().String())}}
	for _, reqProcessor := range ra.cfg.RadiusAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(dcdr, reqVars, cgrRplyNM, rplyNM, opts,
			reqProcessor.Tenant, ra.cfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.fltrs, nil)
		agReq.Vars.Map[MetaRadReqType] = utils.NewLeafNode(MetaRadAuth)
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(req, reqProcessor, agReq, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue) {
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
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.MapStorage{}
	var processed bool
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.RemoteHost: utils.NewLeafNode(req.RemoteAddr().String())}}
	for _, reqProcessor := range ra.cfg.RadiusAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(dcdr, reqVars, cgrRplyNM, rplyNM, opts,
			reqProcessor.Tenant, ra.cfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.fltrs, nil)
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(req, reqProcessor, agReq, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue) {
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
	startTime := time.Now()
	if pass, err := ra.fltrs.Pass(context.TODO(), agReq.Tenant,
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
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone, utils.MetaRadauth} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
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
		rply := new(sessions.V1AuthorizeReply)
		err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1AuthorizeEvent,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.MetaAccounts))
		agReq.setCGRReply(rply, err)
	case utils.MetaInitiate:
		rply := new(sessions.V1InitSessionReply)
		err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1InitiateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesInitiate))
		agReq.setCGRReply(rply, err)
	case utils.MetaUpdate:
		rply := new(sessions.V1UpdateSessionReply)
		err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1UpdateSession,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesUpdate))
		agReq.setCGRReply(rply, err)
	case utils.MetaTerminate:
		var rply string
		err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1TerminateSession,
			cgrEv, &rply)
		agReq.setCGRReply(nil, err)
	case utils.MetaMessage:
		rply := new(sessions.V1ProcessMessageReply)
		err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1ProcessMessage, cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else
		messageS := utils.OptAsBool(cgrEv.APIOpts, utils.OptsSesMessage)
		if messageS {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(messageS)
		agReq.setCGRReply(rply, err)
	case utils.MetaEvent:
		rply := new(sessions.V1ProcessEventReply)
		err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1ProcessEvent,
			cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else
		// if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
		// cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		// }
		agReq.setCGRReply(rply, err)
	case utils.MetaCDRs: // allow this method
	case utils.MetaRadauth:
		if pass, err := radauthReq(reqProcessor.Flags, req, agReq, rpl); err != nil {
			agReq.CGRReply.Map[utils.Error] = utils.NewLeafNode(err.Error())
		} else if !pass {
			agReq.CGRReply.Map[utils.Error] = utils.NewLeafNode(utils.RadauthFailed)
		}
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.GetBool(utils.MetaCDRs) {
		var rplyCDRs string
		if err = ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Map[utils.Error] = utils.NewLeafNode(err.Error())
		}
	}

	if err := agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	endTime := time.Now()
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, Radius reply: %s",
				utils.RadiusAgent, agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, Radius reply: %s",
				utils.RadiusAgent, agReq.Reply))
	}
	if reqProcessor.Flags.Has(utils.MetaDryRun) {
		return true, nil
	}

	rawStatIDs := reqProcessor.Flags.ParamValue(utils.MetaRAStats)
	rawThIDs := reqProcessor.Flags.ParamValue(utils.MetaRAThresholds)

	// Early return if nothing to process.
	if rawStatIDs == "" && rawThIDs == "" {
		return true, nil
	}

	// Clone is needed to prevent data races if requests are sent
	// asynchronously.
	ev := cgrEv.Clone()

	ev.Event[utils.StartTime] = startTime
	ev.Event[utils.EndTime] = endTime
	ev.Event[utils.ProcessingTime] = endTime.Sub(startTime)
	ev.Event[utils.Source] = utils.RadiusAgent
	ev.APIOpts[utils.MetaEventType] = utils.ProcessTime

	if rawStatIDs != "" {
		statIDs := strings.Split(rawStatIDs, utils.ANDSep)
		ev.APIOpts[utils.OptsStatsProfileIDs] = statIDs
		var reply []string
		if err := ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().StatSConns,
			utils.StatSv1ProcessEvent, ev, &reply); err != nil {
			return false, fmt.Errorf("failed to process %s event in %s: %v",
				utils.RadiusAgent, utils.StatS, err)
		}
		// NOTE: ProfileIDs APIOpts key persists for the ThresholdS request,
		// although it would be ignored. Might want to delete it.
	}
	if rawThIDs != "" {
		thIDs := strings.Split(rawThIDs, utils.ANDSep)
		ev.APIOpts[utils.OptsThresholdsProfileIDs] = thIDs
		var reply []string
		if err := ra.cm.Call(context.TODO(), ra.cfg.RadiusAgentCfg().ThresholdSConns,
			utils.ThresholdSv1ProcessEvent, ev, &reply); err != nil {
			return false, fmt.Errorf("failed to process %s event in %s: %v",
				utils.RadiusAgent, utils.ThresholdS, err)
		}
	}
	return true, nil
}

func (ra *RadiusAgent) ListenAndServe(stopChan <-chan struct{}) (err error) {
	errListen := make(chan error, 2)
	for uri, server := range ra.rsAuth {
		ra.wg.Add(1)
		go func(srv *radigo.Server, uri string) {
			defer ra.wg.Done()
			utils.Logger.Info(fmt.Sprintf("<%s> Start listening for auth requests on <%s>", utils.RadiusAgent, uri))
			if err := srv.ListenAndServe(stopChan); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> error <%v>, on ListenAndServe <%s>",
					utils.RadiusAgent, err, uri))
				if strings.Contains(err.Error(), "address already in use") {
					return
				}
				errListen <- err
			}
		}(server, uri)
	}
	for uri, server := range ra.rsAcct {
		ra.wg.Add(1)
		go func(srv *radigo.Server, uri string) {
			defer ra.wg.Done()
			utils.Logger.Info(fmt.Sprintf("<%s> Start listening for acct requests on <%s>", utils.RadiusAgent, uri))
			if err := srv.ListenAndServe(stopChan); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> error <%v>, on ListenAndServe <%s>",
					utils.RadiusAgent, err, uri))
				if strings.Contains(err.Error(), "address already in use") {
					return
				}
				errListen <- err
			}
		}(server, uri)
	}
	err = <-errListen
	return
}

func (ra *RadiusAgent) Wait() { ra.wg.Wait() }
