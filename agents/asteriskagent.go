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
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/aringo"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

const (
	CGRAuthAPP            = "cgrates_auth"
	CGRMaxSessionTime     = "CGRMaxSessionTime"
	ARIStasisStart        = "StasisStart"
	ARIChannelStateChange = "ChannelStateChange"
	ARIChannelDestroyed   = "ChannelDestroyed"
	eventType             = "eventType"
	channelID             = "channelID"
	channelState          = "channelState"
	channelUp             = "Up"
	timestamp             = "timestamp"
	SMAAuthorization      = "SMA_AUTHORIZATION"
	SMASessionStart       = "SMA_SESSION_START"
	SMASessionTerminate   = "SMA_SESSION_TERMINATE"
)

func NewAsteriskAgent(cgrCfg *config.CGRConfig, astConnIdx int,
	smgConn *utils.BiRPCInternalClient) (*AsteriskAgent, error) {
	sma := &AsteriskAgent{cgrCfg: cgrCfg, smg: smgConn,
		eventsCache: make(map[string]*utils.CGREvent)}
	sma.smg.SetClientConn(sma) // pass the connection to SMA back into smg so we can receive the disconnects
	return sma, nil
}

type AsteriskAgent struct {
	cgrCfg      *config.CGRConfig // Separate from smCfg since there can be multiple
	astConnIdx  int
	smg         *utils.BiRPCInternalClient
	astConn     *aringo.ARInGO
	astEvChan   chan map[string]interface{}
	astErrChan  chan error
	eventsCache map[string]*utils.CGREvent // used to gather information about events during various phases
	evCacheMux  sync.RWMutex               // Protect eventsCache
}

func (sma *AsteriskAgent) connectAsterisk() (err error) {
	connCfg := sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx]
	sma.astEvChan = make(chan map[string]interface{})
	sma.astErrChan = make(chan error)
	sma.astConn, err = aringo.NewARInGO(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s",
		connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
		connCfg.User, connCfg.Password, fmt.Sprintf("%s %s", utils.CGRateS, utils.VERSION),
		sma.astEvChan, sma.astErrChan, connCfg.ConnectAttempts, connCfg.Reconnects)
	if err != nil {
		return err
	}
	return nil
}

// Called to start the service
func (sma *AsteriskAgent) ListenAndServe() (err error) {
	if err := sma.connectAsterisk(); err != nil {
		return err
	}
	for {
		select {
		case err = <-sma.astErrChan:
			return
		case astRawEv := <-sma.astEvChan:
			smAsteriskEvent := NewSMAsteriskEvent(astRawEv,
				strings.Split(sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0])
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
	panic("<AsteriskAgent> ListenAndServe out of select")
}

// hangupChannel will disconnect from CGRateS side with congestion reason
func (sma *AsteriskAgent) hangupChannel(channelID string) (err error) {
	_, err = sma.astConn.Call(aringo.HTTP_DELETE, fmt.Sprintf("http://%s/ari/channels/%s",
		sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, channelID),
		url.Values{"reason": {"congestion"}})
	return
}

func (sma *AsteriskAgent) handleStasisStart(ev *SMAsteriskEvent) {
	// Subscribe for channel updates even after we leave Stasis
	if _, err := sma.astConn.Call(aringo.HTTP_POST,
		fmt.Sprintf("http://%s/ari/applications/%s/subscription?eventSource=channel:%s",
			sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address,
			CGRAuthAPP, ev.ChannelID()), nil); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when subscribingto events for channelID: %s",
			utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		// Since we got error, disconnect channel
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	}
	//authorize Session
	authArgs := ev.V1AuthorizeArgs()
	if authArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate auth session arguments",
			utils.AsteriskAgent, ev.ChannelID()))
		return
	}
	var authReply sessions.V1AuthorizeReply
	if err := sma.smg.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authReply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to authorize session for channelID: %s",
			utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	}
	if authReply.MaxUsage != nil && *authReply.MaxUsage == time.Duration(0) {
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	} else if *authReply.MaxUsage != time.Duration(-1) {
		//  Set absolute timeout for non-postpaid calls
		if _, err := sma.astConn.Call(aringo.HTTP_POST,
			fmt.Sprintf("http://%s/ari/channels/%s/variable?variable=%s&value=%d", // Asterisk having issue with variable terminating empty so harcoding param in url
				sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address,
				ev.ChannelID(), CGRMaxSessionTime, int64(authReply.MaxUsage.Seconds()*1000)),
			url.Values{"value": {strconv.FormatFloat(
				authReply.MaxUsage.Seconds()*1000, 'f', -1, 64)}}); err != nil { // Asterisk expects value in ms
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when setting %s for channelID: %s",
				utils.AsteriskAgent, err.Error(), CGRMaxSessionTime, ev.ChannelID()))
			// Since we got error, disconnect channel
			if err := sma.hangupChannel(ev.ChannelID()); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
					utils.AsteriskAgent, err.Error(), ev.ChannelID()))
			}
			return
		}
	}

	// Exit channel from stasis
	if _, err := sma.astConn.Call(aringo.HTTP_POST, fmt.Sprintf("http://%s/ari/channels/%s/continue",
		sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address,
		ev.ChannelID()), nil); err != nil {
	}
	// Done with processing event, cache it for later use
	sma.evCacheMux.Lock()
	sma.eventsCache[ev.ChannelID()] = &authArgs.CGREvent
	sma.evCacheMux.Unlock()
}

// Ussually channelUP
func (sma *AsteriskAgent) handleChannelStateChange(ev *SMAsteriskEvent) {
	if ev.ChannelState() != channelUp {
		return
	}
	sma.evCacheMux.RLock()
	cgrEv, hasIt := sma.eventsCache[ev.ChannelID()]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}
	sma.evCacheMux.Lock()
	err := ev.UpdateCGREvent(cgrEv) // Updates the event directly in the cache
	sma.evCacheMux.Unlock()
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Error: %s when attempting to initiate session for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
					utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	}
	// populate init session args
	initSessionArgs := ev.V1InitSessionArgs(*cgrEv)
	if initSessionArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate init session arguments",
			utils.AsteriskAgent, ev.ChannelID()))
		return
	}

	//initit Session
	var initReply sessions.V1InitSessionReply
	if err := sma.smg.Call(utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Error: %s when attempting to initiate session for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
					utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	} else if initReply.MaxUsage != nil && *initReply.MaxUsage == time.Duration(0) {
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
					utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	}

}

// Channel disconnect
func (sma *AsteriskAgent) handleChannelDestroyed(ev *SMAsteriskEvent) {
	sma.evCacheMux.RLock()
	cgrEv, hasIt := sma.eventsCache[ev.ChannelID()]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}
	sma.evCacheMux.Lock()
	err := ev.UpdateCGREvent(cgrEv) // Updates the event directly in the cache
	sma.evCacheMux.Unlock()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to initiate session for channelID: %s",
			utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
		return
	}
	// populate terminate session args
	tsArgs := ev.V1TerminateSessionArgs(*cgrEv)
	if tsArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate terminate session arguments",
			utils.AsteriskAgent, ev.ChannelID()))
		return
	}

	var reply string
	if err := sma.smg.Call(utils.SessionSv1TerminateSession,
		tsArgs, &reply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to terminate session for channelID: %s",
			utils.AsteriskAgent, err.Error(), ev.ChannelID()))
	}
	if sma.cgrCfg.AsteriskAgentCfg().CreateCDR {
		if err := sma.smg.Call(utils.SessionSv1ProcessCDR, *cgrEv, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error: %s when attempting to process CDR for channelID: %s",
				utils.AsteriskAgent, err.Error(), ev.ChannelID()))
		}
	}

}

// Called to shutdown the service
func (sma *AsteriskAgent) ServiceShutdown() error {
	return nil
}

// Internal method to disconnect session in asterisk
func (sma *AsteriskAgent) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) error {
	channelID := sessions.SMGenericEvent(args.EventStart).GetOriginID(utils.META_DEFAULT)
	if err := sma.hangupChannel(channelID); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Error: %s when attempting to disconnect channelID: %s",
				utils.AsteriskAgent, err.Error(), channelID))
	}
	*reply = utils.OK
	return nil
}

// rpcclient.RpcClientConnection interface
func (sma *AsteriskAgent) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(sma, serviceMethod, args, reply)
}

func (fsa *AsteriskAgent) V1GetActiveSessionIDs(ignParam string, sessionIDs *[]*sessions.SessionID) (err error) {
	return utils.ErrNotImplemented
}
