/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// NewAccountS instantiates the AccountS
func NewAccountS(cfg *config.CGRConfig, fltrS *engine.FilterS, connMgr *engine.ConnManager, dm *engine.DataManager) *AccountS {
	return &AccountS{cfg, fltrS, connMgr, dm}
}

// AccountS operates Accounts
type AccountS struct {
	cfg     *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
	dm      *engine.DataManager
}

// ListenAndServe keeps the service alive
func (aS *AccountS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case rld := <-cfgRld: // configuration was reloaded
			cfgRld <- rld
		}
	}
}

// Shutdown is called to shutdown the service
func (aS *AccountS) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.AccountS))
}

// matchingAccountsForEvent returns the matched Accounts for the given event
// if lked option is passed, each Account will be also locked
//   so it becomes responsibility of upper layers to release the lock
func (aS *AccountS) matchingAccountsForEvent(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	acntIDs []string, lked bool) (acnts utils.AccountsWithWeight, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}
	if len(acntIDs) == 0 {
		var actIDsMp utils.StringSet
		if actIDsMp, err = engine.MatchingItemIDsForEvent(
			ctx,
			evNm,
			aS.cfg.AccountSCfg().StringIndexedFields,
			aS.cfg.AccountSCfg().PrefixIndexedFields,
			aS.cfg.AccountSCfg().SuffixIndexedFields,
			aS.dm,
			utils.CacheAccountsFilterIndexes,
			tnt,
			aS.cfg.AccountSCfg().IndexedSelects,
			aS.cfg.AccountSCfg().NestedFields,
		); err != nil {
			return
		}
		acntIDs = actIDsMp.AsSlice()
	}
	for _, acntID := range acntIDs {
		var refID string
		if lked {
			cacheKey := utils.ConcatenatedKey(utils.CacheAccounts, tnt, acntID)
			refID = guardian.Guardian.GuardIDs("",
				aS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		}
		var qAcnt *utils.Account
		if qAcnt, err = aS.dm.GetAccount(ctx, tnt, acntID); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			unlockAccounts(acnts) // in case of errors will not have unlocks in upper layers
			return
		}
		var pass bool
		if pass, err = aS.fltrS.Pass(ctx, tnt, qAcnt.FilterIDs, evNm); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			unlockAccounts(acnts)
			return
		} else if !pass {
			guardian.Guardian.UnguardIDs(refID)
			continue
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(ctx, qAcnt.Weights,
			aS.fltrS, cgrEv.Tenant, evNm); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			unlockAccounts(acnts)
			return
		}
		acnts = append(acnts, &utils.AccountWithWeight{qAcnt, weight, refID})
	}
	if len(acnts) == 0 {
		return nil, utils.ErrNotFound
	}
	acnts.Sort()
	return
}

// accountsDebit will debit an usage out of multiple accounts
func (aS *AccountS) accountsDebit(ctx *context.Context, acnts []*utils.AccountWithWeight,
	cgrEv *utils.CGREvent, concretes, store bool) (ec *utils.EventCharges, err error) {
	var usage *decimal.Big
	if usage, err = engine.FilterDecimalBigCfgOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), aS.fltrS,
		aS.cfg.AccountSCfg().Opts.Usage); err != nil {
		return
	}
	if usage, err = cgrEv.OptsAsDecimal(usage, utils.OptsAccountsUsage,
		utils.MetaUsage); err != nil {
		return
	}
	dbted := decimal.New(0, 0)
	acntBkps := make([]utils.AccountBalancesBackup, len(acnts))
	for i, acnt := range acnts {
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debits
		}
		acntBkps[i] = acnt.Account.AccountBalancesBackup()
		var ecDbt *utils.EventCharges
		if ecDbt, err = aS.accountDebit(ctx, acnt.Account,
			new(decimal.Big).Copy(usage), cgrEv, concretes, dbted); err != nil {
			if store {
				restoreAccounts(ctx, aS.dm, acnts, acntBkps)
			}
			return
		}
		if ecDbt == nil {
			continue
		}
		if ec == nil { // no debit performed yet
			ec = utils.NewEventCharges()
		}
		if store && acnt.Account.BalancesAltered(acntBkps[i]) {
			if err = aS.dm.SetAccount(ctx, acnt.Account, false); err != nil {
				restoreAccounts(ctx, aS.dm, acnts, acntBkps)
				return
			}
		}
		var used *decimal.Big
		if concretes {
			used = ecDbt.Concretes.Big
		} else {
			used = ecDbt.Abstracts.Big
		}
		usage = utils.SubstractBig(usage, used)
		dbted = utils.SumBig(dbted, used)
		ec.Merge(ecDbt)
	}
	return
}

// accountDebit will debit the usage out of an Account
func (aS *AccountS) accountDebit(ctx *context.Context, acnt *utils.Account, usage *decimal.Big,
	cgrEv *utils.CGREvent, concretes bool, dbted *decimal.Big) (ec *utils.EventCharges, err error) {
	// Find balances matching event
	blcsWithWeight := make(utils.BalancesWithWeight, 0, len(acnt.Balances))
	for _, blnCfg := range acnt.Balances {
		var weight float64
		if weight, err = engine.WeightFromDynamics(ctx, blnCfg.Weights,
			aS.fltrS, cgrEv.Tenant, cgrEv.AsDataProvider()); err != nil {
			return
		}
		blcsWithWeight = append(blcsWithWeight, &utils.BalanceWithWeight{blnCfg, weight})
	}
	blcsWithWeight.Sort()
	var blncOpers []balanceOperator
	if blncOpers, err = newBalanceOperators(ctx, acnt.ID, blcsWithWeight.Balances(), aS.fltrS, aS.connMgr,
		aS.cfg.AccountSCfg().AttributeSConns, aS.cfg.AccountSCfg().RateSConns); err != nil {
		return
	}

	for _, blncOper := range blncOpers {
		debFunc := blncOper.debitAbstracts
		if concretes {
			debFunc = blncOper.debitConcretes
		}
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debit
		}
		var ecDbt *utils.EventCharges
		if ecDbt, err = debFunc(ctx, new(decimal.Big).Copy(usage), cgrEv, dbted); err != nil {
			if err == utils.ErrFilterNotPassingNoCaps ||
				err == utils.ErrNotImplemented {
				err = nil
				continue
			}
			return
		}
		if ecDbt == nil {
			continue // no debit performed
		}
		if ec == nil { // first debit
			ec = utils.NewEventCharges()
		}
		var used *decimal.Big
		if concretes {
			used = ecDbt.Concretes.Big
		} else {
			used = ecDbt.Abstracts.Big
		}
		usage = utils.SubstractBig(usage, used)
		dbted = utils.SumBig(dbted, used)
		ec.Merge(ecDbt)
		ec.Accounts[acnt.ID] = acnt
	}
	return
}

// V1AccountsForEvent returns the matching Accounts for Event
func (aS *AccountS) V1AccountsForEvent(ctx *context.Context, args *utils.CGREvent, aps *[]*utils.Account) (err error) {
	var accIDs []string
	if accIDs, err = engine.FilterStringSliceCfgOpts(ctx, args.Tenant, args.AsDataProvider(), aS.fltrS,
		aS.cfg.AccountSCfg().Opts.AccountIDs); err != nil {
		return
	}
	if accIDs, err = args.OptsAsStringSlice(accIDs, utils.OptsAccountsAccountIDs); err != nil {
		return
	}
	var acnts utils.AccountsWithWeight
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, false); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*aps = acnts.Accounts()
	return
}

// V1MaxAbstracts returns the maximum abstract units for the event, based on matching Accounts
func (aS *AccountS) V1MaxAbstracts(ctx *context.Context, args *utils.CGREvent, eEc *utils.ExtEventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.FilterStringSliceCfgOpts(ctx, args.Tenant, args.AsDataProvider(), aS.fltrS,
		aS.cfg.AccountSCfg().Opts.AccountIDs); err != nil {
		return
	}
	if accIDs, err = args.OptsAsStringSlice(accIDs, utils.OptsAccountsAccountIDs); err != nil {
		return
	}
	var acnts utils.AccountsWithWeight
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, true); err != nil {
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
	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}
	*eEc = *rcvEec
	return
}

// V1DebitAbstracts performs debit for the provided event
func (aS *AccountS) V1DebitAbstracts(ctx *context.Context, args *utils.CGREvent, eEc *utils.ExtEventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.FilterStringSliceCfgOpts(ctx, args.Tenant, args.AsDataProvider(), aS.fltrS,
		aS.cfg.AccountSCfg().Opts.AccountIDs); err != nil {
		return
	}
	if accIDs, err = args.OptsAsStringSlice(accIDs, utils.OptsAccountsAccountIDs); err != nil {
		return
	}
	var acnts utils.AccountsWithWeight
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, true); err != nil {
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
	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}

	*eEc = *rcvEec
	return
}

// V1MaxConcretes returns the maximum concrete units for the event, based on matching Accounts
func (aS *AccountS) V1MaxConcretes(ctx *context.Context, args *utils.CGREvent, eEc *utils.ExtEventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.FilterStringSliceCfgOpts(ctx, args.Tenant, args.AsDataProvider(), aS.fltrS,
		aS.cfg.AccountSCfg().Opts.AccountIDs); err != nil {
		return
	}
	if accIDs, err = args.OptsAsStringSlice(accIDs, utils.OptsAccountsAccountIDs); err != nil {
		return
	}
	var acnts utils.AccountsWithWeight
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, true); err != nil {
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
	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}
	*eEc = *rcvEec
	return
}

// V1DebitConcretes performs debit of concrete units for the provided event
func (aS *AccountS) V1DebitConcretes(ctx *context.Context, args *utils.CGREvent, eEc *utils.ExtEventCharges) (err error) {
	var accIDs []string
	if accIDs, err = engine.FilterStringSliceCfgOpts(ctx, args.Tenant, args.AsDataProvider(), aS.fltrS,
		aS.cfg.AccountSCfg().Opts.AccountIDs); err != nil {
		return
	}
	if accIDs, err = args.OptsAsStringSlice(accIDs, utils.OptsAccountsAccountIDs); err != nil {
		return
	}
	var acnts utils.AccountsWithWeight
	if acnts, err = aS.matchingAccountsForEvent(ctx, args.Tenant,
		args, accIDs, true); err != nil {
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

	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}

	*eEc = *rcvEec
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
