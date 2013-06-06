/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
	"strconv"
	"strings"
	"time"
)

// Event type holding a mapping of all event's proprieties
type FSEvent map[string]string

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
	CALL_DEST_NR       = "Caller-Destination-Number"
	PARK_TIME          = "Caller-Profile-Created-Time"
	START_TIME         = "Caller-Channel-Answered-Time"
	END_TIME           = "Caller-Channel-Hangup-Time"
	NAME               = "Event-Name"
	HEARTBEAT          = "HEARTBEAT"
	ANSWER             = "CHANNEL_ANSWER"
	HANGUP             = "CHANNEL_HANGUP_COMPLETE"
	PARK               = "CHANNEL_PARK"
	AUTH_OK            = "+AUTH_OK"
	DISCONNECT         = "+SWITCH DISCONNECT"
	INSUFFICIENT_FUNDS = "-INSUFFICIENT_FUNDS"
	MISSING_PARAMETER  = "-MISSING_PARAMETER"
	SYSTEM_ERROR       = "-SYSTEM_ERROR"
	MANAGER_REQUEST    = "+MANAGER_REQUEST"
	USERNAME           = "Caller-Username"
)

// Nice printing for the event object.
func (fsev FSEvent) String() (result string) {
	for k, v := range fsev {
		result += fmt.Sprintf("%s = %s\n", k, v)
	}
	result += "=============================================================="
	return
}

// Loads the new event data from a body of text containing the key value proprieties.
// It stores the parsed proprieties in the internal map.
func (fsev FSEvent) New(body string) Event {
	fsev = fsock.FSEventStrToMap(body, nil)
	return fsev
}

func (fsev FSEvent) GetName() string {
	return fsev[NAME]
}
func (fsev FSEvent) GetDirection() string {
	//TODO: implement direction
	return "OUT"
	//return fsev[DIRECTION]
}
func (fsev FSEvent) GetOrigId() string {
	return fsev[ORIG_ID]
}
func (fsev FSEvent) GetSubject() string {
	return utils.FirstNonEmpty(fsev[SUBJECT], fsev[USERNAME])
}
func (fsev FSEvent) GetAccount() string {
	return utils.FirstNonEmpty(fsev[ACCOUNT], fsev[USERNAME])
}

// Charging destination number
func (fsev FSEvent) GetDestination() string {
	return utils.FirstNonEmpty(fsev[DESTINATION], fsev[CALL_DEST_NR])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSEvent) GetCallDestNr() string {
	return fsev[CALL_DEST_NR]
}
func (fsev FSEvent) GetTOR() string {
	return utils.FirstNonEmpty(fsev[TOR], cfg.DefaultTOR)
}
func (fsev FSEvent) GetUUID() string {
	return fsev[UUID]
}
func (fsev FSEvent) GetTenant() string {
	return utils.FirstNonEmpty(fsev[CSTMID], cfg.DefaultTenant)
}
func (fsev FSEvent) GetReqType() string {
	return utils.FirstNonEmpty(fsev[REQTYPE], cfg.DefaultReqType)
}
func (fsev FSEvent) MissingParameter() bool {
	return strings.TrimSpace(fsev.GetDirection()) == "" ||
		strings.TrimSpace(fsev.GetOrigId()) == "" ||
		strings.TrimSpace(fsev.GetSubject()) == "" ||
		strings.TrimSpace(fsev.GetAccount()) == "" ||
		strings.TrimSpace(fsev.GetDestination()) == "" ||
		strings.TrimSpace(fsev.GetTOR()) == "" ||
		strings.TrimSpace(fsev.GetUUID()) == "" ||
		strings.TrimSpace(fsev.GetTenant()) == "" ||
		strings.TrimSpace(fsev.GetCallDestNr()) == ""
}
func (fsev FSEvent) GetFallbackSubj() string {
	return cfg.DefaultSubject
}
func (fsev FSEvent) GetStartTime(field string) (t time.Time, err error) {
	st, err := strconv.ParseInt(fsev[field], 0, 64)
	t = time.Unix(0, st*1000)
	return
}

func (fsev FSEvent) GetEndTime() (t time.Time, err error) {
	st, err := strconv.ParseInt(fsev[END_TIME], 0, 64)
	t = time.Unix(0, st*1000)
	return
}
