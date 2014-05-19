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

package utils

import (
	"net/http"
)

func NewCgrCdrFromHttpReq(req *http.Request) (CgrCdr, error) {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	cgrCdr := make(CgrCdr)
	cgrCdr[CDRHOST] = req.RemoteAddr
	for k, vals := range req.Form {
		cgrCdr[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return cgrCdr, nil
}

type CgrCdr map[string]string

func (cgrCdr CgrCdr) getCgrId() string {
	setupTime, _ := ParseTimeDetectLayout(cgrCdr[SETUP_TIME])
	return Sha1(cgrCdr[ACCID], setupTime.String())
}

func (cgrCdr CgrCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for k, v := range cgrCdr {
		if !IsSliceMember(PrimaryCdrFields, k) {
			extraFields[k] = v
		}
	}
	return extraFields
}

func (cgrCdr CgrCdr) AsStoredCdr() *StoredCdr {
	storCdr := new(StoredCdr)
	storCdr.CgrId = cgrCdr.getCgrId()
	storCdr.TOR = cgrCdr[TOR]
	storCdr.AccId = cgrCdr[ACCID]
	storCdr.CdrHost = cgrCdr[CDRHOST]
	storCdr.CdrSource = cgrCdr[CDRSOURCE]
	storCdr.ReqType = cgrCdr[REQTYPE]
	storCdr.Direction = "*out"
	storCdr.Tenant = cgrCdr[TENANT]
	storCdr.Category = cgrCdr[CATEGORY]
	storCdr.Account = cgrCdr[ACCOUNT]
	storCdr.Subject = cgrCdr[SUBJECT]
	storCdr.Destination = cgrCdr[DESTINATION]
	storCdr.SetupTime, _ = ParseTimeDetectLayout(cgrCdr[SETUP_TIME]) // Not interested to process errors, should do them if necessary in a previous step
	storCdr.AnswerTime, _ = ParseTimeDetectLayout(cgrCdr[ANSWER_TIME])
	storCdr.Usage, _ = ParseDurationWithNanosecs(cgrCdr[USAGE])
	storCdr.ExtraFields = cgrCdr.getExtraFields()
	storCdr.Cost = -1
	return storCdr
}
