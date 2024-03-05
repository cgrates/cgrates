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

package admins

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetDispatcherProfile returns a Dispatcher Profile
func (admS *AdminS) V1GetDispatcherProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error {
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
func (admS *AdminS) V1GetDispatcherProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, dPrfIDs *[]string) (err error) {
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
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*dPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetDispatcherProfiles returns a list of dispatcher profiles registered for a tenant
func (admS *AdminS) V1GetDispatcherProfiles(ctx *context.Context, args *utils.ArgsItemIDs, dspPrfs *[]*engine.DispatcherProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var dspPrfIDs []string
	if err = admS.V1GetDispatcherProfileIDs(ctx, args, &dspPrfIDs); err != nil {
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
func (admS *AdminS) V1GetDispatcherProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
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
	APIOpts map[string]any
}

// SetDispatcherProfile add/update a new Dispatcher Profile
func (admS *AdminS) V1SetDispatcherProfile(ctx *context.Context, args *DispatcherWithAPIOpts, reply *string) error {
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
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetDispatcherProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheDispatcherProfiles,
		args.TenantID(), utils.EmptyString, &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetDispatcherProfile> (DispatchersInstance) Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatchersInstance
	cacheAct := utils.MetaRemove
	if err := admS.CallCache(ctx, utils.FirstNonEmpty(utils.IfaceAsString(args.APIOpts[utils.MetaCache]), cacheAct),
		args.Tenant, utils.CacheDispatchers, args.TenantID(), utils.EmptyString, &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetDispatcherProfile> (DispatcherRoutes) Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherRoutes
	if err := admS.CallCache(ctx, utils.FirstNonEmpty(utils.IfaceAsString(args.APIOpts[utils.MetaCache]), cacheAct),
		args.Tenant, utils.CacheDispatcherRoutes, args.TenantID(),
		utils.ConcatenatedKey(utils.CacheDispatcherProfiles, args.Tenant, args.ID),
		&args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveDispatcherProfile remove a specific Dispatcher Profile
func (admS *AdminS) V1RemoveDispatcherProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
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
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1RemoveDispatcherProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheDispatcherProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetDispatcherHost returns a Dispatcher Host
func (admS *AdminS) V1GetDispatcherHost(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error {
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
func (admS *AdminS) V1GetDispatcherHostIDs(ctx *context.Context, args *utils.ArgsItemIDs, dspHostIDs *[]string) (err error) {
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
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*dspHostIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetDispatcherHosts returns a list of dispatcher hosts registered for a tenant
func (admS *AdminS) V1GetDispatcherHosts(ctx *context.Context, args *utils.ArgsItemIDs, dspHosts *[]*engine.DispatcherHost) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var dspHostIDs []string
	if err = admS.V1GetDispatcherHostIDs(ctx, args, &dspHostIDs); err != nil {
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
func (admS *AdminS) V1GetDispatcherHostsCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
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

// SetDispatcherHost add/update a new Dispatcher Host
func (admS *AdminS) V1SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, reply *string) error {
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
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetDispatcherHost> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheDispatcherHosts,
		args.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveDispatcherHost remove a specific Dispatcher Host
func (admS *AdminS) V1RemoveDispatcherHost(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
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
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1RemoveDispatcherHost> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherProfile
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheDispatcherHosts,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
