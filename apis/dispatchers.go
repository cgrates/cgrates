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

// GetDispatcherProfile returns a Dispatcher Profile
func (admS *AdminSv1) GetDispatcherProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	dpp, err := admS.dm.GetDispatcherProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *dpp
	return nil
}

// GetDispatcherProfileIDs returns list of dispatcherProfile IDs registered for a tenant
func (admS *AdminSv1) GetDispatcherProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, dPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.DispatcherProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	if keys, err = utils.Paginate(keys, limit, offset, maxItems); err != nil {
		return
	}
	*dPrfIDs = make([]string, len(keys))
	for i, key := range keys {
		(*dPrfIDs)[i] = key[lenPrfx:]
	}
	return
}

// GetDispatcherProfiles returns a list of dispatcher profiles registered for a tenant
func (admS *AdminSv1) GetDispatcherProfiles(ctx *context.Context, args *utils.ArgsItemIDs, dspPrfs *[]*engine.DispatcherProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var dspPrfIDs []string
	if err = admS.GetDispatcherProfileIDs(ctx, args, &dspPrfIDs); err != nil {
		return
	}
	*dspPrfs = make([]*engine.DispatcherProfile, 0, len(dspPrfIDs))
	for _, dspPrfID := range dspPrfIDs {
		var dspPrf *engine.DispatcherProfile
		dspPrf, err = admS.dm.GetDispatcherProfile(ctx, tnt, dspPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*dspPrfs = append(*dspPrfs, dspPrf)
	}
	return
}

// GetDispatcherProfilesCount returns the total number of DispatcherProfiles registered for a tenant
// returns ErrNotFound in case of 0 DispatcherProfiles
func (admS *AdminSv1) GetDispatcherProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.DispatcherProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
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

type DispatcherWithAPIOpts struct {
	*engine.DispatcherProfile
	APIOpts map[string]interface{}
}

//SetDispatcherProfile add/update a new Dispatcher Profile
func (admS *AdminSv1) SetDispatcherProfile(ctx *context.Context, args *DispatcherWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.DispatcherProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.SetDispatcherProfile(ctx, args.DispatcherProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheDispatcherProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheDispatcherProfiles,
		args.TenantID(), &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveDispatcherProfile remove a specific Dispatcher Profile
func (admS *AdminSv1) RemoveDispatcherProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveDispatcherProfile(ctx, tnt,
		arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheDispatcherProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheDispatcherProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetDispatcherHost returns a Dispatcher Host
func (admS *AdminSv1) GetDispatcherHost(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	dpp, err := admS.dm.GetDispatcherHost(ctx, tnt, arg.ID, true, false, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *dpp
	return nil
}

// GetDispatcherHostIDs returns list of dispatcherHost IDs registered for a tenant
func (admS *AdminSv1) GetDispatcherHostIDs(ctx *context.Context, args *utils.ArgsItemIDs, dPrfIDs *[]string) (err error) {
	tenant := args.Tenant
	if tenant == utils.EmptyString {
		tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.DispatcherHostPrefix + tenant + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	if keys, err = utils.Paginate(keys, limit, offset, maxItems); err != nil {
		return
	}
	*dPrfIDs = make([]string, len(keys))
	for i, key := range keys {
		(*dPrfIDs)[i] = key[lenPrfx:]
	}
	return
}

// GetDispatcherHosts returns a list of dispatcher hosts registered for a tenant
func (admS *AdminSv1) GetDispatcherHosts(ctx *context.Context, args *utils.ArgsItemIDs, dspHosts *[]*engine.DispatcherHost) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var dspHostIDs []string
	if err = admS.GetDispatcherHostIDs(ctx, args, &dspHostIDs); err != nil {
		return
	}
	*dspHosts = make([]*engine.DispatcherHost, 0, len(dspHostIDs))
	for _, dspHostID := range dspHostIDs {
		var dspHost *engine.DispatcherHost
		dspHost, err = admS.dm.GetDispatcherHost(ctx, tnt, dspHostID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*dspHosts = append(*dspHosts, dspHost)
	}
	return
}

// GetDispatcherHostsCount returns the total number of DispatcherHosts registered for a tenant
// returns ErrNotFound in case of 0 DispatcherHosts
func (admS *AdminSv1) GetDispatcherHostsCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.DispatcherHostPrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
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

//SetDispatcherHost add/update a new Dispatcher Host
func (admS *AdminSv1) SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.DispatcherHost, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.SetDispatcherHost(ctx, args.DispatcherHost); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherHosts and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheDispatcherHosts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheDispatcherHosts,
		args.TenantID(), nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveDispatcherHost remove a specific Dispatcher Host
func (admS *AdminSv1) RemoveDispatcherHost(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveDispatcherHost(ctx, tnt, arg.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherHosts and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheDispatcherHosts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheDispatcherHosts,
		utils.ConcatenatedKey(tnt, arg.ID), nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

/*
func NewDispatcherSv1(dS *dispatchers.DispatcherService) *DispatcherSv1 {
	return &DispatcherSv1{dS: dS}
}

type DispatcherSv1 struct {
	dS *dispatchers.DispatcherService
	ping
}

// // GetProfileForEvent returns the matching dispatcher profile for the provided event
// func (dSv1 DispatcherSv1) GetProfilesForEvent(ctx *context.Context, ev *utils.CGREvent,
// 	dPrfl *engine.DispatcherProfiles) error {
// 	return dSv1.dS.V1GetProfilesForEvent(ctx, ev, dPrfl)
// }

// func (dS *DispatcherSv1) RemoteStatus(args *utils.TenantWithAPIOpts, reply *map[string]interface{}) (err error) {
// 	return dS.dS.DispatcherSv1RemoteStatus(args, reply)
// }

// func (dS *DispatcherSv1) RemotePing(args *utils.CGREvent, reply *string) (err error) {
// 	return dS.dS.DispatcherSv1RemotePing(args, reply)
// }

// func (dS *DispatcherSv1) RemoteSleep(args *utils.DurationArgs, reply *string) (err error) {
// 	return dS.dS.DispatcherSv1RemoteSleep(args, reply)
// }
*/
