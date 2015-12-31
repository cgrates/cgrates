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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Receive CDRs via RPC methods
type CdrsV1 struct {
	CdrSrv *engine.CdrServer
}

// Designed for CGR internal usage
func (self *CdrsV1) ProcessCdr(cdr *engine.CDR, reply *string) error {
	if err := self.CdrSrv.ProcessCdr(cdr); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Designed for external programs feeding CDRs to CGRateS
func (self *CdrsV1) ProcessExternalCdr(cdr *engine.ExternalCDR, reply *string) error {
	if err := self.CdrSrv.ProcessExternalCdr(cdr); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Remotely start mediation with specific runid, runs asynchronously, it's status will be displayed in syslog
func (self *CdrsV1) RateCdrs(attrs utils.AttrRateCdrs, reply *string) error {
	cdrsFltr, err := attrs.AsCDRsFilter(self.CdrSrv.Timezone())
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := self.CdrSrv.RateCDRs(cdrsFltr, attrs.SendToStats); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (self *CdrsV1) LogCallCost(ccl *engine.CallCostLog, reply *string) error {
	if err := self.CdrSrv.LogCallCost(ccl); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
