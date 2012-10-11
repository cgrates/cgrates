/* Provides freeswitch socket communication
 */
package fsock

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/syslog"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var fs *fSock // Used to share FS connection via package globals (singleton)

// Connection to FreeSWITCH Socket
type fSock struct {
	conn               net.Conn
	buffer             *bufio.Reader
	fsaddress, fspaswd string
	eventHandlers      map[string]func(string)
	eventFilters       map[string]string
	apiChan            chan string
	cmdChan            chan string
	reconnects         int
	delayFunc          func() int
	logger             *syslog.Writer
}

// Connects to FS and starts buffering input
func New(fsaddr, fspaswd string, reconnects int, eventHandlers map[string]func(string), eventFilters map[string]string, l *syslog.Writer) error {
	fs = &fSock{fsaddress: fsaddr, fspaswd: fspaswd, eventHandlers: eventHandlers, eventFilters: eventFilters, logger: l}
	fs.apiChan = make(chan string) // Init apichan so we can use it to pass api replies
	fs.reconnects = reconnects
	fs.delayFunc = fib()
	errConn := Connect(reconnects)
	if errConn != nil {
		return errConn
	}
	return nil
}

// Extracts value of a header from anywhere in content string
func headerVal(hdrs, hdr string) string {
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
func urlDecode(hdrVal string) string {
	if valUnescaped, errUnescaping := url.QueryUnescape(hdrVal); errUnescaping == nil {
		hdrVal = valUnescaped
	}
	return hdrVal
}

// Binary string search in slice
func isSliceMember(ss []string, s string) bool {
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
			if filtered && isSliceMember(headers, hdrVal[0]) {
				continue // Loop again since we only work on filtered fields
			}
			fsevent[hdrVal[0]] = urlDecode(strings.TrimSpace(strings.TrimRight(hdrVal[1], "\n")))
		}
	}
	return fsevent
}

// Reads headers until delimiter reached
func readHeaders() (s string, err error) {
	bytesRead := make([]byte, 0)
	var readLine []byte
	for {
		readLine, err = fs.buffer.ReadBytes('\n')
		if err != nil {
			return
		}
		// No Error, add received to localread buffer
		if len(bytes.TrimSpace(readLine)) == 0 {
			break
		}
		bytesRead = append(bytesRead, readLine...)
	}
	return string(bytesRead), nil
}

// Reads the body from buffer, ln is given by content-length of headers
func readBody(ln int) (string, error) {
	bytesRead := make([]byte, ln)
	for i := 0; i < ln; i++ {
		if readByte, err := fs.buffer.ReadByte(); err != nil {
			return "", err
		} else { // No Error, add received to localread buffer
			bytesRead[i] = readByte // Add received line to the local read buffer
		}
	}
	return string(bytesRead), nil
}

// Event is made out of headers and body (if present)
func readEvent() (string, string, error) {
	var hdrs, body string
	var cl int
	var err error

	if hdrs, err = readHeaders(); err != nil {
		return "", "", err
	}
	if !strings.Contains(hdrs, "Content-Length") { //No body
		return hdrs, "", nil
	}
	clStr := headerVal(hdrs, "Content-Length")
	if cl, err = strconv.Atoi(clStr); err != nil {
		return "", "", errors.New("Cannot extract content length")
	}
	if body, err = readBody(cl); err != nil {
		return "", "", err
	}
	return hdrs, body, nil
}

// Checks if socket connected. Can be extended with pings
func Connected() bool {
	if fs.conn == nil {
		return false
	}
	return true
}

// Disconnects from socket
func Disconnect() (err error) {
	if fs.conn != nil {
		err = fs.conn.Close()
	}
	return
}

// Auth to FS
func Auth() error {
	authCmd := fmt.Sprintf("auth %s\n\n", fs.fspaswd)
	fmt.Fprint(fs.conn, authCmd)
	if rply, err := readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK accepted") {
		fmt.Println("Got reply to auth:", rply)
		return errors.New("auth error")
	}
	return nil
}

// Subscribe to events
func EventsPlain(events []string) error {
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
	fmt.Fprint(fs.conn, eventsCmd) // Send command here
	if rply, err := readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK") {
		return errors.New("event error")
	}
	return nil
}

// Enable filters
func filterEvents(filters map[string]string) error {
	if len(filters) == 0 { //Nothing to filter
		return nil
	}
	cmd := "filter"
	for hdr, val := range filters {
		cmd += " " + hdr + " " + val
	}
	cmd += "\n\n"
	fmt.Fprint(fs.conn, cmd)
	if rply, err := readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK") {
		return errors.New("filter error")
	}
	return nil
}

// Connect or reconnect
func Connect(reconnects int) error {
	if Connected() {
		Disconnect()
	}
	var conErr error
	for i := 0; i < reconnects; i++ {
		fs.conn, conErr = net.Dial("tcp", fs.fsaddress)
		if conErr == nil {
			// Connected, init buffer, auth and subscribe to desired events and filters
			fs.buffer = bufio.NewReaderSize(fs.conn, 8192) // reinit buffer
			if authChg, err := readHeaders(); err != nil || !strings.Contains(authChg, "auth/request") {
				return errors.New("No auth challenge received")
			} else if errAuth := Auth(); errAuth != nil { // Auth did not succeed
				return errAuth
			}
			// Subscribe to events handled by event handlers
			handledEvs := make([]string, len(fs.eventHandlers))
			j := 0
			for k := range fs.eventHandlers {
				handledEvs[j] = k
				j++
			}
			if subscribeErr := EventsPlain(handledEvs); subscribeErr != nil {
				return subscribeErr
			}
			if filterErr := filterEvents(fs.eventFilters); filterErr != nil {
				return filterErr
			}
			return nil
		}
		time.Sleep(time.Duration(fs.delayFunc()) * time.Second)
	}
	return conErr
}

// Send API command
func SendApiCmd(cmdStr string) error {
	if !Connected() {
		return errors.New("Not connected to FS")
	}
	cmd := fmt.Sprintf("api %s\n\n", cmdStr)
	fmt.Fprint(fs.conn, cmd)
	resEvent := <-fs.apiChan
	if strings.Contains(resEvent, "-ERR") {
		return errors.New("Command failed")
	}
	return nil
}

// SendMessage command
func SendMsgCmd(uuid string, cmdargs map[string]string) error {
	if len(cmdargs) == 0 {
		return errors.New("Need command arguments")
	}
	if !Connected() {
		return errors.New("Not connected to FS")
	}
	argStr := ""
	for k, v := range cmdargs {
		argStr += fmt.Sprintf("%s:%s\n", k, v)
	}
	fmt.Fprint(fs.conn, fmt.Sprintf("sendmsg %s\n%s\n", uuid, argStr))
	replyTxt := <-fs.cmdChan
	if strings.HasPrefix(replyTxt, "-ERR") {
		return fmt.Errorf("SendMessage: %s", replyTxt)
	}
	return nil
}

// Reads events from socket
func ReadEvents() {
	// Read events from buffer, firing them up further
	for {
		hdr, body, err := readEvent()
		if err != nil {
			fs.logger.Warning("FreeSWITCH connection broken: attemting reconnect")
			connErr := Connect(fs.reconnects)
			if connErr != nil {
				return
			}
			continue // Connection reset
		}
		if strings.Contains(hdr, "api/response") {
			fs.apiChan <- hdr + body
		} else if strings.Contains(hdr, "command/reply") {
			fs.cmdChan <- headerVal(hdr, "Reply-Text")
		}
		if body != "" { // We got a body, could be event, try dispatching it
			dispatchEvent(body)
		}
	}

	return
}

// Dispatch events to handlers in async mode
func dispatchEvent(event string) {
	eventName := headerVal(event, "Event-Name")
	if handlerFunc, hasHandler := fs.eventHandlers[eventName]; hasHandler {
		go handlerFunc(event)
	}
}

// successive Fibonacci numbers.
func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}
