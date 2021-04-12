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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/accounts"

	"github.com/cgrates/cgrates/utils"
)

// GetAccount returns an Account
func (apierSv1 *APIerSv1) GetAccount(arg *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ap, err := apierSv1.DataManager.GetAccount(tnt, arg.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ap
	return nil
}

// GetAccountIDs returns list of action profile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetAccountIDs(args *utils.PaginatorWithTenant, actPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.AccountPrefix + tnt + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*actPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetAccountIDsCount sets in reply var the total number of AccountIDs registered for a tenant
// returns ErrNotFound in case of 0 AccountIDs
func (apierSv1 *APIerSv1) GetAccountIDsCount(args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.AccountPrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetAccount add/update a new Account
func (apierSv1 *APIerSv1) SetAccount(extAp *utils.APIAccountWithOpts, reply *string) error {
	if missing := utils.MissingStructFields(extAp.APIAccount, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if extAp.Tenant == utils.EmptyString {
		extAp.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ap, err := extAp.AsAccount()
	if err != nil {
		return err
	}
	if err := apierSv1.DataManager.SetAccount(ap, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccounts and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAccounts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveAccount remove a specific Account
func (apierSv1 *APIerSv1) RemoveAccount(arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveAccount(tnt, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccounts and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAccounts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewAccountSv1 initializes AccountSv1
func NewAccountSv1(aS *accounts.AccountS) *AccountSv1 {
	return &AccountSv1{aS: aS}
}

// AccountSv1 exports RPC from RLs
type AccountSv1 struct {
	aS *accounts.AccountS
}

// Call implements birpc.ClientConnector interface for internal RPC
func (aSv1 *AccountSv1) Call(ctx *context.Context, serviceMethod string,
	args, reply interface{}) error {
	return utils.APIerRPCCall(aSv1, serviceMethod, args, reply)
}

// Ping return pong if the service is active
func (aSv1 *AccountSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

// AccountsForEvent returns the matching Account for Event
func (aSv1 *AccountSv1) AccountsForEvent(args *utils.ArgsAccountsForEvent,
	aps *[]*utils.Account) (err error) {
	return aSv1.aS.V1AccountsForEvent(args, aps)
}

// MaxAbstracts returns the maximum abstracts for the event, based on matching Account
func (aSv1 *AccountSv1) MaxAbstracts(args *utils.ArgsAccountsForEvent,
	eEc *utils.ExtEventCharges) (err error) {
	return aSv1.aS.V1MaxAbstracts(args, eEc)
}

// DebitAbstracts performs debit for the provided event
func (aSv1 *AccountSv1) DebitAbstracts(args *utils.ArgsAccountsForEvent,
	eEc *utils.ExtEventCharges) (err error) {
	return aSv1.aS.V1DebitAbstracts(args, eEc)
}

// MaxConcretes returns the maximum concretes for the event, based on the matching Account
func (aSv1 *AccountSv1) MaxConcretes(args *utils.ArgsAccountsForEvent,
	eEc *utils.ExtEventCharges) (err error) {
	return aSv1.aS.V1MaxConcretes(args, eEc)
}

// DebitConcretes performs debit of concrete units for the provided event
func (aSv1 *AccountSv1) DebitConcretes(args *utils.ArgsAccountsForEvent,
	eEc *utils.ExtEventCharges) (err error) {
	return aSv1.aS.V1DebitConcretes(args, eEc)
}

// ActionSetBalance performs a set balance action
func (aSv1 *AccountSv1) ActionSetBalance(args *utils.ArgsActSetBalance,
	eEc *string) (err error) {
	return aSv1.aS.V1ActionSetBalance(args, eEc)
}

// ActionRemoveBalance removes a balance from an account
func (aSv1 *AccountSv1) ActionRemoveBalance(args *utils.ArgsActRemoveBalances,
	eEc *string) (err error) {
	return aSv1.aS.V1ActionRemoveBalance(args, eEc)
}
