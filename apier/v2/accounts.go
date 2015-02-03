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
