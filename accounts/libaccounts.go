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
	"errors"
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAccountBalances constructs accountBalances
func newBalanceOperators(ctx *context.Context, acntID string, blnCfgs []*utils.Balance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (blncOpers []balanceOperator, err error) {

	blncOpers = make([]balanceOperator, len(blnCfgs))
	var cncrtBlncs []*concreteBalance
	for i, blnCfg := range blnCfgs { // build the concrete balances
		if blnCfg.Type != utils.MetaConcrete {
			continue
		}
		blncOpers[i] = newConcreteBalanceOperator(ctx, acntID, blnCfg,
			fltrS, connMgr, attrSConns, rateSConns)
		cncrtBlncs = append(cncrtBlncs, blncOpers[i].(*concreteBalance))
	}

	for i, blnCfg := range blnCfgs { // build the abstract balances
		if blnCfg.Type == utils.MetaConcrete {
			continue
		}
		if blncOpers[i], err = newBalanceOperator(ctx, acntID, blnCfg, cncrtBlncs,
			fltrS, connMgr, attrSConns, rateSConns); err != nil {
			return
		}
	}

	return
}

// newBalanceOperator instantiates balanceOperator interface
// cncrtBlncs are needed for abstract balance debits
func newBalanceOperator(ctx *context.Context, acntID string, blncCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (bP balanceOperator, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MetaConcrete:
		return newConcreteBalanceOperator(ctx, acntID, blncCfg, fltrS, connMgr, attrSConns, rateSConns), nil
	case utils.MetaAbstract:
		return newAbstractBalanceOperator(ctx, acntID, blncCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns), nil
	}
}

// balanceOperator is the implementation of a balance type
type balanceOperator interface {
	id() string // balance id
	debitAbstracts(ctx *context.Context, usage *decimal.Big, cgrEv *utils.CGREvent, dbted *decimal.Big) (ec *utils.EventCharges, err error)
	debitConcretes(ctx *context.Context, usage *decimal.Big, cgrEv *utils.CGREvent, dbted *decimal.Big) (ec *utils.EventCharges, err error)
}

// roundUnitsWithIncrements rounds the usage based on increments
func roundUnitsWithIncrements(usage, incrm *decimal.Big) (rndedUsage *decimal.Big) {
	usgMaxIncrm := decimal.WithContext(
		decimal.Context{RoundingMode: decimal.ToZero}).Quo(usage,
		incrm).RoundToInt()
	rndedUsage = utils.MultiplyBig(usgMaxIncrm, incrm)
	return
}

// processAttributeS will process the event with AttributeS
func processAttributeS(ctx *context.Context, connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attrIDs []string) (rplyEv *engine.AttrSProcessEventReply, err error) {
	if len(attrSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AttributeS)
	}
	cgrEv.APIOpts[utils.OptsAttributesAttributeIDs] = attrIDs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaAccounts)
	var tmpReply engine.AttrSProcessEventReply
	if err = connMgr.Call(ctx, attrSConns, utils.AttributeSv1ProcessEvent,
		cgrEv, &tmpReply); err != nil {
		return
	}
	return &tmpReply, nil
}

// rateSCostForEvent will process the event with RateS in order to get the cost
func rateSCostForEvent(ctx *context.Context, connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	rateSConns, rpIDs []string) (rplyCost *utils.RateProfileCost, err error) {
	if len(rateSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	if _, has := cgrEv.APIOpts[utils.OptsRatesRateProfileIDs]; !has {
		cgrEv.APIOpts[utils.OptsRatesRateProfileIDs] = utils.CloneStringSlice(rpIDs)
	}
	var tmpReply utils.RateProfileCost
	if err = connMgr.Call(ctx, rateSConns, utils.RateSv1CostForEvent,
		cgrEv, &tmpReply); err != nil {
		return
	}
	return &tmpReply, nil
}

// costIncrement computes the costIncrement for the event
func costIncrement(ctx *context.Context, cfgCostIncrmts []*utils.CostIncrement,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	// no cost increment in our balance, return a default increment to query to rateS
	if len(cfgCostIncrmts) == 0 {
		return &utils.CostIncrement{
			Increment: utils.NewDecimal(1, 0),
		}, nil
	}
	for _, cIcrm := range cfgCostIncrmts {
		var pass bool
		if pass, err = fltrS.Pass(ctx, tnt, cIcrm.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		costIcrm = cIcrm
		break
	}
	return
}

// unitFactor detects the unitFactor for the event
func unitFactor(ctx *context.Context, cfgUnitFactors []*utils.UnitFactor,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uFcfg := range cfgUnitFactors {
		var pass bool
		if pass, err = fltrS.Pass(ctx, tnt, uFcfg.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		uF = uFcfg
		return
	}
	return
}

// balanceLimit returns the balance limit based on configuration
func balanceLimit(optsCfg map[string]interface{}) (bL *utils.Decimal, err error) {
	if _, isUnlimited := optsCfg[utils.MetaBalanceUnlimited]; isUnlimited {
		return // unlimited is nil pointer
	}
	if lmtIface, has := optsCfg[utils.MetaBalanceLimit]; has {
		flt64Lmt, canCast := lmtIface.(float64)
		if !canCast {
			return nil, errors.New("unsupported *balanceLimit format")
		}
		return utils.NewDecimalFromFloat64(flt64Lmt), nil
	}
	// nothing matched, return default
	bL = utils.NewDecimal(0, 0)
	return
}

// debitConcreteUnits debits concrete units out of concrete balances
// returns utils.ErrInsufficientCredit if complete usage cannot be debited
func debitConcreteUnits(ctx *context.Context, cUnits *decimal.Big,
	acntID string, cncrtBlncs []*concreteBalance,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {

	clnedUnts := cloneUnitsFromConcretes(cncrtBlncs)
	for _, cB := range cncrtBlncs {
		var ecCncrt *utils.EventCharges
		if ecCncrt, err = cB.debitConcretes(ctx, new(decimal.Big).Copy(cUnits), cgrEv, nil); err != nil {
			restoreUnitsFromClones(cncrtBlncs, clnedUnts)
			return nil, err
		}
		if ecCncrt == nil { // no debit performed
			continue
		}
		if ec == nil {
			ec = utils.NewEventCharges()
		}
		ec.Merge(ecCncrt)
		cUnits = utils.SubstractBig(cUnits, ecCncrt.Concretes.Big)
		if cUnits.Cmp(decimal.New(0, 0)) <= 0 {
			return // have debited all, total is smaller or equal to 0
		}
	}
	// we could not debit all, put back what we have debited
	restoreUnitsFromClones(cncrtBlncs, clnedUnts)
	return nil, utils.ErrInsufficientCredit
}

// maxDebitAbstractsFromConcretes will debit the maximum possible abstract units out of concretes
func maxDebitAbstractsFromConcretes(ctx *context.Context, aUnits *decimal.Big,
	acntID string, cncrtBlncs []*concreteBalance,
	connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attributeIDs, rateSConns, rpIDs []string,
	costIcrm *utils.CostIncrement, dbtedAUnts *decimal.Big) (ec *utils.EventCharges, err error) {
	// Init EventCharges
	if costIcrm == nil {
		costIcrm = &utils.CostIncrement{Increment: utils.NewDecimal(1, 0)}
	}
	calculateCost := costIcrm.RecurrentFee == nil && costIcrm.FixedFee == nil
	//var attrIDs []string // will be populated if attributes are processed successfully
	// process AttributeS if needed
	if calculateCost &&
		len(attributeIDs) != 0 && attributeIDs[0] != utils.MetaNone { // cost unknown, apply AttributeS to query from RateS
		var rplyAttrS *engine.AttrSProcessEventReply
		if rplyAttrS, err = processAttributeS(ctx, connMgr, cgrEv, attrSConns,
			attributeIDs); err != nil {
			return
		}
		if len(rplyAttrS.AlteredFields) != 0 { // event was altered
			cgrEv = rplyAttrS.CGREvent
			//attrIDs = rplyAttrS.MatchedProfiles
		}
	}
	// fix the maximum number of iterations
	origConcrtUnts := cloneUnitsFromConcretes(cncrtBlncs) // so we can revert on errors
	paidConcrtUnts := origConcrtUnts                      // so we can revert when higher abstracts are not possible
	var aPaid, aDenied *decimal.Big
	maxItr := config.CgrConfig().AccountSCfg().MaxIterations
	for i := 0; i <= maxItr; i++ {
		if i != 0 {
			restoreUnitsFromClones(cncrtBlncs, origConcrtUnts)
		}
		if i == maxItr {
			utils.Logger.Debug(fmt.Sprintf(
				"<%s> MaximumIncrementsExceeded: %d when debitting abstract units: %s, paid units: %s, deniedUnits: %s, costIncrement: %s, event: %s",
				utils.AccountS, i, aUnits, aPaid, aDenied, utils.ToJSON(costIcrm), utils.ToJSON(cgrEv)))
			return nil, utils.ErrMaxIncrementsExceeded
		}
		if calculateCost {
			var rplyCost *utils.RateProfileCost
			cgrEv.APIOpts[utils.OptsRatesIntervalStart] = dbtedAUnts
			cgrEv.APIOpts[utils.OptsRatesUsage] = aUnits
			if rplyCost, err = rateSCostForEvent(ctx, connMgr, cgrEv, rateSConns, rpIDs); err != nil {
				err = utils.NewErrRateS(err)
				return
			}
			// cleanup the opts
			delete(cgrEv.APIOpts, utils.OptsRatesIntervalStart)
			delete(cgrEv.APIOpts, utils.OptsRatesUsage)
			costIcrm = costIcrm.Clone() // so we don't modify the original
			costIcrm.FixedFee = rplyCost.Cost
		}
		var cUnits *decimal.Big // concrete units to debit
		if costIcrm.FixedFee != nil {
			cUnits = costIcrm.FixedFee.Big
		}
		// RecurrentFee is configured, used it with increments
		if costIcrm.RecurrentFee != nil {
			rcrntCost := utils.MultiplyBig(
				utils.DivideBig(aUnits, costIcrm.Increment.Big),
				costIcrm.RecurrentFee.Big)
			if cUnits == nil {
				cUnits = rcrntCost
			} else {
				cUnits = utils.SumBig(cUnits, rcrntCost)
			}
		}
		aQried := aUnits // so we can detect loops
		var ecDbt *utils.EventCharges
		if ecDbt, err = debitConcreteUnits(ctx, cUnits, acntID, cncrtBlncs, cgrEv); err != nil {
			if err != utils.ErrInsufficientCredit {
				return
			}
			err = nil
			// ErrInsufficientCredit
			aDenied = new(decimal.Big).Copy(aUnits)
			if aPaid == nil { // going backwards
				aUnits = utils.DivideBig( // divide by 2
					aUnits, decimal.New(2, 0))
				aUnits = roundUnitsWithIncrements(aUnits, costIcrm.Increment.Big) // make sure abstracts are multiple of increments
				if aUnits.Cmp(aDenied) >= 0 ||
					aUnits.Cmp(decimal.New(0, 0)) == 0 ||
					aUnits.Cmp(aQried) == 0 { // loop
					break
				}
				continue
			}
		} else { // debit for the usage succeeded
			aPaid = new(decimal.Big).Copy(aUnits)
			paidConcrtUnts = cloneUnitsFromConcretes(cncrtBlncs)
			if ec == nil {
				ec = utils.NewEventCharges()
			}
			ec.Merge(ecDbt)
			if i == 0 { // no estimation done, covering full
				break
			}
		}
		// going upwards
		aUnits = utils.SumBig(aPaid,
			utils.DivideBig(aPaid, decimal.New(2, 0)).RoundToInt())
		if aUnits.Cmp(aDenied) >= 0 {
			aUnits = utils.SumBig(aPaid, costIcrm.Increment.Big)
		}
		aUnits = roundUnitsWithIncrements(aUnits, costIcrm.Increment.Big)
		if aUnits.Cmp(aPaid) <= 0 ||
			aUnits.Cmp(aDenied) >= 0 ||
			aUnits.Cmp(aQried) == 0 { // loop
			break
		}
	}
	// Nothing paid
	if aPaid == nil {
		// since we are erroring, we restore the concerete balances
		aPaid = decimal.New(0, 0)
		ec = utils.NewEventCharges()
	}
	ec.Abstracts = &utils.Decimal{aPaid}
	restoreUnitsFromClones(cncrtBlncs, paidConcrtUnts)
	return
}

// restoreAccounts will restore the accounts in DataDB out of their backups if present
func restoreAccounts(ctx *context.Context, dm *engine.DataManager,
	acnts []*utils.AccountWithWeight, bkps []utils.AccountBalancesBackup) {
	for i, bkp := range bkps {
		if bkp == nil ||
			!acnts[i].Account.BalancesAltered(bkp) {
			continue
		}
		acnts[i].Account.RestoreFromBackup(bkp)
		if err := dm.SetAccount(ctx, acnts[i].Account, false); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> restoring account <%s>",
				utils.AccountS, err, acnts[i].Account.TenantID()))
		}
	}
}

// unlockAccounts is used to unlock the accounts based on their lock identifiers
func unlockAccounts(acnts utils.AccountsWithWeight) {
	for _, lkID := range acnts.LockIDs() {
		guardian.Guardian.UnguardIDs(lkID)
	}
}
