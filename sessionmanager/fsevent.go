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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
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
	Category           = "variable_cgr_tor"
	UUID               = "Unique-ID" // -Unique ID for this call leg
	CSTMID             = "variable_cgr_tenant"
	CALL_DEST_NR       = "Caller-Destination-Number"
	PARK_TIME          = "Caller-Profile-Created-Time"
	SETUP_TIME         = "Caller-Channel-Created-Time"
	ANSWER_TIME        = "Caller-Channel-Answered-Time"
	END_TIME           = "Caller-Channel-Hangup-Time"
	DURATION           = "billsec"
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
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	//TODO: implement direction
	return "*out"
}
func (fsev FSEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[SUBJECT], fsev[USERNAME])
}
func (fsev FSEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[ACCOUNT], fsev[USERNAME])
}

// Charging destination number
func (fsev FSEvent) GetDestination(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[DESTINATION], fsev[CALL_DEST_NR])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSEvent) GetCallDestNr(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CALL_DEST_NR])
}
func (fsev FSEvent) GetTOR(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[Category], config.CgrConfig().DefaultTOR)
}
func (fsev FSEvent) GetCgrId() string {
	setupTime, _ := fsev.GetSetupTime("")
	return utils.Sha1(fsev[UUID], setupTime.String())
}
func (fsev FSEvent) GetUUID() string {
	return fsev[UUID]
}
func (fsev FSEvent) GetTenant(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CSTMID], config.CgrConfig().DefaultTenant)
}
func (fsev FSEvent) GetReqType(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[REQTYPE], config.CgrConfig().DefaultReqType)
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
func (fsev FSEvent) GetSetupTime(fieldName string) (t time.Time, err error) {
	fsSTimeStr, hasKey := fsev[SETUP_TIME]
	if hasKey && fsSTimeStr != "0" {
		// Discard the nanoseconds information since MySQL cannot store them in early versions and csv uses default seconds so cgrid will not corelate
		fsSTimeStr = fsSTimeStr[:len(fsSTimeStr)-6]
	}
	sTimeStr := utils.FirstNonEmpty(fsev[fieldName], fsSTimeStr)
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(sTimeStr)
}
func (fsev FSEvent) GetAnswerTime(fieldName string) (t time.Time, err error) {
	fsATimeStr, hasKey := fsev[ANSWER_TIME]
	if hasKey && fsATimeStr != "0" {
		// Discard the nanoseconds information since MySQL cannot store them in early versions and csv uses default seconds so cgrid will not corelate
		fsATimeStr = fsATimeStr[:len(fsATimeStr)-6]
	}
	aTimeStr := utils.FirstNonEmpty(fsev[fieldName], fsATimeStr)
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(aTimeStr)
}

func (fsev FSEvent) GetEndTime() (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsev[END_TIME])
}

func (fsev FSEvent) GetDuration(fieldName string) (dur time.Duration, err error) {
	durStr := utils.FirstNonEmpty(fsev[fieldName], fsev[DURATION])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		durStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseDurationWithSecs(durStr)
}
