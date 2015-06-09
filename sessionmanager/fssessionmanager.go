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
	"log/syslog"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

// The freeswitch session manager type holding a buffer for the network connection
// and the active sessions
type FSSessionManager struct {
	cfg      *config.SmFsConfig
	conns    map[string]*fsock.FSock // Keep the list here for connection management purposes
	sessions []*Session
	rater    engine.Connector
	cdrs     engine.Connector
}

func NewFSSessionManager(smFsConfig *config.SmFsConfig, rater, cdrs engine.Connector) *FSSessionManager {
	return &FSSessionManager{
		cfg:   smFsConfig,
		conns: make(map[string]*fsock.FSock),
		rater: rater,
		cdrs:  cdrs,
	}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (sm *FSSessionManager) Connect() error {
	eventFilters := map[string]string{"Call-Direction": "inbound"}
	errChan := make(chan error)
	for _, connCfg := range sm.cfg.Connections {
		connId := utils.GenUUID()
		fSock, err := fsock.NewFSock(connCfg.Server, connCfg.Password, connCfg.Reconnects, sm.createHandlers(), eventFilters, engine.Logger.(*syslog.Writer), connId)
		if err != nil {
			return err
		} else if !fSock.Connected() {
			return errors.New("Could not connect to FreeSWITCH")
		} else {
			sm.conns[connId] = fSock
		}
		go func() { // Start reading in own goroutine, return on error
			if err := sm.conns[connId].ReadEvents(); err != nil {
				errChan <- err
			}
		}()
	}
	err := <-errChan // Will keep the Connect locked until the first error in one of the connections
	return err
}

func (sm *FSSessionManager) createHandlers() (handlers map[string][]func(string, string)) {
	cp := func(body, connId string) {
		ev := new(FSEvent).AsEvent(body)
		sm.onChannelPark(ev, connId)
	}
	ca := func(body, connId string) {
		ev := new(FSEvent).AsEvent(body)
		sm.onChannelAnswer(ev, connId)
	}
	ch := func(body, connId string) {
		ev := new(FSEvent).AsEvent(body)
		sm.onChannelHangupComplete(ev)
	}
	return map[string][]func(string, string){
		"CHANNEL_PARK":            []func(string, string){cp},
		"CHANNEL_ANSWER":          []func(string, string){ca},
		"CHANNEL_HANGUP_COMPLETE": []func(string, string){ch},
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
func (sm *FSSessionManager) DisconnectSession(ev engine.Event, connId, notify string) error {
	if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", ev.GetUUID(), notify)); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send disconect api notification to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
		return err
	}
	if notify == INSUFFICIENT_FUNDS {
		if len(sm.cfg.EmptyBalanceContext) != 0 {
			if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s %s\n\n", ev.GetUUID(), ev.GetCallDestNr(utils.META_DEFAULT), sm.cfg.EmptyBalanceContext)); err != nil {
				engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not transfer the call to empty balance context, error: <%s>, connId: %s", err.Error(), connId))
				return err
			}
			return nil
		} else if len(sm.cfg.EmptyBalanceAnnFile) != 0 {
			if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_broadcast %s playback!manager_request::%s aleg\n\n", ev.GetUUID(), sm.cfg.EmptyBalanceAnnFile)); err != nil {
				engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send uuid_broadcast to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
				return err
			}
			return nil
		}
	}
	if err := sm.conns[connId].SendMsgCmd(ev.GetUUID(), map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send disconect msg to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
		return err
	}
	return nil
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
func (sm *FSSessionManager) setMaxCallDuration(uuid, connId string, maxDur time.Duration) error {
	// _, err := fsock.FS.SendApiCmd(fmt.Sprintf("sched_hangup +%d %s\n\n", int(maxDur.Seconds()), uuid))
	_, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n", uuid, int(maxDur.Seconds())))
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send sched_hangup command to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
		return err
	}
	return nil
}

// Queries LCR and sets the cgr_lcr channel variable
func (sm *FSSessionManager) setCgrLcr(ev engine.Event, connId string) error {
	var lcrCost engine.LCRCost
	startTime, err := ev.GetSetupTime(utils.META_DEFAULT)
	if err != nil {
		return err
	}
	cd := &engine.CallDescriptor{
		Direction:   ev.GetDirection(utils.META_DEFAULT),
		Tenant:      ev.GetTenant(utils.META_DEFAULT),
		Category:    ev.GetCategory(utils.META_DEFAULT),
		Subject:     ev.GetSubject(utils.META_DEFAULT),
		Account:     ev.GetAccount(utils.META_DEFAULT),
		Destination: ev.GetDestination(utils.META_DEFAULT),
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(config.CgrConfig().MaxCallDuration),
	}
	if err := sm.rater.GetLCR(cd, &lcrCost); err != nil {
		return err
	}
	supps := []string{}
	for _, supplCost := range lcrCost.SupplierCosts {
		if dtcs, err := utils.NewDTCSFromRPKey(supplCost.Supplier); err != nil {
			return err
		} else if len(dtcs.Subject) != 0 {
			supps = append(supps, dtcs.Subject)
		}
	}
	fsArray := SliceAsFsArray(supps)
	if _, err = sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", ev.GetUUID(), fsArray)); err != nil {
		return err
	}
	return nil
}

// Sends the transfer command to unpark the call to freeswitch
func (sm *FSSessionManager) unparkCall(uuid, connId, call_dest_nb, notify string) {
	_, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send unpark api notification to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
	}
	if _, err = sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, call_dest_nb)); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send unpark api call to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
	}
}

func (sm *FSSessionManager) onChannelPark(ev engine.Event, connId string) {
	if ev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	var maxCallDuration float64 // This will be the maximum duration this channel will be allowed to last
	if err := sm.rater.GetDerivedMaxSessionTime(*ev.AsStoredCdr(), &maxCallDuration); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not get max session time for %s, error: %s", ev.GetUUID(), err.Error()))
	}
	maxCallDur := time.Duration(maxCallDuration)
	if maxCallDur <= sm.cfg.MinCallDuration {
		//engine.Logger.Info(fmt.Sprintf("Not enough credit for trasferring the call %s for %s.", ev.GetUUID(), cd.GetKey(cd.Subject)))
		sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), INSUFFICIENT_FUNDS)
		return
	}
	sm.setMaxCallDuration(ev.GetUUID(), connId, maxCallDur)
	if sm.cfg.ComputeLcr {
		if err := sm.setCgrLcr(ev, connId); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not set LCR for %s, error: %s", ev.GetUUID(), err.Error()))
			sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR)
			return
		}
	}
	sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), AUTH_OK)
}

func (sm *FSSessionManager) onChannelAnswer(ev engine.Event, connId string) {
	if ev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if ev.MissingParameter() {
		sm.DisconnectSession(ev, connId, MISSING_PARAMETER)
	}
	s := NewSession(ev, connId, sm)
	if s != nil {
		sm.sessions = append(sm.sessions, s)
	}
}

func (sm *FSSessionManager) onChannelHangupComplete(ev engine.Event) {
	if ev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if sm.cdrs != nil {
		go sm.ProcessCdr(ev.AsStoredCdr())
	}
	var s *Session
	for i := 0; i < 2; i++ { // Protect us against concurrency, wait a couple of seconds for the answer to be populated before we process hangup
		s = sm.GetSession(ev.GetUUID())
		if s != nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if s == nil { // Not handled by us
		return
	}
	sm.RemoveSession(s.eventStart.GetUUID()) // Unreference it early so we avoid concurrency
	if err := s.Close(ev); err != nil {      // Stop loop, refund advanced charges and save the costs deducted so far to database
		engine.Logger.Err(err.Error())
	}
}

func (sm *FSSessionManager) ProcessCdr(storedCdr *engine.StoredCdr) error {
	var reply string
	if err := sm.cdrs.ProcessCdr(storedCdr, &reply); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", storedCdr.CgrId, storedCdr.AccId, err.Error()))
	}
	return nil
}

func (sm *FSSessionManager) DebitInterval() time.Duration {
	return sm.cfg.DebitInterval
}
func (sm *FSSessionManager) CdrSrv() engine.Connector {
	return sm.cdrs
}

func (sm *FSSessionManager) Rater() engine.Connector {
	return sm.rater
}

// Called when call goes under the minimum duratio threshold, so FreeSWITCH can play an announcement message
func (sm *FSSessionManager) WarnSessionMinDuration(sessionUuid, connId string) {
	if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_broadcast %s %s aleg\n\n", sessionUuid, sm.cfg.LowBalanceAnnFile)); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send uuid_broadcast to freeswitch, error: %s, connection id: %s", err.Error(), connId))
	}
}

func (sm *FSSessionManager) Shutdown() (err error) {
	for connId, fSock := range sm.conns {
		if !fSock.Connected() {
			engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Cannot shutdown sessions, fsock not connected for connection id: %s", connId))
			continue
		}
		engine.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> Shutting down all sessions on connection id: %s", connId))
		if _, err = fSock.SendApiCmd("hupall MANAGER_REQUEST cgr_reqtype prepaid"); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Error on calls shutdown: %s, connection id: %s", err.Error(), connId))
		}
	}
	for guard := 0; len(sm.sessions) > 0 && guard < 20; guard++ {
		time.Sleep(100 * time.Millisecond) // wait for the hungup event to be fired
		engine.Logger.Info(fmt.Sprintf("<SM-FreeSWITC> Shutdown waiting on sessions: %v", sm.sessions))
	}
	return nil
}
