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
	utils "github.com/cgrates/cgrates/cgrcoreutils"
	"github.com/cgrates/cgrates/config"
	"strconv"
	"time"
)

var cfg *config.CGRConfig // Share the configuration with the rest of the package

const (
	// Freswitch event proprities names
	DIRECTION    = "Call-Direction"
	ORIG_ID      = "variable_sip_call_id" //- originator_id - match cdrs
	SUBJECT      = "variable_cgr_subject"
	ACCOUNT      = "variable_cgr_account"
	DESTINATION  = "variable_cgr_destination"
	REQTYPE      = "variable_cgr_reqtype" //prepaid or postpaid
	TOR          = "variable_cgr_tor"
	UUID         = "Unique-ID" // -Unique ID for this call leg
	CSTMID       = "variable_cgr_cstmid"
	CALL_DEST_NR = "Caller-Destination-Number"
	PARK_TIME    = "Caller-Profile-Created-Time"
	START_TIME   = "Caller-Channel-Answered-Time"
	END_TIME     = "Caller-Channel-Hangup-Time"
	NAME         = "Event-Name"
	USERNAME     = "Caller-Username"
)

type FSCdr map[string]string

func (fsev FSCdr) New(body []byte) CDR {
	//fsev = fsock.FSCdrStrToMap(body, nil)
	return fsev
}

func (fsev FSCdr) GetName() string {
	return fsev[NAME]
}
func (fsev FSCdr) GetDirection() string {
	//TODO: implement direction
	return "OUT"
	//return fsev[DIRECTION]
}
func (fsev FSCdr) GetOrigId() string {
	return fsev[ORIG_ID]
}
func (fsev FSCdr) GetSubject() string {
	return utils.FirstNonEmpty(fsev[SUBJECT], fsev[USERNAME])
}
func (fsev FSCdr) GetAccount() string {
	return utils.FirstNonEmpty(fsev[ACCOUNT], fsev[USERNAME])
}

// Charging destination number
func (fsev FSCdr) GetDestination() string {
	return utils.FirstNonEmpty(fsev[DESTINATION], fsev[CALL_DEST_NR])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSCdr) GetCallDestNr() string {
	return fsev[CALL_DEST_NR]
}
func (fsev FSCdr) GetTOR() string {
	return utils.FirstNonEmpty(fsev[TOR], cfg.SMDefaultTOR)
}
func (fsev FSCdr) GetUUID() string {
	return fsev[UUID]
}
func (fsev FSCdr) GetTenant() string {
	return utils.FirstNonEmpty(fsev[CSTMID], cfg.SMDefaultTenant)
}
func (fsev FSCdr) GetReqType() string {
	return utils.FirstNonEmpty(fsev[REQTYPE], cfg.SMDefaultReqType)
}
func (fsev FSCdr) GetExtraParameters() string {
	return ""
}
func (fsev FSCdr) GetFallbackSubj() string {
	return cfg.SMDefaultSubject
}
func (fsev FSCdr) GetStartTime(field string) (t time.Time, err error) {
	st, err := strconv.ParseInt(fsev[field], 0, 64)
	t = time.Unix(0, st*1000)
	return
}

func (fsev FSCdr) GetEndTime() (t time.Time, err error) {
	st, err := strconv.ParseInt(fsev[END_TIME], 0, 64)
	t = time.Unix(0, st*1000)
	return
}
