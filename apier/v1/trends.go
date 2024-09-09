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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetTrendProfile returns a Trend profile
func (apierSv1 *APIerSv1) GetTrendProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.TrendProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	tr, err := apierSv1.DataManager.GetTrendProfile(tnt, arg.ID, true, false, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *tr
	return
}

// GetTrendProfiles returns a list of Trendprofiles
func (apierSv1 *APIerSv1) GetTrendProfiles(ctx *context.Context, arg *engine.TrendProfilesAPI, reply *[]*engine.TrendProfile) (err error) {
	tr, err := apierSv1.DataManager.GetTrendProfiles(arg.Tenant, arg.TpIDs)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = tr
	return
}

// GetTrendProfileIDs returns list of trendProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetTrendProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, trPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.TrendsProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	trIDs := make([]string, len(keys))
	for i, key := range keys {
		trIDs[i] = key[len(prfx):]
	}
	*trPrfIDs = args.PaginateStringSlice(trIDs)
	return
}

// SetTrendProfile alters/creates a TrendProfile
func (apierSv1 *APIerSv1) SetTrendProfile(ctx *context.Context, arg *engine.TrendProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.TrendProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetTrendProfile(arg.TrendProfile); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveTrendProfile remove a specific trend configuration
func (apierSv1 *APIerSv1) RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveTrendProfile(tnt, args.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewTrendSv1 initializes TrendSV1
func NewTrendSv1(trs *engine.TrendS) *TrendSv1 {
	return &TrendSv1{}
}

type TrendSv1 struct {
	trS *engine.TrendS
}
