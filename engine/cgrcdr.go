/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

func NewCgrCdrFromHttpReq(req *http.Request, timezone string) (CgrCdr, error) {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	cgrCdr := make(CgrCdr)
	cgrCdr[utils.Source] = req.RemoteAddr
	for k, vals := range req.Form {
		cgrCdr[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return cgrCdr, nil
}

type CgrCdr map[string]string

func (cgrCdr CgrCdr) getCGRID(timezone string) string {
	if CGRID, hasIt := cgrCdr[utils.CGRID]; hasIt {
		return CGRID
	}
	setupTime, _ := utils.ParseTimeDetectLayout(cgrCdr[utils.SetupTime], timezone)
	return utils.Sha1(cgrCdr[utils.OriginID], setupTime.UTC().String())
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

func (cgrCdr CgrCdr) AsCDR(timezone string) *CDR {
	storCdr := new(CDR)
	storCdr.CGRID = cgrCdr.getCGRID(timezone)
	storCdr.ToR = cgrCdr[utils.ToR]
	storCdr.OriginID = cgrCdr[utils.OriginID]
	storCdr.OriginHost = cgrCdr[utils.OriginHost]
	storCdr.Source = cgrCdr[utils.Source]
	storCdr.RequestType = cgrCdr[utils.RequestType]
	storCdr.Tenant = cgrCdr[utils.Tenant]
	storCdr.Category = cgrCdr[utils.Category]
	storCdr.Account = cgrCdr[utils.Account]
	storCdr.Subject = cgrCdr[utils.Subject]
	storCdr.Destination = cgrCdr[utils.Destination]
	storCdr.SetupTime, _ = utils.ParseTimeDetectLayout(cgrCdr[utils.SetupTime], timezone) // Not interested to process errors, should do them if necessary in a previous step
	storCdr.AnswerTime, _ = utils.ParseTimeDetectLayout(cgrCdr[utils.AnswerTime], timezone)
	storCdr.Usage, _ = utils.ParseDurationWithNanosecs(cgrCdr[utils.Usage])
	storCdr.ExtraFields = cgrCdr.getExtraFields()
	storCdr.Cost = -1
	if costStr, hasIt := cgrCdr[utils.COST]; hasIt {
		storCdr.Cost, _ = strconv.ParseFloat(costStr, 64)
	}
	if ratedStr, hasIt := cgrCdr[utils.RATED]; hasIt {
		storCdr.PreRated, _ = strconv.ParseBool(ratedStr)
	}
	return storCdr
}
