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
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc"
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
	MetaRadAccount     = "*radAccount"
	MetaRadReqCode     = "*radReqCode"
	MetaRadReplyCode   = "*radReplyCode"
	UserPasswordAVP    = "User-Password"
	CHAPPasswordAVP    = "CHAP-Password"
	MSCHAPChallengeAVP = "MS-CHAP-Challenge"
	MSCHAPResponseAVP  = "MS-CHAP-Response"
	MicrosoftVendor    = "Microsoft"
	MSCHAP2SuccessAVP  = "MS-CHAP2-Success"
)

func NewRadiusAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS, connMgr *engine.ConnManager,
	caps *engine.Caps) (*RadiusAgent, error) {
	radAgent := &RadiusAgent{
		cgrCfg:  cgrCfg,
		filterS: filterS,
		connMgr: connMgr,
		caps:    caps,
	}

	// Register RadiusAgent methods whose names start with "V1" under the "AgentV1" object name.
	srv, err := birpc.NewServiceWithMethodsRename(radAgent, utils.AgentV1, true, func(oldFn string) (newFn string) {
		return strings.TrimPrefix(oldFn, "V1")
	})
	if err != nil {
		return nil, err
	}
	radAgent.ctx = context.WithClient(context.TODO(), srv)

	radAgentCfg := cgrCfg.RadiusAgentCfg()
	dts := make(map[string]*radigo.Dictionary, len(radAgentCfg.ClientDictionaries))
	for clntID, dictPath := range radAgentCfg.ClientDictionaries {
		utils.Logger.Info(fmt.Sprintf(
			"<%s> loading dictionary for clientID: <%s> out of path <%s>",
			utils.RadiusAgent, clntID, dictPath))
		if dts[clntID], err = radigo.NewDictionaryFromFoldersWithRFC2865(dictPath); err != nil {
			return nil, err
		}
	}
	dicts := radigo.NewDictionaries(dts)
	secrets := radigo.NewSecrets(radAgentCfg.ClientSecrets)
	radAgent.dacCfg = newRadiusDAClientCfg(dicts, secrets, radAgentCfg)
	radAgent.rsAuth = make(map[string]*radigo.Server, len(radAgentCfg.Listeners))
	radAgent.rsAcct = make(map[string]*radigo.Server, len(radAgentCfg.Listeners))
	for i := range radAgentCfg.Listeners {
		net := radAgentCfg.Listeners[i].Network
		authAddr := radAgentCfg.Listeners[i].AuthAddr
		radAgent.rsAuth[net+"://"+authAddr] = radigo.NewServer(net, authAddr, secrets, dicts,
			map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
				radigo.AccessRequest: radAgent.handleAuth,
				radigo.StatusServer:  radAgent.handleAuth,
			}, nil, utils.Logger)
		acctAddr := radAgentCfg.Listeners[i].AcctAddr
		radAgent.rsAcct[net+"://"+acctAddr] = radigo.NewServer(net, acctAddr, secrets, dicts,
			map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
				radigo.AccountingRequest: radAgent.handleAcct,
				radigo.StatusServer:      radAgent.handleAcct,
			}, nil, utils.Logger)
	}
	return radAgent, nil
}

type RadiusAgent struct {
	sync.RWMutex
	cgrCfg  *config.CGRConfig // reference for future config reloads
	connMgr *engine.ConnManager
	caps    *engine.Caps
	filterS *engine.FilterS
	rsAuth  map[string]*radigo.Server
	rsAcct  map[string]*radigo.Server
	dacCfg  radiusDAClientCfg
	ctx     *context.Context
	sync.WaitGroup
}

// radiusDAClientCfg holds the dictionaries and secrets necessary for initializing Dynamic Authorization Clients in RADIUS (only for
// the clients mentioned in the client_da_addresses map).
// This configuration enables the RadiusAgent to send server-initiated actions, such as Disconnect Requests and CoA requests,
// to manage ongoing user sessions dynamically.
type radiusDAClientCfg struct {
	dicts   *radigo.Dictionaries
	secrets *radigo.Secrets
}

// newRadiusDAClientCfg is a constructor for the radiusDAClientCfg type.
func newRadiusDAClientCfg(dicts *radigo.Dictionaries, secrets *radigo.Secrets,
	radAgentCfg *config.RadiusAgentCfg) radiusDAClientCfg {
	dacDicts := make(map[string]*radigo.Dictionary, len(radAgentCfg.ClientDaAddresses))
	dacSecrets := make(map[string]string, len(radAgentCfg.ClientDaAddresses))
	for client := range radAgentCfg.ClientDaAddresses {
		dacDicts[client] = dicts.GetInstance(client)
		dacSecrets[client] = secrets.GetSecret(client)
	}
	var rdac radiusDAClientCfg
	if len(dacDicts) != 0 {
		rdac.dicts = radigo.NewDictionaries(dacDicts)
	}
	if len(dacSecrets) != 0 {
		rdac.secrets = radigo.NewSecrets(dacSecrets)
	}
	return rdac
}

// handleAuth handles RADIUS Authorization request
func (ra *RadiusAgent) handleAuth(reqPacket *radigo.Packet) (*radigo.Packet, error) {
	if ra.caps.IsLimited() {
		if err := ra.caps.Allocate(); err != nil {
			return reqPacket, err
		}
		defer ra.caps.Deallocate()
	}
	reqPacket.SetAVPValues() // populate string values in AVPs
	replyPacket := reqPacket.Reply()
	replyPacket.Code = radigo.AccessAccept
	cgrReplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	replyNM := utils.NewOrderedNavigableMap()
	opts := utils.MapStorage{}

	varsDataNode := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.RemoteHost: utils.NewLeafNode(reqPacket.RemoteAddr().String()),
			MetaRadReqCode:   utils.NewLeafNode(reqPacket.Code.String()),
			MetaRadReqType:   utils.NewLeafNode(MetaRadAuth),
		},
	}
	radDP := newRADataProvider(reqPacket)
	var processed bool
	var processReqErr error
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(radDP, varsDataNode, cgrReplyNM, replyNM, opts,
			reqProcessor.Tenant, ra.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			ra.filterS, nil)
		var lclProcessed bool
		if lclProcessed, processReqErr = ra.processRequest(reqPacket, reqProcessor, agReq, replyPacket); lclProcessed {
			processed = lclProcessed
		}
		if processReqErr != nil || (lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
			break
		}
	}
	if processReqErr != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: <%v> ignoring request: %s",
			utils.RadiusAgent, processReqErr, utils.ToJSON(reqPacket)))
		return nil, nil
	} else if !processed {
		utils.Logger.Warning(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s",
			utils.RadiusAgent, utils.ToJSON(reqPacket)))
		return nil, nil
	}
	if err := radAppendAttributes(replyPacket, replyNM); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> err: %v, replying to message: %+v",
			utils.RadiusAgent, err, utils.ToIJSON(reqPacket)))
		return nil, err
	}
	return replyPacket, nil
}

// handleAcct processes RADIUS Accounting requests and generates a reply.
// It supports Acct-Status-Type values: Start, Interim-Update, Stop.
func (ra *RadiusAgent) handleAcct(reqPacket *radigo.Packet) (*radigo.Packet, error) {
	if ra.caps.IsLimited() {
		if err := ra.caps.Allocate(); err != nil {
			return nil, err
		}
		defer ra.caps.Deallocate()
	}
	reqPacket.SetAVPValues() // populate string values in AVPs
	replyPacket := reqPacket.Reply()
	replyPacket.Code = radigo.AccountingResponse
	cgrRplyNM := &utils.DataNode{
		Type: utils.NMMapType,
		Map:  map[string]*utils.DataNode{},
	}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.MapStorage{}

	remoteAddr := reqPacket.RemoteAddr().String()
	varsDataNode := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.RemoteHost: utils.NewLeafNode(remoteAddr),
			MetaRadReqType:   utils.NewLeafNode(MetaRadAccount),
			MetaRadReqCode:   utils.NewLeafNode(reqPacket.Code.String()),
		},
	}

	radDP := newRADataProvider(reqPacket)
	radAgentCfg := ra.cgrCfg.RadiusAgentCfg()
	var processed bool
	for _, reqProcessor := range radAgentCfg.RequestProcessors {
		agReq := NewAgentRequest(
			radDP, varsDataNode, cgrRplyNM, rplyNM, opts,
			reqProcessor.Tenant, ra.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(
				reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone,
			),
			ra.filterS, nil,
		)
		lclProcessed, err := ra.processRequest(reqPacket,
			reqProcessor, agReq, replyPacket)
		if err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> error: <%v> ignoring request: %s",
					utils.RadiusAgent, err, utils.ToJSON(reqPacket)))
			return nil, nil
		}
		if lclProcessed {
			processed = true
			if !reqProcessor.Flags.GetBool(utils.MetaContinue) {
				break
			}
		}
	}
	if !processed {
		utils.Logger.Err(
			fmt.Sprintf("<%s> no request processor enabled, ignoring request %s",
				utils.RadiusAgent, utils.ToIJSON(reqPacket)))
		return nil, nil
	}

	// Cache the RADIUS Packet for future CoA/Disconnect Requests.
	if cacheKeyTpl := radAgentCfg.RequestsCacheKey; cacheKeyTpl != nil {
		err := cacheRadiusPacket(reqPacket, remoteAddr, radAgentCfg,
			utils.MapStorage{
				utils.MetaReq:  radDP,
				utils.MetaVars: varsDataNode,
			})
		if err != nil {
			return nil, err
		}
	}

	if err := radAppendAttributes(replyPacket, rplyNM); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> err: %v, replying to message: %s",
			utils.RadiusAgent, err, utils.ToJSON(reqPacket)))
		return nil, err
	}
	return replyPacket, nil
}

// cacheRadiusPacket caches a RADIUS packet if there are client options found for its source address.
func cacheRadiusPacket(packet *radigo.Packet, address string, cfg *config.RadiusAgentCfg,
	dp utils.DataProvider) error {
	// Match address against configured client DA addresses.
	if _, _, err := daRequestAddress(address, cfg.ClientDaAddresses); err != nil {
		return nil // Address did not match any client; proceed without caching.
	}
	cacheKey, err := cfg.RequestsCacheKey.ParseDataProvider(dp)
	if err != nil {
		return fmt.Errorf("failed to parse the RADIUS packet cache key: %w", err)
	}
	if err = engine.Cache.Set(utils.CacheRadiusPackets, cacheKey, packet,
		nil, true, utils.NonTransactional); err != nil {
		return fmt.Errorf("failed to cache RADIUS packet: %w", err)
	}
	return nil
}

// processRequest represents one processor processing the request
func (ra *RadiusAgent) processRequest(req *radigo.Packet, reqProcessor *config.RequestProcessor,
	agReq *AgentRequest, rpl *radigo.Packet) (processed bool, err error) {
	startTime := time.Now()
	if pass, err := ra.filterS.Pass(agReq.Tenant,
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
	var cgrArgs utils.Paginator
	if reqType == utils.MetaAuthorize ||
		reqType == utils.MetaMessage ||
		reqType == utils.MetaEvent {
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(cgrEv.APIOpts); err != nil {
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
		err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
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
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1InitSessionReply)
		err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1InitiateSession,
			initArgs, rply)
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
		err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1UpdateSession,
			updateArgs, rply)
		rply.SetMaxUsageNeeded(updateArgs.UpdateSession)
		agReq.setCGRReply(rply, err)
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
		err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns, utils.SessionSv1TerminateSession,
			terminateArgs, &rply)
		agReq.setCGRReply(nil, err)
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
		err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns,
			utils.SessionSv1ProcessMessage, evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if evArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(evArgs.Debit)
		agReq.setCGRReply(rply, err)
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags:     reqProcessor.Flags.SliceFlags(),
			CGREvent:  cgrEv,
			Paginator: cgrArgs,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns,
			utils.SessionSv1ProcessEvent, evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
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
		if err = ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().SessionSConns,
			utils.SessionSv1ProcessCDR, cgrEv, &rplyCDRs); err != nil {
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
		if err := ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().StatSConns,
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
		if err := ra.connMgr.Call(ra.ctx, ra.cgrCfg.RadiusAgentCfg().ThresholdSConns,
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
		ra.Add(1)
		go func(srv *radigo.Server, uri string) {
			defer ra.Done()
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
		ra.Add(1)
		go func(srv *radigo.Server, uri string) {
			defer ra.Done()
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

// V1DisconnectPeer is needed to satisfy the sessions.BiRPClient interface
func (*RadiusAgent) V1DisconnectPeer(_ *context.Context, _ *utils.DPRArgs, _ *string) error {
	return utils.ErrNotImplemented
}

// V1GetActiveSessionIDs is needed to satisfy the sessions.BiRPClient interface
func (*RadiusAgent) V1GetActiveSessionIDs(_ *context.Context, _ string, _ *[]*sessions.SessionID) error {
	return utils.ErrNotImplemented
}

// V1DisconnectSession remotely disconnects a session by making use of the RADIUS Disconnect Message functionality.
func (ra *RadiusAgent) V1DisconnectSession(_ *context.Context, cgrEv utils.CGREvent, reply *string) error {
	ifaceOriginID, has := cgrEv.Event[utils.OriginID]
	if !has {
		return utils.NewErrMandatoryIeMissing(utils.OriginID)
	}
	originID := utils.IfaceAsString(ifaceOriginID)

	dmrTpl := ra.cgrCfg.RadiusAgentCfg().DMRTemplate
	if optTpl, err := cgrEv.OptAsString(utils.MetaRadDMRTemplate); err == nil {
		dmrTpl = optTpl
	}
	if _, found := ra.cgrCfg.TemplatesCfg()[dmrTpl]; !found {
		return fmt.Errorf("%w: DMR Template %s", utils.ErrNotFound, dmrTpl)
	}

	replyCode, err := ra.sendRadDaReq(radigo.DisconnectRequest, dmrTpl,
		originID, utils.MapStorage(cgrEv.Event), nil)
	if err != nil {
		return err
	}
	switch replyCode {
	case radigo.DisconnectACK:
		*reply = utils.OK
	case radigo.DisconnectNAK:
		return errors.New("received DisconnectNAK from RADIUS client")
	default:
		return errors.New("unexpected reply code")
	}
	return nil
}

// V1AlterSession updates session authorization using RADIUS CoA functionality.
func (ra *RadiusAgent) V1AlterSession(_ *context.Context, cgrEv utils.CGREvent, reply *string) error {
	originID, err := cgrEv.FieldAsString(utils.OriginID)
	if err != nil {
		return fmt.Errorf("could not retrieve OriginID: %w", err)
	}
	if originID == "" {
		return utils.NewErrMandatoryIeMissing(utils.OriginID)
	}
	coaTpl := ra.cgrCfg.RadiusAgentCfg().CoATemplate
	if optTpl, err := cgrEv.OptAsString(utils.MetaRadCoATemplate); err == nil {
		coaTpl = optTpl
	}
	if _, found := ra.cgrCfg.TemplatesCfg()[coaTpl]; !found {
		return fmt.Errorf("%w: CoA Template %s", utils.ErrNotFound, coaTpl)
	}

	replyCode, err := ra.sendRadDaReq(radigo.CoARequest, coaTpl,
		originID, utils.MapStorage(cgrEv.Event), nil)
	if err != nil {
		return err
	}
	switch replyCode {
	case radigo.CoAACK:
		*reply = utils.OK
	case radigo.CoANAK:
		return errors.New("received CoANAK from RADIUS client")
	default:
		return errors.New("unexpected reply code")
	}
	return nil
}

// sendRadDaReq prepares and sends a Radius CoA/Disconnect Request and returns the reply code or an error.
func (ra *RadiusAgent) sendRadDaReq(requestType radigo.PacketCode, requestTemplate, sessionID string,
	requestEv utils.DataProvider, requestVars *utils.DataNode) (radigo.PacketCode, error) {
	cachedPacket, has := engine.Cache.Get(utils.CacheRadiusPackets, sessionID)
	if !has {
		return 0, fmt.Errorf("failed to retrieve packet from cache: %w", utils.ErrNotFound)
	}
	packet := cachedPacket.(*radigo.Packet)

	agReq := NewAgentRequest(
		requestEv, requestVars, nil, nil, nil, nil,
		ra.cgrCfg.GeneralCfg().DefaultTenant,
		ra.cgrCfg.GeneralCfg().DefaultTimezone,
		ra.filterS, map[string]utils.DataProvider{
			utils.MetaOReq: newRADataProvider(packet),
		})
	err := agReq.SetFields(ra.cgrCfg.TemplatesCfg()[requestTemplate])
	if err != nil {
		return 0, fmt.Errorf("could not set attributes: %w", err)
	}

	remoteAddr, remoteHost, err := daRequestAddress(packet.RemoteAddr().String(),
		ra.cgrCfg.RadiusAgentCfg().ClientDaAddresses)
	if err != nil {
		return 0, fmt.Errorf("retrieving remote address failed: %w", err)
	}
	clientOpts := ra.cgrCfg.RadiusAgentCfg().ClientDaAddresses[remoteHost]
	dynAuthClient, err := radigo.NewClient(clientOpts.Transport, remoteAddr,
		ra.dacCfg.secrets.GetSecret(remoteHost),
		ra.dacCfg.dicts.GetInstance(remoteHost),
		ra.cgrCfg.GeneralCfg().ConnectAttempts, nil, utils.Logger)
	if err != nil {
		return 0, fmt.Errorf("dynamic authorization client init failed: %w", err)
	}
	dynAuthReq := dynAuthClient.NewRequest(requestType, 1)
	if err = radAppendAttributes(dynAuthReq, agReq.radDAReq); err != nil {
		return 0, fmt.Errorf("could not append attributes to the request packet: %w", err)
	}
	if clientOpts.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, sending %s for session with ID '%s' to '%s': %s",
				utils.RadiusAgent, requestType, sessionID, remoteAddr, utils.ToJSON(dynAuthReq)))
	}
	dynAuthReply, err := dynAuthClient.SendRequest(dynAuthReq)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	return dynAuthReply.Code, nil
}

// daRequestAddress ranges over the client_da_addresses map and returns the address configured for a
// specific client alongside the host.
func daRequestAddress(remoteAddr string, dynAuthAddresses map[string]config.DAClientOpts) (string, string, error) {
	if len(dynAuthAddresses) == 0 {
		return "", "", utils.ErrNotFound
	}
	remoteHost, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return "", "", err
	}
	for host, opts := range dynAuthAddresses {
		if host == remoteHost {
			address := opts.Host + ":" + strconv.Itoa(opts.Port)
			return address, host, nil
		}
	}
	return "", "", utils.ErrNotFound
}

// V1WarnDisconnect is needed to satisfy the sessions.BiRPClient interface
func (*RadiusAgent) V1WarnDisconnect(_ *context.Context, _ map[string]any, _ *string) error {
	return utils.ErrNotImplemented
}
