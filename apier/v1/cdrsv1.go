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
// Deprecated
func (self *CdrsV1) ProcessCdr(cdr *engine.CDR, reply *string) error {
	return self.ProcessCDR(cdr, reply)
}

// Designed for CGR internal usage
func (self *CdrsV1) ProcessCDR(cdr *engine.CDR, reply *string) error {
	return self.CdrSrv.V1ProcessCDR(cdr, reply)
}

// Designed for external programs feeding CDRs to CGRateS
// Deprecated
func (self *CdrsV1) ProcessExternalCdr(cdr *engine.ExternalCDR, reply *string) error {
	return self.ProcessExternalCDR(cdr, reply)
}

// Designed for external programs feeding CDRs to CGRateS
func (self *CdrsV1) ProcessExternalCDR(cdr *engine.ExternalCDR, reply *string) error {
	if err := self.CdrSrv.ProcessExternalCdr(cdr); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Remotely (re)rating
// Deprecated
func (self *CdrsV1) RateCdrs(attrs utils.AttrRateCdrs, reply *string) error {
	return self.RateCDRs(attrs, reply)
}

// Remotely (re)rating
func (self *CdrsV1) RateCDRs(attrs utils.AttrRateCdrs, reply *string) error {
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

func (self *CdrsV1) StoreSMCost(attr engine.AttrCDRSStoreSMCost, reply *string) error {
	return self.CdrSrv.V1StoreSMCost(attr, reply)
}
