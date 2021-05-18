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

// GetAttributeProfile returns an Attribute Profile based on the tenant and ID received
func (admS *AdminSv1) GetAttributeProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.APIAttributeProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var alsPrf *engine.AttributeProfile
	if alsPrf, err = admS.dm.GetAttributeProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return
	}
	attr := engine.NewAPIAttributeProfile(alsPrf)
	*reply = *attr
	return nil
}

// GetAttributeProfileIDs returns list of attributeProfile IDs registered for a tenant
func (admS *AdminSv1) GetAttributeProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, attrPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep
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
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetAttributeProfileIDsCount returns the total number of AttributeProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 AttributeProfileIDs
func (admS *AdminSv1) GetAttributeProfileIDsCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetAttributeProfile add/update a new Attribute Profile
func (admS *AdminSv1) SetAttributeProfile(ctx *context.Context, arg *engine.AttributeWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.APIAttributeProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	alsPrf, err := arg.APIAttributeProfile.AsAttributeProfile()
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.dm.SetAttributeProfile(ctx, alsPrf, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx,
		map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), alsPrf.Tenant, utils.CacheAttributeProfiles,
		alsPrf.TenantID(), &alsPrf.FilterIDs, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveAttributeProfile remove a specific Attribute Profile based on tenant an ID
func (admS *AdminSv1) RemoveAttributeProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveAttributeProfile(ctx, tnt, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheAttributeProfiles,
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

// AttributeSv1 exports RPC from Attributes service
type AttributeSv1 struct {
	ping
	attrS *engine.AttributeService
}

// GetAttributeForEvent returns matching AttributeProfile for Event
func (alSv1 *AttributeSv1) GetAttributeForEvent(ctx *context.Context, args *engine.AttrArgsProcessEvent,
	reply *engine.AttributeProfile) (err error) {
	return alSv1.attrS.V1GetAttributeForEvent(ctx, args, reply)
}

// ProcessEvent will replace event fields with the ones in matching AttributeProfile
// and return a list of altered fields
func (alSv1 *AttributeSv1) ProcessEvent(ctx *context.Context, args *engine.AttrArgsProcessEvent,
	reply *engine.AttrSProcessEventReply) error {
	return alSv1.attrS.V1ProcessEvent(ctx, args, reply)
}
