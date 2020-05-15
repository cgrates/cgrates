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
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/sipingo"
)

const (
	bufferSize = 5000
)

// NewSIPAgent will construct a SIPAgent
func NewSIPAgent(connMgr *engine.ConnManager, cfg *config.CGRConfig,
	filterS *engine.FilterS) *SIPAgent {
	return &SIPAgent{
		connMgr: connMgr,
		filterS: filterS,
		cfg:     cfg,
	}
}

// SIPAgent is a handler for SIP requests
type SIPAgent struct {
	connMgr  *engine.ConnManager
	filterS  *engine.FilterS
	cfg      *config.CGRConfig
	stopChan chan struct{}
}

// Shutdown will stop the SIPAgent server
func (sa *SIPAgent) Shutdown() {
	close(sa.stopChan)
}

// ListenAndServe will run the SIP handler doing also the connection to listen address
func (sa *SIPAgent) ListenAndServe() (err error) {
	sa.stopChan = make(chan struct{})
	utils.Logger.Info(fmt.Sprintf("<%s> start listening on <%s:%s>",
		utils.SIPAgent, sa.cfg.SIPAgentCfg().ListenNet, sa.cfg.SIPAgentCfg().Listen))
	switch sa.cfg.SIPAgentCfg().ListenNet {
	case utils.TCP:
		return sa.serveTCP(sa.stopChan)
	case utils.UDP:
		return sa.serveUDP(sa.stopChan)
	default:
		return fmt.Errorf("Unecepected protocol %s", sa.cfg.SIPAgentCfg().ListenNet)
	}
}
func (sa *SIPAgent) serveUDP(stop chan struct{}) (err error) {
	var conn net.PacketConn
	if conn, err = net.ListenPacket(utils.UDP, sa.cfg.SIPAgentCfg().Listen); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: %s unable to listen to: %s",
				utils.SIPAgent, err.Error(), sa.cfg.SIPAgentCfg().Listen))
		return
	}

	defer conn.Close()

	buf := make([]byte, bufferSize)
	wg := sync.WaitGroup{}
	for {
		select {
		case <-stop:
			wg.Wait()
			return
		default:
		}
		conn.SetDeadline(time.Now().Add(time.Second))
		var n int
		var saddr net.Addr
		if n, saddr, err = conn.ReadFrom(buf); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			utils.Logger.Err(
				fmt.Sprintf("<%s> error: %s unable to read from: %s",
					utils.SIPAgent, err.Error(), saddr))
			return
		}
		// echo response
		if n < 50 {
			conn.WriteTo(buf[:n], saddr)
			continue
		}
		wg.Add(1)
		go func(message string, conn net.PacketConn) {
			var sipMessage sipingo.Message
			if sipMessage, err = sipingo.NewMessage(message); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s parsing message: %s",
						utils.SIPAgent, err.Error(), message))
				wg.Done()
				return
			}
			var sipAnswer sipingo.Message
			var err error
			if sipAnswer, err = sa.handleMessage(sipMessage, saddr.String()); err != nil {
				wg.Done()
				return
			}
			if _, err = conn.WriteTo([]byte(sipAnswer.String()), saddr); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s sending message: %s",
						utils.SIPAgent, err.Error(), sipAnswer))
				wg.Done()
				return
			}
			wg.Done()
		}(string(buf[:n]), conn)
	}
}

func (sa *SIPAgent) serveTCP(stop chan struct{}) (err error) {
	var l *net.TCPListener
	var addr *net.TCPAddr
	if addr, err = net.ResolveTCPAddr("tcp", sa.cfg.SIPAgentCfg().Listen); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> unable to rezolve TCP Address <%s> because: %s",
				utils.SIPAgent, sa.cfg.SIPAgentCfg().Listen, err.Error()))
		return
	}
	if l, err = net.ListenTCP(utils.TCP, addr); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: %s unable to listen to: %s",
				utils.SIPAgent, err.Error(), sa.cfg.SIPAgentCfg().Listen))
		return
	}

	defer l.Close()

	wg := sync.WaitGroup{}
	for {
		select {
		case <-stop:
			wg.Wait()
			return
		default:
		}
		l.SetDeadline(time.Now().Add(time.Second))
		var conn net.Conn
		if conn, err = l.Accept(); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			utils.Logger.Err(
				fmt.Sprintf("<%s> unable to accept connection because of error %s",
					utils.SIPAgent, err.Error()))
			return
		}
		wg.Add(1)
		go func(conn net.Conn) {
			buf := make([]byte, bufferSize)
			for {
				select {
				case <-stop:
					conn.Close()
					wg.Done()
					return
				default:
				}
				conn.SetReadDeadline(time.Now().Add(time.Second))
				n, err := conn.Read(buf)
				if err != nil {
					continue
				}
				// echo response
				if n < 50 {
					conn.Write(buf[:n])
					continue
				}

				var sipMessage sipingo.Message // recreate map SIP
				if sipMessage, err = sipingo.NewMessage(string(buf[:n])); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s parsing message: %s",
							utils.SIPAgent, err.Error(), string(buf[:n])))
					wg.Done()
					continue
				}
				var sipAnswer sipingo.Message
				if sipAnswer, err = sa.handleMessage(sipMessage, conn.LocalAddr().String()); err != nil {
					continue
				}
				if _, err = conn.Write([]byte(sipAnswer.String())); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s sending message: %s",
							utils.SIPAgent, err.Error(), sipAnswer))
					continue
				}
			}
		}(conn)
	}
}

func (sa *SIPAgent) handleMessage(sipMessage sipingo.Message, remoteHost string) (sipAnswer sipingo.Message, err error) {
	if sipMessage["User-Agent"] != "" {
		sipMessage["User-Agent"] = utils.CGRateS
	}
	sipMessageIface := make(map[string]interface{})
	for k, v := range sipMessage {
		sipMessageIface[k] = v
	}
	dp := utils.MapStorage(sipMessageIface)
	var processed bool
	cgrRplyNM := utils.NavigableMap2{}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.NewOrderedNavigableMap()
	reqVars := utils.NavigableMap2{
		utils.RemoteHost: utils.NewNMData(remoteHost),
		"Method":         utils.NewNMData(sipMessage.MethodFrom("Request")),
	}
	for _, reqProcessor := range sa.cfg.SIPAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(dp, reqVars, &cgrRplyNM, rplyNM,
			opts, reqProcessor.Tenant, sa.cfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			sa.filterS, nil, nil)
		if processed, err = sa.processRequest(reqProcessor, agReq); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing request: %s",
					utils.SIPAgent, err.Error(), utils.ToJSON(agReq)))
			continue
		}
		if !processed {
			continue
		}
		if processed && !reqProcessor.Flags.GetBool(utils.MetaContinue) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s from %s",
				utils.SIPAgent, err.Error(), sipMessage, remoteHost))
		return
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.SIPAgent, sipMessage, remoteHost))
		return
	}
	if err = updateSIPMsgFromNavMap(sipMessage, rplyNM); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s encoding out %s",
				utils.SIPAgent, err.Error(), utils.ToJSON(rplyNM)))
		return
	}
	sipMessage.PrepareReply()
	return sipMessage, nil
}

// processRequest represents one processor processing the request
func (sa *SIPAgent) processRequest(reqProcessor *config.RequestProcessor,
	agReq *AgentRequest) (processed bool, err error) {
	if pass, err := sa.filterS.Pass(agReq.Tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if err = agReq.SetFields(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep)
	opts := config.NMAsMapInterface(agReq.Opts, utils.NestingSep)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuthorize, /*
			utils.MetaInitiate, utils.MetaUpdate,
			utils.MetaTerminate, utils.MetaMessage,
			utils.MetaCDRs, */utils.MetaEvent, utils.META_NONE} {
		if reqProcessor.Flags.HasKey(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	var cgrArgs utils.ExtractedArgs
	if cgrArgs, err = utils.ExtractArgsFromOpts(opts, reqProcessor.Flags.HasKey(utils.MetaDispatchers),
		reqType == utils.MetaAuthorize || reqType == utils.MetaMessage || reqType == utils.MetaEvent); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> args extraction failed because <%s>",
			utils.SIPAgent, err.Error()))
		err = nil // reset the error and continue the processing
	}
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, SIP message: %s",
				utils.SIPAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.META_NONE: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.SIPAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuthorize:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaRoutes),
			reqProcessor.Flags.HasKey(utils.MetaRoutesIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.RoutePaginator,
			reqProcessor.Flags.HasKey(utils.MetaFD),
			opts,
		)
		rply := new(sessions.V1AuthorizeReply)
		err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	// case utils.MetaInitiate:
	// 	initArgs := sessions.NewV1InitSessionArgs(
	// 		reqProcessor.Flags.HasKey(utils.MetaAttributes),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
	// 		reqProcessor.Flags.HasKey(utils.MetaThresholds),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
	// 		reqProcessor.Flags.HasKey(utils.MetaStats),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaStats),
	// 		reqProcessor.Flags.HasKey(utils.MetaResources),
	// 		reqProcessor.Flags.HasKey(utils.MetaAccounts),
	// 		cgrEv, cgrArgs.ArgDispatcher,
	// 		reqProcessor.Flags.HasKey(utils.MetaFD),
	// 		opts)
	// 	rply := new(sessions.V1InitSessionReply)
	// 	err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1InitiateSession,
	// 		initArgs, rply)
	// 	if err = agReq.setCGRReply(rply, err); err != nil {
	// 		return
	// 	}
	// case utils.MetaUpdate:
	// 	updateArgs := sessions.NewV1UpdateSessionArgs(
	// 		reqProcessor.Flags.HasKey(utils.MetaAttributes),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
	// 		reqProcessor.Flags.HasKey(utils.MetaAccounts),
	// 		cgrEv, cgrArgs.ArgDispatcher,
	// 		reqProcessor.Flags.HasKey(utils.MetaFD),
	// 		opts)
	// 	rply := new(sessions.V1UpdateSessionReply)
	// 	err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1UpdateSession,
	// 		updateArgs, rply)
	// 	if err = agReq.setCGRReply(rply, err); err != nil {
	// 		return
	// 	}
	// case utils.MetaTerminate:
	// 	terminateArgs := sessions.NewV1TerminateSessionArgs(
	// 		reqProcessor.Flags.HasKey(utils.MetaAccounts),
	// 		reqProcessor.Flags.HasKey(utils.MetaResources),
	// 		reqProcessor.Flags.HasKey(utils.MetaThresholds),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
	// 		reqProcessor.Flags.HasKey(utils.MetaStats),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaStats),
	// 		cgrEv, cgrArgs.ArgDispatcher,
	// 		reqProcessor.Flags.HasKey(utils.MetaFD),
	// 		opts)
	// 	rply := utils.StringPointer("")
	// 	err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1TerminateSession,
	// 		terminateArgs, rply)
	// 	if err = agReq.setCGRReply(nil, err); err != nil {
	// 		return
	// 	}
	// case utils.MetaMessage:
	// 	evArgs := sessions.NewV1ProcessMessageArgs(
	// 		reqProcessor.Flags.HasKey(utils.MetaAttributes),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaAttributes),
	// 		reqProcessor.Flags.HasKey(utils.MetaThresholds),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaThresholds),
	// 		reqProcessor.Flags.HasKey(utils.MetaStats),
	// 		reqProcessor.Flags.ParamsSlice(utils.MetaStats),
	// 		reqProcessor.Flags.HasKey(utils.MetaResources),
	// 		reqProcessor.Flags.HasKey(utils.MetaAccounts),
	// 		reqProcessor.Flags.HasKey(utils.MetaRoutes),
	// 		reqProcessor.Flags.HasKey(utils.MetaRoutesIgnoreErrors),
	// 		reqProcessor.Flags.HasKey(utils.MetaRoutesEventCost),
	// 		cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.RoutePaginator,
	// 		reqProcessor.Flags.HasKey(utils.MetaFD),
	// 		opts)
	// 	rply := new(sessions.V1ProcessMessageReply)
	// 	err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1ProcessMessage,
	// 		evArgs, rply)
	// 	if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
	// 		cgrEv.Event[utils.Usage] = 0 // avoid further debits
	// 	} else if evArgs.Debit {
	// 		cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
	// 	}
	// 	if err = agReq.setCGRReply(nil, err); err != nil {
	// 		return
	// 	}
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags:         reqProcessor.Flags.SliceFlags(),
			CGREvent:      cgrEv,
			ArgDispatcher: cgrArgs.ArgDispatcher,
			Paginator:     *cgrArgs.RoutePaginator,
			Opts:          opts,
		}
		needMaxUsage := reqProcessor.Flags.HasKey(utils.MetaAuth) ||
			reqProcessor.Flags.HasKey(utils.MetaInit) ||
			reqProcessor.Flags.HasKey(utils.MetaUpdate)
		rply := new(sessions.V1ProcessEventReply)
		err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1ProcessEvent,
			evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needMaxUsage {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
		// case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	// if reqProcessor.Flags.HasKey(utils.MetaCDRs) &&
	// 	!reqProcessor.Flags.HasKey(utils.MetaDryRun) {
	// 	rplyCDRs := utils.StringPointer("")
	// 	if err = sa.connMgr.Call(sa.cfg.SIPAgentCfg().SessionSConns, nil, utils.SessionSv1ProcessCDR,
	// 		&utils.CGREventWithArgDispatcher{CGREvent: cgrEv,
	// 			ArgDispatcher: cgrArgs.ArgDispatcher},
	// 		rplyCDRs); err != nil {
	// 		agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(err.Error()))
	// 	}
	// }
	if err := agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, SIP reply: %s",
				utils.SIPAgent, agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, SIP reply: %s",
				utils.SIPAgent, agReq.Reply))
	}
	return true, nil
}
