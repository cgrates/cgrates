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
package accounts

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// V1AccountsForEvent returns the matching Accounts for Event
func (aS *AccountS) V1AccountsForEvent(ctx *context.Context, args *utils.CGREvent, aps *[]*utils.Account) (err error) {
	var accIDs []string
	if accIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIDs,
		config.AccountsProfileIDsDftOpt, utils.OptsAccountsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var acnts utils.Accounts
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, ignFilters, false); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*aps = acnts.Accounts()
	return
}

// V1MaxAbstracts returns the maximum abstract units for the event, based on matching Accounts
func (aS *AccountS) V1MaxAbstracts(ctx *context.Context, args *utils.CGREvent, eEc *utils.EventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIDs,
		config.AccountsProfileIDsDftOpt, utils.OptsAccountsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var acnts utils.Accounts
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, ignFilters, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccounts(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(ctx, acnts, args, false, false); err != nil {
		return
	}
	*eEc = *procEC
	return
}

// V1DebitAbstracts performs debit for the provided event
func (aS *AccountS) V1DebitAbstracts(ctx *context.Context, args *utils.CGREvent, eEc *utils.EventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIDs,
		config.AccountsProfileIDsDftOpt, utils.OptsAccountsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var acnts utils.Accounts
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, ignFilters, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccounts(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(ctx, acnts, args, false, true); err != nil {
		return
	}
	*eEc = *procEC
	return
}

// V1MaxConcretes returns the maximum concrete units for the event, based on matching Accounts
func (aS *AccountS) V1MaxConcretes(ctx *context.Context, args *utils.CGREvent, eEc *utils.EventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIDs,
		config.AccountsProfileIDsDftOpt, utils.OptsAccountsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var acnts utils.Accounts
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, ignFilters, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccounts(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(ctx, acnts, args, true, false); err != nil {
		return
	}
	*eEc = *procEC
	return
}

// V1DebitConcretes performs debit of concrete units for the provided event
func (aS *AccountS) V1DebitConcretes(ctx *context.Context, args *utils.CGREvent, eEc *utils.EventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIDs,
		config.AccountsProfileIDsDftOpt, utils.OptsAccountsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var acnts utils.Accounts
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, ignFilters, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccounts(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(ctx, acnts, args, true, true); err != nil {
		return
	}
	*eEc = *procEC
	return
}

// V1RefundCharges will refund charges recorded inside EventCharges
func (aS *AccountS) V1RefundCharges(ctx *context.Context, args *utils.APIEventCharges, rply *string) (err error) {
	if err = aS.refundCharges(ctx, args.Tenant, args.EventCharges); err != nil {
		return
	}
	*rply = utils.OK
	return
}

// V1ActionSetBalance performs an update for a specific balance in account
func (aS *AccountS) V1ActionSetBalance(ctx *context.Context, args *utils.ArgsActSetBalance, rply *string) (err error) {
	if args.AccountID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AccountID)
	}
	if len(args.Diktats) == 0 {
		return utils.NewErrMandatoryIeMissing(utils.Diktats)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	if err = guardian.Guardian.Guard(ctx, func(ctx *context.Context) error {
		return actSetAccount(ctx, aS.dm, tnt, args.AccountID, args.Diktats, args.Reset)
	}, aS.cfg.GeneralCfg().LockingTimeout,
		utils.ConcatenatedKey(utils.CacheAccounts, tnt, args.AccountID)); err != nil {
		return
	}

	*rply = utils.OK
	return
}

// V1RemoveBalance removes a balance for a specific account
func (aS *AccountS) V1ActionRemoveBalance(ctx *context.Context, args *utils.ArgsActRemoveBalances, rply *string) (err error) {
	if args.AccountID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AccountID)
	}
	if len(args.BalanceIDs) == 0 {
		return utils.NewErrMandatoryIeMissing(utils.BalanceIDs)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	if err = guardian.Guardian.Guard(ctx, func(ctx *context.Context) error {
		qAcnt, err := aS.dm.GetAccount(ctx, tnt, args.AccountID)
		if err != nil {
			return err
		}
		for _, balID := range args.BalanceIDs {
			delete(qAcnt.Balances, balID)
		}
		return aS.dm.SetAccount(ctx, qAcnt, false)
	}, aS.cfg.GeneralCfg().LockingTimeout,
		utils.ConcatenatedKey(utils.CacheAccounts, tnt, args.AccountID)); err != nil {
		return
	}
	*rply = utils.OK
	return
}

func (aS *AccountS) V1GetAccount(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	ap, err := aS.dm.GetAccount(ctx, tnt, arg.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ap
	return nil
}
