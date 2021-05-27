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
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetActionProfile returns an Action Profile
func (admS *AdminSv1) GetActionProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.ActionProfile) error {
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
func (admS *AdminSv1) GetActionProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, actPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := admS.dm.DataDB().GetKeysForPrefix(ctx, prfx)
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
	*actPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetActionProfileCount sets in reply var the total number of ActionProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 ActionProfileIDs
func (admS *AdminSv1) GetActionProfileCount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetActionProfile add/update a new Action Profile
func (admS *AdminSv1) SetActionProfile(ctx *context.Context, ap *engine.ActionProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(ap.ActionProfile, []string{utils.ID, utils.Actions}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
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
	if err := admS.CallCache(ctx, utils.IfaceAsString(ap.APIOpts[utils.CacheOpt]), ap.Tenant, utils.CacheActionProfiles,
		ap.TenantID(), &ap.FilterIDs, nil, ap.APIOpts); err != nil {
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
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheActionProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewActionSv1 initializes ActionSv1
func NewActionSv1(aS *actions.ActionS) *ActionSv1 {
	return &ActionSv1{aS: aS}
}

// ActionSv1 exports RPC from RLs
type ActionSv1 struct {
	aS *actions.ActionS
	ping
}

// ScheduleActions will be called to schedule actions matching the arguments
func (aSv1 *ActionSv1) ScheduleActions(ctx *context.Context, args *utils.ArgActionSv1ScheduleActions, rpl *string) error {
	return aSv1.aS.V1ScheduleActions(ctx, args, rpl)
}

// ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aSv1 *ActionSv1) ExecuteActions(ctx *context.Context, args *utils.ArgActionSv1ScheduleActions, rpl *string) error {
	return aSv1.aS.V1ExecuteActions(ctx, args, rpl)
}
