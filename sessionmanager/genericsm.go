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

func NewGenericSessionManager(cgrCfg *config.CGRConfig, rater engine.Connector, cdrsrv engine.Connector, timezone string) *GenericSessionManager {
	gsm := &GenericSessionManager{cgrCfg: cgrCfg, rater: rater, cdrsrv: cdrsrv, timezone: timezone, conns: make(map[string]*rpc2.Client), connMux: new(sync.Mutex),
		sessions: make(map[string][]*GenericSession), sessionsMux: new(sync.Mutex), guard: engine.NewGuardianLock()}
	return gsm
}

type GenericSessionManager struct {
	cgrCfg      *config.CGRConfig // Separate from smCfg since there can be multiple
	rater       engine.Connector
	cdrsrv      engine.Connector
	timezone    string
	conns       map[string]*rpc2.Client
	connMux     *sync.Mutex
	sessions    map[string][]*GenericSession //Group sessions per sessionId, multiple runs based on derived charging
	sessionsMux *sync.Mutex
	guard       *engine.GuardianLock // Used to lock on uuid
}

func (self *GenericSessionManager) indexSession(uuid string, s *GenericSession) {
	self.sessionsMux.Lock()
	self.sessions[uuid] = append(self.sessions[uuid], s)
	self.sessionsMux.Unlock()
}

// Remove session from session list, removes all related in case of multiple runs, true if item was found
func (self *GenericSessionManager) unindexSession(uuid string) bool {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	if _, hasIt := self.sessions[uuid]; !hasIt {
		return false
	}
	delete(self.sessions, uuid)
	return true
}

// Returns all sessions handled by the SM
func (self *GenericSessionManager) getSessions() map[string][]*GenericSession {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	return self.sessions
}

// Returns sessions/derived for a specific uuid
func (self *GenericSessionManager) getSession(uuid string) []*GenericSession {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	return self.sessions[uuid]
}

// Handle a new session, pass the connectionId so we can communicate on disconnect request
func (self *GenericSessionManager) sessionStart(evStart SMGenericEvent, connId string) error {
	sessionId := evStart.GetUUID()
	_, err := self.guard.Guard(func() (interface{}, error) { // Lock it on UUID level
		var sessionRuns []*engine.SessionRun
		if err := self.rater.GetSessionRuns(evStart.AsStoredCdr(self.cgrCfg, self.timezone), &sessionRuns); err != nil {
			return nil, err
		} else if len(sessionRuns) == 0 {
			return nil, nil
		}
		stopDebitChan := make(chan struct{})
		for _, sessionRun := range sessionRuns {
			s := &GenericSession{eventStart: evStart, connId: connId, runId: sessionRun.DerivedCharger.RunId, cd: sessionRun.CallDescriptor}
			self.indexSession(sessionId, s)
			if self.cgrCfg.SmGenericConfig.DebitInterval != 0 {
				s.stopDebit = stopDebitChan
				go s.debitLoop(self.cgrCfg.SmGenericConfig.DebitInterval)
			}
		}
		return nil, nil
	}, time.Duration(3)*time.Second, sessionId)
	return err
}

// End a session from outside
func (self *GenericSessionManager) sessionEnd(s *GenericSession, evStop *SMGenericEvent) error {
	_, err := self.guard.Guard(func() (interface{}, error) { // Lock it on UUID level
		if !self.unindexSession(s.eventStart.GetUUID()) { // Unreference it early so we avoid concurrency
			return nil, nil // Did not find the session so no need to close it anymore
		}
		//if err := s.Close(evStop); err != nil { // Stop loop, refund advanced charges and save the costs deducted so far to database
		//	return nil, err
		//}
		return nil, nil
	}, time.Duration(2)*time.Second, s.eventStart.GetUUID())
	return err
}

// Index the client connection so we can use it to communicate back
func (self *GenericSessionManager) OnClientConnect(clnt *rpc2.Client) {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	connId := utils.GenUUID()
	clnt.State.Set(CGR_CONNUUID, connId) // Set unique id for the connection so we can identify it later in requests
	self.conns[connId] = clnt
}

// Unindex the client connection so we can use it to communicate back
func (self *GenericSessionManager) OnClientDisconnect(clnt *rpc2.Client) {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	if connId := getClientConnId(clnt); connId != "" {
		delete(self.conns, connId)
	}
}

// Methods to apply on sessions, mostly exported through RPC/Bi-RPC
//Calculates maximum usage allowed for gevent
func (self *GenericSessionManager) GetMaxUsage(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	gev[utils.EVENT_NAME] = utils.CGR_AUTHORIZATION
	storedCdr := gev.AsStoredCdr(config.CgrConfig(), self.timezone)
	var maxDur float64
	if err := self.rater.GetDerivedMaxSessionTime(storedCdr, &maxDur); err != nil {
		return time.Duration(0), err
	}
	return time.Duration(maxDur), nil
}

func (self *GenericSessionManager) GetLcrSuppliers(gev SMGenericEvent, clnt *rpc2.Client) ([]string, error) {
	gev[utils.EVENT_NAME] = utils.CGR_LCR_REQUEST
	cd, err := gev.AsLcrRequest().AsCallDescriptor(self.timezone)
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
func (self *GenericSessionManager) SessionStart(gev SMGenericEvent, clnt *rpc2.Client) error {
	return nil
}

// Interim updates
func (self *GenericSessionManager) SessionUpdate(gev SMGenericEvent, clnt *rpc2.Client) error {
	return nil
}

// Called on session end, should stop debit loop
func (self *GenericSessionManager) SessionEnd(gev SMGenericEvent, clnt *rpc2.Client) error {
	return nil
}

func (self *GenericSessionManager) ProcessCdr(gev SMGenericEvent) error {
	return nil
}

func (self *GenericSessionManager) Connect() error {
	return nil
}

func (self *GenericSessionManager) Shutdown() error {
	return nil
}
