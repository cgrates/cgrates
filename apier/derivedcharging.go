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

package apier

import (
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Get DerivedChargers applying to our call, appends general configured to account specific ones if that is configured
func (self *ApierV1) DerivedChargers(attrs utils.AttrDerivedChargers, reply *utils.DerivedChargers) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Direction", "Account", "Subject"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if hDc, err := engine.HandleGetDerivedChargers(self.AccountDb, self.Config, attrs); err != nil {
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

func (self *ApierV1) SetDerivedChargers(attrs AttrSetDerivedChargers, reply *utils.DerivedChargers) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Category", "Direction", "Account", "Subject", "DerivedChargers"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.AccountDb.SetDerivedChargers(utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject),
		attrs.DerivedChargers); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	return nil
}

/*
type AttrRemDerivedChargers struct {
	Tenant, Category, Direction, Account, Subject string
}

func (self *ApierV1) RemDerivedChargers(attrs AttrRemDerivedChargers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, ]string{"Tenant", "Category", "Direction", "Account", "Subject"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.DerivedChargersKey(attrs.Tenant, attrs.Category, attrs.Direction, attrs.Account, attrs.Subject), nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
*/
