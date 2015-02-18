/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
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
)

// Get DerivedChargers applying to our call, appends general configured to account specific ones if that is configured
func (self *ApierV1) GetDerivedChargers(attrs utils.AttrDerivedChargers, reply *utils.DerivedChargers) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Direction", "Account", "Subject"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if hDc, err := engine.HandleGetDerivedChargers(self.AccountDb, attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if hDc != nil {
		*reply = hDc
	}
	return nil
}

type AttrSetDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject string
	DerivedChargers                               utils.DerivedChargers
}

func (self *ApierV1) SetDerivedChargers(attrs AttrSetDerivedChargers, reply *string) (err error) {
	if len(attrs.Direction) == 0 {
		attrs.Direction = utils.OUT
	}
	if len(attrs.Tenant) == 0 {
		attrs.Tenant = utils.ANY
	}
	if len(attrs.Category) == 0 {
		attrs.Category = utils.ANY
	}
	if len(attrs.Account) == 0 {
		attrs.Account = utils.ANY
	}
	if len(attrs.Subject) == 0 {
		attrs.Subject = utils.ANY
	}
	for _, dc := range attrs.DerivedChargers {
		if _, err = utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP); err != nil { // Make sure rules are OK before loading in db
			return fmt.Errorf("%s:%s", utils.ERR_PARSER_ERROR, err.Error())
		}
	}
	dcKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)
	if err := self.AccountDb.SetDerivedChargers(dcKey, attrs.DerivedChargers); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if err := self.AccountDb.CacheAccounting([]string{}, []string{}, []string{}, []string{engine.DERIVEDCHARGERS_PREFIX + dcKey}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = utils.OK
	return nil
}

type AttrRemDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject string
}

func (self *ApierV1) RemDerivedChargers(attrs AttrRemDerivedChargers, reply *string) error {
	if len(attrs.Direction) == 0 {
		attrs.Direction = utils.OUT
	}
	if len(attrs.Tenant) == 0 {
		attrs.Tenant = utils.ANY
	}
	if len(attrs.Category) == 0 {
		attrs.Category = utils.ANY
	}
	if len(attrs.Account) == 0 {
		attrs.Account = utils.ANY
	}
	if len(attrs.Subject) == 0 {
		attrs.Subject = utils.ANY
	}
	if err := self.AccountDb.SetDerivedChargers(utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject), nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	if err := self.AccountDb.CacheAccounting([]string{}, []string{}, []string{}, nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	return nil
}
