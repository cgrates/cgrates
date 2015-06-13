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

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrAddRatingSubjectAliases struct {
	Tenant, Subject string
	Aliases         []string
}

type AttrAddAccountAliases struct {
	Tenant, Account string
	Aliases         []string
}

// Retrieve aliases configured for a rating profile subject
func (self *ApierV1) AddRatingSubjectAliases(attrs AttrAddRatingSubjectAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Subject", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aliasesChanged := []string{}
	for _, alias := range attrs.Aliases {
		if err := self.RatingDb.SetRpAlias(utils.RatingSubjectAliasKey(attrs.Tenant, alias), attrs.Subject); err != nil {
			return utils.NewErrServerError(err)
		}
		aliasesChanged = append(aliasesChanged, engine.RP_ALIAS_PREFIX+utils.RatingSubjectAliasKey(attrs.Tenant, alias))
	}
	didNotChange := []string{}
	if err := self.RatingDb.CacheRating(didNotChange, didNotChange, didNotChange, aliasesChanged, didNotChange, didNotChange); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Retrieve aliases configured for a rating profile subject
func (self *ApierV1) GetRatingSubjectAliases(attrs engine.TenantRatingSubject, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if aliases, err := self.RatingDb.GetRPAliases(attrs.Tenant, attrs.Subject, false); err != nil {
		return utils.NewErrServerError(err)
	} else if len(aliases) == 0 { // Need it since otherwise we get some unexpected errrors in the client
		return utils.ErrNotFound
	} else {
		*reply = aliases
	}
	return nil
}

// Retrieve aliases configured for a rating profile subject
func (self *ApierV1) RemRatingSubjectAliases(tenantRatingSubject engine.TenantRatingSubject, reply *string) error {
	if missing := utils.MissingStructFields(&tenantRatingSubject, []string{"Tenant", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.RatingDb.RemoveRpAliases([]*engine.TenantRatingSubject{&tenantRatingSubject}); err != nil {
		return utils.NewErrServerError(err)
	}
	didNotChange := []string{}
	if err := self.RatingDb.CacheRating(didNotChange, didNotChange, didNotChange, nil, didNotChange, didNotChange); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) AddAccountAliases(attrs AttrAddAccountAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aliasesChanged := []string{}
	for _, alias := range attrs.Aliases {
		if err := self.AccountDb.SetAccAlias(utils.AccountAliasKey(attrs.Tenant, alias), attrs.Account); err != nil {
			return utils.NewErrServerError(err)
		}
		aliasesChanged = append(aliasesChanged, engine.ACC_ALIAS_PREFIX+utils.AccountAliasKey(attrs.Tenant, alias))
	}
	didNotChange := []string{}
	if err := self.AccountDb.CacheAccounting(didNotChange, didNotChange, aliasesChanged); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Retrieve aliases configured for an account
func (self *ApierV1) GetAccountAliases(attrs engine.TenantAccount, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if aliases, err := self.AccountDb.GetAccountAliases(attrs.Tenant, attrs.Account, false); err != nil {
		return utils.NewErrServerError(err)
	} else if len(aliases) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = aliases
	}
	return nil
}

// Retrieve aliases configured for a rating profile subject
func (self *ApierV1) RemAccountAliases(tenantAccount engine.TenantAccount, reply *string) error {
	if missing := utils.MissingStructFields(&tenantAccount, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.AccountDb.RemoveAccAliases([]*engine.TenantAccount{&tenantAccount}); err != nil {
		return utils.NewErrServerError(err)
	}
	didNotChange := []string{}
	if err := self.AccountDb.CacheAccounting(didNotChange, didNotChange, nil); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
