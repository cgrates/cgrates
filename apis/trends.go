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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetTrendProfile returns a Trend profile
func (adms *AdminSv1) GetTrendProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.TrendProfile) (err error) {
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
	var keys []string
	if keys, err = adms.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
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
func (admS *AdminSv1) GetTrendProfiles(ctx *context.Context, args *utils.ArgsItemIDs, sqPrfs *[]*engine.TrendProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var sqPrfIDs []string
	if err = admS.GetTrendProfileIDs(ctx, args, &sqPrfIDs); err != nil {
		return
	}
	*sqPrfs = make([]*engine.TrendProfile, 0, len(sqPrfIDs))
	for _, sqPrfID := range sqPrfIDs {
		var sqPrf *engine.TrendProfile
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

// SetTrendProfile alters/creates a TrendProfile
func (adms *AdminSv1) SetTrendProfile(ctx *context.Context, arg *engine.TrendProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.TrendProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err = adms.dm.SetTrendProfile(ctx, arg.TrendProfile); err != nil {
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
	*reply = utils.OK
	return nil
}

// NewTrendSv1 initializes TrendSV1
func NewTrendSv1(trS *engine.TrendS) *TrendSv1 {
	return &TrendSv1{trS: trS}
}

// TrendSv1 exports RPC from RLs
type TrendSv1 struct {
	ping
	trS *engine.TrendS
}

func (trs *TrendSv1) ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries, scheduled *int) error {
	return trs.trS.V1ScheduleQueries(ctx, args, scheduled)
}

func (trs *TrendSv1) GetTrend(ctx *context.Context, args *utils.ArgGetTrend, trend *engine.Trend) error {
	return trs.trS.V1GetTrend(ctx, args, trend)
}

func (trs *TrendSv1) GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, schedTrends *[]utils.ScheduledTrend) error {
	return trs.trS.V1GetScheduledTrends(ctx, args, schedTrends)
}

func (trs *TrendSv1) GetTrendSummary(ctx *context.Context, arg utils.TenantIDWithAPIOpts, reply *engine.TrendSummary) error {
	return trs.trS.V1GetTrendSummary(ctx, arg, reply)
}
