/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetCallCost struct {
	CgrId string // Unique id of the CDR
	RunId string // Run Id
}

// Retrieves the callCost out of CGR logDb
func (apier *ApierV1) GetCallCostLog(attrs AttrGetCallCost, reply *engine.CallCost) error {
	if missing := utils.MissingStructFields(&attrs, []string{"CgrId", "RunId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if cc, err := apier.LogDb.GetCallCostLog(attrs.CgrId, "", attrs.RunId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if cc == nil {
		return fmt.Errorf("NOT_FOUND")
	} else {
		*reply = *cc
	}
	return nil
}

// Retrieves CDRs based on the filters
func (apier *ApierV1) GetCdrs(attrs utils.AttrGetCdrs, reply *[]*utils.CgrExtCdr) error {
	cdrsFltr, err := attrs.AsCdrsFilter()
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if cdrs, _, err := apier.CdrDb.GetStoredCdrs(cdrsFltr); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(cdrs) == 0 {
		*reply = make([]*utils.CgrExtCdr, 0)
	} else {
		for _, cdr := range cdrs {
			*reply = append(*reply, cdr.AsCgrExtCdr())
		}
	}
	return nil
}
