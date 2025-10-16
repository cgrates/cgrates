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
	"github.com/cgrates/cgrates/rankings"
	"github.com/cgrates/cgrates/utils"
)

// GetRankingProfile returns a Ranking profile
func (adms *AdminSv1) GetRankingProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.RankingProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	sCfg, err := adms.dm.GetRankingProfile(ctx, tnt, arg.ID,
		true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sCfg
	return
}

// GetRankingProfileIDs returns list of RankingProfile IDs registered for a tenant
func (adms *AdminSv1) GetRankingProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, rngPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RankingProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := adms.dm.DBConns().GetConn(utils.MetaRankingProfiles)
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
	*rngPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetRankingProfiles returns a list of ranking profiles registered for a tenant
func (admS *AdminSv1) GetRankingProfiles(ctx *context.Context, args *utils.ArgsItemIDs, rgPrfs *[]*utils.RankingProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var sqPrfIDs []string
	if err = admS.GetRankingProfileIDs(ctx, args, &sqPrfIDs); err != nil {
		return
	}
	*rgPrfs = make([]*utils.RankingProfile, 0, len(sqPrfIDs))
	for _, sqPrfID := range sqPrfIDs {
		var rgPrf *utils.RankingProfile
		rgPrf, err = admS.dm.GetRankingProfile(ctx, tnt, sqPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*rgPrfs = append(*rgPrfs, rgPrf)
	}
	return
}

// GetRankingProfilesCount returns the total number of RankingProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 RankingProfileIDs
func (admS *AdminSv1) GetRankingProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RankingProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaRankingProfiles)
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

// SetRankingProfile alters/creates a RankingProfile
func (adms *AdminSv1) SetRankingProfile(ctx *context.Context, arg *utils.RankingProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.RankingProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err = adms.dm.SetRankingProfile(ctx, arg.RankingProfile); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRankingProfiles and store it in database
	loadID := time.Now().UnixNano()
	if err = adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRankingProfiles: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetRankingProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for RankingProfile
	if err = adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheRankingProfiles,
		arg.TenantID(), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRankingProfile remove a specific ranking configuration
func (adms *AdminSv1) RemoveRankingProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveRankingProfile(ctx, tnt, args.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveRankingProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for RankingProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), tnt, utils.CacheRankingProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRankingProfiles and store it in database
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRankingProfiles: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewRankingSv1 initializes the RankingSv1 object.
func NewRankingSv1(rnkS *rankings.RankingS) *RankingSv1 {
	return &RankingSv1{rnkS: rnkS}
}

// RankingSv1 represents the RPC object to register for rankings v1 APIs.
type RankingSv1 struct {
	rnkS *rankings.RankingS
}

// V1ScheduleQueries manually schedules or reschedules ranking queries.
func (rnkS *RankingSv1) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleRankingQueries, scheduled *int) (err error) {
	return rnkS.rnkS.V1ScheduleQueries(ctx, args, scheduled)
}

// V1GetRanking retrieves ranking metrics with optional filtering.
func (rnkS *RankingSv1) V1GetRanking(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, retRanking *utils.Ranking) (err error) {
	return rnkS.rnkS.V1GetRanking(ctx, arg, retRanking)
}

// V1GetSchedule retrieves information about currently scheduled rankings.
func (rnkS *RankingSv1) V1GetSchedule(ctx *context.Context, args *utils.ArgScheduledRankings, schedRankings *[]utils.ScheduledRanking) (err error) {
	return rnkS.rnkS.V1GetSchedule(ctx, args, schedRankings)
}

// V1GetRankingSummary retrieves the most recent ranking summary.
func (rnkS *RankingSv1) V1GetRankingSummary(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.RankingSummary) (err error) {
	return rnkS.rnkS.V1GetRankingSummary(ctx, arg, reply)
}
