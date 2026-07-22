// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// GetTiming returns a TPTiming object
func (apierSv1 *APIerSv1) GetTiming(ctx *context.Context, arg *utils.ArgsGetTimingID, reply *utils.TPTiming) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	tm, err := apierSv1.DataManager.GetTiming(arg.ID, false, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *tm
	return
}

// SetTiming alters/creates a TPTimingWithAPIOpts
func (apierSv1 *APIerSv1) SetTiming(ctx *context.Context, args *utils.TPTimingWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.TPTiming, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	if err := apierSv1.DataManager.SetTiming(args.TPTiming); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheTimings and store it in database
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheTimings: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Timings
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), args.Tenant, utils.CacheTimings,
		args.ID, utils.EmptyString, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}

	*reply = utils.OK
	return nil
}

// RemoveTiming removes a specific TPTimingWithAPIOpts instance
func (apierSv1 *APIerSv1) RemoveTiming(ctx *context.Context, args *utils.TPTimingWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.TPTiming, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveTiming(args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Timings
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheTimings,
		args.ID, utils.EmptyString, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}

	//generate a loadID for CacheTimings and store it in database
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheTimings: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}

	*reply = utils.OK
	return nil
}
