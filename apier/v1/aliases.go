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
	Tenant, Category, Subject string
	Aliases                   []string
}

type AttrAddAccountAliases struct {
	Tenant, Category, Account string
	Aliases                   []string
}

// Add aliases configured for a rating profile subject <Deprecated>
func (self *ApierV1) AddRatingSubjectAliases(attrs AttrAddRatingSubjectAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Subject", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Category == "" {
		attrs.Category = utils.CALL
	}
	aliases := engine.GetAliasService()
	if aliases == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	var ignr string
	for _, alias := range attrs.Aliases {
		als := engine.Alias{Direction: utils.META_OUT, Tenant: attrs.Tenant, Category: attrs.Category, Account: alias, Subject: alias, Context: utils.ALIAS_CONTEXT_RATING,
			Values: engine.AliasValues{&engine.AliasValue{DestinationId: utils.META_ANY,
				Pairs: engine.AliasPairs{"Account": map[string]string{alias: attrs.Subject}, "Subject": map[string]string{alias: attrs.Subject}}, Weight: 10.0}}}
		if err := aliases.Call("AliasesV1.SetAlias", als, &ignr); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return nil
}

// Remove aliases configured for a rating profile subject
func (self *ApierV1) RemRatingSubjectAliases(tenantRatingSubject engine.TenantRatingSubject, reply *string) error {
	if missing := utils.MissingStructFields(&tenantRatingSubject, []string{"Tenant", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aliases := engine.GetAliasService()
	if aliases == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	var reverseAliases map[string][]*engine.Alias
	if err := aliases.Call("AliasesV1.GetReverseAlias", &engine.AttrReverseAlias{Target: "Subject", Alias: tenantRatingSubject.Subject, Context: utils.ALIAS_CONTEXT_RATING}, &reverseAliases); err != nil {
		return utils.NewErrServerError(err)
	}
	var ignr string
	for _, aliass := range reverseAliases {
		for _, alias := range aliass {
			if alias.Tenant != tenantRatingSubject.Tenant {
				continue // From another tenant
			}
			if err := aliases.Call("AliasesV1.RemoveAlias", alias, &ignr); err != nil {
				return utils.NewErrServerError(err)
			}
		}
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) AddAccountAliases(attrs AttrAddAccountAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Category == "" {
		attrs.Category = utils.CALL
	}
	aliases := engine.GetAliasService()
	if aliases == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	var ignr string
	for _, alias := range attrs.Aliases {
		als := engine.Alias{Direction: utils.META_OUT, Tenant: attrs.Tenant, Category: attrs.Category, Account: alias, Subject: alias, Context: utils.ALIAS_CONTEXT_RATING,
			Values: engine.AliasValues{&engine.AliasValue{DestinationId: utils.META_ANY,
				Pairs: engine.AliasPairs{"Account": map[string]string{alias: attrs.Account}, "Subject": map[string]string{alias: attrs.Account}}, Weight: 10.0}}}
		if err := aliases.Call("AliasesV1.SetAlias", als, &ignr); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return nil
}

// Remove aliases configured for an account
func (self *ApierV1) RemAccountAliases(tenantAccount engine.TenantAccount, reply *string) error {
	if missing := utils.MissingStructFields(&tenantAccount, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aliases := engine.GetAliasService()
	if aliases == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	var reverseAliases map[string][]*engine.Alias
	if err := aliases.Call("AliasesV1.GetReverseAlias", &engine.AttrReverseAlias{Target: "Account", Alias: tenantAccount.Account, Context: utils.ALIAS_CONTEXT_RATING}, &reverseAliases); err != nil {
		return utils.NewErrServerError(err)
	}
	var ignr string
	for _, aliass := range reverseAliases {
		for _, alias := range aliass {
			if alias.Tenant != tenantAccount.Tenant {
				continue // From another tenant
			}
			if err := aliases.Call("AliasesV1.RemoveAlias", alias, &ignr); err != nil {
				return utils.NewErrServerError(err)
			}
		}
	}
	*reply = utils.OK
	return nil
}
