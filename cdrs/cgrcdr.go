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
	"github.com/cgrates/cgrates/utils"
	"net/http"
	"strconv"
	"time"
)

func NewCgrCdrFromHttpReq(req *http.Request) (CgrCdr, error) {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	cgrCdr := make(CgrCdr)
	for k, vals := range req.Form {
		cgrCdr[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return cgrCdr, nil
}

type CgrCdr map[string]string

func (cgrCdr CgrCdr) GetCgrId() string {
	return utils.FSCgrId(cgrCdr[ACCID])
}

func (cgrCdr CgrCdr) GetAccId() string {
	return cgrCdr[ACCID]
}

func (cgrCdr CgrCdr) GetCdrHost() string {
	return cgrCdr[CDRHOST]
}

func (cgrCdr CgrCdr) GetCdrSource() string {
	return cgrCdr[CDRSOURCE]
}
func (cgrCdr CgrCdr) GetDirection() string {
	//TODO: implement direction
	return "*out"
}
func (cgrCdr CgrCdr) GetOrigId() string {
	return cgrCdr[CDRHOST]
}
func (cgrCdr CgrCdr) GetSubject() string {
	return cgrCdr[SUBJECT]
}
func (cgrCdr CgrCdr) GetAccount() string {
	return cgrCdr[ACCOUNT]
}

// Charging destination number
func (cgrCdr CgrCdr) GetDestination() string {
	return cgrCdr[DESTINATION]
}

func (cgrCdr CgrCdr) GetTOR() string {
	return cgrCdr[TOR]
}

func (cgrCdr CgrCdr) GetTenant() string {
	return cgrCdr[TENANT]
}
func (cgrCdr CgrCdr) GetReqType() string {
	return cgrCdr[REQTYPE]
}
func (cgrCdr CgrCdr) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for k, v := range cgrCdr {
		if !utils.IsSliceMember(utils.PrimaryCdrFields, k) {
			extraFields[k] = v
		}
	}
	return extraFields
}
func (cgrCdr CgrCdr) GetAnswerTime() (t time.Time, err error) {
	return utils.ParseDate(cgrCdr[TIME_ANSWER])
}

// Extracts duration as considered by the telecom switch
func (cgrCdr CgrCdr) GetDuration() int64 {
	dur, _ := strconv.ParseInt(cgrCdr[DURATION], 0, 64)
	return dur
}
