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
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/osipsdagram"
	"time"
)

func NewOSipsSessionManager(cfg *config.CGRConfig, cdrsrv engine.Connector) (*OsipsSessionManager, error) {
	return &OsipsSessionManager{cfg: cfg, cdrsrv: cdrsrv}, nil
}

type OsipsSessionManager struct {
	cfg    *config.CGRConfig
	rater  engine.Connector
	cdrsrv engine.Connector
}

func (osm *OsipsSessionManager) Connect() (err error) {
	addr := ":2020"
	evsrv, err := osipsdagram.NewEventServer(addr,
		map[string][]func(*osipsdagram.OsipsEvent){
			"E_ACC_CDR": []func(*osipsdagram.OsipsEvent){osm.OnCdr},
		})
	if err != nil {
		fmt.Printf("Cannot initiate OpenSIPS Datagram Server: %s", err.Error())
		return
	}
	engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Started listening for event datagrams at <%s>", addr))
	evsrv.ServeEvents()
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

func (osm *OsipsSessionManager) OnCdr(cdrDagram *osipsdagram.OsipsEvent) {
	engine.Logger.Info(fmt.Sprintf("<SM-OpenSIPSr> Received cdr datagram: %+v", cdrDagram))
	var reply *string
	osipsEv, _ := NewOsipsEvent(cdrDagram)
	osm.cdrsrv.ProcessCdr(osipsEv.AsStoredCdr(), reply)
}
