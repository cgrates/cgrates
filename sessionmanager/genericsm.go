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
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	CGR_CONNUUID = "cgr_connid"
)

// Attempts to get the connId previously set in the client state container
func getClientConnId(clnt *rpc2.Client) string {
	uuid, hasIt := clnt.State.Get(CGR_CONNUUID)
	if !hasIt {
		return ""
	}
	return uuid.(string)
}

func NewGenericSessionManager(cfg *config.SmGenericConfig, rater engine.Connector, cdrsrv engine.Connector, timezone string) *GenericSessionManager {
	gsm := &GenericSessionManager{cfg: cfg, rater: rater, cdrsrv: cdrsrv, timezone: timezone, conns: make(map[string]*rpc2.Client), sessions: NewSessions(), connMutex: new(sync.Mutex)}
	return gsm
}

type GenericSessionManager struct {
	cfg       *config.SmGenericConfig
	rater     engine.Connector
	cdrsrv    engine.Connector
	timezone  string
	conns     map[string]*rpc2.Client
	sessions  *Sessions
	connMutex *sync.Mutex
}

// Index the client connection so we can use it to communicate back
func (self *GenericSessionManager) OnClientConnect(clnt *rpc2.Client) {
	self.connMutex.Lock()
	defer self.connMutex.Unlock()
	connId := utils.GenUUID()
	clnt.State.Set(CGR_CONNUUID, connId) // Set unique id for the connection so we can identify it later in requests
	self.conns[connId] = clnt
}

// Unindex the client connection so we can use it to communicate back
func (self *GenericSessionManager) OnClientDisconnect(clnt *rpc2.Client) {
	self.connMutex.Lock()
	defer self.connMutex.Unlock()
	if connId := getClientConnId(clnt); connId != "" {
		delete(self.conns, connId)
	}
}

// Methods to apply on sessions, mostly exported through RPC/Bi-RPC
//Calculates maximum usage allowed for gevent
func (self *GenericSessionManager) GetMaxUsage(gev GenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	gev[utils.EVENT_NAME] = utils.CGR_AUTHORIZATION
	storedCdr := gev.AsStoredCdr(self.timezone)
	var maxDur float64
	if err := self.rater.GetDerivedMaxSessionTime(storedCdr, &maxDur); err != nil {
		return time.Duration(0), err
	}
	return time.Duration(maxDur), nil
}

func (self *GenericSessionManager) GetLcrSuppliers(gev GenericEvent, clnt *rpc2.Client) ([]string, error) {
	gev[utils.EVENT_NAME] = utils.CGR_LCR_REQUEST
	cd, err := gev.AsLcrRequest(self.timezone).AsCallDescriptor(self.timezone)
	if err != nil {
		return nil, err
	}
	var lcr engine.LCRCost
	if err = self.rater.GetLCR(&engine.AttrGetLcr{CallDescriptor: cd}, &lcr); err != nil {
		return nil, err
	}
	if lcr.HasErrors() {
		lcr.LogErrors()
		return nil, errors.New("LCR_COMPUTE_ERROR")
	}
	return lcr.SuppliersSlice()
}

// Called on session start
func (self *GenericSessionManager) SessionStart(gev GenericEvent, clnt *rpc2.Client) error {
	return nil
}

// Interim updates
func (self *GenericSessionManager) SessionUpdate(gev GenericEvent, clnt *rpc2.Client) error {
	return nil
}

// Called on session end, should stop debit loop
func (self *GenericSessionManager) SessionEnd(gev GenericEvent, clnt *rpc2.Client) error {
	return nil
}

// SessionManager interface methods
func (self *GenericSessionManager) Rater() engine.Connector {
	return self.rater
}

func (self *GenericSessionManager) CdrSrv() engine.Connector {
	return self.cdrsrv
}

func (self *GenericSessionManager) DebitInterval() time.Duration {
	return self.cfg.DebitInterval
}

func (self *GenericSessionManager) DisconnectSession(ev engine.Event, connId, notify string) error {
	return nil
}

func (sm *GenericSessionManager) WarnSessionMinDuration(sessionUuid, connId string) {}

func (self *GenericSessionManager) Sessions() []*Session {
	return self.sessions.getSessions()
}

func (self *GenericSessionManager) Timezone() string {
	return self.timezone
}

func (self *GenericSessionManager) ProcessCdr(gev GenericEvent) error {
	return nil
}

func (self *GenericSessionManager) Connect() error {
	return nil
}

func (self *GenericSessionManager) Shutdown() error {
	return nil
}
