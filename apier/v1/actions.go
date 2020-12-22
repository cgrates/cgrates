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
	"time"

	"github.com/cgrates/cgrates/actions"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetActionProfile returns an Action Profile
func (apierSv1 *APIerSv1) GetActionProfile(arg *utils.TenantIDWithOpts, reply *engine.ActionProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ap, err := apierSv1.DataManager.GetActionProfile(tnt, arg.ID, true, true, utils.NonTransactional)
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
func (apierSv1 *APIerSv1) GetActionProfileIDs(args *utils.PaginatorWithTenant, actPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.ActionProfilePrefix + tnt + utils.CONCATENATED_KEY_SEP
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
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

// GetActionProfileIDsCount sets in reply var the total number of ActionProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 ActionProfileIDs
func (apierSv1 *APIerSv1) GetActionProfileIDsCount(args *utils.TenantWithOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.ActionProfilePrefix + tnt + utils.CONCATENATED_KEY_SEP
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

type ActionProfileWithCache struct {
	*engine.ActionProfileWithOpts
	Cache *string
}

//SetActionProfile add/update a new Action Profile
func (apierSv1 *APIerSv1) SetActionProfile(ap *ActionProfileWithCache, reply *string) error {
	if missing := utils.MissingStructFields(ap.ActionProfile, []string{utils.ID, utils.Actions}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ap.Tenant == utils.EmptyString {
		ap.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}

	if err := apierSv1.DataManager.SetActionProfile(ap.ActionProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheActionProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActionProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(ap.Cache, ap.Tenant, utils.CacheActionProfiles,
		ap.TenantID(), &ap.FilterIDs, nil, ap.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveActionProfile remove a specific Action Profile
func (apierSv1 *APIerSv1) RemoveActionProfile(arg *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveActionProfile(tnt, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheActionProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActionProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(arg.Cache, tnt, utils.CacheActionProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.Opts); err != nil {
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
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (aSv1 *ActionSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(aSv1, serviceMethod, args, reply)
}

// Ping return pong if the service is active
func (aSv1 *ActionSv1) Ping(ign *utils.CGREventWithOpts, reply *string) error {
	*reply = utils.Pong
	return nil
}

// ScheduleActions will be called to schedule actions matching the arguments
func (aSv1 *ActionSv1) ScheduleActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) error {
	return aSv1.aS.V1ExecuteActions(args, rpl)
}

// ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aSv1 *ActionSv1) ExecuteActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) error {
	return aSv1.aS.V1ExecuteActions(args, rpl)
}
