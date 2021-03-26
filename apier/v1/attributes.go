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
func (apierSv1 *APIerSv1) GetAttributeProfile(arg *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var alsPrf *engine.AttributeProfile
	if alsPrf, err = apierSv1.DataManager.GetAttributeProfile(tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*reply = *alsPrf
	return nil
}

// GetAttributeProfileIDs returns list of attributeProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetAttributeProfileIDs(args *utils.PaginatorWithTenant, attrPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep
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
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetAttributeProfileIDsCount sets in reply var the total number of AttributeProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 AttributeProfileIDs
func (apierSv1 *APIerSv1) GetAttributeProfileIDsCount(args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetAttributeProfile add/update a new Attribute Profile
func (apierSv1 *APIerSv1) SetAttributeProfile(alsWrp *engine.AttributeProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(alsWrp.AttributeProfile, []string{utils.ID, utils.Attributes}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if alsWrp.Tenant == utils.EmptyString {
		alsWrp.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	for _, attr := range alsWrp.Attributes {
		if attr.Path == utils.EmptyString {
			return utils.NewErrMandatoryIeMissing("Path")
		}
		for _, sub := range attr.Value {
			if sub.Rules == utils.EmptyString {
				return utils.NewErrMandatoryIeMissing("Rules")
			}
			if err := sub.Compile(); err != nil {
				return utils.NewErrServerError(err)
			}
		}
	}
	if err := apierSv1.DataManager.SetAttributeProfile(alsWrp.AttributeProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}

	if err := apierSv1.CallCache(utils.IfaceAsString(alsWrp.APIOpts[utils.CacheOpt]), alsWrp.Tenant, utils.CacheAttributeProfiles,
		alsWrp.TenantID(), &alsWrp.FilterIDs, alsWrp.Contexts, alsWrp.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveAttributeProfile remove a specific Attribute Profile
func (apierSv1 *APIerSv1) RemoveAttributeProfile(arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveAttributeProfile(tnt, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheAttributeProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewAttributeSv1 returns the RPC Object for AttributeS
func NewAttributeSv1(attrS *engine.AttributeService) *AttributeSv1 {
	return &AttributeSv1{attrS: attrS}
}

// AttributeSv1 exports RPC from RLs
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

// ProcessEvent will replace event fields with the ones in matching AttributeProfile
func (alSv1 *AttributeSv1) ProcessEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttrSProcessEventReply) error {
	return alSv1.attrS.V1ProcessEvent(args, reply)
}

// Ping return pong if the service is active
func (alSv1 *AttributeSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
