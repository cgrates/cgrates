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

// Retrieves the callCost out of CGR logDb
func (apier *ApierV1) GetCallCostLog(attrs utils.AttrGetCallCost, reply *engine.SMCost) error {
	if attrs.CgrId == "" {
		return utils.NewErrMandatoryIeMissing("CgrId")
	}
	if attrs.RunId == "" {
		attrs.RunId = utils.META_DEFAULT
	}
	if smc, err := apier.CdrDb.GetCallCostLog(attrs.CgrId, attrs.RunId); err != nil {
		return utils.NewErrServerError(err)
	} else if smc == nil {
		return utils.ErrNotFound
	} else {
		*reply = *smc
	}
	return nil
}

// Retrieves CDRs based on the filters
func (apier *ApierV1) GetCdrs(attrs utils.AttrGetCdrs, reply *[]*engine.ExternalCDR) error {
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

// Remove Cdrs out of CDR storage
func (apier *ApierV1) RemCdrs(attrs utils.AttrRemCdrs, reply *string) error {
	if len(attrs.CgrIds) == 0 {
		return fmt.Errorf("%s:CgrIds", utils.ErrMandatoryIeMissing.Error())
	}
	if _, _, err := apier.CdrDb.GetCDRs(&utils.CDRsFilter{CGRIDs: attrs.CgrIds}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

// New way of removing CDRs
func (apier *ApierV1) RemoveCDRs(attrs utils.RPCCDRsFilter, reply *string) error {
	cdrsFilter, err := attrs.AsCDRsFilter(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, _, err := apier.CdrDb.GetCDRs(cdrsFilter, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}
