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
	"net"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/sm"
)

func NewDiameterAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (*DiameterAgent, error) {
	da := &DiameterAgent{cgrCfg: cgrCfg, filterS: filterS, connMgr: connMgr}
	dictsPath := cgrCfg.DiameterAgentCfg().DictionariesPath
	if len(dictsPath) != 0 {
		if err := loadDictionaries(dictsPath, utils.DiameterAgent); err != nil {
			return nil, err
		}
	}
	msgTemplates := da.cgrCfg.DiameterAgentCfg().Templates
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

type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	filterS  *engine.FilterS
	connMgr  *engine.ConnManager
	aReqs    int
	aReqsLck sync.RWMutex
}

// ListenAndServe is called when DiameterAgent is started, usually from within cmd/cgr-engine
func (da *DiameterAgent) ListenAndServe() error {
	utils.Logger.Info(fmt.Sprintf("<%s> Start listening on <%s>", utils.DiameterAgent, da.cgrCfg.DiameterAgentCfg().Listen))
	return diam.ListenAndServeNetwork(da.cgrCfg.DiameterAgentCfg().ListenNet, da.cgrCfg.DiameterAgentCfg().Listen, da.handlers(), nil)
}

// Creates the message handlers
func (da *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(da.cgrCfg.DiameterAgentCfg().VendorId),
		ProductName:      datatype.UTF8String(da.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DIAMETER_FIRMWARE_REVISION),
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
				hosts = append(hosts, net.ParseIP(strings.Split(iAddr.String(), utils.HDR_VAL_SEP)[0])) // address came in form x.y.z.t/24
			}
		}
	}
	settings.HostIPAddresses = make([]datatype.Address, len(hosts))
	for i, host := range hosts {
		settings.HostIPAddresses[i] = datatype.Address(host)
	}

	dSM := sm.New(settings)
	if da.cgrCfg.DiameterAgentCfg().SyncedConnReqs {
		dSM.HandleFunc("ALL", da.handleMessage)
	} else {
		dSM.HandleFunc("ALL", da.handleMessageAsync)
	}

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
	reqVars := utils.NavigableMap2{
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
		da.cgrCfg.DiameterAgentCfg().Templates[utils.MetaErr],
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
	if da.cgrCfg.DiameterAgentCfg().ASRTemplate != "" {
		sessID, err := diamDP.FieldAsString([]string{"Session-Id"})
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving Session-Id err: %s, message: %s",
					utils.DiameterAgent, err.Error(), m))
			writeOnConn(c, diamErr)
		}
		// cache message data needed for building up the ASR
		engine.Cache.Set(utils.CacheDiameterMessages, sessID, &diamMsgData{c, m, reqVars},
			nil, true, utils.NonTransactional)
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
	cgrRplyNM := utils.NavigableMap2{}
	rply := utils.NewOrderedNavigableMap() // share it among different processors
	var processed bool
	for _, reqProcessor := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = da.processRequest(
			reqProcessor,
			NewAgentRequest(
				diamDP, reqVars, &cgrRplyNM, rply,
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
	cgrEv := config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuth,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaMessage,
		utils.MetaCDRs, utils.MetaEvent, utils.META_NONE} {
		if reqProcessor.Flags.HasKey(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	cgrArgs := cgrEv.ExtractArgs(reqProcessor.Flags.HasKey(utils.MetaDispatchers),
		reqType == utils.MetaAuth || reqType == utils.MetaMessage || reqType == utils.MetaEvent)
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, diameter message: %s",
				utils.DiameterAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.META_NONE: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, DiameterMessage: %s",
				utils.DiameterAgent, reqProcessor.ID, agReq.Request.String()))
	case utils.MetaAuth:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaSuppliers),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersEventCost),
			cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator,
			reqProcessor.Flags.HasKey(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaSuppliersMaxCost),
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
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			cgrEv, cgrArgs.ArgDispatcher,
			reqProcessor.Flags.HasKey(utils.MetaFD))
		rply := new(sessions.V1InitSessionReply)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1InitiateSession,
			initArgs, rply)
		rply.SetMaxUsageNeeded(initArgs.InitSession)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			cgrEv, cgrArgs.ArgDispatcher,
			reqProcessor.Flags.HasKey(utils.MetaFD))
		rply := new(sessions.V1UpdateSessionReply)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1UpdateSession,
			updateArgs, rply)
		rply.SetMaxUsageNeeded(updateArgs.UpdateSession)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats),
			cgrEv, cgrArgs.ArgDispatcher,
			reqProcessor.Flags.HasKey(utils.MetaFD))
		rply := utils.StringPointer("")
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1TerminateSession,
			terminateArgs, rply)
		if err = agReq.setCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaMessage:
		msgArgs := sessions.NewV1ProcessMessageArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaSuppliers),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersEventCost),
			cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator,
			reqProcessor.Flags.HasKey(utils.MetaFD),
			reqProcessor.Flags.ParamValue(utils.MetaSuppliersMaxCost),
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
			Flags:         reqProcessor.Flags.SliceFlags(),
			CGREvent:      cgrEv,
			ArgDispatcher: cgrArgs.ArgDispatcher,
			Paginator:     *cgrArgs.SupplierPaginator,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1ProcessEvent,
			evArgs, rply)
		needMaxUsage := needsMaxUsage(reqProcessor.Flags.ParamsSlice(utils.MetaRALs))
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needMaxUsage {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		rply.SetMaxUsageNeeded(needMaxUsage)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.HasKey(utils.MetaCDRs) &&
		!reqProcessor.Flags.HasKey(utils.MetaDryRun) {
		rplyCDRs := utils.StringPointer("")
		if err = da.connMgr.Call(da.cgrCfg.DiameterAgentCfg().SessionSConns, da, utils.SessionSv1ProcessCDR,
			&utils.CGREventWithArgDispatcher{CGREvent: cgrEv,
				ArgDispatcher: cgrArgs.ArgDispatcher}, rplyCDRs); err != nil {
			agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(err.Error()))
		}
	}
	if err = agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return
	}
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
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

// rpcclient.ClientConnector interface
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
	msg, has := engine.Cache.Get(utils.CacheDiameterMessages, ssID.(string))
	if !has {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot retrieve message from cache with OriginID: <%s>",
				utils.DiameterAgent, ssID))
		return utils.ErrMandatoryIeMissing
	}
	dmd := msg.(*diamMsgData)
	aReq := NewAgentRequest(
		newDADataProvider(dmd.c, dmd.m),
		dmd.vars, nil, nil, nil,
		da.cgrCfg.GeneralCfg().DefaultTenant,
		da.cgrCfg.GeneralCfg().DefaultTimezone, da.filterS, nil, nil)
	if err = aReq.SetFields(da.cgrCfg.DiameterAgentCfg().Templates[da.cgrCfg.DiameterAgentCfg().ASRTemplate]); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot disconnect session with OriginID: <%s>, err: %s",
				utils.DiameterAgent, ssID, err.Error()))
		return utils.ErrServerError
	}
	m := diam.NewRequest(dmd.m.Header.CommandCode,
		dmd.m.Header.ApplicationID, dmd.m.Dictionary())
	if err = updateDiamMsgFromNavMap(m, aReq.diamreq,
		da.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot disconnect session with OriginID: <%s>, err: %s",
				utils.DiameterAgent, ssID, err.Error()))
		return utils.ErrServerError
	}
	if err = writeOnConn(dmd.c, m); err != nil {
		return utils.ErrServerError
	}
	*reply = utils.OK
	return
}

// V1GetActiveSessionIDs is part of the sessions.BiRPClient
func (da *DiameterAgent) V1GetActiveSessionIDs(ignParam string,
	sessionIDs *[]*sessions.SessionID) error {
	return utils.ErrNotImplemented
}
