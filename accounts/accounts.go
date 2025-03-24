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
	"cmp"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// NewAccountS instantiates the AccountS
func NewAccountS(cfg *config.CGRConfig, fltrS *engine.FilterS,
	connMgr *engine.ConnManager, dm *engine.DataManager) *AccountS {
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

// matchingAccountsForEvent returns the matched Accounts for the given event
// if lked option is passed, each Account will be also locked
//
//	so it becomes responsibility of upper layers to release the lock
func (aS *AccountS) matchingAccountsForEvent(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	accIDs []string, ignoreFilters, lked bool) (accs utils.Accounts, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}
	if len(accIDs) == 0 {
		ignoreFilters = false
		var accIDsMap utils.StringSet
		if accIDsMap, err = engine.MatchingItemIDsForEvent(
			ctx,
			evNm,
			aS.cfg.AccountSCfg().StringIndexedFields,
			aS.cfg.AccountSCfg().PrefixIndexedFields,
			aS.cfg.AccountSCfg().SuffixIndexedFields,
			aS.cfg.AccountSCfg().ExistsIndexedFields,
			aS.cfg.AccountSCfg().NotExistsIndexedFields,
			aS.dm,
			utils.CacheAccountsFilterIndexes,
			tnt,
			aS.cfg.AccountSCfg().IndexedSelects,
			aS.cfg.AccountSCfg().NestedFields,
		); err != nil {
			return
		}
		accIDs = accIDsMap.AsSlice()
	}
	weights := make(map[string]float64) // stores sorting weights by acntID
	for _, accID := range accIDs {
		var refID string
		if lked {
			refID = guardian.Guardian.GuardIDs(utils.EmptyString,
				aS.cfg.GeneralCfg().LockingTimeout,
				utils.ConcatenatedKey(utils.CacheAccounts, tnt, accID)) // RPC caching needs to be atomic
		}
		var acc *utils.Account
		if acc, err = aS.dm.GetAccount(ctx, tnt, accID); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			unlockAccounts(accs) // in case of errors will not have unlocks in upper layers
			return
		}
		if !ignoreFilters {
			var pass bool
			if pass, err = aS.fltrS.Pass(ctx, tnt, acc.FilterIDs, evNm); err != nil {
				guardian.Guardian.UnguardIDs(refID)
				unlockAccounts(accs)
				return
			} else if !pass {
				guardian.Guardian.UnguardIDs(refID)
				continue
			}
		}
		weight, err := engine.WeightFromDynamics(ctx, acc.Weights, aS.fltrS, cgrEv.Tenant, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(refID)
			unlockAccounts(accs)
			return nil, err
		}
		weights[acc.ID] = weight
		accs = append(accs, &utils.AccountWithLock{
			Account: acc,
			LockID:  refID,
		})
	}
	if len(accs) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(accs, func(a, b *utils.AccountWithLock) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	return
}

// accountsDebit will debit an usage out of multiple accounts
// concretes parameter limits the debits to concrete only balances
// store is used for simulate only or complete debit
func (aS *AccountS) accountsDebit(ctx *context.Context, acnts []*utils.AccountWithLock,
	cgrEv *utils.CGREvent, concretes, store bool) (ec *utils.EventCharges, err error) {
	var usage *decimal.Big // total event usage
	if usage, err = engine.GetDecimalBigOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, aS.fltrS, aS.cfg.AccountSCfg().Opts.Usage,
		utils.OptsAccountsUsage, utils.MetaUsage); err != nil {
		return
	}
	dbted := decimal.New(0, 0) // amount debited so far
	acntBkps := make([]utils.AccountBalancesBackup, len(acnts))
	defer func() { // make sure we revert debits if errors occured
		if err != nil && store {
			restoreAccounts(ctx, aS.dm, acnts, acntBkps)
		}
	}()
	cgrEvDP := cgrEv.AsDataProvider()
	for i, acnt := range acnts {
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debits
		}
		acntBkps[i] = acnt.Account.AccountBalancesBackup()
		var ecDbt *utils.EventCharges
		if ecDbt, err = aS.accountDebit(ctx, acnt.Account,
			utils.CloneDecimalBig(usage), cgrEv, concretes, dbted); err != nil {
			return
		}
		if ecDbt == nil { // no balance matched
			continue
		}
		if ec == nil { // no debit performed yet
			ec = utils.NewEventCharges()
		}
		if store && acnt.Account.BalancesAltered(acntBkps[i]) {
			if err = aS.dm.SetAccount(ctx, acnt.Account, false); err != nil {
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
		// check for blockers for every profile
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, acnt.Blockers, aS.fltrS,
			cgrEv.Tenant, cgrEvDP); err != nil {
			return
		}
		// if blockers active, do not debit from the other accounts
		if blocker {
			break
		}
	}
	var forceUsage bool
	values, err := cgrEvDP.FieldAsInterface([]string{utils.MetaOpts})
	if err != nil {
		return
	}
	opts, err := engine.ConvertOptsToMapStringAny(values)
	if err != nil {
		return
	}
	if opt, has := opts[utils.OptsAccountsForceUsage]; has {
		forceUsage, err = utils.IfaceAsBool(opt)
		if err != nil {
			return
		}
	}
	if usage.Cmp(dbted) == 1 && forceUsage {
		err = utils.ErrInsufficientCredit
		return
	}
	return
}

// accountDebit will debit the usage out of an Account
func (aS *AccountS) accountDebit(ctx *context.Context, acnt *utils.Account, usage *decimal.Big,
	cgrEv *utils.CGREvent, concretes bool, dbted *decimal.Big) (ec *utils.EventCharges, err error) {
	// Find balances matching event
	blcsWithWeight := make(utils.BalancesWithWeight, 0, len(acnt.Balances))
	for _, blnCfg := range acnt.Balances {
		if !utils.HasPrefixSlice([]string{utils.MetaConcrete, utils.MetaAbstract}, blnCfg.Type) {
			continue // only concrete and abstracts will participate in debits
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(ctx, blnCfg.Weights,
			aS.fltrS, cgrEv.Tenant, cgrEv.AsDataProvider()); err != nil {
			return
		}
		blcsWithWeight = append(blcsWithWeight, &utils.BalanceWithWeight{Balance: blnCfg, Weight: weight})
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
		if ecDbt, err = debFunc(ctx, utils.CloneDecimalBig(usage), cgrEv, dbted); err != nil {
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
		// check the blocker for every balance in order to continue debiting from balances or not
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, blncOper.balanceCfg().Blockers, aS.fltrS, cgrEv.Tenant, cgrEv.AsDataProvider()); err != nil {
			return
		}
		// if blockers active, do not debit from the other balances
		if blocker {
			break
		}
	}
	return
}

// refundCharges implements the mechanism of refunding the charges into accounts
func (aS *AccountS) refundCharges(ctx *context.Context, tnt string, ecs *utils.EventCharges) (err error) {
	acnts := make(utils.Accounts, 0, len(ecs.Accounts))
	defer unlockAccounts(acnts) // no unlocking in upper layers

	acntsIdxed := make(map[string]*utils.Account) // so we can access Account easier
	alteredAcnts := make(utils.StringSet)         // hold here the list of modified accounts
	for acntID := range ecs.Accounts {
		refID := guardian.Guardian.GuardIDs(utils.EmptyString,
			aS.cfg.GeneralCfg().LockingTimeout,
			utils.ConcatenatedKey(utils.CacheAccounts, tnt, acntID))
		var qAcnt *utils.Account
		if qAcnt, err = aS.dm.GetAccount(ctx, tnt, acntID); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			if err == utils.ErrNotFound { // Account was removed in the mean time
				err = nil
				continue
			}
			return
		}
		acnts = append(acnts, &utils.AccountWithLock{Account: qAcnt, LockID: refID})
		acntsIdxed[acntID] = qAcnt
	}
	acntBkps := make([]utils.AccountBalancesBackup, len(acnts)) // so we can restore in case of issues
	for i, acnt := range acnts {
		acntBkps[i] = acnt.AccountBalancesBackup()
	}
	for _, chrg := range ecs.Charges {
		acntChrg := ecs.Accounting[chrg.ChargingID]
		var uf *utils.UnitFactor
		if acntChrg.UnitFactorID != utils.EmptyString {
			uf = ecs.UnitFactors[acntChrg.UnitFactorID]
		}
		if acntChrg.BalanceID != utils.MetaMockAbstract { // *transabstracts is not a real balance, hence the exception
			refundUnitsOnAccount(
				acntsIdxed[acntChrg.AccountID],
				uncompressUnits(acntChrg.Units, chrg.CompressFactor, acntChrg, uf),
				ecs.Accounts[acntChrg.AccountID].Balances[acntChrg.BalanceID])
			alteredAcnts.Add(acntChrg.AccountID)
		}
		for _, chrgID := range acntChrg.JoinedChargeIDs { // refund extra charges
			extraChrg := ecs.Accounting[chrgID]
			var joinedChargeUf *utils.UnitFactor
			if extraChrg.UnitFactorID != utils.EmptyString {
				joinedChargeUf = ecs.UnitFactors[extraChrg.UnitFactorID]
			}
			refundUnitsOnAccount(
				acntsIdxed[extraChrg.AccountID],
				uncompressUnits(extraChrg.Units, chrg.CompressFactor, extraChrg, joinedChargeUf),
				ecs.Accounts[acntChrg.AccountID].Balances[extraChrg.BalanceID])
			alteredAcnts.Add(extraChrg.AccountID)
		}
	}
	for acntID := range alteredAcnts {
		if err = aS.dm.SetAccount(ctx, acntsIdxed[acntID], false); err != nil {
			restoreAccounts(ctx, aS.dm, acnts, acntBkps)
			return
		}
	}
	return
}
