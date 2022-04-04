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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/accounts"
	"github.com/cgrates/cgrates/utils"
)

// GetAccount returns an Account
func (admS *AdminSv1) GetAccount(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	ap, err := admS.dm.GetAccount(ctx, tnt, arg.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ap
	return nil
}

// GetAccountsIDs returns list of account profile IDs registered for a tenant
func (admS *AdminSv1) GetAccountsIDs(ctx *context.Context, args *utils.ArgsItemIDs, actPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.AccountPrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*actPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetAccounts returns a list of accounts registered for a tenant
func (admS *AdminSv1) GetAccounts(ctx *context.Context, args *utils.ArgsItemIDs, accs *[]*utils.Account) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var accIDs []string
	if err = admS.GetAccountsIDs(ctx, args, &accIDs); err != nil {
		return
	}
	*accs = make([]*utils.Account, 0, len(accIDs))
	for _, accID := range accIDs {
		var acc *utils.Account
		acc, err = admS.dm.GetAccount(ctx, tnt, accID)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*accs = append(*accs, acc)
	}
	return
}

// GetAccountsCount sets in reply var the total number of AccountIDs registered for a tenant
// returns ErrNotFound in case of 0 AccountIDs
func (admS *AdminSv1) GetAccountsCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.AccountPrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetAccount add/update a new Account
func (admS *AdminSv1) SetAccount(ctx *context.Context, args *utils.AccountWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.Account, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.SetAccount(ctx, args.Account, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccountProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheAccounts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheAccounts,
		args.TenantID(), &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveAccount remove a specific Account
func (admS *AdminSv1) RemoveAccount(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveAccount(ctx, tnt, arg.ID,
		true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccountProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheAccounts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheAccounts,
		utils.ConcatenatedKey(tnt, arg.ID), nil, arg.APIOpts); err != nil {
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
	ping
}

// AccountsForEvent returns the matching Account for Event
func (aSv1 *AccountSv1) AccountsForEvent(ctx *context.Context, args *utils.CGREvent,
	aps *[]*utils.Account) (err error) {
	return aSv1.aS.V1AccountsForEvent(ctx, args, aps)
}

// MaxAbstracts returns the maximum abstracts for the event, based on matching Account
func (aSv1 *AccountSv1) MaxAbstracts(ctx *context.Context, args *utils.CGREvent,
	eEc *utils.EventCharges) (err error) {
	return aSv1.aS.V1MaxAbstracts(ctx, args, eEc)
}

// DebitAbstracts performs debit for the provided event
func (aSv1 *AccountSv1) DebitAbstracts(ctx *context.Context, args *utils.CGREvent,
	eEc *utils.EventCharges) (err error) {
	return aSv1.aS.V1DebitAbstracts(ctx, args, eEc)
}

// MaxConcretes returns the maximum concretes for the event, based on the matching Account
func (aSv1 *AccountSv1) MaxConcretes(ctx *context.Context, args *utils.CGREvent,
	eEc *utils.EventCharges) (err error) {
	return aSv1.aS.V1MaxConcretes(ctx, args, eEc)
}

// DebitConcretes performs debit of concrete units for the provided event
func (aSv1 *AccountSv1) DebitConcretes(ctx *context.Context, args *utils.CGREvent,
	eEc *utils.EventCharges) (err error) {
	return aSv1.aS.V1DebitConcretes(ctx, args, eEc)
}

// RefundCharges will refund charges recorded inside EventCharges
func (aSv1 *AccountSv1) RefundCharges(ctx *context.Context,
	args *utils.APIEventCharges, rply *string) (err error) {
	return aSv1.aS.V1RefundCharges(ctx, args, rply)
}

// ActionSetBalance performs a set balance action
func (aSv1 *AccountSv1) ActionSetBalance(ctx *context.Context, args *utils.ArgsActSetBalance,
	eEc *string) (err error) {
	return aSv1.aS.V1ActionSetBalance(ctx, args, eEc)
}

// ActionRemoveBalances removes a blance from an account
func (aSv1 *AccountSv1) ActionRemoveBalances(ctx *context.Context, args *utils.ArgsActRemoveBalances,
	eEc *string) (err error) {
	return aSv1.aS.V1ActionRemoveBalance(ctx, args, eEc)
}

// GetAccount returns an Account
func (aSv1 *AccountSv1) GetAccount(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	return nil
}
