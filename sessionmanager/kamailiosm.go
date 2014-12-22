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
	if err = self.rater.GetDerivedMaxSessionTime(kev.New(""), &remainingDuration); err != nil {
		engine.Logger.Err(fmt.Sprintf("Could not get max session time for %s: %v", kev.GetUUID(), err))
	}
	if remainingDuration == -1.0 { // Unlimited
		return
	}
}

func (self *KamailioSessionManager) Connect() error {
	var err error
	eventHandlers := map[*regexp.Regexp][]func([]byte){
		regexp.MustCompile(".*"): []func([]byte){self.onCgrAuth},
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
func (self *KamailioSessionManager) Shutdown() error {
	return nil
}
