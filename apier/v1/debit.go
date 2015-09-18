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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (self *ApierV1) DebitUsage(usageRecord engine.UsageRecord, reply *string) error {
	if missing := utils.MissingStructFields(&usageRecord, []string{"Account", "Destination", "Usage"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	err := engine.LoadUserProfile(usageRecord, "")
	if err != nil {
		*reply = err.Error()
		return err
	}
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
	if usageRecord.AnswerTime == "" {
		usageRecord.AnswerTime = utils.META_NOW
	}
	cd, err := usageRecord.AsCallDescriptor(self.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var cc engine.CallCost
	if err := self.Responder.Debit(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}
