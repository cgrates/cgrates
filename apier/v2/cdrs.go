/*
Real-time Charging System for Telecom & ISP environments
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

package v2

import (
	"github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Retrieves CDRs based on the filters
func (apier *ApierV2) GetCdrs(attrs utils.RPCCDRsFilter, reply *[]*engine.ExternalCDR) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrs, _, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else if len(cdrs) == 0 {
		*reply = make([]*engine.ExternalCDR, 0)
	} else {
		for _, cdr := range cdrs {
			*reply = append(*reply, cdr.AsExternalCDR())
		}
	}
	return nil
}

func (apier *ApierV2) CountCdrs(attrs utils.RPCCDRsFilter, reply *int64) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrsFltr.Count = true
	if _, count, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = count
	}
	return nil
}

// Receive CDRs via RPC methods, not included with APIer because it has way less dependencies and can be standalone
type CdrsV2 struct {
	v1.CdrsV1
}
