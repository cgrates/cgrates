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
	"errors"

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
	if engine.GetAliasService() == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Subject", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, alias := range attrs.Aliases {
		var ignr string
		if err := engine.GetAliasService().SetAlias(
			engine.Alias{Direction: utils.META_OUT, Tenant: attrs.Tenant, Account: utils.META_ANY, Subject: attrs.Subject, Group: utils.ALIAS_GROUP_RP,
				Values: engine.AliasValues{&engine.AliasValue{DestinationId: utils.META_ANY, Alias: alias, Weight: 10.0}}}, &ignr); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return nil
}

/*
// Retrieve aliases configured for a rating profile subject
func (self *ApierV1) RemRatingSubjectAliases(tenantRatingSubject engine.TenantRatingSubject, reply *string) error {
	if engine.GetAliasService() == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	if missing := utils.MissingStructFields(&tenantRatingSubject, []string{"Tenant", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.RatingDb.RemoveRpAliases([]*engine.TenantRatingSubject{&tenantRatingSubject}, false); err != nil {
		if err == utils.ErrNotFound {
			return err
		}
		return utils.NewErrServerError(err)
	}

	// cache refresh not needed, synched in RemoveRpAliases
	*reply = utils.OK
	return nil
}
*/

func (self *ApierV1) AddAccountAliases(attrs AttrAddAccountAliases, reply *string) error {
	if engine.GetAliasService() == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, alias := range attrs.Aliases {
		var ignr string
		if err := engine.GetAliasService().SetAlias(
			engine.Alias{Direction: utils.META_OUT, Tenant: attrs.Tenant, Account: attrs.Account, Subject: utils.META_ANY, Group: utils.ALIAS_GROUP_ACC,
				Values: engine.AliasValues{&engine.AliasValue{DestinationId: utils.META_ANY, Alias: alias, Weight: 10.0}}}, &ignr); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return nil
}

/*
// Retrieve aliases configured for a rating profile subject
func (self *ApierV1) RemAccountAliases(tenantAccount engine.TenantAccount, reply *string) error {
	if engine.GetAliasService() == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	if missing := utils.MissingStructFields(&tenantAccount, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.RatingDb.RemoveAccAliases([]*engine.TenantAccount{&tenantAccount}, false); err != nil {
		if err == utils.ErrNotFound {
			return err
		}
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
*/
