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
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Event type holding a mapping of all event's proprieties
type FSEvent struct {
	Fields map[string]string
}

var (
	eventBodyRE = regexp.MustCompile(`"(.*?)":\s+"(.*?)"`) // for parsing the proprieties
)

const (
	// Freswitch event proprities names
	DIRECTION          = "Call-Direction"
	ORIG_ID            = "variable_sip_call_id" //- originator_id - match cdrs
	SUBJECT            = "variable_cgr_subject"
	ACCOUNT            = "variable_cgr_account"
	DESTINATION        = "variable_cgr_destination"
	REQTYPE            = "variable_cgr_reqtype" //prepaid or postpaid
	TOR                = "variable_cgr_tor"
	UUID               = "Unique-ID" // -Unique ID for this call leg
	CSTMID             = "variable_cgr_cstmid"
	CALL_DEST_NB       = "Caller-Destination-Number"
	PARK_TIME          = "Event-Date-Timestamp"
	START_TIME         = "variable_answer_epoch"
	END_TIME           = "variable_end_epoch"
	NAME               = "Event-Name"
	HEARTBEAT          = "HEARTBEAT"
	ANSWER             = "CHANNEL_ANSWER"
	HANGUP             = "CHANNEL_HANGUP_COMPLETE"
	PARK               = "CHANNEL_PARK"
	REQTYPE_PREPAID    = "prepaid"
	REQTYPE_POSTPAID   = "postpaid"
	AUTH_OK            = "+AUTH_OK"
	INSUFFICIENT_FUNDS = "-INSUFFICIENT_FUNDS"
	MISSING_PARAMETER  = "-MISSING_PARAMETER"
	SYSTEM_ERROR       = "-SYSTEM_ERROR"
)

// Nice printing for the event object.
func (fsev *FSEvent) String() (result string) {
	for k, v := range fsev.Fields {
		result += fmt.Sprintf("%s = %s\n", k, v)
	}
	result += "=============================================================="
	return
}

// Loads the new event data from a body of text containing the key value proprieties.
// It stores the parsed proprieties in the internal map.
func (fsev *FSEvent) New(body string) Event {
	fsev.Fields = make(map[string]string)
	for _, fields := range eventBodyRE.FindAllStringSubmatch(body, -1) {
		if len(fields) == 3 {
			fsev.Fields[fields[1]] = fields[2]
		} else {
			timespans.Logger.Err(fmt.Sprintf("malformed event field: %v", fields))
		}
	}
	return fsev
}

func (fsev *FSEvent) GetName() string {
	return fsev.Fields[NAME]
}
func (fsev *FSEvent) GetDirection() string {
	//TODO: temporary hack
	return "OUT"
	//return fsev.Fields[DIRECTION]
}
func (fsev *FSEvent) GetOrigId() string {
	return fsev.Fields[ORIG_ID]
}
func (fsev *FSEvent) GetSubject() string {
	return fsev.Fields[SUBJECT]
}
func (fsev *FSEvent) GetAccount() string {
	return fsev.Fields[ACCOUNT]
}
func (fsev *FSEvent) GetDestination() string {
	return fsev.Fields[DESTINATION]
}
func (fsev *FSEvent) GetTOR() string {
	return fsev.Fields[TOR]
}
func (fsev *FSEvent) GetUUID() string {
	return fsev.Fields[UUID]
}
func (fsev *FSEvent) GetTenant() string {
	return fsev.Fields[CSTMID]
}
func (fsev *FSEvent) GetCallDestNb() string {
	return fsev.Fields[CALL_DEST_NB]
}
func (fsev *FSEvent) GetReqType() string {
	return fsev.Fields[REQTYPE]
}
func (fsev *FSEvent) MissingParameter() bool {
	return strings.TrimSpace(fsev.GetDirection()) == "" ||
		strings.TrimSpace(fsev.GetOrigId()) == "" ||
		strings.TrimSpace(fsev.GetSubject()) == "" ||
		strings.TrimSpace(fsev.GetAccount()) == "" ||
		strings.TrimSpace(fsev.GetDestination()) == "" ||
		strings.TrimSpace(fsev.GetTOR()) == "" ||
		strings.TrimSpace(fsev.GetUUID()) == "" ||
		strings.TrimSpace(fsev.GetTenant()) == "" ||
		strings.TrimSpace(fsev.GetCallDestNb()) == ""
}
func (fsev *FSEvent) GetStartTime(field string) (t time.Time, err error) {
	st, err := strconv.ParseInt(fsev.Fields[field], 0, 64)
	t = time.Unix(0, st*1000)
	return
}

func (fsev *FSEvent) GetEndTime() (t time.Time, err error) {
	st, err := strconv.ParseInt(fsev.Fields[END_TIME], 0, 64)
	t = time.Unix(0, st*1000)
	return
}
