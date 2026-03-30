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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/aringo"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// constants used by AsteriskAgent
const (
	CGRAuthAPP               = "cgrates_auth"
	CGRMaxSessionTime        = "CGRMaxSessionTime"
	CGRSupplier              = "CGRSupplier"
	ARIStasisStart           = "StasisStart"
	ARIChannelStateChange    = "ChannelStateChange"
	ARIChannelDestroyed      = "ChannelDestroyed"
	ARIRESTResponse          = "RESTResponse"
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
	connMgr *engine.ConnManager) (*AsteriskAgent, error) {
	sma := &AsteriskAgent{
		cgrCfg:      cgrCfg,
		astConnIdx:  astConnIdx,
		connMgr:     connMgr,
		eventsCache: make(map[string]*utils.CGREventWithArgDispatcher),
	}
	return sma, nil
}

// ARIConnector abstracts the transport layer (HTTP or WebSocket) for sending ARI commands.
type ARIConnector interface {
	Call(method, uri string, queryStr map[string]string, body []byte) (aringo.RESTResponse, error)
}

// AsteriskAgent used to cominicate with asterisk
type AsteriskAgent struct {
	cgrCfg      *config.CGRConfig // Separate from smCfg since there can be multiple
	connMgr     *engine.ConnManager
	astConnIdx  int
	astConn     ARIConnector
	astEvChan   chan map[string]any
	astErrChan  chan error
	eventsCache map[string]*utils.CGREventWithArgDispatcher // used to gather information about events during various phases
	evCacheMux  sync.RWMutex                                // Protect eventsCache
}

func (sma *AsteriskAgent) connectAsterisk() (err error) {
	connCfg := sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx]
	sma.astEvChan = make(chan map[string]any)
	sma.astErrChan = make(chan error)
	if connCfg.AriWebSocket {
		sma.astConn, err = aringo.NewARInGO(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s",
			connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
			connCfg.User, connCfg.Password, fmt.Sprintf("%s@%s", utils.CGRateS, utils.VERSION),
			sma.astEvChan, sma.astErrChan, nil, connCfg.ConnectAttempts, connCfg.Reconnects,
			0, utils.FibDuration)
	} else {
		sma.astConn, err = aringo.NewARInGOV1(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s",
			connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
			connCfg.User, connCfg.Password, connCfg.Address, fmt.Sprintf("%s@%s", utils.CGRateS, utils.VERSION),
			sma.astEvChan, sma.astErrChan, nil, connCfg.ConnectAttempts, connCfg.Reconnects,
			0, utils.FibDuration)
	}
	if err != nil {
		return err
	}
	utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to Asterisk at: <%s>", utils.AsteriskAgent, connCfg.Address))
	return nil
}

func (sma *AsteriskAgent) filterEventTypes() {
	filterEv := struct {
		Allowed []map[string]string `json:"allowed"`
	}{
		Allowed: []map[string]string{
			{"type": ARIStasisStart},
			{"type": ARIChannelStateChange},
			{"type": ARIChannelDestroyed},
			{"type": ARIRESTResponse},
		}}
	body, err := json.Marshal(filterEv)
	if err != nil {
		utils.Logger.Warning(err.Error())
		return
	}
	if _, err := sma.astConn.Call(aringo.HTTP_PUT,
		fmt.Sprintf("applications/%s/eventFilter", CGRAuthAPP), nil, body); err != nil {
		utils.Logger.Warning(err.Error())
	}
}

// ListenAndServe is called to start the service
func (sma *AsteriskAgent) ListenAndServe() (err error) {
	if err := sma.connectAsterisk(); err != nil {
		return err
	}
	sma.filterEventTypes()
	for {
		select {
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
		fmt.Sprintf("channels/%s/variable", chanID), // Asterisk having issue with variable terminating empty so harcoding param in url
		map[string]string{"variable": vrblName, "value": vrblVal}, nil); err != nil {
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
	if _, err := sma.astConn.Call(aringo.HTTP_DELETE, fmt.Sprintf("channels/%s", channelID),
		map[string]string{"reason": "congestion"}, nil); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> failed disconnecting channel <%s>, err: %s",
				utils.AsteriskAgent, channelID, err.Error()))
	}
}

func (sma *AsteriskAgent) handleStasisStart(ev *SMAsteriskEvent) {
	// Subscribe for channel updates even after we leave Stasis
	if _, err := sma.astConn.Call(aringo.HTTP_POST,
		fmt.Sprintf("applications/%s/subscription", CGRAuthAPP),
		map[string]string{"eventSource": fmt.Sprintf("channel:%s", ev.ChannelID())}, nil); err != nil {
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
		if authReply.MaxUsage == time.Duration(0) {
			sma.hangupChannel(ev.ChannelID(), "")
			return
		}
		//  Set absolute timeout for non-postpaid calls
		if !sma.setChannelVar(ev.ChannelID(), CGRMaxSessionTime,
			strconv.Itoa(int(authReply.MaxUsage.Milliseconds()))) {
			return
		}
	}
	if authReply.ResourceAllocation != nil {
		if !sma.setChannelVar(ev.ChannelID(),
			ARICGRResourceAllocation, *authReply.ResourceAllocation) {
			return
		}
	}
	if authReply.Suppliers != nil {
		for i, spl := range authReply.Suppliers.SortedSuppliers {
			if !sma.setChannelVar(ev.ChannelID(),
				CGRSupplier+strconv.Itoa(i+1), spl.SupplierID) {
				return
			}
		}
	}
	// Exit channel from stasis
	if _, err := sma.astConn.Call(
		aringo.HTTP_POST,
		fmt.Sprintf("channels/%s/continue", ev.ChannelID()), nil, nil); err != nil {
	}
	// Done with processing event, cache it for later use
	sma.evCacheMux.Lock()
	sma.eventsCache[ev.ChannelID()] = &utils.CGREventWithArgDispatcher{
		CGREvent:      authArgs.CGREvent,
		ArgDispatcher: authArgs.ArgDispatcher,
	}
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

	// Update the cached event with new channel state.
	sma.evCacheMux.Lock()
	err := ev.UpdateCGREvent(cgrEvDisp.CGREvent, sma.cgrCfg.AsteriskAgentCfg().AlterableFields)
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
	} else if initSessionArgs.InitSession && initReply.MaxUsage == time.Duration(0) {
		sma.hangupChannel(ev.ChannelID(), "")
		return
	}
}

// Channel disconnect
func (sma *AsteriskAgent) handleChannelDestroyed(ev *SMAsteriskEvent) {
	sma.evCacheMux.RLock()
	cgrEvDisp, hasIt := sma.eventsCache[ev.ChannelID()]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}

	// Update the cached event with new channel state.
	sma.evCacheMux.Lock()
	err := ev.UpdateCGREvent(cgrEvDisp.CGREvent, sma.cgrCfg.AsteriskAgentCfg().AlterableFields)
	sma.evCacheMux.Unlock()
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s when attempting to destroy session for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		return
	}

	// populate terminate session args
	tsArgs := ev.V1TerminateSessionArgs(*cgrEvDisp)
	if tsArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate terminate session arguments",
			utils.AsteriskAgent, ev.ChannelID()))
		return
	}

	var reply string
	if err := sma.connMgr.Call(sma.cgrCfg.AsteriskAgentCfg().SessionSConns, sma,
		utils.SessionSv1TerminateSession,
		tsArgs, &reply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to terminate session for channelID: %s",
			utils.AsteriskAgent, err.Error(), ev.ChannelID()))
	}
	if sma.cgrCfg.AsteriskAgentCfg().CreateCDR {
		if err := sma.connMgr.Call(sma.cgrCfg.AsteriskAgentCfg().SessionSConns, sma,
			utils.SessionSv1ProcessCDR,
			cgrEvDisp, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to process CDR for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
	}

}

// ServiceShutdown is called to shutdown the service
func (sma *AsteriskAgent) ServiceShutdown() error {
	return nil
}

// V1DisconnectSession is internal method to disconnect session in asterisk
func (sma *AsteriskAgent) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) error {
	channelID := engine.NewMapEvent(args.EventStart).GetStringIgnoreErrors(utils.OriginID)
	sma.hangupChannel(channelID, "")
	*reply = utils.OK
	return nil
}

// Call implements birpc.ClientConnector interface
func (sma *AsteriskAgent) Call(ctx *context.Context, serviceMethod string, args any, reply any) error {
	return utils.RPCCall(sma, serviceMethod, args, reply)
}

// V1GetActiveSessionIDs is internal method to  get all active sessions in asterisk
func (sma *AsteriskAgent) V1GetActiveSessionIDs(ignParam string,
	sessionIDs *[]*sessions.SessionID) error {
	var slMpIface []map[string]any // decode the result from ari into a slice of map[string]any
	restResp, err := sma.astConn.Call(aringo.HTTP_GET, "channels", nil, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(restResp.MessageBody), &slMpIface); err != nil {
		return err
	}
	var sIDs []*sessions.SessionID
	for _, mpIface := range slMpIface {
		sIDs = append(sIDs, &sessions.SessionID{
			OriginHost: strings.Split(sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0],
			OriginID:   mpIface["id"].(string)},
		)
	}
	*sessionIDs = sIDs
	return nil

}
