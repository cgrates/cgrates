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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Retrieves the callCost out of CGR logDb
func (apierSv1 *APIerSv1) GetEventCost(ctx *context.Context, attrs *utils.AttrGetCallCost, reply *engine.EventCost) error {
	if attrs.CgrId == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("CgrId")
	}
	if attrs.RunId == utils.EmptyString {
		attrs.RunId = utils.MetaDefault
	}
	cdrFltr := &utils.CDRsFilter{
		CGRIDs: []string{attrs.CgrId},
		RunIDs: []string{attrs.RunId},
	}
	if cdrs, _, err := apierSv1.CdrDb.GetCDRs(cdrFltr, false); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	} else if len(cdrs) == 0 ||
		cdrs[0].CostDetails == nil { // to avoid nil pointer dereference
		return utils.ErrNotFound
	} else {
		*reply = *cdrs[0].CostDetails
	}
	return nil
}

// Retrieves CDRs based on the filters
func (apierSv1 *APIerSv1) GetCDRs(ctx *context.Context, attrs *utils.AttrGetCdrs, reply *[]*engine.ExternalCDR) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrs, _, err := apierSv1.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
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

// New way of removing CDRs
func (apierSv1 *APIerSv1) RemoveCDRs(ctx *context.Context, attrs *utils.RPCCDRsFilter, reply *string) error {
	cdrsFilter, err := attrs.AsCDRsFilter(apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, _, err := apierSv1.CdrDb.GetCDRs(cdrsFilter, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func NewCDRsV1(CDRs *engine.CDRServer) *CDRsV1 {
	return &CDRsV1{CDRs: CDRs}
}

// Receive CDRs via RPC methods
type CDRsV1 struct {
	CDRs *engine.CDRServer
}

// ProcessCDR will process a CDR in CGRateS internal format
func (cdrSv1 *CDRsV1) ProcessCDR(ctx *context.Context, cdr *engine.CDRWithAPIOpts, reply *string) error {
	return cdrSv1.CDRs.V1ProcessCDR(ctx, cdr, reply)
}

// ProcessEvent will process an Event based on the flags attached
func (cdrSv1 *CDRsV1) ProcessEvent(ctx *context.Context, arg *engine.ArgV1ProcessEvent, reply *string) error {
	return cdrSv1.CDRs.V1ProcessEvent(ctx, arg, reply)
}

// ProcessEvent will process multiple events Event based on the flags attached.
func (cdrSv1 *CDRsV1) ProcessEvents(ctx *context.Context, arg *engine.ArgV1ProcessEvents, reply *string) error {
	return cdrSv1.CDRs.V1ProcessEvents(ctx, arg, reply)
}

// ProcessExternalCDR will process a CDR in external format
func (cdrSv1 *CDRsV1) ProcessExternalCDR(ctx *context.Context, cdr *engine.ExternalCDRWithAPIOpts, reply *string) error {
	return cdrSv1.CDRs.V1ProcessExternalCDR(ctx, cdr, reply)
}

// RateCDRs can re-/rate remotely CDRs
func (cdrSv1 *CDRsV1) RateCDRs(ctx *context.Context, arg *engine.ArgRateCDRs, reply *string) error {
	return cdrSv1.CDRs.V1RateCDRs(ctx, arg, reply)
}

// ReprocessCDRs can reprocess remotely CDRs
func (cdrSv1 *CDRsV1) ReprocessCDRs(ctx *context.Context, arg *engine.ArgRateCDRs, reply *string) error {
	return cdrSv1.CDRs.ReprocessCDRs(ctx, arg, reply)
}

// StoreSMCost will store
func (cdrSv1 *CDRsV1) StoreSessionCost(ctx *context.Context, attr *engine.AttrCDRSStoreSMCost, reply *string) error {
	return cdrSv1.CDRs.V1StoreSessionCost(ctx, attr, reply)
}

func (cdrSv1 *CDRsV1) GetCDRsCount(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *int64) error {
	return cdrSv1.CDRs.V1GetCDRsCount(ctx, args, reply)
}

func (cdrSv1 *CDRsV1) GetCDRs(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *[]*engine.CDR) error {
	return cdrSv1.CDRs.V1GetCDRs(ctx, *args, reply)
}
