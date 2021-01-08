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
	"strconv"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (apierSv1 *APIerSv1) GetMaxUsage(usageRecord *engine.UsageRecordWithOpts, maxUsage *int64) error {
	if apierSv1.Responder == nil {
		return utils.NewErrNotConnected(utils.RALService)
	}
	if usageRecord.ToR == utils.EmptyString {
		usageRecord.ToR = utils.MetaVoice
	}
	if usageRecord.RequestType == utils.EmptyString {
		usageRecord.RequestType = apierSv1.Config.GeneralCfg().DefaultReqType
	}
	if usageRecord.Tenant == utils.EmptyString {
		usageRecord.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if usageRecord.Category == utils.EmptyString {
		usageRecord.Category = apierSv1.Config.GeneralCfg().DefaultCategory
	}
	if usageRecord.Subject == utils.EmptyString {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.SetupTime == utils.EmptyString {
		usageRecord.SetupTime = utils.MetaNow
	}
	if usageRecord.Usage == utils.EmptyString {
		usageRecord.Usage = strconv.FormatFloat(
			apierSv1.Config.GeneralCfg().MaxCallDuration.Seconds(), 'f', -1, 64)
	}
	cd, err := usageRecord.AsCallDescriptor(apierSv1.Config.GeneralCfg().DefaultTimezone, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var maxDur time.Duration
	if err := apierSv1.Responder.GetMaxSessionTime(&engine.CallDescriptorWithOpts{
		CallDescriptor: cd,
		Opts:           usageRecord.Opts,
	}, &maxDur); err != nil {
		return err
	}
	if maxDur == time.Duration(-1) {
		*maxUsage = -1
		return nil
	}
	*maxUsage = maxDur.Nanoseconds()
	return nil
}
