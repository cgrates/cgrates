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
	// Freswitch event property names
	FS_CDR_MAP      = "variables"
	FS_DIRECTION    = "direction"
	FS_SUBJECT      = "cgr_subject"
	FS_ACCOUNT      = "cgr_account"
	FS_DESTINATION  = "cgr_destination"
	FS_REQTYPE      = "cgr_reqtype" //prepaid or postpaid
	FS_TOR          = "cgr_tor"
	FS_UUID         = "uuid" // -Unique ID for this call leg
	FS_CSTMID       = "cgr_tenant"
	FS_CALL_DEST_NR = "dialed_extension"
	FS_PARK_TIME    = "start_epoch"
	FS_ANSWER_TIME  = "answer_epoch"
	FS_HANGUP_TIME  = "end_epoch"
	FS_DURATION     = "billsec"
	FS_USERNAME     = "user_name"
	FS_IP           = "sip_local_network_addr"
)

type FSCdr map[string]string

func (fsCdr FSCdr) New(body []byte) (utils.CDR, error) {
	fsCdr = make(map[string]string)
	var tmp map[string]interface{}
	var err error
	if err = json.Unmarshal(body, &tmp); err == nil {
		if variables, ok := tmp[FS_CDR_MAP]; ok {
			if variables, ok := variables.(map[string]interface{}); ok {
				for k, v := range variables {
					fsCdr[k] = v.(string)
				}
			}
			return fsCdr, nil
		}
	}
	return nil, err
}

func (fsCdr FSCdr) GetCgrId() string {
	return utils.FSCgrId(fsCdr[FS_UUID])
}
func (fsCdr FSCdr) GetAccId() string {
	return fsCdr[FS_UUID]
}
func (fsCdr FSCdr) GetCdrHost() string {
	return fsCdr[FS_IP]
}
func (fsCdr FSCdr) GetDirection() string {
	//TODO: implement direction, not related to FS_DIRECTION but traffic towards or from subject/account
	return "*out"
}
func (fsCdr FSCdr) GetSubject() string {
	return utils.FirstNonEmpty(fsCdr[FS_SUBJECT], fsCdr[FS_USERNAME])
}
func (fsCdr FSCdr) GetAccount() string {
	return utils.FirstNonEmpty(fsCdr[FS_ACCOUNT], fsCdr[FS_USERNAME])
}

// Charging destination number
func (fsCdr FSCdr) GetDestination() string {
	return utils.FirstNonEmpty(fsCdr[FS_DESTINATION], fsCdr[FS_CALL_DEST_NR])
}

func (fsCdr FSCdr) GetTOR() string {
	return utils.FirstNonEmpty(fsCdr[FS_TOR], cfg.DefaultTOR)
}

func (fsCdr FSCdr) GetTenant() string {
	return utils.FirstNonEmpty(fsCdr[FS_CSTMID], cfg.DefaultTenant)
}
func (fsCdr FSCdr) GetReqType() string {
	return utils.FirstNonEmpty(fsCdr[FS_REQTYPE], cfg.DefaultReqType)
}
func (fsCdr FSCdr) GetExtraFields() map[string]string {
	extraFields := make(map[string]string, len(cfg.CDRSExtraFields))
	for _, field := range cfg.CDRSExtraFields {
		extraFields[field] = fsCdr[field]
	}
	return extraFields
}
func (fsCdr FSCdr) GetAnswerTime() (t time.Time, err error) {
	//ToDo: Make sure we work with UTC instead of local time
	at, err := strconv.ParseInt(fsCdr[FS_ANSWER_TIME], 0, 64)
	t = time.Unix(at, 0)
	return
}
func (fsCdr FSCdr) GetHangupTime() (t time.Time, err error) {
	hupt, err := strconv.ParseInt(fsCdr[FS_HANGUP_TIME], 0, 64)
	t = time.Unix(hupt, 0)
	return
}

// Extracts duration as considered by the telecom switch
func (fsCdr FSCdr) GetDuration() int64 {
	dur, _ := strconv.ParseInt(fsCdr[FS_DURATION], 0, 64)
	return dur
}

func (fsCdr FSCdr) Store() (result string, err error) {
	result += fsCdr.GetCgrId() + "|"
	result += fsCdr.GetAccId() + "|"
	result += fsCdr.GetCdrHost() + "|"
	result += fsCdr.GetDirection() + "|"
	result += fsCdr.GetSubject() + "|"
	result += fsCdr.GetAccount() + "|"
	result += fsCdr.GetDestination() + "|"
	result += fsCdr.GetTOR() + "|"
	result += fsCdr.GetAccId() + "|"
	result += fsCdr.GetTenant() + "|"
	result += fsCdr.GetReqType() + "|"
	st, err := fsCdr.GetAnswerTime()
	if err != nil {
		return "", err
	}
	result += strconv.FormatInt(st.UnixNano(), 10) + "|"
	et, err := fsCdr.GetHangupTime()
	if err != nil {
		return "", err
	}
	result += strconv.FormatInt(et.UnixNano(), 10) + "|"
	result += strconv.FormatInt(fsCdr.GetDuration(), 10) + "|"
	return
}

func (fsCdr FSCdr) Restore(input string) error {
	return errors.New("Not implemented")
}
