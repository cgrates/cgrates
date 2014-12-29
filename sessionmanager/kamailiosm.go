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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
	"log/syslog"
	"regexp"
	"time"
)

func NewKamailioSessionManager(cfg *config.CGRConfig, rater, cdrsrv engine.Connector) (*KamailioSessionManager, error) {
	ksm := &KamailioSessionManager{cgrCfg: cfg, rater: rater, cdrsrv: cdrsrv}
	return ksm, nil
}

type KamailioSessionManager struct {
	cgrCfg *config.CGRConfig
	rater  engine.Connector
	cdrsrv engine.Connector
	kea    *kamevapi.KamEvapi
}

func (self *KamailioSessionManager) onCgrAuth(evData []byte) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		engine.Logger.Info(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", evData, err.Error()))
	}
	var remainingDuration float64
	if err = self.rater.GetDerivedMaxSessionTime(kev.AsEvent(""), &remainingDuration); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-Kamailio> Could not get max session time for %s, error: %s", kev.GetUUID(), err.Error()))
	}
	if kar, err := kev.AsKamAuthReply(remainingDuration, err); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed building auth reply %s", err.Error()))
	} else if err = self.kea.Send(kar.String()); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed sending auth reply %s", err.Error()))
	}
}

func (self *KamailioSessionManager) onCallStart(evData []byte) {
	_, err := NewKamEvent(evData)
	if err != nil {
		engine.Logger.Info(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", evData, err.Error()))
	}
}

func (self *KamailioSessionManager) onCallEnd(evData []byte) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		engine.Logger.Info(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", evData, err.Error()))
	}
	go self.ProcessCdr(kev)
}

func (self *KamailioSessionManager) Connect() error {
	var err error
	eventHandlers := map[*regexp.Regexp][]func([]byte){
		regexp.MustCompile("CGR_AUTH_REQUEST"): []func([]byte){self.onCgrAuth},
		regexp.MustCompile("CGR_CALL_START"):   []func([]byte){self.onCallStart},
		regexp.MustCompile("CGR_CALL_END"):     []func([]byte){self.onCallEnd},
	}
	if self.kea, err = kamevapi.NewKamEvapi(self.cgrCfg.KamailioEvApiAddr, self.cgrCfg.KamailioReconnects, eventHandlers, engine.Logger.(*syslog.Writer)); err != nil {
		return err
	}
	if err := self.kea.ReadEvents(); err != nil {
		return err
	}
	return errors.New("<SM-Kamailio> Stopped reading events")
}

func (self *KamailioSessionManager) DisconnectSession(uuid, notify, destnr string) {
	return
}
func (self *KamailioSessionManager) RemoveSession(uuid string) {
	return
}
func (self *KamailioSessionManager) MaxDebit(cd *engine.CallDescriptor, cc *engine.CallCost) error {
	return nil
}
func (self *KamailioSessionManager) GetDebitPeriod() time.Duration {
	var nilDuration time.Duration
	return nilDuration
}
func (self *KamailioSessionManager) GetDbLogger() engine.LogStorage {
	return nil
}
func (self *KamailioSessionManager) ProcessCdr(ev utils.Event) {
	if self.cdrsrv == nil {
		return
	}
	storedCdr := ev.AsStoredCdr()
	var reply string
	if err := self.cdrsrv.ProcessCdr(storedCdr, &reply); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", storedCdr.CgrId, storedCdr.AccId, err.Error()))
	}
}
func (self *KamailioSessionManager) Shutdown() error {
	return nil
}
