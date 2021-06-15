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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetChargerProfile returns a Charger Profile
func (adms *AdminSv1) GetChargerProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.ChargerProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if cpp, err := adms.dm.GetChargerProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *cpp
	}
	return nil
}

// GetChargerProfileIDs returns list of chargerProfile IDs registered for a tenant
func (adms *AdminSv1) GetChargerProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, chPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ChargerProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := adms.dm.DataDB().GetKeysForPrefix(ctx, prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*chPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetChargerProfileCount returns the total number of ChargerProfiles registered for a tenant
// returns ErrNotFound in case of 0 ChargerProfiles
func (admS *AdminSv1) GetChargerProfileCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx,
		utils.ChargerProfilePrefix+tnt+utils.ConcatenatedKeySep); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

type ChargerWithAPIOpts struct {
	*engine.ChargerProfile
	APIOpts map[string]interface{}
}

//SetChargerProfile add/update a new Charger Profile
func (adms *AdminSv1) SetChargerProfile(ctx *context.Context, arg *ChargerWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.ChargerProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.SetChargerProfile(ctx, arg.ChargerProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheChargerProfiles and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheChargerProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ChargerProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheChargerProfiles,
		arg.TenantID(), &arg.FilterIDs, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveChargerProfile remove a specific Charger Profile
func (adms *AdminSv1) RemoveChargerProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveChargerProfile(ctx, tnt,
		arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheChargerProfiles and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheChargerProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ChargerProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheChargerProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewChargerSv1 returns the RPC Object for ChargerS
func NewChargerSv1(cS *engine.ChargerService) *ChargerSv1 {
	return &ChargerSv1{cS: cS}
}

// ChargerSv1 Exports RPC from ChargerS
type ChargerSv1 struct {
	ping
	cS *engine.ChargerService
}

// GetChargersForEvent returns matching ChargerProfile for Event
func (cSv1 *ChargerSv1) GetChargersForEvent(ctx *context.Context, cgrEv *utils.CGREvent,
	reply *engine.ChargerProfiles) error {
	return cSv1.cS.V1GetChargersForEvent(ctx, cgrEv, reply)
}

// ProcessEvent
func (cSv1 *ChargerSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *[]*engine.ChrgSProcessEventReply) error {
	return cSv1.cS.V1ProcessEvent(ctx, args, reply)
}
