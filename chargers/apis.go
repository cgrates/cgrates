// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package chargers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// V1ProcessEvent will process the event received via API and return list of events forked
func (cS *ChargerS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *[]*ChrgSProcessEventReply) (err error) {
	if args == nil ||
		args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = cS.cfg.GeneralCfg().DefaultTenant
	}
	rply, err := cS.processEvent(ctx, tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = rply
	return
}

// V1GetChargersForEvent exposes the list of ordered matching ChargingProfiles for an event
func (cS *ChargerS) V1GetChargersForEvent(ctx *context.Context, args *utils.CGREvent,
	rply *[]*utils.ChargerProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = cS.cfg.GeneralCfg().DefaultTenant
	}
	cPs, err := cS.matchingChargerProfilesForEvent(ctx, tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*rply = cPs
	return
}
