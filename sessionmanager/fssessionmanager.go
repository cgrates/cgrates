/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"
)

// The freeswitch session manager type holding a buffer for the network connection,
// the active sessions, and a session delegate doing specific actions on every session.
type FSSessionManager struct {
	conn            net.Conn
	buf             *bufio.Reader
	sessions        []*Session
	sessionDelegate *SessionDelegate
	postgresLogger  *PostgresLogger
	address, pass   string
}

func NewFSSessionManager(db *sql.DB) *FSSessionManager {
	return &FSSessionManager{postgresLogger: &PostgresLogger{db}}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events in json format.
func (sm *FSSessionManager) Connect(ed *SessionDelegate, address, pass string)  {
	if ed == nil {
		log.Fatal("Please provide a non nil SessionDelegate")
	}
	sm.sessionDelegate = ed
	sm.address = address
	sm.pass = pass
	if sm.conn != nil {
		// in case it is a reconnect
		sm.conn.Close()
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Print("Could not connect to freeswitch server!")
		return
	}
	log.Print("Successfuly connected to freeswitch! ")
	sm.conn = conn
	sm.buf = bufio.NewReaderSize(conn, 8192)
	fmt.Fprint(conn, fmt.Sprintf("auth %s\n\n", pass))
	fmt.Fprint(conn, "event json all\n\n")
	go func() {
		for {
			sm.readNextEvent()
		}
	}()
}

// Reads from freeswitch server buffer until it encounters a '}',
// than it creates an event object and calls the appropriate method
// on the session delegate.
func (sm *FSSessionManager) readNextEvent() (ev Event) {
	body, err := sm.buf.ReadString('}')
	if err != nil {
		log.Print("Could not read from freeswitch connection!")
		// wait until a sec
		time.Sleep(5 * time.Second)
		// try to reconnect
		sm.Connect(sm.sessionDelegate, sm.address, sm.pass)
	}
	ev = new(FSEvent).New(body)
	switch ev.GetName() {
	case HEARTBEAT:
		sm.OnHeartBeat(ev)
	case ANSWER:
		sm.OnChannelAnswer(ev)
	case HANGUP:
		sm.OnChannelHangupComplete(ev)
	default:
		sm.OnOther(ev)
	}
	return
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
func (sm *FSSessionManager) DisconnectSession(s *Session) {
	fmt.Fprint(sm.conn, fmt.Sprintf("SendMsg %s\ncall-command: hangup\nhangup-cause: MANAGER_REQUEST\n\n", s.uuid))
	s.Close()
}

// Called on freeswitch's hearbeat event
func (sm *FSSessionManager) OnHeartBeat(ev Event) {
	if sm.sessionDelegate != nil {
		sm.sessionDelegate.OnHeartBeat(ev)
	} else {
		log.Print("♥")
	}
}

// Called on freeswitch's answer event
func (sm *FSSessionManager) OnChannelAnswer(ev Event) {
	if sm.sessionDelegate != nil {
		s := NewSession(ev, sm)
		sm.sessions = append(sm.sessions, s)
		sm.sessionDelegate.OnChannelAnswer(ev, s)
	} else {
		log.Print("answer")
	}
}

// Called on freeswitch's hangup event
func (sm *FSSessionManager) OnChannelHangupComplete(ev Event) {
	s := sm.GetSession(ev.GetUUID())
	if sm.sessionDelegate != nil {
		sm.sessionDelegate.OnChannelHangupComplete(ev, s)
		s.SaveOperations()
	} else {
		log.Print("HangupComplete")
	}
	if s != nil {

		s.Close()
	}
}

// Called on freeswitch's events not processed by the session manger,
// for logging purposes (maybe).
func (sm *FSSessionManager) OnOther(ev Event) {
	//log.Printf("Other event: %s", ev.GetName())
}

func (sm *FSSessionManager) GetSessionDelegate() *SessionDelegate {
	return sm.sessionDelegate
}

func (sm *FSSessionManager) GetDbLogger() *PostgresLogger {
	return sm.postgresLogger
}
