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
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/utils"
)

// GetAttributeProfile returns an Attribute Profile based on the tenant and ID received
func (admS *AdminSv1) GetAttributeProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.APIAttributeProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if attrPrf, err := admS.dm.GetAttributeProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		attr := utils.NewAPIAttributeProfile(attrPrf)
		*reply = *attr
	}
	return nil
}

// GetAttributeProfileIDs returns list of attributeProfile IDs registered for a tenant
func (admS *AdminSv1) GetAttributeProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, attrPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaAttributeProfiles)
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
	*attrPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetAttributeProfiles returns a list of attribute profiles registered for a tenant
func (admS *AdminSv1) GetAttributeProfiles(ctx *context.Context, args *utils.ArgsItemIDs, attrPrfs *[]*utils.APIAttributeProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var attrPrfIDs []string
	if err = admS.GetAttributeProfileIDs(ctx, args, &attrPrfIDs); err != nil {
		return
	}
	*attrPrfs = make([]*utils.APIAttributeProfile, 0, len(attrPrfIDs))
	for _, attrPrfID := range attrPrfIDs {
		var ap *utils.AttributeProfile
		ap, err = admS.dm.GetAttributeProfile(ctx, tnt, attrPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		attr := utils.NewAPIAttributeProfile(ap)
		*attrPrfs = append(*attrPrfs, attr)
	}
	return
}

// GetAttributeProfilesCount returns the total number of AttributeProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 AttributeProfileIDs
func (admS *AdminSv1) GetAttributeProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaAttributeProfiles)
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

// SetAttributeProfile add/update a new Attribute Profile
func (admS *AdminSv1) SetAttributeProfile(ctx *context.Context, arg *utils.APIAttributeProfileWithAPIOpts, reply *string) error {
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
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetAttributeProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), alsPrf.Tenant, utils.CacheAttributeProfiles,
		alsPrf.TenantID(), utils.EmptyString, &alsPrf.FilterIDs, arg.APIOpts); err != nil {
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
		true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveAttributeProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheAttributeProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewAttributeSv1 initializes the AttributeSv1 object.
func NewAttributeSv1(atrs *attributes.AttributeS) *AttributeSv1 {
	return &AttributeSv1{atrs: atrs}
}

// AttributeSv1 represents the RPC object to register for attributes v1 APIs.
type AttributeSv1 struct {
	atrs *attributes.AttributeS
}

// V1GetAttributeForEvent returns the AttributeProfile that matches the event
func (atrS *AttributeSv1) V1GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent, attrPrf *utils.APIAttributeProfile) (err error) {
	return atrS.V1GetAttributeForEvent(ctx, args, attrPrf)
}

// V1ProcessEvent proccess the event and returns the result
func (atrS *AttributeSv1) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, attrEvntRpl *attributes.AttrSProcessEventReply) (err error) {
	return atrS.V1ProcessEvent(ctx, args, attrEvntRpl)
}
