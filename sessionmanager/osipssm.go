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
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/osipsdagram"
)

/*
E_ACC_EVENT
method::INVITE
from_tag::87d02470
to_tag::a671a98
callid::05dac0aaa716c9814f855f0e8fee6936@0:0:0:0:0:0:0:0
sip_code::200
sip_reason::OK
time::1430579770
cgr_reqtype::*pseudoprepaid
cgr_account::1002
cgr_subject::1002
cgr_destination::1002
originalUri::sip:1002@172.16.254.77
duration::

#
E_ACC_EVENT
method::BYE
from_tag::a671a98
to_tag::87d02470
callid::05dac0aaa716c9814f855f0e8fee6936@0:0:0:0:0:0:0:0
sip_code::200
sip_reason::OK
time::1430579797
cgr_reqtype::
cgr_account::
cgr_subject::
cgr_destination::
originalUri::sip:1002@172.16.254.1:5060;transport=udp;registering_acc=172_16_254_77
duration::

E_ACC_MISSED_EVENT
method::INVITE
from_tag::1d5efcc1
to_tag::
callid::c0965d3f42c720397ca1a5be9619c2ef@0:0:0:0:0:0:0:0
sip_code::404
sip_reason::Not Found
time::1430579759
cgr_reqtype::*pseudoprepaid
cgr_account::1002
cgr_subject::1002
cgr_destination::1002
originalUri::sip:1002@172.16.254.77
duration::

*/

func NewOSipsSessionManager(smOsipsCfg *config.SmOsipsConfig, rater, cdrsrv engine.Connector) (*OsipsSessionManager, error) {
	osm := &OsipsSessionManager{cfg: smOsipsCfg, rater: rater, cdrsrv: cdrsrv, cdrStartEvents: make(map[string]*OsipsEvent)}
	osm.eventHandlers = map[string][]func(*osipsdagram.OsipsEvent){
		"E_OPENSIPS_START":   []func(*osipsdagram.OsipsEvent){osm.onOpensipsStart}, // Raised when OpenSIPS starts so we can register our event handlers
		"E_ACC_CDR":          []func(*osipsdagram.OsipsEvent){osm.onCdr},           // Raised if cdr_flag is configured
		"E_ACC_MISSED_EVENT": []func(*osipsdagram.OsipsEvent){osm.onCdr},           // Raised if evi_missed_flag is configured
		"E_ACC_EVENT":        []func(*osipsdagram.OsipsEvent){osm.onAccEvent},      // Raised if evi_flag is configured and not cdr_flag containing start/stop events
	}
	return osm, nil
}

type OsipsSessionManager struct {
	cfg             *config.SmOsipsConfig
	rater           engine.Connector
	cdrsrv          engine.Connector
	eventHandlers   map[string][]func(*osipsdagram.OsipsEvent)
	evSubscribeStop chan struct{}                         // Reference towards the channel controlling subscriptions, keep it as reference so we do not need to copy it
	stopServing     chan struct{}                         // Stop serving datagrams
	miConn          *osipsdagram.OsipsMiDatagramConnector // Pool of connections used to various OpenSIPS servers, keep reference towards events received so we can issue commands always to the same remote
	sessions        []*Session
	cdrStartEvents  map[string]*OsipsEvent // Used when building CDRs, ToDo: secure access to map
}

// Called when firing up the session manager, will stay connected for the duration of the daemon running
func (osm *OsipsSessionManager) Connect() (err error) {
	osm.stopServing = make(chan struct{})
	if osm.miConn, err = osipsdagram.NewOsipsMiDatagramConnector(osm.cfg.MiAddr, osm.cfg.Reconnects); err != nil {
		return fmt.Errorf("Cannot connect to OpenSIPS at %s, error: %s", osm.cfg.MiAddr, err.Error())
	}
	osm.evSubscribeStop = make(chan struct{})
	defer func() { osm.evSubscribeStop <- struct{}{} }() // Stop subscribing on disconnect
	go osm.SubscribeEvents(osm.evSubscribeStop)
	evsrv, err := osipsdagram.NewEventServer(osm.cfg.ListenUdp, osm.eventHandlers)
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Cannot initialize datagram server, error: <%s>", err.Error()))
		return
	}
	engine.Logger.Info(fmt.Sprintf("<SM-OpenSIPS> Listening for datagram events at <%s>", osm.cfg.ListenUdp))
	evsrv.ServeEvents(osm.stopServing) // Will break through stopServing on error in other places
	return errors.New("<SM-OpenSIPS> Stopped reading events")
}

// Removes a session on call end
func (osm *OsipsSessionManager) RemoveSession(uuid string) {
	for i, ss := range osm.sessions {
		if ss.eventStart.GetUUID() == uuid {
			osm.sessions = append(osm.sessions[:i], osm.sessions[i+1:]...)
			return
		}
	}
}

// DebitInterval will give out the frequence of the debits sent to engine
func (osm *OsipsSessionManager) DebitInterval() time.Duration {
	return osm.cfg.DebitInterval
}

// Returns the connection to local cdr database, used by session to log it's final costs
func (osm *OsipsSessionManager) CdrSrv() engine.Connector {
	return osm.cdrsrv
}

// Returns connection to rater/controller
func (osm *OsipsSessionManager) Rater() engine.Connector {
	return osm.rater
}

// Part of the session manager interface, not really used with OpenSIPS now
func (osm *OsipsSessionManager) WarnSessionMinDuration(sessionUuid, connId string) {
	return
}

// Called on session manager shutdown, could add more cleanup actions in the future
func (osm *OsipsSessionManager) Shutdown() error {
	return nil
}

// Process the CDR with CDRS component
func (osm *OsipsSessionManager) ProcessCdr(storedCdr *engine.StoredCdr) error {
	var reply string
	return osm.cdrsrv.ProcessCdr(storedCdr, &reply)
}

// Disconnects the session
func (osm *OsipsSessionManager) DisconnectSession(ev engine.Event, connId, notify string) error {
	sessionIds := ev.GetSessionIds()
	if len(sessionIds) != 2 {
		errMsg := fmt.Sprintf("Failed disconnecting session for event: %+v, notify: %s, dialogId: %v", ev, notify, sessionIds)
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> " + errMsg))
		return errors.New(errMsg)
	}
	cmd := fmt.Sprintf(":dlg_end_dlg:\n%s\n%s\n\n", sessionIds[0], sessionIds[1])
	if reply, err := osm.miConn.SendCommand([]byte(cmd)); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed disconnecting session for event: %+v, notify: %s, dialogId: %v, error: <%s>", ev, notify, sessionIds, err))
		return err
	} else if !bytes.HasPrefix(reply, []byte("200 OK")) {
		errStr := fmt.Sprintf("Failed disconnecting session for event: %+v, notify: %s, dialogId: %v", ev, notify, sessionIds)
		engine.Logger.Err("<SM-OpenSIPS> " + errStr)
		return errors.New(errStr)
	}
	return nil
}

// Automatic subscribe to OpenSIPS for events, trigered on Connect or OpenSIPS restart
func (osm *OsipsSessionManager) SubscribeEvents(evStop chan struct{}) error {
	if err := osm.subscribeEvents(); err != nil { // Init subscribe
		close(osm.stopServing) // Do not serve anymore since we got errors on subscribing
	}
	for {
		select {
		case <-evStop: // Break this loop from outside
			return nil
		case <-time.After(osm.cfg.EventsSubscribeInterval): // Subscribe on interval
			if err := osm.subscribeEvents(); err != nil {
				close(osm.stopServing) // Order stop serving, do not return here since we will block the channel consuming
			}
		}
	}
}

// One subscribe attempt to OpenSIPS
func (osm *OsipsSessionManager) subscribeEvents() error {
	subscribeInterval := osm.cfg.EventsSubscribeInterval + time.Duration(1)*time.Second // Avoid concurrency on expiry
	listenAddrSplt := strings.Split(osm.cfg.ListenUdp, ":")
	portListen := listenAddrSplt[1]
	addrListen := listenAddrSplt[0]
	if len(addrListen) == 0 { //Listen on all addresses, try finding out from mi connection
		if localAddr := osm.miConn.LocallAddr(); localAddr != nil {
			addrListen = strings.Split(localAddr.String(), ":")[0]
		}
	}
	for eventName := range osm.eventHandlers {
		if eventName == "E_OPENSIPS_START" { // Do not subscribe for start since this should be hardcoded
			continue
		}
		cmd := fmt.Sprintf(":event_subscribe:\n%s\nudp:%s:%s\n%d\n", eventName, addrListen, portListen, int(subscribeInterval.Seconds()))
		if reply, err := osm.miConn.SendCommand([]byte(cmd)); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed subscribing to OpenSIPS at address: <%s>, error: <%s>", osm.cfg.MiAddr, err))
			return err
		} else if !bytes.HasPrefix(reply, []byte("200 OK")) {
			engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed subscribing to OpenSIPS at address: <%s>", osm.cfg.MiAddr))
			return errors.New("Failed subscribing to OpenSIPS events")
		}
	}
	return nil
}

// Triggered opensips_start  event
func (osm *OsipsSessionManager) onOpensipsStart(cdrDagram *osipsdagram.OsipsEvent) {
	osm.evSubscribeStop <- struct{}{}         // Cancel previous subscribes
	osm.evSubscribeStop = make(chan struct{}) // Create a fresh communication channel
	go osm.SubscribeEvents(osm.evSubscribeStop)
}

// Triggered by CDR event
func (osm *OsipsSessionManager) onCdr(cdrDagram *osipsdagram.OsipsEvent) {
	osipsEv, _ := NewOsipsEvent(cdrDagram)
	if err := osm.ProcessCdr(osipsEv.AsStoredCdr()); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", osipsEv.GetCgrId(), osipsEv.GetUUID(), err.Error()))
	}
}

// Triggered by ACC_EVENT
func (osm *OsipsSessionManager) onAccEvent(osipsDgram *osipsdagram.OsipsEvent) {
	osipsEv, _ := NewOsipsEvent(osipsDgram)
	if osipsEv.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if osipsDgram.AttrValues["method"] == "INVITE" { // Call start
		if err := osm.callStart(osipsEv); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing CALL_START out of %+v, error: <%s>", osipsDgram, err.Error()))
		}
		if err := osm.processCdrStart(osipsEv); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing cdr start out of %+v, error: <%s>", osipsDgram, err.Error()))
		}
	} else if osipsDgram.AttrValues["method"] == "BYE" {
		if err := osm.callEnd(osipsEv); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing CALL_END out of %+v, error: <%s>", osipsDgram, err.Error()))
		}
		if err := osm.processCdrStop(osipsEv); err != nil {
			engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> Failed processing cdr stop out of %+v, error: <%s>", osipsDgram, err.Error()))
		}
	}
}

// Handler of call start event. Mostly starts a session if needed
func (osm *OsipsSessionManager) callStart(osipsEv *OsipsEvent) error {
	if osipsEv.MissingParameter() {
		if err := osm.DisconnectSession(osipsEv, "", utils.ERR_MANDATORY_IE_MISSING); err != nil {
			return err
		}
		return errors.New(utils.ERR_MANDATORY_IE_MISSING)
	}
	s := NewSession(osipsEv, "", osm)
	if s != nil {
		osm.sessions = append(osm.sessions, s)
	}
	return nil
}

// Handler for callEnd. Mostly removes a session if needed
func (osm *OsipsSessionManager) callEnd(osipsEv *OsipsEvent) error {
	s := osm.getSession(osipsEv.GetUUID())
	if s == nil { // Not handled by us
		return nil
	}
	osm.RemoveSession(s.eventStart.GetUUID()) // Unreference it early so we avoid concurrency
	origEvent := s.eventStart.(*OsipsEvent)   // Need a complete event for methods in close
	if err := origEvent.updateDurationFromEvent(osipsEv); err != nil {
		return err
	}
	if origEvent.MissingParameter() {
		return errors.New(utils.ERR_MANDATORY_IE_MISSING)
	}
	if err := s.Close(origEvent); err != nil { // Stop loop, refund advanced charges and save the costs deducted so far to database
		return err
	}
	return nil
}

// Records the event start in case of received so we can create CDR out of it
func (osm *OsipsSessionManager) processCdrStart(osipsEv *OsipsEvent) error {
	if !osm.cfg.CreateCdr {
		return nil
	}
	if dialogId := osipsEv.DialogId(); dialogId == "" {
		return errors.New("Missing dialog_id")
	} else {
		osm.cdrStartEvents[dialogId] = osipsEv
	}
	return nil
}

// processCdrStop builds the complete CDR out of eventStart+eventStop and sends it to the CDRS component
func (osm *OsipsSessionManager) processCdrStop(osipsEv *OsipsEvent) error {
	if osm.cdrsrv == nil {
		return nil
	}
	var osipsEvStart *OsipsEvent
	var hasIt bool
	if dialogId := osipsEv.DialogId(); dialogId == "" {
		return errors.New("Missing dialog_id")
	} else if osipsEvStart, hasIt = osm.cdrStartEvents[dialogId]; !hasIt {
		return errors.New("Missing event start info")
	} else {
		delete(osm.cdrStartEvents, dialogId) // Cleanup the event once we got it
	}
	if err := osipsEvStart.updateDurationFromEvent(osipsEv); err != nil {
		return err
	}
	return osm.ProcessCdr(osipsEvStart.AsStoredCdr())
}

// Searches and return the session with the specifed uuid
func (osm *OsipsSessionManager) getSession(uuid string) *Session {
	for _, s := range osm.sessions {
		if s.eventStart.GetUUID() == uuid {
			return s
		}
	}
	return nil
}
