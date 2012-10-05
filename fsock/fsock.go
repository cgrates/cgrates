/* Provides freeswitch socket communication
 */
package fsock

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var FS *FSock // Used to share FS connection via package globals

// Extracts value of a header from anywhere in content string
func HeaderVal(hdrs, hdr string) string {
	var hdrSIdx, hdrEIdx int
	if hdrSIdx = strings.Index(hdrs, hdr); hdrSIdx == -1 {
		return ""
	} else if hdrEIdx = strings.Index(hdrs[hdrSIdx:], "\n"); hdrEIdx == -1 {
		hdrEIdx = len(hdrs[hdrSIdx:])
	}
	splt := strings.SplitN(hdrs[hdrSIdx:hdrSIdx+hdrEIdx], ": ", 2)
	if len(splt) != 2 {
		return ""
	}
	return strings.TrimSpace(strings.TrimRight(splt[1], "\n"))
}

// FS event header values are urlencoded. Use this to decode them. On error, use original value
func UrlDecode(hdrVal string) string {
	if valUnescaped, errUnescaping := url.QueryUnescape(hdrVal); errUnescaping == nil {
		hdrVal = valUnescaped
	}
	return hdrVal
}

// Binary string search in slice
func IsSliceMember(ss []string, s string) bool {
	sort.Strings(ss)
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return true
	}
	return false
}

// Convert fseventStr into fseventMap
func FSEventStrToMap(fsevstr string, headers []string) map[string]string {
	fsevent := make(map[string]string)
	filtered := false
	if len(headers) != 0 {
		filtered = true
	}
	for _, strLn := range strings.Split(fsevstr, "\n") {
		if hdrVal := strings.SplitN(strLn, ": ", 2); len(hdrVal) == 2 {
			if filtered && IsSliceMember(headers, hdrVal[0]) {
				continue // Loop again since we only work on filtered fields
			}
			fsevent[hdrVal[0]] = UrlDecode(strings.TrimSpace(strings.TrimRight(hdrVal[1], "\n")))
		}
	}
	return fsevent
}

// Connects to FS and starts buffering input
func NewFSock(fsaddr, fspaswd string, reconnects int, eventHandlers map[string]func(string)) (*FSock, error) {
	eventFilters := make(map[string]string)
	fsock := FSock{fsaddress: fsaddr, fspaswd: fspaswd, eventHandlers: eventHandlers, eventFilters: eventFilters}
	fsock.apiChan = make(chan string) // Init apichan so we can use it to pass api replies
	errConn := fsock.Connect(reconnects)
	if errConn != nil {
		return nil, errConn
	}
	return &fsock, nil
}

// Connection to FreeSWITCH Socket
type FSock struct {
	conn   net.Conn
	buffer *bufio.Reader
	fsaddress,
	fspaswd string
	eventHandlers map[string]func(string)
	eventFilters  map[string]string
	apiChan       chan string
}

// Reads headers until delimiter reached
func (self *FSock) readHeaders() (s string, err error) {
	bytesRead := make([]byte, 0)
	var readLine []byte
	for {
		readLine, err = self.buffer.ReadBytes('\n')
		if err != nil {
			return
		}
		// No Error, add received to localread buffer
		if readLine[0] == '\n' {
			break
		}
		bytesRead = append(bytesRead, readLine...)
	}
	return string(bytesRead), nil
}

// Reads the body from buffer, ln is given by content-length of headers
func (self *FSock) readBody(ln int) (s string, err error) {
	bytesRead := make([]byte, ln)
	n, err := self.buffer.Read(bytesRead)
	if err != nil {
		return
	}
	if n != ln {
		err = errors.New("Could not read whole body")
		return
	}
	return string(bytesRead), nil
}

// Event is made out of headers and body (if present)
func (self *FSock) readEvent() (string, string, error) {
	var hdrs, body string
	var cl int
	var err error
	if hdrs, err = self.readHeaders(); err != nil {
		return "", "", err
	}
	if !strings.Contains(hdrs, "Content-Length") { //No body
		return hdrs, "", nil
	}
	clStr := HeaderVal(hdrs, "Content-Length")
	if cl, err = strconv.Atoi(clStr); err != nil {
		return "", "", errors.New("Cannot extract content length")
	}
	if body, err = self.readBody(cl); err != nil {
		return "", "", errors.New("Cannot extract body")
	}
	return hdrs, body, nil
}

// Checks if socket connected. Can be extended with pings
func (self *FSock) Connected() bool {
	if self.conn == nil {
		return false
	}
	return true
}

// Disconnects from socket
func (self *FSock) Disconnect() {
	if self.conn != nil {
		self.conn.Close()
	}
}

// Auth to FS
func (self *FSock) Auth() error {
	authCmd := fmt.Sprintf("auth %s\n\n", self.fspaswd)
	fmt.Fprint(self.conn, authCmd)
	if rply, err := self.readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK accepted") {
		fmt.Println("Got reply to auth:", rply)
		return errors.New("auth error")
	}
	return nil
}

// Subscribe to events
func (self *FSock) EventsPlain(events []string) error {
	if len(events) == 0 {
		return nil
	}
	eventsCmd := "event plain"
	for _, ev := range events {
		if ev == "ALL" {
			eventsCmd = "event plain all"
			break
		}
		eventsCmd += " " + ev
	}
	eventsCmd += "\n\n"
	fmt.Fprint(self.conn, eventsCmd) // Send command here
	if rply, err := self.readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK") {
		return errors.New("event error")
	}
	return nil
}

// Enable filters
func (self *FSock) FilterEvents(filters map[string]string) error {
	if len(filters) == 0 { //Nothing to filter
		return nil
	}
	cmd := "filter"
	for hdr, val := range filters {
		cmd += " " + hdr + " " + val
	}
	cmd += "\n\n"
	fmt.Fprint(self.conn, cmd)
	if rply, err := self.readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK") {
		return errors.New("filter error")
	}
	return nil
}

// Connect or reconnect
func (self *FSock) Connect(reconnects int) error {
	if self.Connected() {
		self.Disconnect()
	}
	var conErr error
	for i := 0; i < reconnects; i++ {
		fmt.Println("Attempting FS connect")
		self.conn, conErr = net.Dial("tcp", self.fsaddress)
		if conErr == nil {
			// Connected, init buffer, auth and subscribe to desired events and filters
			self.buffer = bufio.NewReaderSize(self.conn, 8192) // reinit buffer
			if authChg, err := self.readHeaders(); err != nil || !strings.Contains(authChg, "auth/request") {
				return errors.New("No auth challenge received")
			} else if errAuth := self.Auth(); errAuth != nil { // Auth did not succeed
				return errAuth
			}
			// Subscribe to events handled by event handlers
			handledEvs := make([]string, len(self.eventHandlers))
			j := 0
			for k := range self.eventHandlers {
				handledEvs[j] = k
				j++
			}
			if subscribeErr := self.EventsPlain(handledEvs); subscribeErr != nil {
				return subscribeErr
			} else if filterErr := self.FilterEvents(self.eventFilters); filterErr != nil {
				return filterErr
			}
			return nil
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return conErr
}

// Send API command
func (self *FSock) SendApiCmd(cmdStr string) error {
	if !self.Connected() {
		return errors.New("Not connected to FS")
	}
	cmd := fmt.Sprintf("api %s\n\n", cmdStr)
	fmt.Fprint(self.conn, cmd)
	resEvent := <-self.apiChan
	if strings.Contains(resEvent, "-ERR") {
		return errors.New("Command failed")
	}
	return nil
}

// Reads events from socket
func (self *FSock) ReadEvents() {
	// Read events from buffer, firing them up further
	for {
		hdr, body, err := self.readEvent()
		if err != nil {
			fmt.Println("WARNING: got error while reading events: ", err.Error())
			connErr := self.Connect(3)
			if connErr != nil {
				fmt.Println("FSock reader - cannot connect to FS")
				return
			}
			continue // Connection reset
		}
		if strings.Contains(hdr, "api/response") {
			self.apiChan <- hdr + body
		}
		if body != "" { // We got a body, could be event, try dispatching it
			self.DispatchEvent(body)
		}
	}
	fmt.Println("Exiting ReadEvents")
	return
}

// Dispatch events to handlers in async mode
func (self *FSock) DispatchEvent(event string) {
	eventName := HeaderVal(event, "Event-Name")
	if handlerFunc, hasHandler := self.eventHandlers[eventName]; hasHandler {
		go handlerFunc(event)
	}
}
