/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package apis

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/utils"
)

// GetChargerProfile returns a Charger Profile
func (adms *AdminSv1) GetChargerProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ChargerProfile) error {
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
func (adms *AdminSv1) GetChargerProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, chPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ChargerProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = adms.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*chPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetChargerProfiles returns a list of charger profiles registered for a tenant
func (admS *AdminSv1) GetChargerProfiles(ctx *context.Context, args *utils.ArgsItemIDs, chrgPrfs *[]*utils.ChargerProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var chrgPrfIDs []string
	if err = admS.GetChargerProfileIDs(ctx, args, &chrgPrfIDs); err != nil {
		return
	}
	*chrgPrfs = make([]*utils.ChargerProfile, 0, len(chrgPrfIDs))
	for _, chrgPrfID := range chrgPrfIDs {
		var chgrPrf *utils.ChargerProfile
		chgrPrf, err = admS.dm.GetChargerProfile(ctx, tnt, chrgPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*chrgPrfs = append(*chrgPrfs, chgrPrf)
	}
	return
}

// GetChargerProfilesCount returns the total number of ChargerProfiles registered for a tenant
// returns ErrNotFound in case of 0 ChargerProfiles
func (admS *AdminSv1) GetChargerProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ChargerProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetChargerProfile add/update a new Charger Profile
func (adms *AdminSv1) SetChargerProfile(ctx *context.Context, arg *utils.ChargerProfileWithAPIOpts, reply *string) error {
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
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetChargerProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for ChargerProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheChargerProfiles,
		arg.TenantID(), utils.EmptyString, &arg.FilterIDs, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveChargerProfile remove a specific Charger Profile
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
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveChargerProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for ChargerProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheChargerProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewChargerSv1 initializes the ChargerSv1 object.
func NewChargerSv1(chgs *chargers.ChargerS) *ChargerSv1 {
	return &ChargerSv1{chgs: chgs}
}

// ChargerSv1 represents the RPC object to register for chargers v1 APIs.
type ChargerSv1 struct {
	chgs *chargers.ChargerS
}

// V1ProcessEvent will process the event received via API and return list of events forked
func (chgS *ChargerSv1) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*chargers.ChrgSProcessEventReply) (err error) {
	return chgS.chgs.V1ProcessEvent(ctx, args, reply)
}

// V1GetChargersForEvent exposes the list of ordered matching ChargingProfiles for an event
func (chgS *ChargerSv1) V1GetChargersForEvent(ctx *context.Context, args *utils.CGREvent, rply *[]*utils.ChargerProfile) (err error) {
	return chgS.chgs.V1GetChargersForEvent(ctx, args, rply)
}
