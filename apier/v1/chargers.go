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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetChargerProfile returns a Charger Profile
func (APIerSv1 *APIerSv1) GetChargerProfile(arg *utils.TenantID, reply *engine.ChargerProfile) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if cpp, err := APIerSv1.DataManager.GetChargerProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *cpp
	}
	return nil
}

// GetChargerProfileIDs returns list of chargerProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetChargerProfileIDs(args *utils.TenantArgWithPaginator, chPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.ChargerProfilePrefix + args.Tenant + ":"
	keys, err := APIerSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
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

type ChargerWithCache struct {
	*engine.ChargerProfile
	Cache *string
}

//SetChargerProfile add/update a new Charger Profile
func (APIerSv1 *APIerSv1) SetChargerProfile(arg *ChargerWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg.ChargerProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetChargerProfile(arg.ChargerProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheChargerProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheChargerProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ChargerProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheChargerProfiles,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveChargerProfile remove a specific Charger Profile
func (APIerSv1 *APIerSv1) RemoveChargerProfile(arg *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveChargerProfile(arg.Tenant,
		arg.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheChargerProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheChargerProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ChargerProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheChargerProfiles,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewChargerSv1(cS *engine.ChargerService) *ChargerSv1 {
	return &ChargerSv1{cS: cS}
}

// Exports RPC from ChargerS
type ChargerSv1 struct {
	cS *engine.ChargerService
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (cSv1 *ChargerSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(cSv1, serviceMethod, args, reply)
}

func (cSv1 *ChargerSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// GetChargerForEvent  returns matching ChargerProfile for Event
func (cSv1 *ChargerSv1) GetChargersForEvent(cgrEv *utils.CGREventWithArgDispatcher,
	reply *engine.ChargerProfiles) error {
	return cSv1.cS.V1GetChargersForEvent(cgrEv, reply)
}

// ProcessEvent
func (cSv1 *ChargerSv1) ProcessEvent(args *utils.CGREventWithArgDispatcher,
	reply *[]*engine.ChrgSProcessEventReply) error {
	return cSv1.cS.V1ProcessEvent(args, reply)
}
