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
	SUBJECT            = "variable_cgr_subject"
	ACCOUNT            = "variable_cgr_account"
	DESTINATION        = "variable_cgr_destination"
	REQTYPE            = "variable_cgr_reqtype" //prepaid or postpaid
	TOR                = "variable_cgr_tor"
	UUID               = "Unique-ID" // -Unique ID for this call leg
	CSTMID             = "variable_cgr_tenant"
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
func (fsev FSEvent) GetDirection(fieldName string) string {
	//TODO: implement direction
	return "*out"
	//return fsev[DIRECTION]
}
func (fsev FSEvent) GetSubject(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[SUBJECT], fsev[USERNAME])
}
func (fsev FSEvent) GetAccount(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[ACCOUNT], fsev[USERNAME])
}

// Charging destination number
func (fsev FSEvent) GetDestination(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[DESTINATION], fsev[CALL_DEST_NR])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSEvent) GetCallDestNr(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CALL_DEST_NR])
}
func (fsev FSEvent) GetTOR(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[TOR], cfg.DefaultTOR)
}
func (fsev FSEvent) GetUUID() string {
	return fsev[UUID]
}
func (fsev FSEvent) GetTenant(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CSTMID], cfg.DefaultTenant)
}
func (fsev FSEvent) GetReqType(fieldName string) string {
	return utils.FirstNonEmpty(fsev[fieldName], fsev[REQTYPE], cfg.DefaultReqType)
}
func (fsev FSEvent) MissingParameter() bool {
	return strings.TrimSpace(fsev.GetDirection("")) == "" ||
		strings.TrimSpace(fsev.GetSubject("")) == "" ||
		strings.TrimSpace(fsev.GetAccount("")) == "" ||
		strings.TrimSpace(fsev.GetDestination("")) == "" ||
		strings.TrimSpace(fsev.GetTOR("")) == "" ||
		strings.TrimSpace(fsev.GetUUID()) == "" ||
		strings.TrimSpace(fsev.GetTenant("")) == "" ||
		strings.TrimSpace(fsev.GetCallDestNr("")) == ""
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
