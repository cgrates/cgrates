/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

package engine

import (
	"github.com/cgrates/cgrates/utils"
	"net/http"
	"strconv"
)

func NewCgrCdrFromHttpReq(req *http.Request) (CgrCdr, error) {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	cgrCdr := make(CgrCdr)
	cgrCdr[utils.CDRHOST] = req.RemoteAddr
	for k, vals := range req.Form {
		cgrCdr[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return cgrCdr, nil
}

type CgrCdr map[string]string

func (cgrCdr CgrCdr) getCgrId() string {
	if cgrId, hasIt := cgrCdr[utils.CGRID]; hasIt {
		return cgrId
	}
	setupTime, _ := utils.ParseTimeDetectLayout(cgrCdr[utils.SETUP_TIME])
	return utils.Sha1(cgrCdr[utils.ACCID], setupTime.UTC().String())
}

func (cgrCdr CgrCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for k, v := range cgrCdr {
		if !utils.IsSliceMember(utils.PrimaryCdrFields, k) {
			extraFields[k] = v
		}
	}
	return extraFields
}

func (cgrCdr CgrCdr) AsStoredCdr() *StoredCdr {
	storCdr := new(StoredCdr)
	storCdr.CgrId = cgrCdr.getCgrId()
	storCdr.TOR = cgrCdr[utils.TOR]
	storCdr.AccId = cgrCdr[utils.ACCID]
	storCdr.CdrHost = cgrCdr[utils.CDRHOST]
	storCdr.CdrSource = cgrCdr[utils.CDRSOURCE]
	storCdr.ReqType = cgrCdr[utils.REQTYPE]
	storCdr.Direction = "*out"
	storCdr.Tenant = cgrCdr[utils.TENANT]
	storCdr.Category = cgrCdr[utils.CATEGORY]
	storCdr.Account = cgrCdr[utils.ACCOUNT]
	storCdr.Subject = cgrCdr[utils.SUBJECT]
	storCdr.Destination = cgrCdr[utils.DESTINATION]
	storCdr.SetupTime, _ = utils.ParseTimeDetectLayout(cgrCdr[utils.SETUP_TIME]) // Not interested to process errors, should do them if necessary in a previous step
	storCdr.Pdd, _ = utils.ParseDurationWithSecs(cgrCdr[utils.PDD])
	storCdr.AnswerTime, _ = utils.ParseTimeDetectLayout(cgrCdr[utils.ANSWER_TIME])
	storCdr.Usage, _ = utils.ParseDurationWithSecs(cgrCdr[utils.USAGE])
	storCdr.Supplier = cgrCdr[utils.SUPPLIER]
	storCdr.DisconnectCause = cgrCdr[utils.DISCONNECT_CAUSE]
	storCdr.ExtraFields = cgrCdr.getExtraFields()
	storCdr.Cost = -1
	if costStr, hasIt := cgrCdr[utils.COST]; hasIt {
		storCdr.Cost, _ = strconv.ParseFloat(costStr, 64)
	}
	if ratedStr, hasIt := cgrCdr[utils.RATED]; hasIt {
		storCdr.Rated, _ = strconv.ParseBool(ratedStr)
	}
	return storCdr
}
