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
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetRankingProfile returns a Ranking profile
func (apierSv1 *APIerSv1) GetRankingProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.RankingProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	sg, err := apierSv1.DataManager.GetRankingProfile(tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sg
	return
}

// GetRankingProfileIDs returns list of rankingProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetRankingProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, sgPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.RankingsProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	sgIDs := make([]string, len(keys))
	for i, key := range keys {
		sgIDs[i] = key[len(prfx):]
	}
	*sgPrfIDs = args.PaginateStringSlice(sgIDs)
	return
}

// SetRankingProfile alters/creates a RankingProfile
func (apierSv1 *APIerSv1) SetRankingProfile(ctx *context.Context, arg *engine.RankingProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.RankingProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetRankingProfile(arg.RankingProfile); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRankingProfiles and store it in database
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRankingProfiles: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<SetRankingProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for RankingProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheRankingProfiles,
		arg.TenantID(), utils.EmptyString, nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRankingProfile remove a specific ranking configuration
func (apierSv1 *APIerSv1) RemoveRankingProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveRankingProfile(tnt, args.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<RemoveRankingProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for RankingProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheRankingProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRankingProfiles and store it in database
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRankingProfiles: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewRankingSv1 initializes RankingSV1
func NewRankingSv1(rnks *engine.RankingS) *RankingSv1 {
	return &RankingSv1{
		rnkS: rnks,
	}
}

type RankingSv1 struct {
	rnkS *engine.RankingS
}

// GetRankingSummary returns summary of the last updated ranking
func (rnks *RankingSv1) GetRankingSummary(ctx *context.Context, arg utils.TenantIDWithAPIOpts, reply *engine.RankingSummary) error {
	return rnks.rnkS.V1GetRankingSummary(ctx, arg, reply)
}
