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
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/aringo"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// constants used by AsteriskAgent
const (
	CGRAuthAPP               = "cgrates_auth"
	CGRMaxSessionTime        = "CGRMaxSessionTime"
	CGRRoute                 = "CGRRoute"
	ARIStasisStart           = "StasisStart"
	ARIChannelStateChange    = "ChannelStateChange"
	ARIChannelDestroyed      = "ChannelDestroyed"
	eventType                = "eventType"
	channelID                = "channelID"
	channelState             = "channelState"
	channelUp                = "Up"
	timestamp                = "timestamp"
	SMAAuthorization         = "SMA_AUTHORIZATION"
	SMASessionStart          = "SMA_SESSION_START"
	SMASessionTerminate      = "SMA_SESSION_TERMINATE"
	ARICGRResourceAllocation = "CGRResourceAllocation"
)

// NewAsteriskAgent constructs a new Asterisk Agent
func NewAsteriskAgent(cgrCfg *config.CGRConfig, astConnIdx int,
	connMgr *engine.ConnManager) *AsteriskAgent {
	sma := &AsteriskAgent{
		cgrCfg:      cgrCfg,
		astConnIdx:  astConnIdx,
		connMgr:     connMgr,
		eventsCache: make(map[string]*utils.CGREvent),
	}
	return sma
}

// AsteriskAgent used to cominicate with asterisk
type AsteriskAgent struct {
	cgrCfg      *config.CGRConfig // Separate from smCfg since there can be multiple
	connMgr     *engine.ConnManager
	astConnIdx  int
	astConn     *aringo.ARInGO
	astEvChan   chan map[string]interface{}
	astErrChan  chan error
	eventsCache map[string]*utils.CGREvent // used to gather information about events during various phases
	evCacheMux  sync.RWMutex               // Protect eventsCache
}

func (sma *AsteriskAgent) connectAsterisk(stopChan <-chan struct{}) (err error) {
	connCfg := sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx]
	sma.astEvChan = make(chan map[string]interface{})
	sma.astErrChan = make(chan error)
	sma.astConn, err = aringo.NewARInGO(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s",
		connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
		connCfg.User, connCfg.Password, fmt.Sprintf("%s@%s", utils.CGRateS, utils.Version),
		sma.astEvChan, sma.astErrChan, stopChan, connCfg.ConnectAttempts, connCfg.Reconnects)
	return
}

// ListenAndServe is called to start the service
func (sma *AsteriskAgent) ListenAndServe(stopChan <-chan struct{}) (err error) {
	if err = sma.connectAsterisk(stopChan); err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to Asterisk at: <%s>",
		utils.AsteriskAgent, sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address))
	for {
		select {
		case <-stopChan:
		case err = <-sma.astErrChan:
			return
		case astRawEv := <-sma.astEvChan:
			smAsteriskEvent := NewSMAsteriskEvent(astRawEv,
				strings.Split(sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0],
				sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Alias)

			switch smAsteriskEvent.EventType() {
			case ARIStasisStart:
				go sma.handleStasisStart(smAsteriskEvent)
			case ARIChannelStateChange:
				go sma.handleChannelStateChange(smAsteriskEvent)
			case ARIChannelDestroyed:
				go sma.handleChannelDestroyed(smAsteriskEvent)
			}
		}
	}
}

// setChannelVar will set the value of a variable
func (sma *AsteriskAgent) setChannelVar(chanID string, vrblName, vrblVal string) (success bool) {
	if _, err := sma.astConn.Call(aringo.HTTP_POST,
		fmt.Sprintf("http://%s/ari/channels/%s/variable?variable=%s&value=%s", // Asterisk having issue with variable terminating empty so harcoding param in url
			sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address,
			chanID, vrblName, vrblVal),
		nil); err != nil {
		// Since we got error, disconnect channel
		sma.hangupChannel(chanID,
			fmt.Sprintf("<%s> error: <%s> setting <%s> for channelID: <%s>",
				utils.AsteriskAgent, err.Error(), vrblName, chanID))
		return
	}
	return true
}

// hangupChannel will disconnect from CGRateS side with congestion reason
func (sma *AsteriskAgent) hangupChannel(channelID, warnMsg string) {
	if warnMsg != "" {
		utils.Logger.Warning(warnMsg)
	}
	if _, err := sma.astConn.Call(aringo.HTTP_DELETE, fmt.Sprintf("http://%s/ari/channels/%s",
		sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, channelID),
		url.Values{"reason": {"congestion"}}); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> failed disconnecting channel <%s>, err: %s",
				utils.AsteriskAgent, channelID, err.Error()))
	}
	return
}

func (sma *AsteriskAgent) handleStasisStart(ev *SMAsteriskEvent) {
	// Subscribe for channel updates even after we leave Stasis
	if _, err := sma.astConn.Call(aringo.HTTP_POST,
		fmt.Sprintf("http://%s/ari/applications/%s/subscription?eventSource=channel:%s",
			sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address,
			CGRAuthAPP, ev.ChannelID()), nil); err != nil {
		// Since we got error, disconnect channel
		sma.hangupChannel(ev.ChannelID(),
			fmt.Sprintf("<%s> error: %s subscribing for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		return
	}
	//authorize Session
	authArgs := ev.V1AuthorizeArgs()
	if authArgs == nil {
		sma.hangupChannel(ev.ChannelID(),
			fmt.Sprintf("<%s> event: %s cannot generate auth session arguments",
				utils.AsteriskAgent, ev.ChannelID()))
		return
	}
	var authReply sessions.V1AuthorizeReply
	if err := sma.connMgr.Call(sma.cgrCfg.AsteriskAgentCfg().SessionSConns, sma,
		utils.SessionSv1AuthorizeEvent, authArgs, &authReply); err != nil {
		sma.hangupChannel(ev.ChannelID(),
			fmt.Sprintf("<%s> error: %s authorizing session for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		return
	}
	if authReply.Attributes != nil {
		for _, fldName := range authReply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if _, has := authReply.Attributes.CGREvent.Event[fldName]; !has {
				continue //maybe removed
			}
			fldVal, err := authReply.Attributes.CGREvent.FieldAsString(fldName)
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> error <%s> extracting attribute field: <%s>",
						utils.AsteriskAgent, err.Error(), fldName))
			}
			if !sma.setChannelVar(ev.ChannelID(), fldName, fldVal) {
				return
			}
		}
	}
	if authArgs.GetMaxUsage {
		if authReply.MaxUsage == nil || *authReply.MaxUsage == time.Duration(0) {
			sma.hangupChannel(ev.ChannelID(), "")
			return
		}
		//  Set absolute timeout for non-postpaid calls
		if !sma.setChannelVar(ev.ChannelID(), CGRMaxSessionTime,
			strconv.Itoa(int(authReply.MaxUsage.Seconds()*1000))) {
			return
		}
	}
	if authReply.ResourceAllocation != nil {
		if !sma.setChannelVar(ev.ChannelID(),
			ARICGRResourceAllocation, *authReply.ResourceAllocation) {
			return
		}
	}
	if authReply.RouteProfiles != nil {
		for i, route := range authReply.RouteProfiles.RouteIDs() {
			if !sma.setChannelVar(ev.ChannelID(),
				CGRRoute+strconv.Itoa(i+1), route) {
				return
			}
		}
	}
	// Exit channel from stasis
	if _, err := sma.astConn.Call(
		aringo.HTTP_POST,
		fmt.Sprintf("http://%s/ari/channels/%s/continue",
			sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address,
			ev.ChannelID()), nil); err != nil {
	}
	// Done with processing event, cache it for later use
	sma.evCacheMux.Lock()
	sma.eventsCache[ev.ChannelID()] = authArgs.CGREvent
	sma.evCacheMux.Unlock()
}

// Ussually channelUP
func (sma *AsteriskAgent) handleChannelStateChange(ev *SMAsteriskEvent) {
	if ev.ChannelState() != channelUp {
		return
	}
	sma.evCacheMux.RLock()
	cgrEvDisp, hasIt := sma.eventsCache[ev.ChannelID()]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}
	sma.evCacheMux.Lock()
	err := ev.UpdateCGREvent(cgrEvDisp) // Updates the event directly in the cache
	sma.evCacheMux.Unlock()
	if err != nil {
		sma.hangupChannel(ev.ChannelID(),
			fmt.Sprintf("<%s> error: %s when attempting to initiate session for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		return
	}
	// populate init session args
	initSessionArgs := ev.V1InitSessionArgs(*cgrEvDisp)
	if initSessionArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate init session arguments",
			utils.AsteriskAgent, ev.ChannelID()))
		return
	}

	//initit Session
	var initReply sessions.V1InitSessionReply
	if err := sma.connMgr.Call(sma.cgrCfg.AsteriskAgentCfg().SessionSConns, sma,
		utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		sma.hangupChannel(ev.ChannelID(),
			fmt.Sprintf("<%s> error: %s when attempting to initiate session for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		return
	}
	if initSessionArgs.InitSession && (initReply.MaxUsage == nil || *initReply.MaxUsage == time.Duration(0)) {
		sma.hangupChannel(ev.ChannelID(), "")
		return
	}
}

// Channel disconnect
func (sma *AsteriskAgent) handleChannelDestroyed(ev *SMAsteriskEvent) {
	chID := ev.ChannelID()
	sma.evCacheMux.RLock()
	cgrEvDisp, hasIt := sma.eventsCache[chID]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}
	sma.evCacheMux.Lock()
	delete(sma.eventsCache, chID)       // delete the event from cache as we do not need to keep it here forever
	err := ev.UpdateCGREvent(cgrEvDisp) // Updates the event directly in the cache
	sma.evCacheMux.Unlock()
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s when attempting to destroy session for channelID: %s",
				utils.AsteriskAgent, err.Error(), chID))
		return
	}
	// populate terminate session args
	tsArgs := ev.V1TerminateSessionArgs(*cgrEvDisp)
	if tsArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate terminate session arguments",
			utils.AsteriskAgent, chID))
		return
	}

	var reply string
	if err := sma.connMgr.Call(sma.cgrCfg.AsteriskAgentCfg().SessionSConns, sma,
		utils.SessionSv1TerminateSession,
		tsArgs, &reply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to terminate session for channelID: %s",
			utils.AsteriskAgent, err.Error(), chID))
	}
	if sma.cgrCfg.AsteriskAgentCfg().CreateCDR {
		if err := sma.connMgr.Call(sma.cgrCfg.AsteriskAgentCfg().SessionSConns, sma,
			utils.SessionSv1ProcessCDR,
			cgrEvDisp, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to process CDR for channelID: %s",
				utils.AsteriskAgent, err.Error(), chID))
		}
	}

}

// Call implements rpcclient.ClientConnector interface
func (sma *AsteriskAgent) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(sma, serviceMethod, args, reply)
}

// V1DisconnectSession is internal method to disconnect session in asterisk
func (sma *AsteriskAgent) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) error {
	channelID := engine.NewMapEvent(args.EventStart).GetStringIgnoreErrors(utils.OriginID)
	sma.hangupChannel(channelID, "")
	*reply = utils.OK
	return nil
}

// V1GetActiveSessionIDs is internal method to  get all active sessions in asterisk
func (sma *AsteriskAgent) V1GetActiveSessionIDs(ignParam string,
	sessionIDs *[]*sessions.SessionID) error {
	var slMpIface []map[string]interface{} // decode the result from ari into a slice of map[string]interface{}
	if byts, err := sma.astConn.Call(
		aringo.HTTP_GET,
		fmt.Sprintf("http://%s/ari/channels",
			sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address),
		nil); err != nil {
		return err
	} else if err := json.Unmarshal(byts, &slMpIface); err != nil {
		return err
	}
	var sIDs []*sessions.SessionID
	for _, mpIface := range slMpIface {
		sIDs = append(sIDs, &sessions.SessionID{
			OriginHost: strings.Split(sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0],
			OriginID:   utils.IfaceAsString(mpIface["id"]),
		})
	}
	*sessionIDs = sIDs
	return nil

}

// V1ReAuthorize is used to implement the sessions.BiRPClient interface
func (*AsteriskAgent) V1ReAuthorize(originID string, reply *string) (err error) {
	return utils.ErrNotImplemented
}

// V1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (*AsteriskAgent) V1DisconnectPeer(args *utils.DPRArgs, reply *string) (err error) {
	return utils.ErrNotImplemented
}

// V1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (sma *AsteriskAgent) V1WarnDisconnect(args map[string]interface{}, reply *string) (err error) {
	return utils.ErrNotImplemented
}

// CallBiRPC is part of utils.BiRPCServer interface to help internal connections do calls over rpcclient.ClientConnector interface
func (sma *AsteriskAgent) CallBiRPC(clnt rpcclient.ClientConnector, serviceMethod string, args interface{}, reply interface{}) error {
	return utils.BiRPCCall(sma, clnt, serviceMethod, args, reply)
}

// BiRPCv1DisconnectSession is internal method to disconnect session in asterisk
func (sma *AsteriskAgent) BiRPCv1DisconnectSession(clnt rpcclient.ClientConnector, args utils.AttrDisconnectSession, reply *string) error {
	return sma.V1DisconnectSession(args, reply)
}

// BiRPCv1GetActiveSessionIDs is internal method to  get all active sessions in asterisk
func (sma *AsteriskAgent) BiRPCv1GetActiveSessionIDs(clnt rpcclient.ClientConnector, ignParam string, sessionIDs *[]*sessions.SessionID) error {
	return sma.V1GetActiveSessionIDs(ignParam, sessionIDs)

}

// BiRPCv1ReAuthorize is used to implement the sessions.BiRPClient interface
func (sma *AsteriskAgent) BiRPCv1ReAuthorize(clnt rpcclient.ClientConnector, originID string, reply *string) (err error) {
	return sma.V1ReAuthorize(originID, reply)
}

// BiRPCv1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (sma *AsteriskAgent) BiRPCv1DisconnectPeer(clnt rpcclient.ClientConnector, args *utils.DPRArgs, reply *string) (err error) {
	return sma.V1DisconnectPeer(args, reply)
}

// BiRPCv1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (sma *AsteriskAgent) BiRPCv1WarnDisconnect(clnt rpcclient.ClientConnector, args map[string]interface{}, reply *string) (err error) {
	return sma.V1WarnDisconnect(args, reply)
}

// Handlers is used to implement the rpcclient.BiRPCConector interface
func (sma *AsteriskAgent) Handlers() map[string]interface{} {
	return map[string]interface{}{
		utils.SessionSv1DisconnectSession: func(clnt *rpc2.Client, args utils.AttrDisconnectSession, rply *string) error {
			return sma.BiRPCv1DisconnectSession(clnt, args, rply)
		},
		utils.SessionSv1GetActiveSessionIDs: func(clnt *rpc2.Client, args string, rply *[]*sessions.SessionID) error {
			return sma.BiRPCv1GetActiveSessionIDs(clnt, args, rply)
		},
		utils.SessionSv1ReAuthorize: func(clnt *rpc2.Client, args string, rply *string) (err error) {
			return sma.BiRPCv1ReAuthorize(clnt, args, rply)
		},
		utils.SessionSv1DisconnectPeer: func(clnt *rpc2.Client, args *utils.DPRArgs, rply *string) (err error) {
			return sma.BiRPCv1DisconnectPeer(clnt, args, rply)
		},
		utils.SessionSv1WarnDisconnect: func(clnt *rpc2.Client, args map[string]interface{}, rply *string) (err error) {
			return sma.BiRPCv1WarnDisconnect(clnt, args, rply)
		},
	}
}
