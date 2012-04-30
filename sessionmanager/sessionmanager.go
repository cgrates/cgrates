package sessionmanager

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
)

type Event struct {
	body string
}

func (ev *Event) GetField(field string) (result string, err error) {
	if re, err := regexp.Compile(fmt.Sprintf(`"%s":\s+"(.*?)"`, field)); err == nil {
		results := re.FindStringSubmatch(ev.body)
		if len(results) > 1 {
			result = results[1]
		}
	}
	return
}

type SessionManager struct {
	conn net.Conn
}

func (sm *SessionManager) Connect(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Could not connect to freeswitch server!")
	}
	sm.conn = conn
	fmt.Fprint(sm.conn, "auth ClueCon\n\n")
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
	ev = &Event{}
	ev.body = body
	return
}

func (ssm *SessionManager) onHeartBeat(event string) {
	log.Printf("hear beat event: %s", event)
}
