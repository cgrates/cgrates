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
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

func NewCgrCdrFromHttpReq(req *http.Request) (CgrCdr, error) {
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

func (cgrCdr CgrCdr) getCGRID() string {
	if CGRID, hasIt := cgrCdr[utils.CGRID]; hasIt {
		return CGRID
	}
	return utils.Sha1(cgrCdr[utils.OriginID], cgrCdr[utils.OriginHost])
}

func (cgrCdr CgrCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for k, v := range cgrCdr {
		if !utils.MainCDRFields.Has(k) {
			extraFields[k] = v
		}
	}
	return extraFields
}

func (cgrCdr CgrCdr) AsCDR(timezone string) (storCdr *CDR, err error) {
	storCdr = &CDR{
		CGRID:       cgrCdr.getCGRID(),
		RunID:       cgrCdr[utils.RunID],
		OriginHost:  cgrCdr[utils.OriginHost],
		Source:      cgrCdr[utils.Source],
		OriginID:    cgrCdr[utils.OriginID],
		ToR:         cgrCdr[utils.ToR],
		RequestType: cgrCdr[utils.RequestType],
		Tenant:      cgrCdr[utils.Tenant],
		Category:    cgrCdr[utils.Category],
		Account:     cgrCdr[utils.Account],
		Subject:     cgrCdr[utils.Subject],
		Destination: cgrCdr[utils.Destination],
		ExtraFields: cgrCdr.getExtraFields(),
		ExtraInfo:   cgrCdr[utils.ExtraInfo],
		CostSource:  cgrCdr[utils.CostSource],
	}
	if orderID, hasIt := cgrCdr[utils.OrderID]; hasIt {
		if storCdr.OrderID, err = strconv.ParseInt(orderID, 10, 64); err != nil {
			return nil, err
		}
	}
	storCdr.SetupTime, err = utils.ParseTimeDetectLayout(cgrCdr[utils.SetupTime], timezone) // Not interested to process errors, should do them if necessary in a previous step
	if err != nil {
		return nil, err
	}
	storCdr.AnswerTime, err = utils.ParseTimeDetectLayout(cgrCdr[utils.AnswerTime], timezone)
	if err != nil {
		return nil, err
	}
	storCdr.Usage, err = utils.ParseDurationWithNanosecs(cgrCdr[utils.Usage])
	if err != nil {
		return nil, err
	}
	if partial, hasIt := cgrCdr[utils.Partial]; hasIt {
		if storCdr.Partial, err = strconv.ParseBool(partial); err != nil {
			return nil, err
		}
	}
	if ratedStr, hasIt := cgrCdr[utils.PreRated]; hasIt {
		if storCdr.PreRated, err = strconv.ParseBool(ratedStr); err != nil {
			return nil, err
		}
	}
	storCdr.Cost = -1
	if costStr, hasIt := cgrCdr[utils.COST]; hasIt {
		if storCdr.Cost, err = strconv.ParseFloat(costStr, 64); err != nil {
			return nil, err
		}
	}
	if costDetails, hasIt := cgrCdr[utils.CostDetails]; hasIt {
		if err = json.Unmarshal([]byte(costDetails), &storCdr.CostDetails); err != nil {
			return nil, err
		}
	}
	return
}
