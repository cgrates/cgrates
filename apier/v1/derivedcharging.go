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
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Get DerivedChargers applying to our call, appends general configured to account specific ones if that is configured
func (self *ApierV1) GetDerivedChargers(attrs utils.AttrDerivedChargers, reply *utils.DerivedChargers) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Direction", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if hDc, err := engine.HandleGetDerivedChargers(self.DataManager, &attrs); err != nil {
		return utils.NewErrServerError(err)
	} else if hDc != nil {
		*reply = *hDc
	}
	return nil
}

type AttrSetDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject, DestinationIds string
	DerivedChargers                                               []*utils.DerivedCharger
	Overwrite                                                     bool // Do not overwrite if present in redis
}

func (self *ApierV1) SetDerivedChargers(attrs AttrSetDerivedChargers, reply *string) (err error) {
	if len(attrs.DerivedChargers) == 0 {
		return utils.NewErrMandatoryIeMissing("DerivedChargers")
	}
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
			return fmt.Errorf("%s:%s", utils.ErrParserError.Error(), err.Error())
		}
	}
	dcKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)
	if !attrs.Overwrite {
		if exists, err := self.DataManager.HasData(utils.DERIVEDCHARGERS_PREFIX, dcKey, ""); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	dstIds := strings.Split(attrs.DestinationIds, utils.INFIELD_SEP)
	dcs := &utils.DerivedChargers{DestinationIDs: utils.NewStringMap(dstIds...), Chargers: attrs.DerivedChargers}
	if err := self.DataManager.DataDB().SetDerivedChargers(dcKey, dcs, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := self.DataManager.CacheDataFromDB(utils.DERIVEDCHARGERS_PREFIX, []string{dcKey}, true); err != nil {
		return utils.NewErrServerError(err)
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
	if err := self.DataManager.DataDB().SetDerivedChargers(utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject), nil, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := self.DataManager.CacheDataFromDB(utils.DERIVEDCHARGERS_PREFIX,
		[]string{utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
