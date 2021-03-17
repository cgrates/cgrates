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
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
	"github.com/fiorix/go-diameter/v4/diam/sm"
	"github.com/fiorix/go-diameter/v4/diam/sm/smpeer"
)

const (
	all = "ALL"
	raa = "RAA"
	dpa = "DPA"
)

// NewDiameterAgent initializes a new DiameterAgent
func NewDiameterAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (*DiameterAgent, error) {
	da := &DiameterAgent{
		cgrCfg:  cgrCfg,
		filterS: filterS,
		connMgr: connMgr,
		raa:     make(map[string]chan *diam.Message),
		dpa:     make(map[string]chan *diam.Message),
		peers:   make(map[string]diam.Conn),
	}
	dictsPath := cgrCfg.DiameterAgentCfg().DictionariesPath
	if len(dictsPath) != 0 {
		if err := loadDictionaries(dictsPath, utils.DiameterAgent); err != nil {
			return nil, err
		}
	}
	msgTemplates := da.cgrCfg.TemplatesCfg()
	// Inflate *template field types
	for _, procsr := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		if tpls, err := config.InflateTemplates(procsr.RequestFields, msgTemplates); err != nil {
			return nil, err
		} else if tpls != nil {
			procsr.RequestFields = tpls
		}
		if tpls, err := config.InflateTemplates(procsr.ReplyFields, msgTemplates); err != nil {
			return nil, err
		} else if tpls != nil {
			procsr.ReplyFields = tpls
		}
	}
	return da, nil
}

// DiameterAgent describes the diameter server
type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	filterS  *engine.FilterS
	connMgr  *engine.ConnManager
	aReqs    int
	aReqsLck sync.RWMutex
	raa      map[string]chan *diam.Message
	raaLck   sync.RWMutex

	peersLck sync.Mutex
	peers    map[string]diam.Conn // peer index by OriginHost;OriginRealm
	dpa      map[string]chan *diam.Message
	dpaLck   sync.RWMutex
}

// ListenAndServe is called when DiameterAgent is started, usually from within cmd/cgr-engine
func (da *DiameterAgent) ListenAndServe(stopChan <-chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> Start listening on <%s>", utils.DiameterAgent, da.cgrCfg.DiameterAgentCfg().Listen))
	srv := &diam.Server{
		Network: da.cgrCfg.DiameterAgentCfg().ListenNet,
		Addr:    da.cgrCfg.DiameterAgentCfg().Listen,
		Handler: da.handlers(),
		Dict:    nil,
	}
	// used to control the server state
	var lsn net.Listener
	if lsn, err = diam.MultistreamListen(utils.FirstNonEmpty(srv.Network, utils.TCP),
		utils.FirstNonEmpty(srv.Addr, ":3868")); err != nil {
		return
	}
	errChan := make(chan error)
	go func() {
		errChan <- srv.Serve(lsn)
	}()
	select {
	case err = <-errChan:
		return
	case <-stopChan:
		return lsn.Close()
	}
}

// Creates the message handlers
func (da *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(da.cgrCfg.DiameterAgentCfg().VendorID),
		ProductName:      datatype.UTF8String(da.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DiameterFirmwareRevision),
	}
	hosts := disectDiamListen(da.cgrCfg.DiameterAgentCfg().Listen)
	if len(hosts) == 0 {
		interfaces, err := net.Interfaces()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error : %v, when quering interfaces for address", utils.DiameterAgent, err))
		}
		for _, inter := range interfaces {
			addrs, err := inter.Addrs()
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> error: %+v, when taking address from interface: %+v",
					utils.DiameterAgent, err, inter.Name))
				continue
			}
			for _, iAddr := range addrs {
				hosts = append(hosts, net.ParseIP(strings.Split(iAddr.String(), utils.HDRValSep)[0])) // address came in form x.y.z.t/24
			}
		}
	}
	settings.HostIPAddresses = make([]datatype.Address, len(hosts))
	for i, host := range hosts {
		settings.HostIPAddresses[i] = datatype.Address(host)
	}

	dSM := sm.New(settings)
	if da.cgrCfg.DiameterAgentCfg().SyncedConnReqs {
		dSM.HandleFunc(all, da.handleMessage)
		dSM.HandleFunc(raa, da.handleRAA)
		dSM.HandleFunc(dpa, da.handleDPA)
	} else {
		dSM.HandleFunc(all, da.handleMessageAsync)
		dSM.HandleFunc(raa, func(c diam.Conn, m *diam.Message) { go da.handleRAA(c, m) })
		dSM.HandleFunc(dpa, func(c diam.Conn, m *diam.Message) { go da.handleDPA(c, m) })
	}
	go da.handleConns(dSM.HandshakeNotify())
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Err(fmt.Sprintf("<%s> sm error: %v", utils.DiameterAgent, err))
		}
	}()
	return dSM
}

// handleMessageAsync will dispatch the message into it's own goroutine
func (da *DiameterAgent) handleMessageAsync(c diam.Conn, m *diam.Message) {
	go da.handleMessage(c, m)
}

// handleALL is the handler of all messages coming in via Diameter
func (da *DiameterAgent) handleMessage(c diam.Conn, m *diam.Message) {
	dApp, err := m.Dictionary().App(m.Header.ApplicationID)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> decoding app: %d, err: %s",
			utils.DiameterAgent, m.Header.ApplicationID, err.Error()))
		writeOnConn(c, diamBareErr(m, diam.NoCommonApplication))
		return
	}
	dCmd, err := m.Dictionary().FindCommand(
		m.Header.ApplicationID,
		m.Header.CommandCode)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> decoding app: %d, command %d, err: %s",
			utils.DiameterAgent, m.Header.ApplicationID, m.Header.CommandCode, err.Error()))
		writeOnConn(c, diamBareErr(m, diam.CommandUnsupported))
		return
	}
	diamDP := newDADataProvider(c, m)
	reqVars := utils.NavigableMap{
		utils.OriginHost:  utils.NewNMData(da.cgrCfg.DiameterAgentCfg().OriginHost), // used in templates
		utils.OriginRealm: utils.NewNMData(da.cgrCfg.DiameterAgentCfg().OriginRealm),
		utils.ProductName: utils.NewNMData(da.cgrCfg.DiameterAgentCfg().ProductName),
		utils.MetaApp:     utils.NewNMData(dApp.Name),
		utils.MetaAppID:   utils.NewNMData(dApp.ID),
		utils.MetaCmd:     utils.NewNMData(dCmd.Short + "R"),
		utils.RemoteHost:  utils.NewNMData(c.RemoteAddr().String()),
	}
	// build the negative error answer
	diamErr, err := diamErr(
		m, diam.UnableToComply, reqVars,
		da.cgrCfg.TemplatesCfg()[utils.MetaErr],
		da.cgrCfg.GeneralCfg().DefaultTenant,
		da.cgrCfg.GeneralCfg().DefaultTimezone,
		da.filterS)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s building errDiam for message: %s",
				utils.DiameterAgent, err.Error(), m))
		writeOnConn(c, diamBareErr(m, diam.CommandUnsupported))
		return
	}
	// cache message for ASR
	if da.cgrCfg.DiameterAgentCfg().ASRTemplate != "" ||
		da.cgrCfg.DiameterAgentCfg().RARTemplate != "" {
		sessID, err := diamDP.FieldAsString([]string{"Session-Id"})
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving Session-Id err: %s, message: %s",
					utils.DiameterAgent, err.Error(), m))
			writeOnConn(c, diamErr)
		}
		// cache message data needed for building up the ASR
		if errCh := engine.Cache.Set(utils.CacheDiameterMessages, sessID, &diamMsgData{c, m, reqVars},
			nil, true, utils.NonTransactional); errCh != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed message: %s to set Cache: %s", utils.DiameterAgent, m, errCh.Error()))
			writeOnConn(c, diamErr)
			return
		}
	}
	// handle MaxActiveReqs
	if da.cgrCfg.DiameterAgentCfg().ConcurrentReqs != -1 {
		da.aReqsLck.Lock()
		if da.aReqs == da.cgrCfg.DiameterAgentCfg().ConcurrentReqs {
			utils.Logger.Err(
				fmt.Sprintf("<%s> denying request due to maximum active requests reached: %d, message: %s",
					utils.DiameterAgent, da.cgrCfg.DiameterAgentCfg().ConcurrentReqs, m))
			writeOnConn(c, diamErr)
			da.aReqsLck.Unlock()
			return
		}
		da.aReqs++
		da.aReqsLck.Unlock()
		defer func() { // schedule decrement when returning out of function
			da.aReqsLck.Lock()
			da.aReqs--
			da.aReqsLck.Unlock()
		}()
	}
	cgrRplyNM := utils.NavigableMap{}
	rply := utils.NewOrderedNavigableMap() // share it among different processors
	opts := utils.NewOrderedNavigableMap()
	var processed bool
	for _, reqProcessor := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = da.processRequest(
			reqProcessor,
			NewAgentRequest(
				diamDP, reqVars, &cgrRplyNM, rply, opts,
				reqProcessor.Tenant, da.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(reqProcessor.Timezone,
					da.cgrCfg.GeneralCfg().DefaultTimezone),
				da.filterS, nil, nil))
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s",
				utils.DiameterAgent, err.Error(), m))
		writeOnConn(c, diamErr)
		return
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.DiameterAgent, m, c.RemoteAddr()))
		writeOnConn(c, diamErr)
		return
	}
	a, err := diamAnswer(m, 0, false,
		rply, da.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> err: %s, replying to message: %+v",
				utils.DiameterAgent, err.Error(), m))
		writeOnConn(c, diamErr)
		return
	}
	writeOnConn(c, a)
}

func (da *DiameterAgent) processRequest(reqProcessor *config.RequestProcessor,
	agReq *AgentRequest) (processed bool, err error) {
	if pass, err := da.filterS.Pass(agReq.Tenant,
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
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	var cgrArgs utils.Paginator
	if reqType == utils.MetaAuthorize || reqType == utils.MetaMessage || reqType == utils.MetaEvent {
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(cgrEv.Opts); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> args extraction failed because <%s>",
				utils.DiameterAgent, err.Error()))
			err = nil // reset the error and continue the processing
		}
	}

	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, diameter message: %s",
				utils.DiameterAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, DiameterMessage: %s",
				utils.DiameterAgent, reqProcessor.ID, agReq.Request.String()))
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
			cgrEv, cgrArgs,
			reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1AuthorizeReply)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1AuthorizeEvent,
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
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1InitiateSession,
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
		rply.SetMaxUsageNeeded(updateArgs.UpdateSession)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1UpdateSession,
			updateArgs, rply)
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
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1TerminateSession,
			terminateArgs, &rply)
		if err = agReq.setCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaMessage:
		msgArgs := sessions.NewV1ProcessMessageArgs(
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
			cgrEv, cgrArgs,
			reqProcessor.Flags.Has(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1ProcessMessageReply)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1ProcessMessage,
			msgArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if msgArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(msgArgs.Debit)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags:     reqProcessor.Flags.SliceFlags(),
			Paginator: cgrArgs,
			CGREvent:  cgrEv,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1ProcessEvent,
			evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.GetBool(utils.MetaCDRs) &&
		!reqProcessor.Flags.Has(utils.MetaDryRun) {
		var rplyCDRs string
		if err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(err.Error()))
		}
	}
	if err = agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, Diameter reply: %s",
				utils.DiameterAgent, agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, Diameter reply: %s",
				utils.DiameterAgent, agReq.Reply))
	}
	return true, nil
}

// Call implements rpcclient.ClientConnector interface
func (da *DiameterAgent) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(da, serviceMethod, args, reply)
}

// V1DisconnectSession is part of the sessions.BiRPClient
func (da *DiameterAgent) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error) {
	ssID, has := args.EventStart[utils.OriginID]
	if !has {
		utils.Logger.Info(
			fmt.Sprintf("<%s> cannot disconnect session, missing OriginID in event: %s",
				utils.DiameterAgent, utils.ToJSON(args.EventStart)))
		return utils.ErrMandatoryIeMissing
	}
	originID := ssID.(string)
	switch da.cgrCfg.DiameterAgentCfg().ForcedDisconnect {
	case utils.MetaNone:
		*reply = utils.OK
		return
	case utils.MetaASR:
		return da.sendASR(originID, reply)
	case utils.MetaRAR:
		return da.V1ReAuthorize(originID, reply)
	default:
		return fmt.Errorf("Unsupported request type <%s>", da.cgrCfg.DiameterAgentCfg().ForcedDisconnect)
	}
}

// V1GetActiveSessionIDs is part of the sessions.BiRPClient
func (da *DiameterAgent) V1GetActiveSessionIDs(ignParam string,
	sessionIDs *[]*sessions.SessionID) error {
	return utils.ErrNotImplemented
}

func (da *DiameterAgent) sendASR(originID string, reply *string) (err error) {
	msg, has := engine.Cache.Get(utils.CacheDiameterMessages, originID)
	if !has {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot retrieve message from cache with OriginID: <%s>",
				utils.DiameterAgent, originID))
		return utils.ErrMandatoryIeMissing
	}
	dmd := msg.(*diamMsgData)
	aReq := NewAgentRequest(
		newDADataProvider(dmd.c, dmd.m),
		dmd.vars, nil, nil, nil, nil,
		da.cgrCfg.GeneralCfg().DefaultTenant,
		da.cgrCfg.GeneralCfg().DefaultTimezone, da.filterS, nil, nil)
	if err = aReq.SetFields(da.cgrCfg.TemplatesCfg()[da.cgrCfg.DiameterAgentCfg().ASRTemplate]); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot disconnect session with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	m := diam.NewRequest(dmd.m.Header.CommandCode,
		dmd.m.Header.ApplicationID, dmd.m.Dictionary())
	if err = updateDiamMsgFromNavMap(m, aReq.diamreq,
		da.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot disconnect session with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	if err = writeOnConn(dmd.c, m); err != nil {
		return utils.ErrServerError
	}
	*reply = utils.OK
	return
}

// V1ReAuthorize  sends a rar message to diameter client
func (da *DiameterAgent) V1ReAuthorize(originID string, reply *string) (err error) {
	if originID == "" {
		utils.Logger.Info(
			fmt.Sprintf("<%s> cannot send RAR, missing session ID",
				utils.DiameterAgent))
		return utils.ErrMandatoryIeMissing
	}
	msg, has := engine.Cache.Get(utils.CacheDiameterMessages, originID)
	if !has {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot retrieve message from cache with OriginID: <%s>",
				utils.DiameterAgent, originID))
		return utils.ErrMandatoryIeMissing
	}
	dmd := msg.(*diamMsgData)
	aReq := NewAgentRequest(
		newDADataProvider(dmd.c, dmd.m),
		dmd.vars, nil, nil, nil, nil,
		da.cgrCfg.GeneralCfg().DefaultTenant,
		da.cgrCfg.GeneralCfg().DefaultTimezone, da.filterS, nil, nil)
	if err = aReq.SetFields(da.cgrCfg.TemplatesCfg()[da.cgrCfg.DiameterAgentCfg().RARTemplate]); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot send RAR with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	m := diam.NewRequest(diam.ReAuth,
		dmd.m.Header.ApplicationID, dmd.m.Dictionary())
	if err = updateDiamMsgFromNavMap(m, aReq.diamreq,
		da.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot send RAR with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	raaCh := make(chan *diam.Message, 1)
	da.raaLck.Lock()
	da.raa[originID] = raaCh
	da.raaLck.Unlock()
	defer func() {
		da.raaLck.Lock()
		delete(da.raa, originID)
		da.raaLck.Unlock()
	}()
	if err = writeOnConn(dmd.c, m); err != nil {
		return utils.ErrServerError
	}
	select {
	case raa := <-raaCh:
		var avps []*diam.AVP
		if avps, err = raa.FindAVPsWithPath([]interface{}{avp.ResultCode}, dict.UndefinedVendorID); err != nil {
			return
		}
		if len(avps) == 0 {
			return fmt.Errorf("Missing AVP")
		}
		var data interface{}
		if data, err = diamAVPAsIface(avps[0]); err != nil {
			return
		} else if data != uint32(diam.Success) {
			return fmt.Errorf("Wrong result code: <%v>", data)
		}
	case <-time.After(time.Second):
		return utils.ErrTimedOut
	}
	*reply = utils.OK
	return
}

// handleRAA is used to handle all Re-Authorize Answers that are received
func (da *DiameterAgent) handleRAA(c diam.Conn, m *diam.Message) {
	avp, err := m.FindAVP(avp.SessionID, dict.UndefinedVendorID)
	if err != nil {
		return
	}
	originID, err := diamAVPAsString(avp)
	if err != nil {
		return
	}
	da.raaLck.Lock()
	ch, has := da.raa[originID]
	da.raaLck.Unlock()
	if !has {
		return
	}
	ch <- m
}

// handleConns is used to handle all conns that are connected to the agent
// it register the connection so it can be used to send a DPR
func (da *DiameterAgent) handleConns(peers <-chan diam.Conn) {
	for c := range peers {
		meta, _ := smpeer.FromContext(c.Context())
		key := string(meta.OriginHost + utils.ConcatenatedKeySep + meta.OriginRealm)
		da.peersLck.Lock()
		da.peers[key] = c // store in peers table
		da.peersLck.Unlock()
		go func(c diam.Conn, key string) {
			// wait for disconnect notification
			<-c.(diam.CloseNotifier).CloseNotify()
			da.peersLck.Lock()
			delete(da.peers, key) // remove from peers table
			da.peersLck.Unlock()
		}(c, key)
	}
}

// handleDPA is used to handle all DisconnectPeer Answers that are received
func (da *DiameterAgent) handleDPA(c diam.Conn, m *diam.Message) {
	meta, _ := smpeer.FromContext(c.Context())
	key := string(meta.OriginHost + utils.ConcatenatedKeySep + meta.OriginRealm)

	da.dpaLck.Lock()
	ch, has := da.dpa[key]
	da.dpaLck.Unlock()
	if !has {
		return
	}
	ch <- m
	c.Close()
}

// V1DisconnectPeer  sends a DPR meseage to diameter client
func (da *DiameterAgent) V1DisconnectPeer(args *utils.DPRArgs, reply *string) (err error) {
	if args == nil {
		utils.Logger.Info(
			fmt.Sprintf("<%s> cannot send DPR, missing arrguments",
				utils.DiameterAgent))
		return utils.ErrMandatoryIeMissing
	}

	if args.DisconnectCause < 0 || args.DisconnectCause > 2 {
		return errors.New("WRONG_DISCONNECT_CAUSE")
	}
	m := diam.NewRequest(diam.DisconnectPeer,
		diam.CHARGING_CONTROL_APP_ID, dict.Default)
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(args.OriginHost))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(args.OriginRealm))
	m.NewAVP(avp.DisconnectCause, avp.Mbit, 0, datatype.Enumerated(args.DisconnectCause))

	key := args.OriginHost + utils.ConcatenatedKeySep + args.OriginRealm

	dpaCh := make(chan *diam.Message, 1)
	da.dpaLck.Lock()
	da.dpa[key] = dpaCh
	da.dpaLck.Unlock()
	defer func() {
		da.dpaLck.Lock()
		delete(da.dpa, key)
		da.dpaLck.Unlock()
	}()
	da.peersLck.Lock()
	conn, has := da.peers[key]
	da.peersLck.Unlock()
	if !has {
		return utils.ErrNotFound
	}
	if err = writeOnConn(conn, m); err != nil {
		return utils.ErrServerError
	}
	select {
	case dpa := <-dpaCh:
		var avps []*diam.AVP
		if avps, err = dpa.FindAVPsWithPath([]interface{}{avp.ResultCode}, dict.UndefinedVendorID); err != nil {
			return
		}
		if len(avps) == 0 {
			return fmt.Errorf("Missing AVP")
		}
		var data interface{}
		if data, err = diamAVPAsIface(avps[0]); err != nil {
			return
		} else if data != uint32(diam.Success) {
			return fmt.Errorf("Wrong result code: <%v>", data)
		}
	case <-time.After(10 * time.Second):
		return utils.ErrTimedOut
	}
	*reply = utils.OK
	return
}

// V1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (*DiameterAgent) V1WarnDisconnect(args map[string]interface{}, reply *string) (err error) {
	return utils.ErrNotImplemented
}

// CallBiRPC is part of utils.BiRPCServer interface to help internal connections do calls over rpcclient.ClientConnector interface
func (da *DiameterAgent) CallBiRPC(clnt rpcclient.ClientConnector, serviceMethod string, args interface{}, reply interface{}) error {
	return utils.BiRPCCall(da, clnt, serviceMethod, args, reply)
}

// BiRPCv1DisconnectSession is internal method to disconnect session in asterisk
func (da *DiameterAgent) BiRPCv1DisconnectSession(clnt rpcclient.ClientConnector, args utils.AttrDisconnectSession, reply *string) error {
	return da.V1DisconnectSession(args, reply)
}

// BiRPCv1GetActiveSessionIDs is internal method to  get all active sessions in asterisk
func (da *DiameterAgent) BiRPCv1GetActiveSessionIDs(clnt rpcclient.ClientConnector, ignParam string,
	sessionIDs *[]*sessions.SessionID) error {
	return da.V1GetActiveSessionIDs(ignParam, sessionIDs)

}

// BiRPCv1ReAuthorize is used to implement the sessions.BiRPClient interface
func (da *DiameterAgent) BiRPCv1ReAuthorize(clnt rpcclient.ClientConnector, originID string, reply *string) (err error) {
	return da.V1ReAuthorize(originID, reply)
}

// BiRPCv1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (da *DiameterAgent) BiRPCv1DisconnectPeer(clnt rpcclient.ClientConnector, args *utils.DPRArgs, reply *string) (err error) {
	return da.V1DisconnectPeer(args, reply)
}

// BiRPCv1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (da *DiameterAgent) BiRPCv1WarnDisconnect(clnt rpcclient.ClientConnector, args map[string]interface{}, reply *string) (err error) {
	return da.V1WarnDisconnect(args, reply)
}

// Handlers is used to implement the rpcclient.BiRPCConector interface
func (da *DiameterAgent) Handlers() map[string]interface{} {
	return map[string]interface{}{
		utils.SessionSv1DisconnectSession: func(clnt *rpc2.Client, args utils.AttrDisconnectSession, rply *string) error {
			return da.BiRPCv1DisconnectSession(clnt, args, rply)
		},
		utils.SessionSv1GetActiveSessionIDs: func(clnt *rpc2.Client, args string, rply *[]*sessions.SessionID) error {
			return da.BiRPCv1GetActiveSessionIDs(clnt, args, rply)
		},
		utils.SessionSv1ReAuthorize: func(clnt *rpc2.Client, args string, rply *string) (err error) {
			return da.BiRPCv1ReAuthorize(clnt, args, rply)
		},
		utils.SessionSv1DisconnectPeer: func(clnt *rpc2.Client, args *utils.DPRArgs, rply *string) (err error) {
			return da.BiRPCv1DisconnectPeer(clnt, args, rply)
		},
		utils.SessionSv1WarnDisconnect: func(clnt *rpc2.Client, args map[string]interface{}, rply *string) (err error) {
			return da.BiRPCv1WarnDisconnect(clnt, args, rply)
		},
	}
}
