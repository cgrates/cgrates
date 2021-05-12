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

// GetSupplierProfile returns a Supplier configuration
func (APIerSv1 *APIerSv1) GetSupplierProfile(arg utils.TenantID, reply *engine.SupplierProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if spp, err := APIerSv1.DataManager.GetSupplierProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *spp
	}
	return nil
}

// GetSupplierProfileIDs returns list of supplierProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetSupplierProfileIDs(args utils.TenantArgWithPaginator, sppPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.SupplierProfilePrefix + args.Tenant + ":"
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
	*sppPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

type SupplierWithCache struct {
	*engine.SupplierProfile
	Cache *string
}

//SetSupplierProfile add a new Supplier configuration
func (APIerSv1 *APIerSv1) SetSupplierProfile(args *SupplierWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args.SupplierProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetSupplierProfile(args.SupplierProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheSupplierProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheSupplierProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheSupplierProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveSupplierProfile remove a specific Supplier configuration
func (APIerSv1 *APIerSv1) RemoveSupplierProfile(args *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveSupplierProfile(args.Tenant, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheSupplierProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheSupplierProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheSupplierProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewSupplierSv1(splS *engine.SupplierService) *SupplierSv1 {
	return &SupplierSv1{splS: splS}
}

// Exports RPC from SupplierS
type SupplierSv1 struct {
	splS *engine.SupplierService
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (splv1 *SupplierSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(splv1, serviceMethod, args, reply)
}

// GetSuppliers returns sorted list of suppliers for Event
func (splv1 *SupplierSv1) GetSuppliers(args *engine.ArgsGetSuppliers,
	reply *engine.SortedSuppliers) error {
	return splv1.splS.V1GetSuppliers(args, reply)
}

// GetSuppliersProfiles returns a list of suppliers profiles that match for Event
func (splv1 *SupplierSv1) GetSupplierProfilesForEvent(args *utils.CGREventWithArgDispatcher,
	reply *[]*engine.SupplierProfile) error {
	return splv1.splS.V1GetSupplierProfilesForEvent(args, reply)
}

func (splv1 *SupplierSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
