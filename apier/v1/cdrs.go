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
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Retrieves the callCost out of CGR logDb
func (apier *ApierV1) GetEventCost(attrs utils.AttrGetCallCost, reply *engine.EventCost) error {
	if attrs.CgrId == "" {
		return utils.NewErrMandatoryIeMissing("CgrId")
	}
	if attrs.RunId == "" {
		attrs.RunId = utils.META_DEFAULT
	}
	cdrFltr := &utils.CDRsFilter{
		CGRIDs: []string{attrs.CgrId},
		RunIDs: []string{attrs.RunId},
	}
	if cdrs, _, err := apier.CdrDb.GetCDRs(cdrFltr, false); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	} else if len(cdrs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *cdrs[0].CostDetails
	}
	return nil
}

// Retrieves CDRs based on the filters
func (apier *ApierV1) GetCdrs(attrs utils.AttrGetCdrs, reply *[]*engine.ExternalCDR) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrs, _, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return err
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
	*reply = utils.OK
	return nil
}

// New way of removing CDRs
func (apier *ApierV1) RemoveCDRs(attrs utils.RPCCDRsFilter, reply *string) error {
	cdrsFilter, err := attrs.AsCDRsFilter(apier.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, _, err := apier.CdrDb.GetCDRs(cdrsFilter, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// New way of (re-)rating CDRs
func (apier *ApierV1) RateCDRs(attrs utils.AttrRateCDRs, reply *string) error {
	if apier.CDRs == nil {
		return errors.New("CDRS_NOT_ENABLED")
	}
	return apier.CDRs.Call("CDRsV1.RateCDRs", attrs, reply)
}
