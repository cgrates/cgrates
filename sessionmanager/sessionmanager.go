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
	"fmt"
	"github.com/rif/cgrates/timespans"
	"log"
	"net"
)

var (
	storageGetter, _ = timespans.NewRedisStorage("tcp:127.0.0.1:6379", 10)
)

type SessionManager struct {
	buf      *bufio.Reader
	sessions []*Session
}

func (sm *SessionManager) Connect(address, pass string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Could not connect to freeswitch server!")
	}
	sm.buf = bufio.NewReaderSize(conn, 8192)
	fmt.Fprint(conn, fmt.Sprintf("auth %s\n\n", pass))
	fmt.Fprint(conn, "event json all\n\n")
}

func (sm *SessionManager) ReadNextEvent() (ev *Event) {
	body, err := sm.buf.ReadString('}')
	if err != nil {
		log.Print("Could not read from freeswitch connection!")
	}
	ev = NewEvent(body)
	switch ev.Fields["Event-Name"] {
	case "HEARTBEAT":
		sm.OnHeartBeat(ev)
	case "CHANNEL_ANSWER":
		sm.OnChannelAnswer(ev)
	case "CHANNEL_HANGUP_COMPLETE":
		sm.OnChannelHangupComplete(ev)
	default:
		sm.OnOther(ev)
	}
	return
}

func (sm *SessionManager) GetSession(uuid string) *Session {
	for _, s := range sm.sessions {
		if s.uuid == uuid {
			return s
		}
	}
	return nil
}

func (sm *SessionManager) OnHeartBeat(ev *Event) {
	log.Print("heartbeat")
}

func (sm *SessionManager) OnChannelAnswer(ev *Event) {
	s := NewSession(ev)
	log.Printf("answer: %v", s)
}

func (sm *SessionManager) OnChannelHangupComplete(ev *Event) {
	s := sm.GetSession(ev.Fields[UUID])
	s.Close()
}

func (sm *SessionManager) OnOther(ev *Event) {
	//log.Printf("Other event: %s", ev.Fields["Event-Name"])
}
