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
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// GetIPProfile returns an IP configuration
func (s *AdminSv1) GetIPProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.IPProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	ipp, err := s.dm.GetIPProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *ipp
	return nil
}

// GetIPProfileIDs returns list of IPProfile IDs registered for a tenant
func (s *AdminSv1) GetIPProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, ippIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.IPProfilesPrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	keys, err := s.dm.DataDB().GetKeysForPrefix(ctx, prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	limit, offset, maxItems, err := utils.GetPaginateOpts(args.APIOpts)
	if err != nil {
		return err
	}
	*ippIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return err
}

// GetIPProfiles returns a list of IPProfiles registered for a tenant.
func (s *AdminSv1) GetIPProfiles(ctx *context.Context, args *utils.ArgsItemIDs, ipps *[]*utils.IPProfile) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	var ippIDs []string
	if err := s.GetIPProfileIDs(ctx, args, &ippIDs); err != nil {
		return err
	}
	*ipps = make([]*utils.IPProfile, 0, len(ippIDs))
	for _, ippID := range ippIDs {
		ipp, err := s.dm.GetIPProfile(ctx, tnt, ippID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*ipps = append(*ipps, ipp)
	}
	return nil
}

// GetIPProfilesCount returns the total number of IPProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 IPProfileIDs
func (s *AdminSv1) GetIPProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.IPProfilesPrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	keys, err := s.dm.DataDB().GetKeysForPrefix(ctx, prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return nil
}

// SetIPProfile adds a new IP configuration.
func (s *AdminSv1) SetIPProfile(ctx *context.Context, arg *utils.IPProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.IPProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = s.cfg.GeneralCfg().DefaultTenant
	}
	if err := s.dm.SetIPProfile(ctx, arg.IPProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheIPProfiles and CacheIPs and store it in database
	//make 1 insert for both IPProfile and IPs instead of 2
	loadID := time.Now().UnixNano()
	if err := s.dm.SetLoadIDs(ctx,
		map[string]int64{utils.CacheIPProfiles: loadID,
			utils.CacheIPs: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if s.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<SetIPProfile> Delaying cache call for %v", s.cfg.GeneralCfg().CachingDelay))
		time.Sleep(s.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for IPProfile
	if err := s.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheIPProfiles,
		arg.TenantID(), utils.EmptyString, &arg.FilterIDs, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveIPProfile remove a specific IP configuration.
func (s *AdminSv1) RemoveIPProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	if err := s.dm.RemoveIPProfile(ctx, tnt, arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if s.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<RemoveIPProfile> Delaying cache call for %v", s.cfg.GeneralCfg().CachingDelay))
		time.Sleep(s.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for IPProfile
	if err := s.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheIPProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheIPProfiles and CacheIPs and store it in database
	//make 1 insert for both IPProfile and IPs instead of 2
	loadID := time.Now().UnixNano()
	if err := s.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheIPProfiles: loadID, utils.CacheIPs: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
