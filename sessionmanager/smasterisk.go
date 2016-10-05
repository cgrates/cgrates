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
package sessionmanager

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cgrates/aringo"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
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

func NewSMAsterisk(cgrCfg *config.CGRConfig, astConnIdx int, smgConn *utils.BiRPCInternalClient) (*SMAsterisk, error) {
	sma := &SMAsterisk{cgrCfg: cgrCfg, smg: *smgConn, eventsCache: make(map[string]*SMGenericEvent)}
	sma.smg.SetClientConn(sma) // pass the connection to SMA back into smg so we can receive the disconnects
	return sma, nil
}

type SMAsterisk struct {
	cgrCfg      *config.CGRConfig // Separate from smCfg since there can be multiple
	astConnIdx  int
	smg         utils.BiRPCInternalClient
	astConn     *aringo.ARInGO
	astEvChan   chan map[string]interface{}
	astErrChan  chan error
	eventsCache map[string]*SMGenericEvent // used to gather information about events during various phases
	evCacheMux  sync.RWMutex               // Protect eventsCache
}

func (sma *SMAsterisk) connectAsterisk() (err error) {
	connCfg := sma.cgrCfg.SMAsteriskCfg().AsteriskConns[sma.astConnIdx]
	sma.astEvChan = make(chan map[string]interface{})
	sma.astErrChan = make(chan error)
	sma.astConn, err = aringo.NewARInGO(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s", connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
		connCfg.User, connCfg.Password, fmt.Sprintf("%s %s", utils.CGRateS, utils.VERSION), sma.astEvChan, sma.astErrChan, connCfg.ConnectAttempts, connCfg.Reconnects)
	if err != nil {
		return err
	}
	return nil
}

// Called to start the service
func (sma *SMAsterisk) ListenAndServe() (err error) {
	if err := sma.connectAsterisk(); err != nil {
		return err
	}
	for {
		select {
		case err = <-sma.astErrChan:
			return
		case astRawEv := <-sma.astEvChan:
			smAsteriskEvent := NewSMAsteriskEvent(astRawEv, strings.Split(sma.cgrCfg.SMAsteriskCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0])
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
	panic("<SMAsterisk> ListenAndServe out of select")
}

// hangupChannel will disconnect from CGRateS side with congestion reason
func (sma *SMAsterisk) hangupChannel(channelID string) (err error) {
	_, err = sma.astConn.Call(aringo.HTTP_DELETE, fmt.Sprintf("http://%s/ari/channels/%s",
		sma.cgrCfg.SMAsteriskCfg().AsteriskConns[sma.astConnIdx].Address, channelID),
		url.Values{"reason": {"congestion"}})
	return
}

func (sma *SMAsterisk) handleStasisStart(ev *SMAsteriskEvent) {
	// Subscribe for channel updates even after we leave Stasis
	if _, err := sma.astConn.Call(aringo.HTTP_POST, fmt.Sprintf("http://%s/ari/applications/%s/subscription?eventSource=channel:%s",
		sma.cgrCfg.SMAsteriskCfg().AsteriskConns[sma.astConnIdx].Address, CGRAuthAPP, ev.ChannelID()), nil); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when subscribing to events for channelID: %s", err.Error(), ev.ChannelID()))
		// Since we got error, disconnect channel
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	}
	// Query the SMG via RPC for maxUsage
	var maxUsage float64
	smgEv := ev.AsSMGenericEvent()
	if err := sma.smg.Call("SMGenericV1.MaxUsage", *smgEv, &maxUsage); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to authorize session for channelID: %s", err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	}
	if maxUsage == 0 {
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	} else if maxUsage != -1 {
		//  Set absolute timeout for non-postpaid calls
		if _, err := sma.astConn.Call(aringo.HTTP_POST, fmt.Sprintf("http://%s/ari/channels/%s/variable?variable=%s", // Asterisk having issue with variable terminating empty so harcoding param in url
			sma.cgrCfg.SMAsteriskCfg().AsteriskConns[sma.astConnIdx].Address, ev.ChannelID(), CGRMaxSessionTime),
			url.Values{"value": {strconv.FormatFloat(maxUsage*1000, 'f', -1, 64)}}); err != nil { // Asterisk expects value in ms
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when setting %s for channelID: %s", err.Error(), CGRMaxSessionTime, ev.ChannelID()))
			// Since we got error, disconnect channel
			if err := sma.hangupChannel(ev.ChannelID()); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
			}
			return
		}
	}

	// Exit channel from stasis
	if _, err := sma.astConn.Call(aringo.HTTP_POST, fmt.Sprintf("http://%s/ari/channels/%s/continue",
		sma.cgrCfg.SMAsteriskCfg().AsteriskConns[sma.astConnIdx].Address, ev.ChannelID()), nil); err != nil {
	}
	// Done with processing event, cache it for later use
	sma.evCacheMux.Lock()
	sma.eventsCache[ev.ChannelID()] = smgEv
	sma.evCacheMux.Unlock()
}

// Ussually channelUP
func (sma *SMAsterisk) handleChannelStateChange(ev *SMAsteriskEvent) {
	if ev.ChannelState() != channelUp {
		return
	}
	sma.evCacheMux.RLock()
	smgEv, hasIt := sma.eventsCache[ev.ChannelID()]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}
	sma.evCacheMux.Lock()
	err := ev.UpdateSMGEvent(smgEv) // Updates the event directly in the cache
	sma.evCacheMux.Unlock()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to initiate session for channelID: %s", err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	}
	var maxUsage float64
	if err := sma.smg.Call("SMGenericV1.InitiateSession", *smgEv, &maxUsage); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to initiate session for channelID: %s", err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	} else if maxUsage == 0 {
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	}
}

// Channel disconnect
func (sma *SMAsterisk) handleChannelDestroyed(ev *SMAsteriskEvent) {
	sma.evCacheMux.RLock()
	smgEv, hasIt := sma.eventsCache[ev.ChannelID()]
	sma.evCacheMux.RUnlock()
	if !hasIt { // Not handled by us
		return
	}
	sma.evCacheMux.Lock()
	err := ev.UpdateSMGEvent(smgEv) // Updates the event directly in the cache
	sma.evCacheMux.Unlock()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to initiate session for channelID: %s", err.Error(), ev.ChannelID()))
		if err := sma.hangupChannel(ev.ChannelID()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), ev.ChannelID()))
		}
		return
	}
	var reply string
	if err := sma.smg.Call("SMGenericV1.TerminateSession", *smgEv, &reply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to terminate session for channelID: %s", err.Error(), ev.ChannelID()))
	}
	if sma.cgrCfg.SMAsteriskCfg().CreateCDR {
		if err := sma.smg.Call("SMGenericV1.ProcessCDR", *smgEv, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to process CDR for channelID: %s", err.Error(), ev.ChannelID()))
		}
	}
}

// Called to shutdown the service
func (sma *SMAsterisk) ServiceShutdown() error {
	return nil
}

// Internal method to disconnect session in asterisk
func (sma *SMAsterisk) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) error {
	channelID := SMGenericEvent(args.EventStart).GetUUID()
	if err := sma.hangupChannel(channelID); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMAsterisk> Error: %s when attempting to disconnect channelID: %s", err.Error(), channelID))
	}
	*reply = utils.OK
	return nil
}

// rpcclient.RpcClientConnection interface
func (sma *SMAsterisk) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method
	method := reflect.ValueOf(sma).MethodByName(parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version in the method
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
