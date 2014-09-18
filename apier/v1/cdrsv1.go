/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

// Receive CDRs via RPC methods
type CDRSV1 struct {
	CdrSrv *engine.CDRS
}

func (cdrsrv *CDRSV1) ProcessCdr(cdr *utils.StoredCdr, reply *string) error {
	if cdrsrv.CdrSrv == nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, "CDRS_NOT_RUNNING")
	}
	if err := cdrsrv.CdrSrv.ProcessCdr(cdr); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = utils.OK
	return nil
}
