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
	"github.com/cgrates/cgrates/trends"
	"github.com/cgrates/cgrates/utils"
)

// GetTrendProfile returns a Trend profile
func (adms *AdminSv1) GetTrendProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.TrendProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	sCfg, err := adms.dm.GetTrendProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sCfg
	return
}

// GetTrendProfileIDs returns list of TrendProfile IDs registered for a tenant
func (adms *AdminSv1) GetTrendProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, stsPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.TrendProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := adms.dm.DBConns().GetConn(utils.MetaTrendProfiles)
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
	*stsPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetTrendProfiles returns a list of stats profiles registered for a tenant
func (admS *AdminSv1) GetTrendProfiles(ctx *context.Context, args *utils.ArgsItemIDs, sqPrfs *[]*utils.TrendProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var sqPrfIDs []string
	if err = admS.GetTrendProfileIDs(ctx, args, &sqPrfIDs); err != nil {
		return
	}
	*sqPrfs = make([]*utils.TrendProfile, 0, len(sqPrfIDs))
	for _, sqPrfID := range sqPrfIDs {
		var sqPrf *utils.TrendProfile
		sqPrf, err = admS.dm.GetTrendProfile(ctx, tnt, sqPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*sqPrfs = append(*sqPrfs, sqPrf)
	}
	return
}

// GetTrendProfilesCount returns the total number of TrendProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 TrendProfileIDs
func (admS *AdminSv1) GetTrendProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.TrendProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaTrendProfiles)
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

// SetTrendProfile alters/creates a TrendProfile
func (adms *AdminSv1) SetTrendProfile(ctx *context.Context, arg *utils.TrendProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.TrendProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err = adms.dm.SetTrendProfile(ctx, arg.TrendProfile); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheTrendProfiles and store it in database
	loadID := time.Now().UnixNano()
	if err = adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheTrendProfiles: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetTrendProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for TrendProfile
	if err = adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheTrendProfiles,
		arg.TenantID(), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveTrendProfile remove a specific stat configuration
func (adms *AdminSv1) RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveTrendProfile(ctx, tnt, args.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveTrendProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for TrendProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), tnt, utils.CacheTrendProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheTrendProfiles and store it in database
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheTrendProfiles: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewTrendSv1 initializes the TrendSv1 object.
func NewTrendSv1(trndS *trends.TrendS) *TrendSv1 {
	return &TrendSv1{trndS: trndS}
}

// TrendSv1 represents the RPC object to register for trends v1 APIs.
type TrendSv1 struct {
	trndS *trends.TrendS
}

// V1ScheduleQueries manually schedules or reschedules trend queries.
func (tS *TrendSv1) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries, scheduled *int) (err error) {
	return tS.trndS.V1ScheduleQueries(ctx, args, scheduled)
}

// V1GetTrend retrieves trend metrics with optional time and index filtering.
func (tS *TrendSv1) V1GetTrend(ctx *context.Context, arg *utils.ArgGetTrend, retTrend *utils.Trend) (err error) {
	return tS.trndS.V1GetTrend(ctx, arg, retTrend)
}

// V1GetScheduledTrends retrieves information about currently scheduled trends.
func (tS *TrendSv1) V1GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, schedTrends *[]utils.ScheduledTrend) (err error) {
	return tS.trndS.V1GetScheduledTrends(ctx, args, schedTrends)
}

// V1GetTrendSummary retrieves the most recent trend summary.
func (tS *TrendSv1) V1GetTrendSummary(ctx *context.Context, arg utils.TenantIDWithAPIOpts, reply *utils.TrendSummary) (err error) {
	return tS.trndS.V1GetTrendSummary(ctx, arg, reply)
}
