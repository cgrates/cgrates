package sessionmanager

import (
	"bufio"
	"fmt"
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
		ev.Fields[fields[1]] = fields[2]
	}
	return
}

func (ssm *SessionManager) onHeartBeat(event string) {
	log.Printf("hear beat event: %s", event)
}
