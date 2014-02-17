/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"strings"
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
	connector   engine.Connector
	debitPeriod time.Duration
	loggerDB    engine.LogStorage
}

func NewFSSessionManager(storage engine.LogStorage, connector engine.Connector, debitPeriod time.Duration) *FSSessionManager {
	return &FSSessionManager{loggerDB: storage, connector: connector, debitPeriod: debitPeriod}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (sm *FSSessionManager) Connect(cgrCfg *config.CGRConfig) (err error) {
	cfg = cgrCfg // make config global
	eventFilters := map[string]string{"Call-Direction": "inbound"}
	if fsock.FS, err = fsock.NewFSock(cfg.FreeswitchServer, cfg.FreeswitchPass, cfg.FreeswitchReconnects, sm.createHandlers(), eventFilters, engine.Logger.(*syslog.Writer)); err != nil {
		return err
	} else if !fsock.FS.Connected() {
		return errors.New("Cannot connect to FreeSWITCH")
	}
	fsock.FS.ReadEvents()
	return errors.New("stopped reading events")
}

func (sm *FSSessionManager) createHandlers() (handlers map[string][]func(string)) {
	hb := func(body string) {
		ev := new(FSEvent).New(body)
		sm.OnHeartBeat(ev)
	}
	cp := func(body string) {
		ev := new(FSEvent).New(body)
		sm.OnChannelPark(ev)
	}
	ca := func(body string) {
		ev := new(FSEvent).New(body)
		sm.OnChannelAnswer(ev)
	}
	ch := func(body string) {
		ev := new(FSEvent).New(body)
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
		if s.uuid == uuid {
			return s
		}
	}
	return nil
}

// Disconnects a session by sending hangup command to freeswitch
func (sm *FSSessionManager) DisconnectSession(s *Session, notify string) {
	// engine.Logger.Debug(fmt.Sprintf("Session: %+v", s.uuid))
	_, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", s.uuid, notify))
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("could not send disconect api notification to freeswitch: %v", err))
	}
	err = fsock.FS.SendMsgCmd(s.uuid, map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}) // without + sign
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("could not send disconect msg to freeswitch: %v", err))
	}
	return
}

// Remove session from sessin list
func (sm *FSSessionManager) RemoveSession(s *Session) {
	for i, ss := range sm.sessions {
		if ss == s {
			sm.sessions = append(sm.sessions[:i], sm.sessions[i+1:]...)
			return
		}
	}
}

// Sets the call timeout valid of starting of the call
func (sm *FSSessionManager) setMaxCallDuration(uuid string, maxDur time.Duration) error {
	_, err := fsock.FS.SendApiCmd(fmt.Sprintf("sched_hangup +%d %s\n\n", int(maxDur.Seconds()), uuid))
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
		engine.Logger.Err("could not send unpark api notification to freeswitch")
	}
	_, err = fsock.FS.SendApiCmd(fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, call_dest_nb))
	if err != nil {
		engine.Logger.Err("could not send unpark api call to freeswitch")
	}
}

func (sm *FSSessionManager) OnHeartBeat(ev Event) {
	engine.Logger.Info("freeswitch ♥")
}

func (sm *FSSessionManager) OnChannelPark(ev Event) {
	//engine.Logger.Info("freeswitch park")
	startTime, err := ev.GetStartTime(PARK_TIME)
	if err != nil {
		engine.Logger.Err("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}
	// if there is no account configured leave the call alone
	if !utils.IsSliceMember([]string{utils.PREPAID, utils.PSEUDOPREPAID}, strings.TrimSpace(ev.GetReqType(""))) {
		return // we unpark only prepaid and pseudoprepaid calls
	}
	if ev.MissingParameter() {
		sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(""), MISSING_PARAMETER)
		engine.Logger.Err(fmt.Sprintf("Missing parameter for %s", ev.GetUUID()))
		return
	}
	cd := engine.CallDescriptor{
		Direction:   ev.GetDirection(""),
		Tenant:      ev.GetTenant(""),
		TOR:         ev.GetTOR(""),
		Subject:     ev.GetSubject(""),
		Account:     ev.GetAccount(""),
		Destination: ev.GetDestination(""),
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(cfg.SMMaxCallDuration),
	}
	var remainingDurationFloat float64
	err = sm.connector.GetMaxSessionTime(cd, &remainingDurationFloat)
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("Could not get max session time for %s: %v", ev.GetUUID(), err))
		sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(""), SYSTEM_ERROR)
		return
	}
	remainingDuration := time.Duration(remainingDurationFloat)
	//engine.Logger.Info(fmt.Sprintf("Remaining duration: %v", remainingDuration))
	if remainingDuration == 0 {
		//engine.Logger.Info(fmt.Sprintf("Not enough credit for trasferring the call %s for %s.", ev.GetUUID(), cd.GetKey(cd.Subject)))
		sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(""), INSUFFICIENT_FUNDS)
		return
	}
	sm.setMaxCallDuration(ev.GetUUID(), remainingDuration)
	sm.unparkCall(ev.GetUUID(), ev.GetCallDestNr(""), AUTH_OK)
}

func (sm *FSSessionManager) OnChannelAnswer(ev Event) {
	//engine.Logger.Info("<SessionManager> FreeSWITCH answer.")
	// Make sure cgr_type is enforced even if not set by FreeSWITCH
	if _, err := fsock.FS.SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_reqtype %s\n\n", ev.GetUUID(), ev.GetReqType(""))); err != nil {
		engine.Logger.Err(fmt.Sprintf("Error on attempting to overwrite cgr_type in chan variables: %v", err))
	}
	s := NewSession(ev, sm)
	if s != nil {
		sm.sessions = append(sm.sessions, s)
	}
}

func (sm *FSSessionManager) OnChannelHangupComplete(ev Event) {
	//engine.Logger.Info("<SessionManager> FreeSWITCH hangup.")
	s := sm.GetSession(ev.GetUUID())
	if s == nil { // Not handled by us
		return
	}
	defer s.Close(ev) // Stop loop and save the costs deducted so far to database
	if ev.GetReqType("") == utils.POSTPAID {
		startTime, err := ev.GetStartTime(START_TIME)
		if err != nil {
			engine.Logger.Crit("Error parsing postpaid call start time from event")
			return
		}
		endTime, err := ev.GetEndTime()
		if err != nil {
			engine.Logger.Crit("Error parsing postpaid call start time from event")
			return
		}
		cd := engine.CallDescriptor{
			Direction:    ev.GetDirection(""),
			Tenant:       ev.GetTenant(""),
			TOR:          ev.GetTOR(""),
			Subject:      ev.GetSubject(""),
			Account:      ev.GetAccount(""),
			LoopIndex:    0,
			CallDuration: endTime.Sub(startTime),
			Destination:  ev.GetDestination(""),
			TimeStart:    startTime,
			TimeEnd:      endTime,
		}
		cc := &engine.CallCost{}
		err = sm.connector.Debit(cd, cc)
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("Error making the general debit for postpaid call: %v", ev.GetUUID()))
			return
		}
		s.CallCosts = append(s.CallCosts, cc)
		return
	}

	if s == nil || len(s.CallCosts) == 0 {
		return // why would we have 0 callcosts
	}
	lastCC := s.CallCosts[len(s.CallCosts)-1]
	// put credit back
	var hangupTime time.Time
	var err error
	if hangupTime, err = ev.GetEndTime(); err != nil {
		engine.Logger.Err("Error parsing answer event hangup time, using time.Now!")
		hangupTime = time.Now()
	}
	end := lastCC.Timespans[len(lastCC.Timespans)-1].TimeEnd
	refundDuration := end.Sub(hangupTime)
	//initialRefundDuration := refundDuration
	var refundIncrements engine.Increments
	for i := len(lastCC.Timespans) - 1; i >= 0; i-- {
		ts := lastCC.Timespans[i]
		tsDuration := ts.GetDuration()
		if refundDuration <= tsDuration {
			lastRefundedIncrementIndex := 0
			for j := len(ts.Increments) - 1; j >= 0; j-- {
				increment := ts.Increments[j]
				if increment.Duration <= refundDuration {
					refundIncrements = append(refundIncrements, increment)
					refundDuration -= increment.Duration
					lastRefundedIncrementIndex = j
				}
			}
			ts.SplitByIncrement(lastRefundedIncrementIndex)
			break // do not go to other timespans
		} else {
			refundIncrements = append(refundIncrements, ts.Increments...)
			// remove the timespan entirely
			lastCC.Timespans[i] = nil
			lastCC.Timespans = lastCC.Timespans[:i]
			// continue to the next timespan with what is left to refund
			refundDuration -= tsDuration
		}
	}
	// show only what was actualy refunded (stopped in timespan)
	// engine.Logger.Info(fmt.Sprintf("Refund duration: %v", initialRefundDuration-refundDuration))
	if len(refundIncrements) > 0 {
		cd := &engine.CallDescriptor{
			Direction:   lastCC.Direction,
			Tenant:      lastCC.Tenant,
			TOR:         lastCC.TOR,
			Subject:     lastCC.Subject,
			Account:     lastCC.Account,
			Destination: lastCC.Destination,
			Increments:  refundIncrements,
			// FallbackSubject: lastCC.FallbackSubject, // TODO: check how to best add it
		}
		var response float64
		err := sm.connector.RefundIncrements(*cd, &response)
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("Debit cents failed: %v", err))
		}
	}
	cost := refundIncrements.GetTotalCost()
	lastCC.Cost -= cost
	// engine.Logger.Info(fmt.Sprintf("Rambursed %v cents", cost))

}

func (sm *FSSessionManager) LoopAction(s *Session, cd *engine.CallDescriptor) (cc *engine.CallCost) {
	cc = &engine.CallCost{}
	err := sm.connector.MaxDebit(*cd, cc)
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("Could not complete debit opperation: %v", err))
		// disconnect session
		s.sessionManager.DisconnectSession(s, SYSTEM_ERROR)
	}
	// engine.Logger.Debug(fmt.Sprintf("Result of MaxDebit call: %v", cc))
	if cc.GetDuration() == 0 || err != nil {
		// engine.Logger.Info(fmt.Sprintf("No credit left: Disconnect %v", s))
		sm.DisconnectSession(s, INSUFFICIENT_FUNDS)
		return
	}
	s.CallCosts = append(s.CallCosts, cc)
	return
}

func (sm *FSSessionManager) GetDebitPeriod() time.Duration {
	return sm.debitPeriod
}

func (sm *FSSessionManager) GetDbLogger() engine.LogStorage {
	return sm.loggerDB
}

func (sm *FSSessionManager) Shutdown() (err error) {
	if fsock.FS == nil || !fsock.FS.Connected() {
		return errors.New("Cannot shutdown sessions, fsock not connected")
	}
	engine.Logger.Info("Shutting down all sessions...")
	cmdKillPrepaid := "hupall MANAGER_REQUEST cgr_reqtype prepaid"
	cmdKillPostpaid := "hupall MANAGER_REQUEST cgr_reqtype postpaid"
	for _, cmd := range []string{cmdKillPrepaid, cmdKillPostpaid} {
		if _, err = fsock.FS.SendApiCmd(cmd); err != nil {
			engine.Logger.Err(fmt.Sprintf("Error on calls shutdown: %s", err))
			return
		}
	}
	for guard := 0; len(sm.sessions) > 0 && guard < 20; guard++ {
		time.Sleep(100 * time.Millisecond) // wait for the hungup event to be fired
		engine.Logger.Info(fmt.Sprintf("<SessionManager> Shutdown waiting on sessions: %v", sm.sessions))
	}
	return
}
