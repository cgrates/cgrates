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

package cdrs

import (
	"encoding/json"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/rater"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
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
	CALL_DEST_NR       = "Caller-Destination-Number"
	PARK_TIME          = "Caller-Profile-Created-Time"
	START_TIME         = "Caller-Channel-Answered-Time"
	END_TIME           = "Caller-Channel-Hangup-Time"
	NAME               = "Event-Name"
	HEARTBEAT          = "HEARTBEAT"
	ANSWER             = "CHANNEL_ANSWER"
	HANGUP             = "CHANNEL_HANGUP_COMPLETE"
	PARK               = "CHANNEL_PARK"
	REQTYPE_PREPAID    = "prepaid"
	REQTYPE_POSTPAID   = "postpaid"
	AUTH_OK            = "+AUTH_OK"
	DISCONNECT         = "+SWITCH DISCONNECT"
	INSUFFICIENT_FUNDS = "-INSUFFICIENT_FUNDS"
	MISSING_PARAMETER  = "-MISSING_PARAMETER"
	SYSTEM_ERROR       = "-SYSTEM_ERROR"
	MANAGER_REQUEST    = "+MANAGER_REQUEST"
	USERNAME           = "Caller-Username"
)

var cfg *config.CGRConfig

// Returns first non empty string out of vals. Useful to extract defaults
func firstNonEmpty(vals ...string) string {
	for _, val := range vals {
		if len(val) != 0 {
			return val
		}
	}
	return ""
}

func GetName(vars map[string]string) string {
	return vars[NAME]
}
func GetDirection(vars map[string]string) string {
	//TODO: implement direction
	return "OUT"
	//return vars[DIRECTION]
}
func GetOrigId(vars map[string]string) string {
	return vars[ORIG_ID]
}
func GetSubject(vars map[string]string) string {
	return firstNonEmpty(vars[SUBJECT], vars[USERNAME])
}
func GetAccount(vars map[string]string) string {
	return firstNonEmpty(vars[ACCOUNT], vars[USERNAME])
}

// Charging destination number
func GetDestination(vars map[string]string) string {
	return firstNonEmpty(vars[DESTINATION], vars[CALL_DEST_NR])
}

// Original dialed destination number, useful in case of unpark
func GetCallDestNr(vars map[string]string) string {
	return vars[CALL_DEST_NR]
}
func GetTOR(vars map[string]string) string {
	return firstNonEmpty(vars[TOR], cfg.SMDefaultTOR)
}
func GetUUID(vars map[string]string) string {
	return vars[UUID]
}
func GetTenant(vars map[string]string) string {
	return firstNonEmpty(vars[CSTMID], cfg.SMDefaultTenant)
}
func GetReqType(vars map[string]string) string {
	return firstNonEmpty(vars[REQTYPE], cfg.SMDefaultReqType)
}
func GetFallbackSubj(vars map[string]string) string {
	return cfg.SMDefaultSubject
}
func GetStartTime(vars map[string]string, field string) (t time.Time, err error) {
	st, err := strconv.ParseInt(vars[field], 0, 64)
	t = time.Unix(0, st*1000)
	return
}

func GetEndTime() (vars map[string]string, t time.Time, err error) {
	st, err := strconv.ParseInt(vars[END_TIME], 0, 64)
	t = time.Unix(0, st*1000)
	return
}

type CDR struct {
	Variables map[string]string
}

func cdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	cdr := CDR{}
	if err := json.Unmarshal(body, &cdr); err == nil {

	} else {
		rater.Logger.Err(fmt.Sprintf("CDRCAPTOR: Could not unmarshal cdr: %v", err))
	}
}

func startCaptiuringCDRs() {
	http.HandleFunc("/cdr", cdrHandler)
	http.ListenAndServe(":8080", nil)
}
