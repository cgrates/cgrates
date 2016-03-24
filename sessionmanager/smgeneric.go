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
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var ErrPartiallyExecuted = errors.New("Partially executed")

func NewSMGeneric(cgrCfg *config.CGRConfig, rater engine.Connector, cdrsrv engine.Connector, timezone string, extconns *SMGExternalConnections) *SMGeneric {
	gsm := &SMGeneric{cgrCfg: cgrCfg, rater: rater, cdrsrv: cdrsrv, extconns: extconns, timezone: timezone,
		sessions: make(map[string][]*SMGSession), sessionsMux: new(sync.Mutex), guard: engine.NewGuardianLock()}
	return gsm
}

type SMGeneric struct {
	cgrCfg      *config.CGRConfig // Separate from smCfg since there can be multiple
	rater       engine.Connector
	cdrsrv      engine.Connector
	timezone    string
	sessions    map[string][]*SMGSession //Group sessions per sessionId, multiple runs based on derived charging
	extconns    *SMGExternalConnections  // Reference towards external connections manager
	sessionsMux *sync.Mutex              // Locks sessions map
	guard       *engine.GuardianLock     // Used to lock on uuid
}

func (self *SMGeneric) indexSession(uuid string, s *SMGSession) {
	self.sessionsMux.Lock()
	self.sessions[uuid] = append(self.sessions[uuid], s)
	self.sessionsMux.Unlock()
}

// Remove session from session list, removes all related in case of multiple runs, true if item was found
func (self *SMGeneric) unindexSession(uuid string) bool {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	if _, hasIt := self.sessions[uuid]; !hasIt {
		return false
	}
	delete(self.sessions, uuid)
	return true
}

// Returns all sessions handled by the SM
func (self *SMGeneric) getSessions() map[string][]*SMGSession {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	return self.sessions
}

// Returns sessions/derived for a specific uuid
func (self *SMGeneric) getSession(uuid string) []*SMGSession {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	return self.sessions[uuid]
}

// Handle a new session, pass the connectionId so we can communicate on disconnect request
func (self *SMGeneric) sessionStart(evStart SMGenericEvent, connId string) error {
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
			s := &SMGSession{eventStart: evStart, connId: connId, runId: sessionRun.DerivedCharger.RunID, timezone: self.timezone,
				rater: self.rater, cdrsrv: self.cdrsrv, cd: sessionRun.CallDescriptor}
			self.indexSession(sessionId, s)
			//utils.Logger.Info(fmt.Sprintf("<SMGeneric> Starting session: %s, runId: %s", sessionId, s.runId))
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
func (self *SMGeneric) sessionEnd(sessionId string, usage time.Duration) error {
	_, err := self.guard.Guard(func() (interface{}, error) { // Lock it on UUID level
		ss := self.getSession(sessionId)
		if len(ss) == 0 { // Not handled by us
			return nil, nil
		}
		if !self.unindexSession(sessionId) { // Unreference it early so we avoid concurrency
			return nil, nil // Did not find the session so no need to close it anymore
		}
		for idx, s := range ss {
			//utils.Logger.Info(fmt.Sprintf("<SMGeneric> Ending session: %s, runId: %s", sessionId, s.runId))
			if idx == 0 && s.stopDebit != nil {
				close(s.stopDebit) // Stop automatic debits
			}
			aTime, err := s.eventStart.GetAnswerTime(utils.META_DEFAULT, self.cgrCfg.DefaultTimezone)
			if err != nil || aTime.IsZero() {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not retrieve answer time for session: %s, runId: %s, aTime: %+v, error: %s",
					sessionId, s.runId, aTime, err.Error()))
			}
			if err := s.close(aTime.Add(usage)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not close session: %s, runId: %s, error: %s", sessionId, s.runId, err.Error()))
			}
			if err := s.saveOperations(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not save session: %s, runId: %s, error: %s", sessionId, s.runId, err.Error()))
			}
		}
		return nil, nil
	}, time.Duration(2)*time.Second, sessionId)
	return err
}

// Methods to apply on sessions, mostly exported through RPC/Bi-RPC
//Calculates maximum usage allowed for gevent
func (self *SMGeneric) GetMaxUsage(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	gev[utils.EVENT_NAME] = utils.CGR_AUTHORIZATION
	storedCdr := gev.AsStoredCdr(config.CgrConfig(), self.timezone)
	var maxDur float64
	if err := self.rater.GetDerivedMaxSessionTime(storedCdr, &maxDur); err != nil {
		return time.Duration(0), err
	}
	return time.Duration(maxDur), nil
}

func (self *SMGeneric) GetLcrSuppliers(gev SMGenericEvent, clnt *rpc2.Client) ([]string, error) {
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

// Execute debits for usage/maxUsage
func (self *SMGeneric) SessionUpdate(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	evLastUsed, err := gev.GetLastUsed(utils.META_DEFAULT)
	if err != nil && err != utils.ErrNotFound {
		return nilDuration, err
	}
	evMaxUsage, err := gev.GetMaxUsage(utils.META_DEFAULT, self.cgrCfg.MaxCallDuration)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrMandatoryIeMissing
		}
		return nilDuration, err
	}
	evUuid := gev.GetUUID()
	for _, s := range self.getSession(evUuid) {
		if maxDur, err := s.debit(evMaxUsage, evLastUsed); err != nil {
			return nilDuration, err
		} else if maxDur < evMaxUsage {
			evMaxUsage = maxDur
		}
	}
	return evMaxUsage, nil
}

// Called on session start
func (self *SMGeneric) SessionStart(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	if err := self.sessionStart(gev, getClientConnId(clnt)); err != nil {
		return nilDuration, err
	}
	return self.SessionUpdate(gev, clnt)
}

// Called on session end, should stop debit loop
func (self *SMGeneric) SessionEnd(gev SMGenericEvent, clnt *rpc2.Client) error {
	usage, err := gev.GetUsage(utils.META_DEFAULT)
	if err != nil {
		if err != utils.ErrNotFound {
			return err

		}
		lastUsed, err := gev.GetLastUsed(utils.META_DEFAULT)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrMandatoryIeMissing
			}
			return err
		}
		var s *SMGSession
		for _, s = range self.getSession(gev.GetUUID()) {
			break
		}
		if s == nil {
			return nil
		}
		usage = s.TotalUsage() + lastUsed
	}
	if err := self.sessionEnd(gev.GetUUID(), usage); err != nil {
		return err
	}
	return nil
}

// Processes one time events (eg: SMS)
func (self *SMGeneric) ChargeEvent(gev SMGenericEvent, clnt *rpc2.Client) (maxDur time.Duration, err error) {
	var sessionRuns []*engine.SessionRun
	if err := self.rater.GetSessionRuns(gev.AsStoredCdr(self.cgrCfg, self.timezone), &sessionRuns); err != nil {
		return nilDuration, err
	} else if len(sessionRuns) == 0 {
		return nilDuration, nil
	}
	var maxDurInit bool // Avoid differences between default 0 and received 0
	for _, sR := range sessionRuns {
		cc := new(engine.CallCost)
		if err = self.rater.MaxDebit(sR.CallDescriptor, cc); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not Debit CD: %+v, RunID: %s, error: %s", sR.CallDescriptor, sR.DerivedCharger.RunID, err.Error()))
			break
		}
		sR.CallCosts = append(sR.CallCosts, cc) // Save it so we can revert on issues
		if ccDur := cc.GetDuration(); ccDur == 0 {
			err = utils.ErrInsufficientCredit
			break
		} else if !maxDurInit || ccDur < maxDur {
			maxDur = ccDur
		}
	}
	if err != nil { // Refund the ones already taken since we have error on one of the debits
		for _, sR := range sessionRuns {
			if len(sR.CallCosts) == 0 {
				continue
			}
			cc := sR.CallCosts[0]
			if len(sR.CallCosts) > 1 {
				for _, ccSR := range sR.CallCosts {
					cc.Merge(ccSR)
				}
			}
			// collect increments
			var refundIncrements engine.Increments
			cc.Timespans.Decompress()
			for _, ts := range cc.Timespans {
				refundIncrements = append(refundIncrements, ts.Increments...)
			}
			// refund cc
			if len(refundIncrements) > 0 {
				cd := &engine.CallDescriptor{
					Direction:   cc.Direction,
					Tenant:      cc.Tenant,
					Category:    cc.Category,
					Subject:     cc.Subject,
					Account:     cc.Account,
					Destination: cc.Destination,
					TOR:         cc.TOR,
					Increments:  refundIncrements,
				}
				cd.Increments.Compress()
				utils.Logger.Info(fmt.Sprintf("Refunding session run callcost: %s", utils.ToJSON(cd)))
				var response float64
				err := self.rater.RefundIncrements(cd, &response)
				if err != nil {
					return nilDuration, err
				}
			}
		}
		return nilDuration, err
	}
	var withErrors bool
	for _, sR := range sessionRuns {
		if len(sR.CallCosts) == 0 {
			continue
		}
		cc := sR.CallCosts[0]
		if len(sR.CallCosts) > 1 {
			for _, ccSR := range sR.CallCosts[1:] {
				cc.Merge(ccSR)
			}
		}
		cc.Round()
		roundIncrements := cc.GetRoundIncrements()
		if len(roundIncrements) != 0 {
			cd := cc.CreateCallDescriptor()
			cd.Increments = roundIncrements
			var response float64
			if err := self.rater.RefundRounding(cd, &response); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM> ERROR failed to refund rounding: %v", err))
			}
		}

		var reply string
		if err := self.cdrsrv.LogCallCost(&engine.CallCostLog{
			CgrId:          gev.GetCgrId(self.timezone),
			Source:         utils.SESSION_MANAGER_SOURCE,
			RunId:          sR.DerivedCharger.RunID,
			CallCost:       cc,
			CheckDuplicate: true,
		}, &reply); err != nil && err != utils.ErrExists {
			withErrors = true
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not save CC: %+v, RunID: %s error: %s", cc, sR.DerivedCharger.RunID, err.Error()))
		}
	}
	if withErrors {
		return nilDuration, ErrPartiallyExecuted
	}
	return maxDur, nil
}

func (self *SMGeneric) ProcessCdr(gev SMGenericEvent) error {
	cdr := gev.AsStoredCdr(self.cgrCfg, self.timezone)
	if cdr.Usage == 0 {
		var s *SMGSession
		for _, s = range self.getSession(gev.GetUUID()) {
			break
		}
		if s != nil {
			cdr.Usage = s.TotalUsage()
		}
	}
	var reply string
	if err := self.cdrsrv.ProcessCdr(cdr, &reply); err != nil {
		return err
	}
	return nil
}

func (self *SMGeneric) Connect() error {
	return nil
}

// Used by APIer to retrieve sessions
func (self *SMGeneric) Sessions() map[string][]*SMGSession {
	return self.getSessions()
}

func (self *SMGeneric) Timezone() string {
	return self.timezone
}

// System shutdown
func (self *SMGeneric) Shutdown() error {
	for ssId := range self.getSessions() { // Force sessions shutdown
		self.sessionEnd(ssId, time.Duration(self.cgrCfg.MaxCallDuration))
	}
	return nil
}
