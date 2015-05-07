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
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"time"
)

// Returns MaxSessionTime in seconds, -1 for no limit
func (self *ApierV1) GetMaxSessionTime(auth engine.MaxUsageReq, maxSessionTime *float64) error {
	if auth.TOR == "" {
		auth.TOR = utils.VOICE
	}
	if auth.ReqType == "" {
		auth.ReqType = self.Config.DefaultReqType
	}
	if auth.Direction == "" {
		auth.Direction = utils.OUT
	}
	if auth.Tenant == "" {
		auth.Tenant = self.Config.DefaultTenant
	}
	if auth.Category == "" {
		auth.Category = self.Config.DefaultCategory
	}
	if auth.Subject == "" {
		auth.Subject = auth.Account
	}
	if auth.Subject == "" {
		auth.Subject = auth.Account
	}
	if auth.SetupTime == "" {
		auth.SetupTime = utils.META_NOW
	}
	if auth.Usage == "" {
		auth.Usage = strconv.FormatFloat(self.Config.MaxCallDuration.Seconds(), 'f', -1, 64)
	}
	storedCdr, err := auth.AsStoredCdr()
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	var maxDur float64
	if err := self.Responder.GetDerivedMaxSessionTime(*storedCdr, &maxDur); err != nil {
		return err
	}
	if maxDur == -1.0 {
		*maxSessionTime = -1.0
		return nil
	}
	*maxSessionTime = time.Duration(maxDur).Seconds()
	return nil
}
