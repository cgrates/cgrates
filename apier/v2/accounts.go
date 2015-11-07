/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

package v2

import (
	"fmt"
	"math"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (self *ApierV2) GetAccounts(attr utils.AttrGetAccounts, reply *[]*engine.Account) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenanat")
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIds) == 0 {
		if accountKeys, err = self.AccountDb.GetKeysForPrefix(utils.ACCOUNT_PREFIX + utils.ConcatenatedKey(attr.Tenant)); err != nil {
			return err
		}
	} else {
		for _, acntId := range attr.AccountIds {
			if len(acntId) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			accountKeys = append(accountKeys, utils.ACCOUNT_PREFIX+utils.ConcatenatedKey(attr.Tenant, acntId))
		}
	}
	if len(accountKeys) == 0 {
		return nil
	}
	var limitedAccounts []string
	if attr.Limit != 0 {
		max := math.Min(float64(attr.Offset+attr.Limit), float64(len(accountKeys)))
		limitedAccounts = accountKeys[attr.Offset:int(max)]
	} else {
		limitedAccounts = accountKeys[attr.Offset:]
	}
	retAccounts := make([]*engine.Account, 0)
	for _, acntKey := range limitedAccounts {
		if acnt, err := self.AccountDb.GetAccount(acntKey[len(utils.ACCOUNT_PREFIX):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		} else if acnt != nil {
			retAccounts = append(retAccounts, acnt)
		}
	}
	*reply = retAccounts
	return nil
}

// Get balance
func (self *ApierV2) GetAccount(attr *utils.AttrGetAccount, reply *engine.Account) error {
	tag := fmt.Sprintf("%s:%s", attr.Tenant, attr.Account)
	account, err := self.AccountDb.GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = *account
	return nil
}
