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
	"regexp"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/sipingo"
)

const (
	bufferSize      = 5000
	ackMethod       = "ACK"
	inviteMethod    = "INVITE"
	requestHeader   = "Request"
	callIDHeader    = "Call-ID"
	fromHeader      = "From"
	sipServerErr    = "SIP/2.0 500 Internal Server Error"
	userAgentHeader = "User-Agent"
	method          = "Method"
)

var (
	sipTagRgx = regexp.MustCompile(`tag=([^ ,;>]*)`)
)

// NewSIPAgent will construct a SIPAgent
func NewSIPAgent(connMgr *engine.ConnManager, cfg *config.CGRConfig,
	filterS *engine.FilterS) (sa *SIPAgent, err error) {
	sa = &SIPAgent{
		connMgr:  connMgr,
		filterS:  filterS,
		cfg:      cfg,
		ackMap:   make(map[string]chan struct{}),
		stopChan: make(chan struct{}),
	}
	msgTemplates := sa.cfg.TemplatesCfg()
	// Inflate *template field types
	for _, procsr := range sa.cfg.SIPAgentCfg().RequestProcessors {
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
	return
}

// SIPAgent is a handler for SIP requests
type SIPAgent struct {
	connMgr  *engine.ConnManager
	filterS  *engine.FilterS
	cfg      *config.CGRConfig
	stopChan chan struct{}
	ackMap   map[string]chan struct{}
	ackLocks sync.RWMutex
}

// Shutdown will stop the SIPAgent server
func (sa *SIPAgent) Shutdown() {
	sa.ackLocks.Lock()
	for _, ch := range sa.ackMap { // close all ack
		close(ch)
	}
	sa.ackLocks.Unlock()
	close(sa.stopChan)
}

// ListenAndServe will run the SIP handler doing also the connection to listen address
func (sa *SIPAgent) ListenAndServe() (err error) {
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
func (sa *SIPAgent) InitStopChan() {
	sa.stopChan = make(chan struct{})
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
		go func(message string, saddr net.Addr, conn net.PacketConn) {
			sa.answerMessage(message, saddr.String(), func(ans []byte) (werr error) {
				_, werr = conn.WriteTo(ans, saddr)
				return
			}) // do not log the received error because is already logged in function so for now just ignore it
			wg.Done()
		}(string(buf[:n]), saddr, conn)
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

				sa.answerMessage(string(buf[:n]), conn.LocalAddr().String(), func(ans []byte) (werr error) {
					_, werr = conn.Write(ans)
					return
				}) // do not log the received error because is already logged in function so for now just ignore it
			}
		}(conn)
	}
}

func (sa *SIPAgent) answerMessage(messageStr, addr string, write func(ans []byte) error) (err error) {
	var sipMessage sipingo.Message // recreate map SIP
	if sipMessage, err = sipingo.NewMessage(messageStr); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s parsing message: %s",
				utils.SIPAgent, err.Error(), messageStr))
		return // do we need to return error in case we can't parse the message?
	}
	tags := sipTagRgx.FindStringSubmatch(sipMessage[fromHeader])
	// in case we get a wrong sip message ( without tag in the From header) the next line should panic
	key := utils.ConcatenatedKey(sipMessage[callIDHeader], tags[1])
	method := sipMessage.MethodFrom(requestHeader)
	if ackMethod == method {
		if sa.cfg.SIPAgentCfg().RetransmissionTimer == 0 { // ignore ACK
			return
		}
		sa.ackLocks.Lock()
		if stopChan, has := sa.ackMap[key]; has {
			close(stopChan)
			sa.ackLocks.Unlock()
			return
		}
		sa.ackLocks.Unlock() // log the message if we did not find it in the map
	}
	var sipAnswer sipingo.Message
	if sipAnswer = sa.handleMessage(sipMessage, addr); len(sipAnswer) == 0 {
		return // do not write the message if we do not have anything to reply
	}
	ans := []byte(sipAnswer.String())
	if err = write(ans); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s sending message: %s",
				utils.SIPAgent, err.Error(), sipAnswer))
		return
	}
	// because we expext to send codes from 300-699 we wait for the ACK every time
	if method != inviteMethod || // only invitest need ACK
		sa.cfg.SIPAgentCfg().RetransmissionTimer == 0 {
		return // disabled ACK
	}
	stopChan := make(chan struct{})
	sa.ackLocks.Lock()
	sa.ackMap[key] = stopChan
	sa.ackLocks.Unlock()
	go func(stopChan chan struct{}, a []byte) {
		for {
			select {
			case <-time.After(sa.cfg.SIPAgentCfg().RetransmissionTimer):
				if err = write(ans); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s sending message: %s",
							utils.SIPAgent, err.Error(), sipAnswer))
					return
				}
			case <-stopChan:
				sa.ackLocks.Lock()
				delete(sa.ackMap, key)
				sa.ackLocks.Unlock()
				return
			}
		}
	}(stopChan, ans)
	return
}

func (sa *SIPAgent) handleMessage(sipMessage sipingo.Message, remoteHost string) (sipAnswer sipingo.Message) {
	if sipMessage[userAgentHeader] != "" {
		sipMessage[userAgentHeader] = fmt.Sprintf("%s@%s", utils.CGRateS, utils.Version)
	}
	sipMessageIface := make(map[string]interface{})
	for k, v := range sipMessage {
		sipMessageIface[k] = v
	}
	dp := utils.MapStorage(sipMessageIface)
	var processed bool
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rplyNM := utils.NewOrderedNavigableMap()
	opts := utils.MapStorage{}
	reqVars := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.RemoteHost: utils.NewLeafNode(remoteHost),
			method:           utils.NewLeafNode(sipMessage.MethodFrom(requestHeader)),
		},
	}
	// build the negative error answer
	sErr, err := sipErr(
		dp, sipMessage.Clone(), reqVars,
		sa.cfg.TemplatesCfg()[utils.MetaErr],
		sa.cfg.GeneralCfg().DefaultTenant,
		sa.cfg.GeneralCfg().DefaultTimezone,
		sa.filterS)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s building errSIP for message: %s",
				utils.SIPAgent, err.Error(), sipMessage))
		return bareSipErr(sipMessage, sipServerErr)
	}

	for _, reqProcessor := range sa.cfg.SIPAgentCfg().RequestProcessors {
		agReq := NewAgentRequest(dp, reqVars, cgrRplyNM, rplyNM,
			opts, reqProcessor.Tenant, sa.cfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone,
				config.CgrConfig().GeneralCfg().DefaultTimezone),
			sa.filterS, nil)
		var lclProcessed bool
		if lclProcessed, err = sa.processRequest(reqProcessor, agReq); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing request: %s",
					utils.SIPAgent, err.Error(), utils.ToJSON(agReq)))
			continue
		}
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
			break
		}
	}
	if err != nil { // write err message on conection 500 Server Error
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s from %s",
				utils.SIPAgent, err.Error(), sipMessage, remoteHost))
		return sErr
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.SIPAgent, sipMessage, remoteHost))
		return
	}
	if rplyNM.Empty() { // if we do not populate the reply with any field we do not send any reply back
		return
	}
	if err = updateSIPMsgFromNavMap(sipMessage, rplyNM); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s encoding out %s",
				utils.SIPAgent, err.Error(), utils.ToJSON(rplyNM)))
		return sErr
	}
	sipMessage.PrepareReply()
	return sipMessage
}

// processRequest represents one processor processing the request
func (sa *SIPAgent) processRequest(reqProcessor *config.RequestProcessor,
	agReq *AgentRequest) (processed bool, err error) {
	if pass, err := sa.filterS.Pass(context.TODO(), agReq.Tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if err = agReq.SetFields(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuthorize, /*
			utils.MetaInitiate, utils.MetaUpdate,
			utils.MetaTerminate, utils.MetaMessage,
			utils.MetaCDRs, */utils.MetaEvent, utils.MetaNone} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}

	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: %s, SIP message: %s",
				utils.SIPAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.SIPAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuthorize:
		rply := new(sessions.V1AuthorizeReply)
		err = sa.connMgr.Call(context.TODO(), sa.cfg.SIPAgentCfg().SessionSConns, utils.SessionSv1AuthorizeEvent,
			cgrEv, rply)
		rply.SetMaxUsageNeeded(utils.OptAsBool(cgrEv.APIOpts, utils.OptsAccountS))
		agReq.setCGRReply(rply, err)
	case utils.MetaEvent:
		rply := new(sessions.V1ProcessEventReply)
		err = sa.connMgr.Call(context.TODO(), sa.cfg.SIPAgentCfg().SessionSConns, utils.SessionSv1ProcessEvent,
			cgrEv, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else
		// if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
		// cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		// }
		agReq.setCGRReply(rply, err)
	}
	if err := agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
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
