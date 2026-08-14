// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package accounts

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAccountBalances constructs accountBalances
// expects the balances to be ordered and it will keep them that way
func newBalanceOperators(ctx *context.Context, cfg *config.CGRConfig, acntID string, blnCfgs []*utils.Balance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (blncOpers []balanceOperator, err error) {

	blncOpers = make([]balanceOperator, len(blnCfgs))
	var cncrtBlncs []*concreteBalance
	for i, blnCfg := range blnCfgs { // build the concrete balances
		if blnCfg.Type != utils.MetaConcrete {
			continue
		}
		blncOpers[i] = newConcreteBalanceOperator(ctx, cfg, acntID, blnCfg,
			fltrS, connMgr, attrSConns, rateSConns)
		cncrtBlncs = append(cncrtBlncs, blncOpers[i].(*concreteBalance))
	}

	for i, blnCfg := range blnCfgs { // build the abstract balances
		if blnCfg.Type != utils.MetaAbstract {
			continue
		}
		blncOpers[i] = newAbstractBalanceOperator(ctx, cfg, acntID, blnCfg, cncrtBlncs,
			fltrS, connMgr, attrSConns, rateSConns)
	}
	return
}

// newBalanceOperator instantiates balanceOperator interface
// cncrtBlncs are needed for abstract balance debits
func newBalanceOperator(ctx *context.Context, cfg *config.CGRConfig, acntID string, blncCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (bP balanceOperator, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MetaConcrete:
		return newConcreteBalanceOperator(ctx, cfg, acntID, blncCfg, fltrS, connMgr, attrSConns, rateSConns), nil
	case utils.MetaAbstract:
		return newAbstractBalanceOperator(ctx, cfg, acntID, blncCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns), nil
	}
}

// balanceOperator is the implementation of a balance type
type balanceOperator interface {
	id() string // balance id
	debitAbstracts(ctx *context.Context, usage *utils.Decimal, cgrEv *utils.CGREvent, dbted *utils.Decimal) (ec *utils.EventCharges, err error)
	debitConcretes(ctx *context.Context, usage *utils.Decimal, cgrEv *utils.CGREvent, dbted *utils.Decimal) (ec *utils.EventCharges, err error)
	balanceCfg() (bal *utils.Balance)
}

// roundUnitsWithIncrements rounds the usage based on increments
func roundUnitsWithIncrements(usage, incrm *utils.Decimal) *utils.Decimal {
	usgMaxIncrm := decimal.WithContext(
		decimal.Context{RoundingMode: decimal.ToZero}).Quo(usage.Big,
		incrm.Big).RoundToInt()
	return utils.MultiplyDecimal(utils.NewDecimalFromBig(usgMaxIncrm), incrm)
}

// processAttributeS will process the event with AttributeS
func processAttributeS(ctx *context.Context, connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attrIDs []string) (rplyEv *attributes.ProcessEventReply, err error) {
	if len(attrSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AttributeS)
	}
	cgrEv.APIOpts[utils.OptsAttributesProfileIDs] = attrIDs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaAccounts)
	var tmpReply attributes.ProcessEventReply
	if err = connMgr.Call(ctx, attrSConns, utils.AttributeSv1ProcessEvent,
		cgrEv, &tmpReply); err != nil {
		return
	}
	return &tmpReply, nil
}

// ratesCostForEvent will process the event with RateS in order to get the cost
func ratesCostForEvent(ctx *context.Context, connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	rateSConns, rpIDs []string) (_ *utils.RateProfileCost, err error) {
	if len(rateSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	tmp := cgrEv.APIOpts[utils.OptsRatesProfileIDs]
	cgrEv.APIOpts[utils.OptsRatesProfileIDs] = slices.Clone(rpIDs)
	var tmpReply utils.RateProfileCost
	err = connMgr.Call(ctx, rateSConns, utils.RateSv1CostForEvent,
		cgrEv, &tmpReply)
	cgrEv.APIOpts[utils.OptsRatesProfileIDs] = tmp
	if err != nil {
		return
	}
	return &tmpReply, nil
}

// costIncrement computes the costIncrement for the event
func costIncrement(ctx *context.Context, cfgCostIncrmts []*utils.CostIncrement,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	// no cost increment in our balance, return a default increment to query to rateS
	for _, costIcrm = range cfgCostIncrmts {
		var pass bool
		if pass, err = fltrS.Pass(ctx, tnt, costIcrm.FilterIDs, ev); err != nil {
			return
		} else if pass {
			return
		}
	}
	return &utils.CostIncrement{
		Increment: utils.NewDecimal(1, 0),
	}, nil
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
func balanceLimit(optsCfg map[string]any) (bL *utils.Decimal, err error) {
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
func debitConcreteUnits(ctx *context.Context, cUnits *utils.Decimal,
	acntID string, cncrtBlncs []*concreteBalance,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {

	clnedUnts := cloneUnitsFromConcretes(cncrtBlncs)
	for _, cB := range cncrtBlncs {
		var ecCncrt *utils.EventCharges
		if ecCncrt, err = cB.debitConcretes(ctx, cUnits.Clone(), cgrEv, nil); err != nil {
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
		cUnits = utils.SubstractDecimal(cUnits, ecCncrt.Concretes)
		if cUnits.Compare(utils.NewDecimal(0, 0)) <= 0 {
			return // have debited all, total is smaller or equal to 0
		}
	}
	// we could not debit all, put back what we have debited
	restoreUnitsFromClones(cncrtBlncs, clnedUnts)
	return nil, utils.ErrInsufficientCredit
}

// maxDebitAbstractsFromConcretes will debit the maximum possible abstract units out of concretes
func maxDebitAbstractsFromConcretes(ctx *context.Context, aUnits *utils.Decimal,
	acntID string, cncrtBlncs []*concreteBalance,
	connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attributeIDs, rateSConns, rpIDs []string,
	costIcrm *utils.CostIncrement, dbtedAUnts *utils.Decimal, maxItr int) (ec *utils.EventCharges, err error) {
	// Init EventCharges
	calculateCost := costIcrm.RecurrentFee == nil && costIcrm.FixedFee == nil
	// process AttributeS if needed
	if calculateCost &&
		len(attributeIDs) != 0 && attributeIDs[0] != utils.MetaNone { // cost unknown, apply AttributeS to query from RateS
		var rplyAttrS *attributes.ProcessEventReply
		if rplyAttrS, err = processAttributeS(ctx, connMgr, cgrEv, attrSConns,
			attributeIDs); err != nil {
			return
		}
		if len(rplyAttrS.AlteredFields) != 0 { // event was altered
			cgrEv = rplyAttrS.CGREvent // local change, not reflected out
		}
	}
	origConcrtUnts := cloneUnitsFromConcretes(cncrtBlncs) // so we can revert on errors
	paidConcrtUnts := origConcrtUnts                      // so we can revert when higher abstracts are not possible
	var aPaid, aDenied *utils.Decimal
	ec = utils.NewEventCharges() //
	for i := 0; i <= maxItr; i++ {
		if i != 0 {
			restoreUnitsFromClones(cncrtBlncs, origConcrtUnts)
		}
		if i == maxItr {
			utils.Logger.Err(fmt.Sprintf(
				"<%s> MaximumIncrementsExceeded: %d when debitting abstract units: %s, paid units: %s, deniedUnits: %s, costIncrement: %s, event: %s",
				utils.AccountS, i, aUnits, aPaid, aDenied, utils.ToJSON(costIcrm), utils.ToJSON(cgrEv)))
			return nil, utils.ErrMaxIncrementsExceeded
		}
		if calculateCost {
			var rplyCost *utils.RateProfileCost
			cgrEv.APIOpts[utils.OptsRatesIntervalStart] = dbtedAUnts
			cgrEv.APIOpts[utils.OptsRatesUsage] = aUnits
			if rplyCost, err = ratesCostForEvent(ctx, connMgr, cgrEv, rateSConns, rpIDs); err != nil {
				err = utils.NewErrRateS(err)
				return
			}
			// cleanup the opts
			delete(cgrEv.APIOpts, utils.OptsRatesIntervalStart)
			delete(cgrEv.APIOpts, utils.OptsRatesUsage)
			costIcrm = costIcrm.Clone() // so we don't modify the original
			costIcrm.FixedFee = rplyCost.Cost
		}
		var cUnits *utils.Decimal // concrete units to debit
		if costIcrm.FixedFee != nil {
			cUnits = costIcrm.FixedFee
		}
		// RecurrentFee is configured, used it with increments
		if costIcrm.RecurrentFee != nil {
			rcrntCost := utils.MultiplyDecimal(
				utils.DivideDecimal(aUnits, costIcrm.Increment),
				costIcrm.RecurrentFee)
			if cUnits == nil {
				cUnits = rcrntCost
			} else {
				cUnits = utils.SumDecimal(cUnits, rcrntCost)
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
			aDenied = aUnits.Clone()
			if aPaid == nil { // going backwards
				aUnits = utils.DivideDecimal( // divide by 2
					aUnits, utils.NewDecimal(2, 0))
				aUnits = roundUnitsWithIncrements(aUnits, costIcrm.Increment) // make sure abstracts are multiple of increments
				if aUnits.Compare(aDenied) >= 0 ||
					aUnits.Compare(utils.NewDecimal(0, 0)) == 0 ||
					aUnits.Compare(aQried) == 0 { // loop
					break
				}
				continue
			}
		} else { // debit for the usage succeeded
			aPaid = aUnits.Clone()
			paidConcrtUnts = cloneUnitsFromConcretes(cncrtBlncs)
			ec = utils.NewEventCharges() // so we can keep the record of the last sucessful debit
			ec.Merge(ecDbt)              // merge the last debit
			if i == 0 {                  // no estimation done, covering full
				break
			}
		}
		// going upwards
		aPaidDiv := utils.DivideDecimal(aPaid, utils.NewDecimal(2, 0))
		aPaidDiv.RoundToInt()
		aUnits = utils.SumDecimal(aPaid, aPaidDiv)
		if aUnits.Compare(aDenied) >= 0 {
			aUnits = utils.SumDecimal(aPaid, costIcrm.Increment)
		}
		aUnits = roundUnitsWithIncrements(aUnits, costIcrm.Increment)
		if aUnits.Compare(aPaid) <= 0 ||
			aUnits.Compare(aDenied) >= 0 ||
			aUnits.Compare(aQried) == 0 { // loop
			break
		}
	}
	// Nothing paid
	if aPaid == nil {
		// since we are erroring, we restore the concerete balances
		aPaid = utils.NewDecimal(0, 0)
		ec = utils.NewEventCharges()
	}
	ec.Abstracts = aPaid
	restoreUnitsFromClones(cncrtBlncs, paidConcrtUnts)
	return
}

// restoreAccounts will restore the accounts in DB out of their backups if present
func restoreAccounts(ctx *context.Context, dm *engine.DataManager,
	acnts []*utils.Account, bkps []utils.AccountBalancesBackup) {
	for i, bkp := range bkps {
		if bkp == nil ||
			!acnts[i].BalancesAltered(bkp) {
			continue
		}
		acnts[i].RestoreFromBackup(bkp)
		if err := dm.SetAccount(ctx, acnts[i], false); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> restoring account <%s>",
				utils.AccountS, err, acnts[i].TenantID()))
		}
	}
}

// uncompressUnits returns the uncompressed value of the units if compressFactor is provided
func uncompressUnits(units *utils.Decimal, cmprsFctr int, acntChrg *utils.AccountCharge, uf *utils.UnitFactor) (tU *utils.Decimal) {
	tU = units
	if cmprsFctr > 1 {
		tU = utils.MultiplyDecimal(tU, utils.NewDecimal(int64(cmprsFctr), 0))
	}
	// check it this has unit factor
	if uf != nil {
		tU = utils.MultiplyDecimal(tU, uf.Factor)
	}
	return
}

// refundUnitsOnAccount is responsible for returning the units back to the balance
// origBlnc is used for both it's ID as well as as a configuration backup in case when the balance is not longer present
func refundUnitsOnAccount(acnt *utils.Account, units *utils.Decimal, origBlnc *utils.Balance) {
	if _, has := acnt.Balances[origBlnc.ID]; has {
		acnt.Balances[origBlnc.ID].Units = utils.SumDecimal(acnt.Balances[origBlnc.ID].Units, units)
	} else {
		acnt.Balances[origBlnc.ID] = origBlnc.Clone()
		acnt.Balances[origBlnc.ID].Units = units.Clone()
	}
}

// AccountScDebitAbstracts is a wrapper to unify processing from the client side from multiple subsystems
func AccountScDebitAbstracts(ctx *context.Context, fltrS *engine.FilterS,
	connsCfg map[string][]*config.DynamicConns, connMgr *engine.ConnManager,
	cgrEv *utils.CGREvent) (dbt *utils.EventCharges, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, connsCfg, utils.MetaAccounts,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		err = utils.NewErrNotConnected(utils.AccountS)
		return
	}
	dbt = &utils.EventCharges{}
	if err = connMgr.Call(ctx, conns,
		utils.AccountSv1DebitAbstracts, cgrEv, dbt); err != nil {
		return
	}
	return
}
