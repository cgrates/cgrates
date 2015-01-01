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
	"bufio"
	"errors"
	"fmt"
	"log/syslog"
	"net"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

var cfg *config.CGRConfig // Share the configuration with the rest of the package

// The freeswitch session manager type holding a buffer for the network connection
// and the active sessions
type FSSessionManager struct {
	conn        net.Conn
	buf         *bufio.Reader
	sessions    []*Session
	rater       engine.Connector
	cdrs        engine.Connector
	debitPeriod time.Duration
	loggerDB    engine.LogStorage
}

func NewFSSessionManager(cgrCfg *config.CGRConfig, storage engine.LogStorage, rater, cdrs engine.Connector, debitPeriod time.Duration) *FSSessionManager {
	cfg = cgrCfg // make config global
	return &FSSessionManager{loggerDB: storage, rater: rater, cdrs: cdrs, debitPeriod: debitPeriod}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (sm *FSSessionManager) Connect() (err error) {
	eventFilters := map[string]string{"Call-Direction": "inbound"}
	if fsock.FS, err = fsock.NewFSock(cfg.FreeswitchServer, cfg.FreeswitchPass, cfg.FreeswitchReconnects, sm.createHandlers(), eventFilters, engine.Logger.(*syslog.Writer)); err != nil {
		return err
	} else if !fsock.FS.Connected() {
		return errors.New("Cannot connect to FreeSWITCH")
	}
	if err := fsock.FS.ReadEvents(); err != nil {
		return err
	}
	return errors.New("<SessionManager> - Stopped reading events")
}

func (sm *FSSessionManager) createHandlers() (handlers map[string][]func(string)) {
	hb := func(body string) {
		ev := new(FSEvent).AsEvent(body)
		sm.OnHeartBeat(ev)
	}
	cp := func(body string) {
		ev := new(FSEvent).AsEvent(body)
		sm.OnChannelPark(ev)
	}
	ca := func(body string) {
		ev := new(FSEvent).AsEvent(body)
		sm.OnChannelAnswer(ev)
	}
	ch := func(body string) {
		ev := new(FSEvent).AsEvent(body)
		sm.OnChannelHangupComplete(ev)
	}
	return map[string][]func(string){
		"HEARTBEAT":               []func(string){hb},
		"CHANNEL_PARK":            []func(string){cp},
		"CHANNEL_ANSWER":          []func(string){ca},
		"CHANNEL_HANGUP_COMPLETE": []func(string){ch},
	}
}

// Searches and return the session with the specifed uuid
func (sm *FSSessionManager) GetSession(uuid string) *Session {
	for _, s := range sm.sessions {
		if s.eventStart.GetUUID() == uuid {
			return s
		}
	}
	return nil
}

// Disconnects a session by sending hangup command to freeswitch
func (sm *FSSessionManager) DisconnectSession(ev utils.Event, notify string) {
	if _, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", ev.GetUUID(), notify)); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> Could not send disconect api notification to freeswitch: %s", err.Error()))
	}
	if notify == INSUFFICIENT_FUNDS {
		if len(cfg.FSEmptyBalanceContext) != 0 {
			if _, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_transfer %s %s %s\n\n", ev.GetUUID(), ev.GetCallDestNr(utils.META_DEFAULT), cfg.FSEmptyBalanceContext)); err != nil {
				engine.Logger.Err("<SessionManager> Could not transfer the call to empty balance context")
			}
			return
		} else if len(cfg.FSEmptyBalanceAnnFile) != 0 {
			if _, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_broadcast %s playback!manager_request::%s aleg\n\n", ev.GetUUID(), cfg.FSEmptyBalanceAnnFile)); err != nil {
				engine.Logger.Err(fmt.Sprintf("<SessionManager> Could not send uuid_broadcast to freeswitch: %s", err.Error()))
			}
			return
		}
	}
	if err := fsock.FS.SendMsgCmd(ev.GetUUID(), map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> Could not send disconect msg to freeswitch: %v", err))
	}
	return
}

// Remove session from session list, removes all related in case of multiple runs
func (sm *FSSessionManager) RemoveSession(uuid string) {
	for i, ss := range sm.sessions {
		if ss.eventStart.GetUUID() == uuid {
			sm.sessions = append(sm.sessions[:i], sm.sessions[i+1:]...)
			return
		}
	}
}

// Sets the call timeout valid of starting of the call
func (sm *FSSessionManager) setMaxCallDuration(uuid string, maxDur time.Duration) error {
	// _, err := fsock.FS.SendApiCmd(fmt.Sprintf("sched_hangup +%d %s\n\n", int(maxDur.Seconds()), uuid))
	_, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n", uuid, int(maxDur.Seconds())))
	if err != nil {
		engine.Logger.Err("could not send sched_hangup command to freeswitch")
		return err
	}
	return nil
}

// Sends the transfer command to unpark the call to freeswitch
func (sm *FSSessionManager) unparkCall(uuid, call_dest_nb, notify string) {
	_, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	if err != nil {
		engine.Logger.Err("<SessionManager> Could not send unpark api notification to freeswitch")
	}
	if _, err = fsock.FS.SendApiCmd(fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, call_dest_nb)); err != nil {
		engine.Logger.Err("<SessionManager> Could not send unpark api call to freeswitch")
	}
}

func (sm *FSSessionManager) OnHeartBeat(ev utils.Event) {
	engine.Logger.Info("freeswitch ♥")
}

func (sm *FSSessionManager) OnChannelPark(ev utils.Event) {
	var maxCallDuration time.Duration // This will be the maximum duration this channel will be allowed to last
	var durInitialized bool
	attrsDC := utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT), Category: ev.GetCategory(utils.META_DEFAULT), Direction: ev.GetDirection(utils.META_DEFAULT),
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	var dcs utils.DerivedChargers
	if err := sm.rater.GetDerivedChargers(attrsDC, &dcs); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> OnPark: could not get derived charging for event %s: %s", ev.GetUUID(), err.Error()))
		sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR) // We unpark on original destination
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
		startTime, err := ev.GetAnswerTime(PARK_TIME)
		if err != nil {
			engine.Logger.Err("Error parsing answer event start time, using time.Now!")
			startTime = time.Now()
		}
		if ev.MissingParameter() {
			sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(dc.DestinationField), MISSING_PARAMETER)
			engine.Logger.Err(fmt.Sprintf("Missing parameter for %s", ev.GetUUID()))
			return
		}
		cd := engine.CallDescriptor{
			Direction:   ev.GetDirection(dc.DirectionField),
			Tenant:      ev.GetTenant(dc.TenantField),
			Category:    ev.GetCategory(dc.CategoryField),
			Subject:     ev.GetSubject(dc.SubjectField),
			Account:     ev.GetAccount(dc.AccountField),
			Destination: ev.GetDestination(dc.DestinationField),
			TimeStart:   startTime,
			TimeEnd:     startTime.Add(cfg.SMMaxCallDuration),
		}
		var remainingDurationFloat float64
		err = sm.rater.GetMaxSessionTime(cd, &remainingDurationFloat)
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("Could not get max session time for %s: %v", ev.GetUUID(), err))
			sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(""), SYSTEM_ERROR) // We unpark on original destination
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
	if maxCallDuration <= cfg.SMMinCallDuration {
		//engine.Logger.Info(fmt.Sprintf("Not enough credit for trasferring the call %s for %s.", ev.GetUUID(), cd.GetKey(cd.Subject)))
		sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(utils.META_DEFAULT), INSUFFICIENT_FUNDS)
		return
	}
	sm.setMaxCallDuration(ev.GetUUID(), maxCallDuration)
	sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(utils.META_DEFAULT), AUTH_OK)
}

func (sm *FSSessionManager) OnChannelAnswer(ev utils.Event) {
	if ev.MissingParameter() {
		sm.DisconnectSession(ev, MISSING_PARAMETER)
	}
	s := NewSession(ev, sm)
	if s != nil {
		sm.sessions = append(sm.sessions, s)
	}
}

func (sm *FSSessionManager) OnChannelHangupComplete(ev utils.Event) {
	go sm.processCdr(ev.AsStoredCdr())
	s := sm.GetSession(ev.GetUUID())
	if s == nil { // Not handled by us
		return
	}
	sm.RemoveSession(s.eventStart.GetUUID()) // Unreference it early so we avoid concurrency
	if err := s.Close(ev); err != nil {      // Stop loop, refund advanced charges and save the costs deducted so far to database
		engine.Logger.Err(err.Error())
	}
}

func (sm *FSSessionManager) processCdr(storedCdr *utils.StoredCdr) error {
	if sm.cdrs != nil {
		var reply string
		if err := sm.cdrs.ProcessCdr(storedCdr, &reply); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", storedCdr.CgrId, storedCdr.AccId, err.Error()))
		}

	}
	return nil
}

func (sm *FSSessionManager) GetDebitPeriod() time.Duration {
	return sm.debitPeriod
}

func (sm *FSSessionManager) MaxDebit(cd *engine.CallDescriptor, cc *engine.CallCost) error {
	return sm.rater.MaxDebit(*cd, cc)
}

func (sm *FSSessionManager) GetDbLogger() engine.LogStorage {
	return sm.loggerDB
}

func (sm *FSSessionManager) Rater() engine.Connector {
	return sm.rater
}

func (sm *FSSessionManager) Shutdown() (err error) {
	if fsock.FS == nil || !fsock.FS.Connected() {
		return errors.New("Cannot shutdown sessions, fsock not connected")
	}
	engine.Logger.Info("Shutting down all sessions...")
	if _, err = fsock.FS.SendApiCmd("hupall MANAGER_REQUEST cgr_reqtype prepaid"); err != nil {
		engine.Logger.Err(fmt.Sprintf("Error on calls shutdown: %s", err))
	}
	for guard := 0; len(sm.sessions) > 0 && guard < 20; guard++ {
		time.Sleep(100 * time.Millisecond) // wait for the hungup event to be fired
		engine.Logger.Info(fmt.Sprintf("<SessionManager> Shutdown waiting on sessions: %v", sm.sessions))
	}
	return
}
