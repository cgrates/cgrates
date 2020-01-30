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
func (self *APIerSv1) GetMaxUsage(usageRecord engine.UsageRecordWithArgDispatcher, maxUsage *int64) error {
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.VOICE
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = self.Config.GeneralCfg().DefaultReqType
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = self.Config.GeneralCfg().DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = self.Config.GeneralCfg().DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.SetupTime == "" {
		usageRecord.SetupTime = utils.META_NOW
	}
	if usageRecord.Usage == "" {
		usageRecord.Usage = strconv.FormatFloat(
			self.Config.MaxCallDuration.Seconds(), 'f', -1, 64)
	}
	cd, err := usageRecord.AsCallDescriptor(self.Config.GeneralCfg().DefaultTimezone, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var maxDur time.Duration
	if err := self.Responder.GetMaxSessionTime(&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd,
		ArgDispatcher: usageRecord.ArgDispatcher}, &maxDur); err != nil {
		return err
	}
	if maxDur == time.Duration(-1) {
		*maxUsage = -1
		return nil
	}
	*maxUsage = maxDur.Nanoseconds()
	return nil
}
