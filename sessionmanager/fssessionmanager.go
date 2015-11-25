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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
	"github.com/cgrates/rpcclient"
)

func NewFSSessionManager(smFsConfig *config.SmFsConfig, rater, cdrs rpcclient.RpcClientConnection, timezone string) *FSSessionManager {
	return &FSSessionManager{
		cfg:         smFsConfig,
		conns:       make(map[string]*fsock.FSock),
		senderPools: make(map[string]*fsock.FSockPool),
		rater:       rater,
		cdrsrv:      cdrs,
		sessions:    NewSessions(),
		timezone:    timezone,
	}
}

// The freeswitch session manager type holding a buffer for the network connection
// and the active sessions
type FSSessionManager struct {
	cfg         *config.SmFsConfig
	conns       map[string]*fsock.FSock     // Keep the list here for connection management purposes
	senderPools map[string]*fsock.FSockPool // Keep sender pools here
	rater       rpcclient.RpcClientConnection
	cdrsrv      rpcclient.RpcClientConnection

	sessions *Sessions
	timezone string
}

func (sm *FSSessionManager) createHandlers() map[string][]func(string, string) {
	ca := func(body, connId string) {
		ev := new(FSEvent).AsEvent(body)
		sm.onChannelAnswer(ev, connId)
	}
	ch := func(body, connId string) {
		ev := new(FSEvent).AsEvent(body)
		sm.onChannelHangupComplete(ev)
	}
	handlers := map[string][]func(string, string){
		"CHANNEL_ANSWER":          []func(string, string){ca},
		"CHANNEL_HANGUP_COMPLETE": []func(string, string){ch},
	}
	if sm.cfg.SubscribePark {
		cp := func(body, connId string) {
			ev := new(FSEvent).AsEvent(body)
			sm.onChannelPark(ev, connId)
		}
		handlers["CHANNEL_PARK"] = []func(string, string){cp}
	}
	return handlers
}

// Sets the call timeout valid of starting of the call
func (sm *FSSessionManager) setMaxCallDuration(uuid, connId string, maxDur time.Duration) error {
	// _, err := fsock.FS.SendApiCmd(fmt.Sprintf("sched_hangup +%d %s\n\n", int(maxDur.Seconds()), uuid))
	_, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n", uuid, int(maxDur.Seconds())))
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send sched_hangup command to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
		return err
	}
	return nil
}

// Queries LCR and sets the cgr_lcr channel variable
func (sm *FSSessionManager) setCgrLcr(ev engine.Event, connId string) error {
	var lcrCost engine.LCRCost
	startTime, err := ev.GetSetupTime(utils.META_DEFAULT, sm.timezone)
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
	if err := sm.rater.Call("Responder.GetLCR", &engine.AttrGetLcr{CallDescriptor: cd}, &lcrCost); err != nil {
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

func (sm *FSSessionManager) onChannelPark(ev engine.Event, connId string) {
	fsev := ev.(FSEvent)
	if ev.GetReqType(utils.META_DEFAULT) == utils.META_NONE || fsev[IGNOREPARK] == "true" { // Do not process this request
		return
	}
	var maxCallDuration float64 // This will be the maximum duration this channel will be allowed to last
	if err := sm.rater.Call("Responder.GetDerivedMaxSessionTime", ev.AsStoredCdr(config.CgrConfig().DefaultTimezone), &maxCallDuration); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not get max session time for %s, error: %s", ev.GetUUID(), err.Error()))
	}
	if maxCallDuration != -1 { // For calls different than unlimited, set limits
		maxCallDur := time.Duration(maxCallDuration)
		if maxCallDur <= sm.cfg.MinCallDuration {
			//utils.Logger.Info(fmt.Sprintf("Not enough credit for trasferring the call %s for %s.", ev.GetUUID(), cd.GetKey(cd.Subject)))
			sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), INSUFFICIENT_FUNDS)
			return
		}
		sm.setMaxCallDuration(ev.GetUUID(), connId, maxCallDur)
	}
	// ComputeLcr
	if ev.ComputeLcr() {
		cd, err := fsev.AsCallDescriptor()
		if err != nil {
			utils.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> LCR_PREPROCESS_ERROR: %s", err.Error()))
			sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR)
			return
		}
		var lcr engine.LCRCost
		if err = sm.Rater().Call("Responder.GetLCR", &engine.AttrGetLcr{CallDescriptor: cd}, &lcr); err != nil {
			utils.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> LCR_API_ERROR: %s", err.Error()))
			sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR)
		}
		if lcr.HasErrors() {
			lcr.LogErrors()
			sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR)
			return
		}
		if supps, err := lcr.SuppliersSlice(); err != nil {
			utils.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> LCR_ERROR: %s", err.Error()))
			sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR)
			return
		} else {
			fsArray := SliceAsFsArray(supps)
			if _, err = sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n", ev.GetUUID(), utils.CGR_SUPPLIERS, fsArray)); err != nil {
				utils.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> LCR_ERROR: %s", err.Error()))
				sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), SYSTEM_ERROR)
			}
		}
	}
	sm.unparkCall(ev.GetUUID(), connId, ev.GetCallDestNr(utils.META_DEFAULT), AUTH_OK)
}

// Sends the transfer command to unpark the call to freeswitch
func (sm *FSSessionManager) unparkCall(uuid, connId, call_dest_nb, notify string) {
	_, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send unpark api notification to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
	}
	if _, err = sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, call_dest_nb)); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send unpark api call to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
	}
}

func (sm *FSSessionManager) onChannelAnswer(ev engine.Event, connId string) {
	if ev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if ev.MissingParameter(sm.timezone) {
		sm.DisconnectSession(ev, connId, MISSING_PARAMETER)
	}
	s := NewSession(ev, connId, sm)
	if s != nil {
		sm.sessions.indexSession(s)
	}
}

func (sm *FSSessionManager) onChannelHangupComplete(ev engine.Event) {
	if ev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	var s *Session
	for i := 0; i < 2; i++ { // Protect us against concurrency, wait a couple of seconds for the answer to be populated before we process hangup
		s = sm.sessions.getSession(ev.GetUUID())
		if s != nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if s != nil { // Handled by us, cleanup here
		if err := sm.sessions.removeSession(s, ev); err != nil {
			utils.Logger.Err(err.Error())
		}
	}
	if sm.cfg.CreateCdr {
		sm.ProcessCdr(ev.AsStoredCdr(config.CgrConfig().DefaultTimezone))
	}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (sm *FSSessionManager) Connect() error {
	eventFilters := map[string]string{"Call-Direction": "inbound"}
	errChan := make(chan error)
	for _, connCfg := range sm.cfg.Connections {
		connId := utils.GenUUID()
		fSock, err := fsock.NewFSock(connCfg.Server, connCfg.Password, connCfg.Reconnects, sm.createHandlers(), eventFilters, utils.Logger.(*syslog.Writer), connId)
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
		if fsSenderPool, err := fsock.NewFSockPool(5, connCfg.Server, connCfg.Password, 1,
			make(map[string][]func(string, string)), make(map[string]string), utils.Logger.(*syslog.Writer), connId); err != nil {
			return fmt.Errorf("Cannot connect FreeSWITCH senders pool, error: %s", err.Error())
		} else if fsSenderPool == nil {
			return errors.New("Cannot connect FreeSWITCH senders pool.")
		} else {
			sm.senderPools[connId] = fsSenderPool
		}
		if sm.cfg.ChannelSyncInterval != 0 { // Schedule running of the callsync
			go func() {
				for { // Schedule sync channels to run repetately
					time.Sleep(sm.cfg.ChannelSyncInterval)
					sm.SyncSessions()
				}

			}()
		}
	}
	err := <-errChan // Will keep the Connect locked until the first error in one of the connections
	return err
}

// Disconnects a session by sending hangup command to freeswitch
func (sm *FSSessionManager) DisconnectSession(ev engine.Event, connId, notify string) error {
	if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", ev.GetUUID(), notify)); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send disconect api notification to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
		return err
	}
	if notify == INSUFFICIENT_FUNDS {
		if len(sm.cfg.EmptyBalanceContext) != 0 {
			if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s %s\n\n", ev.GetUUID(), ev.GetCallDestNr(utils.META_DEFAULT), sm.cfg.EmptyBalanceContext)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not transfer the call to empty balance context, error: <%s>, connId: %s", err.Error(), connId))
				return err
			}
			return nil
		} else if len(sm.cfg.EmptyBalanceAnnFile) != 0 {
			if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_broadcast %s playback!manager_request::%s aleg\n\n", ev.GetUUID(), sm.cfg.EmptyBalanceAnnFile)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send uuid_broadcast to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
				return err
			}
			return nil
		}
	}
	if err := sm.conns[connId].SendMsgCmd(ev.GetUUID(), map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send disconect msg to freeswitch, error: <%s>, connId: %s", err.Error(), connId))
		return err
	}
	return nil
}

func (sm *FSSessionManager) ProcessCdr(storedCdr *engine.StoredCdr) error {
	var reply string
	if err := sm.cdrsrv.Call("Responder.ProcessCdr", storedCdr, &reply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", storedCdr.CgrId, storedCdr.AccId, err.Error()))
	}
	return nil
}

func (sm *FSSessionManager) DebitInterval() time.Duration {
	return sm.cfg.DebitInterval
}

func (sm *FSSessionManager) CdrSrv() rpcclient.RpcClientConnection {
	return sm.cdrsrv
}

func (sm *FSSessionManager) Rater() rpcclient.RpcClientConnection {
	return sm.rater
}

func (sm *FSSessionManager) Sessions() []*Session {
	return sm.sessions.getSessions()
}

func (sm *FSSessionManager) Timezone() string {
	return sm.timezone
}

// Called when call goes under the minimum duratio threshold, so FreeSWITCH can play an announcement message
func (sm *FSSessionManager) WarnSessionMinDuration(sessionUuid, connId string) {
	if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_broadcast %s %s aleg\n\n", sessionUuid, sm.cfg.LowBalanceAnnFile)); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Could not send uuid_broadcast to freeswitch, error: %s, connection id: %s", err.Error(), connId))
	}
}

func (sm *FSSessionManager) Shutdown() (err error) {
	for connId, fSock := range sm.conns {
		if !fSock.Connected() {
			utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Cannot shutdown sessions, fsock not connected for connection id: %s", connId))
			continue
		}
		utils.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> Shutting down all sessions on connection id: %s", connId))
		if _, err = fSock.SendApiCmd("hupall MANAGER_REQUEST cgr_reqtype *prepaid"); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Error on calls shutdown: %s, connection id: %s", err.Error(), connId))
		}
	}
	for i := 0; len(sm.sessions.getSessions()) > 0 && i < 20; i++ {
		time.Sleep(100 * time.Millisecond) // wait for the hungup event to be fired
		utils.Logger.Info(fmt.Sprintf("<SM-FreeSWITCH> Shutdown waiting on sessions: %v", sm.sessions))
	}
	return nil
}

// Sync sessions with FS
/*
map[secure: hostname:CgrDev1 callstate:ACTIVE callee_num:1002 initial_dest:1002 state:CS_EXECUTE dialplan:XML read_codec:SPEEX initial_ip_addr:127.0.0.1 write_codec:SPEEX write_bit_rate:44000
call_uuid:3427e500-10e5-4864-a589-e306b70419a2 presence_id: initial_cid_name:1001 context:default read_rate:32000 read_bit_rate:44000 callee_direction:SEND initial_context:default created:2015-06-15 18:48:13
dest:1002 callee_name:Outbound Call direction:inbound ip_addr:127.0.0.1 sent_callee_name:Outbound Call write_rate:32000 presence_data: sent_callee_num:1002 created_epoch:1434386893 cid_name:1001 application:sched_hangup
application_data:+10800 alloted_timeout uuid:3427e500-10e5-4864-a589-e306b70419a2 name:sofia/cgrtest/1001@127.0.0.1 cid_num:1001 initial_cid_num:1001 initial_dialplan:XML]
*/
func (sm *FSSessionManager) SyncSessions() error {
	for connId, senderPool := range sm.senderPools {
		fsConn, err := senderPool.PopFSock()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Error on syncing active calls, senderPool: %+v, error: %s", senderPool, err.Error()))
			continue
		}
		activeChanStr, err := fsConn.SendApiCmd("show channels")
		senderPool.PushFSock(fsConn)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Error on syncing active calls, senderPool: %+v, error: %s", senderPool, err.Error()))
			continue
		}
		aChans := fsock.MapChanData(activeChanStr)
		if len(aChans) == 0 && strings.HasPrefix(activeChanStr, "uuid,direction") { // Failed converting output from FS
			utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Syncing active calls, failed converting output from FS: %s", activeChanStr))
			continue
		}
		for _, session := range sm.sessions.getSessions() {
			if session.connId != connId { // This session belongs to another connectionId
				continue
			}
			var stillActive bool
			for _, fsAChan := range aChans {
				if fsAChan["call_uuid"] == session.eventStart.GetUUID() { // Channel still active
					stillActive = true
					break
				}
			}
			if stillActive { // No need to do anything since the channel is still there
				continue
			}
			utils.Logger.Warning(fmt.Sprintf("<SM-FreeSWITCH> Sync active channels, stale session detected, uuid: %s", session.eventStart.GetUUID()))
			fsev := session.eventStart.(FSEvent)
			now := time.Now()
			aTime, _ := fsev.GetAnswerTime("", sm.timezone)
			dur := now.Sub(aTime)
			fsev[END_TIME] = now.String()
			fsev[DURATION] = strconv.FormatFloat(dur.Seconds(), 'f', -1, 64)
			if err := sm.sessions.removeSession(session, fsev); err != nil { // Stop loop, refund advanced charges and save the costs deducted so far to database
				utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> Error on removing stale session with uuid: %s, error: %s", session.eventStart.GetUUID(), err.Error()))
				continue
			}
		}
	}
	return nil
}
