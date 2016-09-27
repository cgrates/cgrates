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
func (self *ApierV1) GetMaxUsage(usageRecord engine.UsageRecord, maxUsage *float64) error {
	err := engine.LoadUserProfile(&usageRecord, "ExtraFields")
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.VOICE
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = self.Config.DefaultReqType
	}
	if usageRecord.Direction == "" {
		usageRecord.Direction = utils.OUT
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = self.Config.DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = self.Config.DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.SetupTime == "" {
		usageRecord.SetupTime = utils.META_NOW
	}
	if usageRecord.Usage == "" {
		usageRecord.Usage = strconv.FormatFloat(self.Config.MaxCallDuration.Seconds(), 'f', -1, 64)
	}
	storedCdr, err := usageRecord.AsStoredCdr(self.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var maxDur float64
	if err := self.Responder.GetDerivedMaxSessionTime(storedCdr, &maxDur); err != nil {
		return err
	}
	if maxDur == -1.0 {
		*maxUsage = -1.0
		return nil
	}
	*maxUsage = time.Duration(maxDur).Seconds()
	return nil
}
