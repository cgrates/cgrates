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

package v2

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"strings"
)

type AttrGetAccountIds struct {
	Page         int
	ItemsPerPage int
	SearchTerm   string
}

func (self *ApierV2) GetAccountIds(attrs AttrGetAccountIds, reply *[]string) error {
	prefix := engine.ACCOUNT_PREFIX
	if attrs.SearchTerm != "" {
		prefix += "*" + attrs.SearchTerm
	}
	accountKeys, err := self.AccountDb.GetKeysForPrefix(prefix)
	if err != nil {
		return err
	}
	*reply = accountKeys
	return nil
}

type AttrGetAccounts struct {
	Tenant  string
	Account string
	Offset  int // Set the item offset
	Limit   int // Limit number of items retrieved

}

func (self *ApierV2) GetAccounts(attr AttrGetAccounts, reply *[]*engine.Account) error {
	searchKeyPrefix := engine.ACCOUNT_PREFIX + utils.OUT + ":"
	if len(attr.Tenant) != 0 { // ToDO: Update here as soon as redis 2.8 becomes last supported platform
		searchKeyPrefix += attr.Tenant + ":"
		if len(attr.Account) != 0 {
			searchKeyPrefix += attr.Account
		}
	}
	accountKeys, err := self.AccountDb.GetKeysForPrefix(searchKeyPrefix)
	if err != nil {
		return err
	} else if len(accountKeys) == 0 {
		return nil
	}
	if len(attr.Tenant) == 0 && len(attr.Account) != 0 { // Since redis version lower than 2.8 does not support masked searches in middle, we filter records out here
		filteredAccounts := make([]string, 0)
		for _, acntKey := range accountKeys {
			if strings.HasSuffix(acntKey, ":"+attr.Account) {
				filteredAccounts = append(filteredAccounts, acntKey)
			}
		}
		accountKeys = filteredAccounts
	}
	var limitedAccounts []string
	if attr.Limit != 0 {
		limitedAccounts = accountKeys[attr.Offset : attr.Offset+attr.Limit]
	} else {
		limitedAccounts = accountKeys[attr.Offset:]
	}
	retAccounts := make([]*engine.Account, len(limitedAccounts))
	for idx, acntKey := range limitedAccounts {
		retAccounts[idx], err = self.AccountDb.GetAccount(acntKey[len(engine.ACCOUNT_PREFIX):])
		if err != nil {
			return err
		}
	}
	*reply = retAccounts
	return nil
}
