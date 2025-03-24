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
