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
func (admS *AdminSv1) GetAttributeProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) (err error) {
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
		*reply = *attrPrf
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
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
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
func (admS *AdminSv1) GetAttributeProfiles(ctx *context.Context, args *utils.ArgsItemIDs, attrPrfs *[]*engine.AttributeProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var attrPrfIDs []string
	if err = admS.GetAttributeProfileIDs(ctx, args, &attrPrfIDs); err != nil {
		return
	}
	*attrPrfs = make([]*engine.AttributeProfile, 0, len(attrPrfIDs))
	for _, attrPrfID := range attrPrfIDs {
		var ap *engine.AttributeProfile
		ap, err = admS.dm.GetAttributeProfile(ctx, tnt, attrPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*attrPrfs = append(*attrPrfs, ap)
	}
	return
}

// GetAttributeProfileCount returns the total number of AttributeProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 AttributeProfileIDs
func (admS *AdminSv1) GetAttributeProfileCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
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

//SetAttributeProfile add/update a new Attribute Profile
func (admS *AdminSv1) SetAttributeProfile(ctx *context.Context, arg *engine.AttributeProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.AttributeProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	if len(arg.Attributes) == 0 {
		return utils.NewErrMandatoryIeMissing("Attributes")
	}
	if err := admS.dm.SetAttributeProfile(ctx, arg.AttributeProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx,
		map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheAttributeProfiles,
		arg.TenantID(), &arg.FilterIDs, arg.APIOpts); err != nil {
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
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheAttributeProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewAttributeSv1 returns the RPC Object for AttributeS
func NewAttributeSv1(attrS *engine.AttributeS) *AttributeSv1 {
	return &AttributeSv1{attrS: attrS}
}

// AttributeSv1 exports RPC from Attributes service
type AttributeSv1 struct {
	ping
	attrS *engine.AttributeS
}

// GetAttributeForEvent returns matching AttributeProfile for Event
func (alSv1 *AttributeSv1) GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AttributeProfile) (err error) {
	return alSv1.attrS.V1GetAttributeForEvent(ctx, args, reply)
}

// ProcessEvent will replace event fields with the ones in matching AttributeProfile
// and return a list of altered fields
func (alSv1 *AttributeSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) error {
	return alSv1.attrS.V1ProcessEvent(ctx, args, reply)
}
