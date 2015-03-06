/*
Real-time Charging System for Telecom & ISP environments
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
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/osipsdagram"
)

func NewOSipsSessionManager(cfg *config.CGRConfig, rater, cdrsrv engine.Connector) (*OsipsSessionManager, error) {
	osm := &OsipsSessionManager{cgrCfg: cfg, rater: rater, cdrsrv: cdrsrv}
	osm.eventHandlers = map[string][]func(*osipsdagram.OsipsEvent){
		"E_OPENSIPS_START": []func(*osipsdagram.OsipsEvent){osm.OnOpensipsStart},
		"E_ACC_CDR":        []func(*osipsdagram.OsipsEvent){osm.onCdr},
		"E_CGR_AUTHORIZE":  []func(*osipsdagram.OsipsEvent){osm.OnAuthorize},
	}
	return osm, nil
}

type OsipsSessionManager struct {
	cgrCfg          *config.CGRConfig
	rater           engine.Connector
	cdrsrv          engine.Connector
	eventHandlers   map[string][]func(*osipsdagram.OsipsEvent)
	evSubscribeStop *chan struct{} // Reference towards the channel controlling subscriptions, keep it as reference so we do not need to copy it
	stopServing     chan struct{}  // Stop serving datagrams
	miConn          *osipsdagram.OsipsMiDatagramConnector
}

func (osm *OsipsSessionManager) Connect() (err error) {
	osm.stopServing = make(chan struct{})
	if osm.miConn, err = osipsdagram.NewOsipsMiDatagramConnector(osm.cgrCfg.OsipsMiAddr, osm.cgrCfg.OsipsReconnects); err != nil {
		return fmt.Errorf("Cannot connect to OpenSIPS at %s, error: %s", osm.cgrCfg.OsipsMiAddr, err.Error())
	}
	evSubscribeStop := make(chan struct{})
	osm.evSubscribeStop = &evSubscribeStop
	defer close(*osm.evSubscribeStop) // Stop subscribing on disconnect
	go osm.SubscribeEvents(evSubscribeStop)
	evsrv, err := osipsdagram.NewEventServer(osm.cgrCfg.OsipsListenUdp, osm.eventHandlers)
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Cannot initialize datagram server, error: <%s>", err.Error()))
		return
	}
	engine.Logger.Info(fmt.Sprintf("<SM-OpenSIPS> Listening for datagram events at <%s>", osm.cgrCfg.OsipsListenUdp))
	evsrv.ServeEvents(osm.stopServing) // Will break through stopServing on error in other places
	return errors.New("<SM-OpenSIPS> Stopped reading events")
}

func (osm *OsipsSessionManager) DisconnectSession(ev utils.Event, cgrId, notify string) {
	return
}
func (osm *OsipsSessionManager) RemoveSession(uuid string) {
	return
}
func (osm *OsipsSessionManager) MaxDebit(cd *engine.CallDescriptor, cc *engine.CallCost) error {
	return nil
}
func (osm *OsipsSessionManager) DebitInterval() time.Duration {
	var nilDuration time.Duration
	return nilDuration
}
func (osm *OsipsSessionManager) DbLogger() engine.LogStorage {
	return nil
}
func (osm *OsipsSessionManager) Rater() engine.Connector {
	return osm.rater
}
func (osm *OsipsSessionManager) WarnSessionMinDuration(sessionUuid, connId string) {
	return
}
func (osm *OsipsSessionManager) Shutdown() error {
	return nil
}

// Event Handlers

// Automatic subscribe to OpenSIPS for events, trigered on Connect or OpenSIPS restart
func (osm *OsipsSessionManager) SubscribeEvents(evStop chan struct{}) error {
	for {
		select {
		case <-evStop: // Break this loop from outside
			return nil
		default:
			subscribeInterval := osm.cgrCfg.OsipsEvSubscInterval + time.Duration(1)*time.Second // Avoid concurrency on expiry
			listenAddrSplt := strings.Split(osm.cgrCfg.OsipsListenUdp, ":")
			portListen := listenAddrSplt[1]
			addrListen := listenAddrSplt[0]
			if len(addrListen) == 0 { //Listen on all addresses, try finding out from mi connection
				if localAddr := osm.miConn.LocallAddr(); localAddr != nil {
					addrListen = strings.Split(localAddr.String(), ":")[0]
				}
			}
			for eventName := range osm.eventHandlers {
				if eventName == "E_OPENSIPS_START" { // Do not subscribe for start since this should be hardcoded
					continue
				}
				cmd := fmt.Sprintf(":event_subscribe:\n%s\nudp:%s:%s\n%d\n", eventName, addrListen, portListen, int(subscribeInterval.Seconds()))
				success := false
				for attempts := 0; attempts < osm.cgrCfg.OsipsReconnects; attempts++ {
					if reply, err := osm.miConn.SendCommand([]byte(cmd)); err == nil && bytes.HasPrefix(reply, []byte("200 OK")) {
						success = true
						break
					}
					time.Sleep(time.Duration((attempts+1)/2) * time.Second) // Allow OpenSIPS to recover from errors
					continue                                                // Try again
				}
				if !success {
					engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Shutting down, failed subscribing to OpenSIPS at address: <%s>", osm.cgrCfg.OsipsMiAddr))
					close(osm.stopServing) // Do not serve anymore since we got errors on subscribing
					return errors.New("Failed subscribing to OpenSIPS events")
				}
			}
			time.Sleep(osm.cgrCfg.OsipsEvSubscInterval)
		}
	}
}

func (osm *OsipsSessionManager) OnOpensipsStart(cdrDagram *osipsdagram.OsipsEvent) {
	close(*osm.evSubscribeStop) // Cancel previous subscribes
	evStop := make(chan struct{})
	osm.evSubscribeStop = &evStop
	go osm.SubscribeEvents(evStop)
}

func (osm *OsipsSessionManager) onCdr(cdrDagram *osipsdagram.OsipsEvent) {
	osipsEv, _ := NewOsipsEvent(cdrDagram)
	if err := osm.ProcessCdr(osipsEv.AsStoredCdr()); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", osipsEv.GetCgrId(), osipsEv.GetUUID(), err.Error()))
	}

}

func (osm *OsipsSessionManager) ProcessCdr(storedCdr *utils.StoredCdr) error {
	var reply string
	return osm.cdrsrv.ProcessCdr(storedCdr, &reply)
}

// Process Authorize request from OpenSIPS and communicate back maxdur
func (osm *OsipsSessionManager) OnAuthorize(osipsDagram *osipsdagram.OsipsEvent) {
	ev, _ := NewOsipsEvent(osipsDagram)
	if ev.MissingParameter() {
		cmdNotify := fmt.Sprintf(":cache_store:\nlocal\n%s/cgr_notify\n%s\n2\n\n", ev.GetUUID(), utils.ERR_MANDATORY_IE_MISSING)
		if reply, err := osm.miConn.SendCommand([]byte(cmdNotify)); err != nil || !bytes.HasPrefix(reply, []byte("200 OK")) {
			engine.Logger.Err(fmt.Sprintf("Failed setting cgr_notify variable for accid: %s, err: %v, reply: %s", ev.GetUUID(), err, string(reply)))
		}
		return
	}
	var maxCallDuration time.Duration // This will be the maximum duration this channel will be allowed to last
	var durInitialized bool
	attrsDC := utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT), Category: ev.GetCategory(utils.META_DEFAULT), Direction: ev.GetDirection(utils.META_DEFAULT),
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	var dcs utils.DerivedChargers
	if err := osm.rater.GetDerivedChargers(attrsDC, &dcs); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> OnAuthorize: could not get derived charging for event %s: %s", ev.GetUUID(), err.Error()))
		cmdNotify := fmt.Sprintf(":cache_store:\nlocal\n%s/cgr_notify\n%s\n2\n\n", ev.GetUUID(), utils.ERR_SERVER_ERROR)
		if reply, err := osm.miConn.SendCommand([]byte(cmdNotify)); err != nil || !bytes.HasPrefix(reply, []byte("200 OK")) {
			engine.Logger.Err(fmt.Sprintf("Failed setting cgr_notify variable for accid: %s, err: %v, reply: %s", ev.GetUUID(), err, string(reply)))
		}
		return
	}
	dcs, _ = dcs.AppendDefaultRun()
	for _, dc := range dcs {
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if fltrPass, _ := ev.PassesFieldFilter(dcRunFilter); !fltrPass {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched
			continue
		}
		startTime, err := ev.GetSetupTime(utils.META_DEFAULT)
		if err != nil {
			engine.Logger.Err("Error parsing answer event start time, using time.Now!")
			startTime = time.Now()
		}
		cd := engine.CallDescriptor{
			Direction:   ev.GetDirection(dc.DirectionField),
			Tenant:      ev.GetTenant(dc.TenantField),
			Category:    ev.GetCategory(dc.CategoryField),
			Subject:     ev.GetSubject(dc.SubjectField),
			Account:     ev.GetAccount(dc.AccountField),
			Destination: ev.GetDestination(dc.DestinationField),
			TimeStart:   startTime,
			TimeEnd:     startTime.Add(osm.cgrCfg.SMMaxCallDuration),
		}
		var remainingDurationFloat float64
		err = osm.rater.GetMaxSessionTime(cd, &remainingDurationFloat)
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("Could not get max session time for %s: %v", ev.GetUUID(), err))
			cmdNotify := fmt.Sprintf(":cache_store:\nlocal\n%s/cgr_notify\n%s\n2\n\n", ev.GetUUID(), utils.ERR_SERVER_ERROR)
			if reply, err := osm.miConn.SendCommand([]byte(cmdNotify)); err != nil || !bytes.HasPrefix(reply, []byte("200 OK")) {
				engine.Logger.Err(fmt.Sprintf("Failed setting cgr_notify variable for accid: %s, err: %v, reply: %s", ev.GetUUID(), err, string(reply)))
			}
			return
		}
		remainingDuration := time.Duration(remainingDurationFloat)
		// Set maxCallDuration, smallest out of all forked sessions
		if !durInitialized { // first time we set it /not initialized yet
			maxCallDuration = remainingDuration
			durInitialized = true
		} else if maxCallDuration > remainingDuration {
			maxCallDuration = remainingDuration
		}
	}
	if maxCallDuration <= osm.cgrCfg.SMMinCallDuration {
		cmdNotify := fmt.Sprintf(":cache_store:\nlocal\n%s/cgr_notify\n%s\n2\n\n", ev.GetUUID(), OSIPS_INSUFFICIENT_FUNDS)
		if reply, err := osm.miConn.SendCommand([]byte(cmdNotify)); err != nil || !bytes.HasPrefix(reply, []byte("200 OK")) {
			engine.Logger.Err(fmt.Sprintf("Failed setting cgr_notify variable for accid: %s, err: %v, reply: %s", ev.GetUUID(), err, string(reply)))
		}
		return
	}
	cmdMaxDur := fmt.Sprintf(":cache_store:\nlocal\n%s/cgr_maxdur\n%d\n\n", ev.GetUUID(), int(maxCallDuration.Seconds()))
	if reply, err := osm.miConn.SendCommand([]byte(cmdMaxDur)); err != nil || !bytes.HasPrefix(reply, []byte("200 OK")) {
		engine.Logger.Err(fmt.Sprintf("Failed setting cgr_maxdur variable for accid: %s, err: %v, reply: %s", ev.GetUUID(), err, string(reply)))
	}
	cmdNotify := fmt.Sprintf(":cache_store:\nlocal\n%s/cgr_notify\n%s\n", ev.GetUUID(), OSIPS_AUTH_OK)
	if reply, err := osm.miConn.SendCommand([]byte(cmdNotify)); err != nil || !bytes.HasPrefix(reply, []byte("200 OK")) {
		engine.Logger.Err(fmt.Sprintf("Failed setting cgr_notify variable for accid: %s, err: %v, reply: %s", ev.GetUUID(), err, string(reply)))
	}
}
