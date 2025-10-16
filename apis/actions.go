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
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/utils"
)

// GetActionProfile returns an Action Profile
func (admS *AdminSv1) GetActionProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ActionProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	ap, err := admS.dm.GetActionProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ap
	return nil
}

// GetActionProfileIDs returns list of action profile IDs registered for a tenant
func (admS *AdminSv1) GetActionProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, actPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaActionProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = dataDB.GetKeysForPrefix(ctx, prfx); err != nil {
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
	*actPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetActionProfiles returns a list of action profiles registered for a tenant
func (admS *AdminSv1) GetActionProfiles(ctx *context.Context, args *utils.ArgsItemIDs, actPrfs *[]*utils.ActionProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var actPrfIDs []string
	if err = admS.GetActionProfileIDs(ctx, args, &actPrfIDs); err != nil {
		return
	}
	*actPrfs = make([]*utils.ActionProfile, 0, len(actPrfIDs))
	for _, actPrfID := range actPrfIDs {
		var ap *utils.ActionProfile
		ap, err = admS.dm.GetActionProfile(ctx, tnt, actPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*actPrfs = append(*actPrfs, ap)
	}
	return
}

// GetActionProfilesCount sets in reply var the total number of ActionProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 ActionProfileIDs
func (admS *AdminSv1) GetActionProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaActionProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = dataDB.GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetActionProfile add/update a new Action Profile
func (admS *AdminSv1) SetActionProfile(ctx *context.Context, ap *utils.ActionProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(ap.ActionProfile, []string{utils.ID, utils.Actions}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for i := range ap.ActionProfile.Actions {
		if ap.ActionProfile.Actions[i] == nil {
			continue
		}
		for j := range ap.ActionProfile.Actions[i].Diktats { // if there are diktats, make sure their ID exists
			if missing := utils.MissingStructFields(ap.ActionProfile.Actions[i].
				Diktats[j], []string{utils.ID}); len(missing) != 0 {
				return utils.NewErrMandatoryIeMissing(missing...)
			}
		}
	}
	if ap.Tenant == utils.EmptyString {
		ap.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}

	if err := admS.dm.SetActionProfile(ctx, ap.ActionProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheActionProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheActionProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetActionProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(ap.APIOpts[utils.MetaCache]), ap.Tenant, utils.CacheActionProfiles,
		ap.TenantID(), utils.EmptyString, &ap.FilterIDs, ap.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveActionProfile remove a specific Action Profile
func (admS *AdminSv1) RemoveActionProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveActionProfile(ctx, tnt, arg.ID,
		true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheActionProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheActionProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveActionProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheActionProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewActionSv1 initializes the ActionSv1 object.
func NewActionSv1(acts *actions.ActionS) *ActionSv1 {
	return &ActionSv1{acts: acts}
}

// ActionSv1 represents the RPC object to register for actions v1 APIs.
type ActionSv1 struct {
	acts *actions.ActionS
}

// ScheduleActions will be called to schedule actions matching the arguments
func (aS *ActionSv1) ScheduleActions(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	return aS.acts.V1ScheduleActions(ctx, args, rpl)
}

// ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aS *ActionSv1) ExecuteActions(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	return aS.acts.V1ExecuteActions(ctx, args, rpl)
}
