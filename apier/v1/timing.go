/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNEtS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// GetTiming returns a TPTiming object
func (apierSv1 *APIerSv1) GetTiming(arg *utils.ArgsGetTimingID, reply *utils.TPTiming) (err error) {
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
func (apierSv1 *APIerSv1) SetTiming(args *utils.TPTimingWithAPIOpts, reply *string) error {
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
		args.ID, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}

	*reply = utils.OK
	return nil
}

// RemoveTiming removes a specific TPTimingWithAPIOpts instance
func (apierSv1 *APIerSv1) RemoveTiming(args *utils.TPTimingWithAPIOpts, reply *string) error {
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
		args.ID, nil, nil, args.APIOpts); err != nil {
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
