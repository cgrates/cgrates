package fsock

import (
	"bufio"
	"os"
	"testing"
)

const (
	HEADER = `Content-Length: 564
Content-Type: text/event-plain

`
	BODY = `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2012-10-05%2013%3A41%3A38
Event-Date-GMT: Fri,%2005%20Oct%202012%2011%3A41%3A38%20GMT
Event-Date-Timestamp: 1349437298012866
Event-Calling-File: switch_scheduler.c
Event-Calling-Function: switch_scheduler_execute
Event-Calling-Line-Number: 65
Event-Sequence: 34263
Task-ID: 2
Task-Desc: heartbeat
Task-Group: core
Task-Runtime: 1349437318

extra data
`
)

func TestHeaders(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("Error creating pype!")
	}
	fs = &fSock{}
	fs.buffer = bufio.NewReader(r)
	w.Write([]byte(HEADER))
	h, err := readHeaders()
	if err != nil || h != "Content-Length: 564\nContent-Type: text/event-plain\n" {
		t.Error("Error parsing headers: ", h, err)
	}
}

func TestEvent(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("Error creating pype!")
	}
	fs = &fSock{}
	fs.buffer = bufio.NewReader(r)
	w.Write([]byte(HEADER + BODY))
	h, b, err := readEvent()
	if err != nil || h != HEADER[:len(HEADER)-1] || len(b) != 564 {
		t.Error("Error parsing event: ", h, b, err)
	}
}

func TestHeaderVal(t *testing.T) {
	h := headerVal(BODY, "Event-Date-GMT")
	if h != "Fri,%2005%20Oct%202012%2011%3A41%3A38%20GMT" {
		t.Error("Header val error: ", h)
	}
}

func TestEventToMapUnfiltered(t *testing.T) {
	fields := FSEventStrToMap(BODY, nil)
	if fields["Event-Name"] != "RE_SCHEDULE" {
		t.Error("Event not parsed correctly: ", fields)
	}
	if len(fields) != 17 {
		t.Error("Incorrect number of event fields: ", len(fields))
	}
}

func TestEventToMapFiltered(t *testing.T) {
	fields := FSEventStrToMap(BODY, []string{"Event-Name", "Task-Group", "Event-Date-GMT"})
	if fields["Event-Date-Local"] != "2012-10-05 13:41:38" {
		t.Error("Event not parsed correctly: ", fields)
	}
	if len(fields) != 14 {
		t.Error("Incorrect number of event fields: ", len(fields))
	}
}

func BenchmarkHeaderVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		headerVal(HEADER, "Content-Length")
		headerVal(BODY, "Event-Date-Loca")
	}
}
