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

// GetAttributeProfile returns an Attribute Profile
func (APIerSv1 *APIerSv1) GetAttributeProfile(arg utils.TenantIDWithArgDispatcher, reply *engine.AttributeProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if alsPrf, err := APIerSv1.DataManager.GetAttributeProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *alsPrf
	}
	return nil
}

// GetAttributeProfileIDs returns list of attributeProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetAttributeProfileIDs(args utils.TenantArgWithPaginator, attrPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.AttributeProfilePrefix + args.Tenant + ":"
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
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

type AttributeWithCache struct {
	*engine.AttributeProfile
	Cache *string
	*utils.ArgDispatcher
}

//SetAttributeProfile add/update a new Attribute Profile
func (APIerSv1 *APIerSv1) SetAttributeProfile(alsWrp *AttributeWithCache, reply *string) error {
	if missing := utils.MissingStructFields(alsWrp.AttributeProfile, []string{"Tenant", "ID", "Attributes"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(alsWrp.Attributes) != 0 {
		for _, attr := range alsWrp.Attributes {
			for _, sub := range attr.Value {
				if sub.Rules == "" {
					return utils.NewErrMandatoryIeMissing("Rules")
				}
				if err := sub.Compile(); err != nil {
					return utils.NewErrServerError(err)
				}
			}
		}
	}
	if err := APIerSv1.DataManager.SetAttributeProfile(alsWrp.AttributeProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	args := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  alsWrp.TenantID(),
	}
	if err := APIerSv1.CallCache(alsWrp.Tenant, GetCacheOpt(alsWrp.Cache), args); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveAttributeProfile remove a specific Attribute Profile
func (APIerSv1 *APIerSv1) RemoveAttributeProfile(arg *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveAttributeProfile(arg.Tenant, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	args := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  utils.ConcatenatedKey(arg.Tenant, arg.ID),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), args); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewAttributeSv1(attrS *engine.AttributeService) *AttributeSv1 {
	return &AttributeSv1{attrS: attrS}
}

// Exports RPC from RLs
type AttributeSv1 struct {
	attrS *engine.AttributeService
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (alSv1 *AttributeSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(alSv1, serviceMethod, args, reply)
}

// GetAttributeForEvent  returns matching AttributeProfile for Event
func (alSv1 *AttributeSv1) GetAttributeForEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttributeProfile) (err error) {
	return alSv1.attrS.V1GetAttributeForEvent(args, reply)
}

// ProcessEvent will replace event fields with the ones in maching AttributeProfile
func (alSv1 *AttributeSv1) ProcessEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttrSProcessEventReply) error {
	return alSv1.attrS.V1ProcessEvent(args, reply)
}

func (alSv1 *AttributeSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
