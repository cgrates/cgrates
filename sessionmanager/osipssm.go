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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/osipsdagram"
	"strings"
	"time"
)

func NewOSipsSessionManager(cfg *config.CGRConfig, cdrsrv engine.Connector) (*OsipsSessionManager, error) {
	osm := &OsipsSessionManager{cgrCfg: cfg, cdrsrv: cdrsrv}
	osm.eventHandlers = map[string][]func(*osipsdagram.OsipsEvent){
		"E_OPENSIPS_START": []func(*osipsdagram.OsipsEvent){osm.OnOpensipsStart},
		"E_ACC_CDR":        []func(*osipsdagram.OsipsEvent){osm.OnCdr},
	}
	return osm, nil
}

type OsipsSessionManager struct {
	cgrCfg          *config.CGRConfig
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

func (osm *OsipsSessionManager) DisconnectSession(uuid string, notify string) {
	return
}
func (osm *OsipsSessionManager) RemoveSession(uuid string) {
	return
}
func (osm *OsipsSessionManager) MaxDebit(cd *engine.CallDescriptor, cc *engine.CallCost) error {
	return nil
}
func (osm *OsipsSessionManager) GetDebitPeriod() time.Duration {
	var nilDuration time.Duration
	return nilDuration
}
func (osm *OsipsSessionManager) GetDbLogger() engine.LogStorage {
	return nil
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
					addrListen = strings.Split(localAddr.String(), ":")[1]
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
					close(osm.stopServing) // Do not serve anymore since we got errors on subscribing
					return errors.New("Failed subscribing to OpenSIPS events")
				}
			}
			time.Sleep(osm.cgrCfg.OsipsEvSubscInterval)
		}
	}
	return nil
}

func (osm *OsipsSessionManager) OnOpensipsStart(cdrDagram *osipsdagram.OsipsEvent) {
	close(*osm.evSubscribeStop) // Cancel previous subscribes
	evStop := make(chan struct{})
	osm.evSubscribeStop = &evStop
	go osm.SubscribeEvents(evStop)
}

func (osm *OsipsSessionManager) OnCdr(cdrDagram *osipsdagram.OsipsEvent) {
	var reply string
	osipsEv, _ := NewOsipsEvent(cdrDagram)
	storedCdr := osipsEv.AsStoredCdr()
	if err := osm.cdrsrv.ProcessCdr(storedCdr, &reply); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", storedCdr.CgrId, storedCdr.AccId, err.Error()))
	}
}
