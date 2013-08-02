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
	"errors"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"time"
)

const (
	CDR_MAP      = "variables"
	DIRECTION    = "direction"
	ORIG_ID      = "id"
	SUBJECT      = "subject"
	ACCOUNT      = "account"
	DESTINATION  = "destination"
	REQTYPE      = "reqtype" //prepaid or postpaid
	TOR          = "tor"
	UUID         = "uuid" // -Unique ID for this call leg
	CSTMID       = "tenant"
	CALL_DEST_NR = "dialed_extension"
	PARK_TIME    = "start_epoch"
	ANSWER_TIME  = "time_answer"
	HANGUP_TIME  = "time_hangup"
	DURATION     = "duration"
	USERNAME     = "user_name"
	IP           = "sip_local_network_addr"
)

type GenCdr map[string]string

func (genCdr GenCdr) New(body []byte) (utils.CDR, error) {
	genCdr = make(map[string]string)
	var tmp map[string]interface{}
	var err error
	if err = json.Unmarshal(body, &tmp); err == nil {
		if variables, ok := tmp[CDR_MAP]; ok {
			if variables, ok := variables.(map[string]interface{}); ok {
				for k, v := range variables {
					genCdr[k] = v.(string)
				}
			}
			return genCdr, nil
		}
	}
	return nil, err
}

func (genCdr GenCdr) GetCgrId() string {
	return utils.FSCgrId(genCdr[UUID])
}
func (genCdr GenCdr) GetAccId() string {
	return genCdr[UUID]
}
func (genCdr GenCdr) GetCdrHost() string {
	return genCdr[FS_IP]
}
func (genCdr GenCdr) GetDirection() string {
	//TODO: implement direction
	return "*out"
}
func (genCdr GenCdr) GetOrigId() string {
	return genCdr[ORIG_ID]
}
func (genCdr GenCdr) GetSubject() string {
	return utils.FirstNonEmpty(genCdr[SUBJECT], genCdr[USERNAME])
}
func (genCdr GenCdr) GetAccount() string {
	return utils.FirstNonEmpty(genCdr[ACCOUNT], genCdr[USERNAME])
}

// Charging destination number
func (genCdr GenCdr) GetDestination() string {
	return utils.FirstNonEmpty(genCdr[DESTINATION], genCdr[CALL_DEST_NR])
}

func (genCdr GenCdr) GetTOR() string {
	return utils.FirstNonEmpty(genCdr[TOR], cfg.DefaultTOR)
}

func (genCdr GenCdr) GetTenant() string {
	return utils.FirstNonEmpty(genCdr[CSTMID], cfg.DefaultTenant)
}
func (genCdr GenCdr) GetReqType() string {
	return utils.FirstNonEmpty(genCdr[REQTYPE], cfg.DefaultReqType)
}
func (genCdr GenCdr) GetExtraFields() map[string]string {
	extraFields := make(map[string]string, len(cfg.CDRSExtraFields))
	for _, field := range cfg.CDRSExtraFields {
		extraFields[field] = genCdr[field]
	}
	return extraFields
}
func (genCdr GenCdr) GetFallbackSubj() string {
	return cfg.DefaultSubject
}
func (genCdr GenCdr) GetAnswerTime() (t time.Time, err error) {
	st, err := strconv.ParseInt(genCdr[ANSWER_TIME], 0, 64)
	t = time.Unix(0, st*1000)
	return
}
func (genCdr GenCdr) GetHangupTime() (t time.Time, err error) {
	st, err := strconv.ParseInt(genCdr[HANGUP_TIME], 0, 64)
	t = time.Unix(0, st*1000)
	return
}

// Extracts duration as considered by the telecom switch
func (genCdr GenCdr) GetDuration() int64 {
	dur, _ := strconv.ParseInt(genCdr[DURATION], 0, 64)
	return dur
}

func (genCdr GenCdr) Store() (result string, err error) {
	result += genCdr.GetCgrId() + "|"
	result += genCdr.GetAccId() + "|"
	result += genCdr.GetCdrHost() + "|"
	result += genCdr.GetDirection() + "|"
	result += genCdr.GetOrigId() + "|"
	result += genCdr.GetSubject() + "|"
	result += genCdr.GetAccount() + "|"
	result += genCdr.GetDestination() + "|"
	result += genCdr.GetTOR() + "|"
	result += genCdr.GetAccId() + "|"
	result += genCdr.GetTenant() + "|"
	result += genCdr.GetReqType() + "|"
	st, err := genCdr.GetAnswerTime()
	if err != nil {
		return "", err
	}
	result += strconv.FormatInt(st.UnixNano(), 10) + "|"
	et, err := genCdr.GetHangupTime()
	if err != nil {
		return "", err
	}
	result += strconv.FormatInt(et.UnixNano(), 10) + "|"
	result += strconv.FormatInt(genCdr.GetDuration(), 10) + "|"
	result += genCdr.GetFallbackSubj() + "|"
	return
}

func (genCdr GenCdr) Restore(input string) error {
	return errors.New("Not implemented")
}
