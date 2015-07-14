/*
Real-time Charging System for Telecom & ISP environments
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
	out, err := engine.LoadUserProfile(usageRecord)
	if err != nil {
		return err
	}
	usageRecord = out.(engine.UsageRecord)
	if usageRecord.TOR == "" {
		usageRecord.TOR = utils.VOICE
	}
	if usageRecord.ReqType == "" {
		usageRecord.ReqType = self.Config.DefaultReqType
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
	storedCdr, err := usageRecord.AsStoredCdr()
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
