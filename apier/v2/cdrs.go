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
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Retrieves CDRs based on the filters
func (apier *ApierV2) GetCdrs(attrs utils.RpcCdrsFilter, reply *[]*utils.CgrCdrOut) error {
	cdrsFltr, err := attrs.AsCdrsFilter()
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if cdrs, _, err := apier.CdrDb.GetStoredCdrs(cdrsFltr); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(cdrs) == 0 {
		*reply = make([]*utils.CgrCdrOut, 0)
	} else {
		for _, cdr := range cdrs {
			*reply = append(*reply, cdr.AsCgrCdrOut())
		}
	}
	return nil
}
