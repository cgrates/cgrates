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
	"regexp"
)

type Event struct {
	Fields map[string]string
}

func NewEvent() (ev *Event) {
	return &Event{Fields: make(map[string]string)}
}

var (
	storageGetter, _ = timespans.NewRedisStorage("tcp:127.0.0.1:6379", 10)
)

func (ev *Event) String() (result string) {
	for k, v := range ev.Fields {
		result += fmt.Sprintf("%s = %s\n", k, v)
	}
	result += "=============================================================="
	return
}

type SessionManager struct {
	conn        net.Conn
	eventBodyRE *regexp.Regexp
	sessions    []*Session
}

func (sm *SessionManager) Connect(address, pass string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Could not connect to freeswitch server!")
	}
	sm.conn = conn
	sm.eventBodyRE, _ = regexp.Compile(`"(.*?)":\s+"(.*?)"`)
	fmt.Fprint(sm.conn, fmt.Sprintf("auth %s\n\n", pass))
	fmt.Fprint(sm.conn, "event json all\n\n")
}

func (sm *SessionManager) Close() {
	sm.conn.Close()
}

func (sm *SessionManager) ReadNextEvent() (ev *Event) {
	body, err := bufio.NewReader(sm.conn).ReadString('}')
	if err != nil {
		log.Print("Could not read from freeswitch connection!")
	}
	ev = NewEvent()
	for _, fields := range sm.eventBodyRE.FindAllStringSubmatch(body, -1) {
		if len(fields) == 3 {
			ev.Fields[fields[1]] = fields[2]
		} else {
			log.Printf("malformed event field: %v", fields)
		}
	}
	return
}

func (ssm *SessionManager) onHeartBeat(event string) {
	log.Printf("hear beat event: %s", event)
}
