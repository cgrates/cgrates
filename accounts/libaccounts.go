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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAccountBalances constructs accountBalances
func newAccountBalances(acnt *utils.AccountProfile,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (acntBlncs *accountBalances, err error) {
	blncs := make(utils.Balances, len(acnt.Balances)) // Temporary code to pass the build please update this accordingly
	for _, bal := range acnt.Balances {
		blncs = append(blncs, bal)
	}
	blncs.Sort()
	acntBlncs = &accountBalances{blnCfgs: blncs, connMgr: connMgr, attrSConns: attrSConns, rateSConns: rateSConns}
	// populate typIdx
	for i, blnCfg := range blncs {
		acntBlncs.typIdx[blnCfg.Type] = append(acntBlncs.typIdx[blnCfg.Type], i)
	}
	// populate cncrtBlncs
	acntBlncs.cncrtBlncs = make([]*concreteBalance, len(acntBlncs.typIdx[utils.MetaConcrete]))
	for i, blncIdx := range acntBlncs.typIdx[utils.MetaConcrete] {
		acntBlncs.cncrtBlncs[i] = newConcreteBalanceOperator(acntBlncs.blnCfgs[blncIdx],
			acntBlncs.cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns).(*concreteBalance)
		acntBlncs.opers[acntBlncs.blnCfgs[blncIdx].ID] = acntBlncs.cncrtBlncs[i]
	}
	// populate opers
	for _, blnCfg := range acntBlncs.blnCfgs {
		if blnCfg.Type == utils.MetaConcrete { // already computed above
			continue
		}
		if acntBlncs.opers[blnCfg.ID], err = newBalanceOperator(blnCfg,
			acntBlncs.cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns); err != nil {
			return
		}
	}
	return
}

// accountBalances implements processing of the events centralized
type accountBalances struct {
	blnCfgs    []*utils.Balance           // ordered list of balance configurations
	typIdx     map[string][]int           // index based on type
	cncrtBlncs []*concreteBalance         // concrete balances so we can pass them to the newBalanceOperator
	opers      map[string]balanceOperator // map[blncID]balanceOperator

	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
	attrSConns,
	rateSConns []string
}

// newBalanceOperator instantiates balanceOperator interface
// cncrtBlncs are needed for abstract balance debits
func newBalanceOperator(blncCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (bP balanceOperator, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MetaConcrete:
		return newConcreteBalanceOperator(blncCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns), nil
	case utils.MetaAbstract:
		return newAbstractBalanceOperator(blncCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns), nil
	}
}

// balanceOperator is the implementation of a balance type
type balanceOperator interface {
	debitUsage(usage *utils.Decimal, cgrEv *utils.CGREvent) (dbted *utils.Decimal, ec *utils.EventCharges, err error)
}

// roundUsageWithIncrements rounds the usage based on increments
func roundedUsageWithIncrements(usage, incrm *decimal.Big) (rndedUsage *decimal.Big) {
	usgMaxIncrm := decimal.WithContext(
		decimal.Context{RoundingMode: decimal.ToZero}).Quo(usage,
		incrm).RoundToInt()
	rndedUsage = utils.MultiplyBig(usgMaxIncrm, incrm)
	return
}

// processAttributeS will process the event with AttributeS
func processAttributeS(connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attrIDs []string) (rplyEv *engine.AttrSProcessEventReply, err error) {
	if len(attrSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AttributeS)
	}
	var procRuns *int
	if val, has := cgrEv.Opts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			procRuns = utils.IntPointer(int(v))
		}
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.FirstNonEmpty(
			engine.MapEvent(cgrEv.Opts).GetStringIgnoreErrors(utils.OptsContext),
			utils.MetaAccountS)),
		CGREvent:     cgrEv,
		AttributeIDs: attrIDs,
		ProcessRuns:  procRuns,
	}
	err = connMgr.Call(attrSConns, nil, utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv)
	return
}

// rateSCostForEvent will process the event with RateS in order to get the cost
func rateSCostForEvent(connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	rateSConns, rpIDs []string) (rplyCost *engine.RateProfileCost, err error) {
	if len(rateSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	err = connMgr.Call(rateSConns, nil, utils.RateSv1CostForEvent,
		&utils.ArgsCostForEvent{CGREvent: cgrEv, RateProfileIDs: rpIDs}, &rplyCost)
	return
}

// costIncrement computes the costIncrement for the event
func costIncrement(cfgCostIncrmts []*utils.CostIncrement,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	for _, cIcrm := range cfgCostIncrmts {
		var pass bool
		if pass, err = fltrS.Pass(tnt, cIcrm.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		costIcrm = cIcrm
		break
	}
	if costIcrm == nil {
		costIcrm = new(utils.CostIncrement)
	}
	if costIcrm.Increment == nil {
		costIcrm.Increment = utils.NewDecimal(1, 0)
	}
	if costIcrm.RecurrentFee == nil {
		costIcrm.RecurrentFee = utils.NewDecimal(-1, 0)
	}
	return
}

// unitFactor detects the unitFactor for the event
func unitFactor(cfgUnitFactors []*utils.UnitFactor,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uF = range cfgUnitFactors {
		var pass bool
		if pass, err = fltrS.Pass(tnt, uF.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
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

// debitUsageFromConcrete attempts to debit the usage out of concrete balances
// returns utils.ErrInsufficientCredit if complete usage cannot be debitted
func debitUsageFromConcretes(cncrtBlncs []*concreteBalance, usage *utils.Decimal,
	costIcrm *utils.CostIncrement, cgrEv *utils.CGREvent,
	connMgr *engine.ConnManager, rateSConns, rpIDs []string) (err error) {
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil {
		var rplyCost *engine.RateProfileCost
		if rplyCost, err = rateSCostForEvent(connMgr, cgrEv, rateSConns, rpIDs); err != nil {
			return
		}
		costIcrm = costIcrm.Clone() // so we don't modify the original
		costIcrm.FixedFee = utils.NewDecimalFromFloat64(rplyCost.Cost)
	}
	var tCost *decimal.Big
	if costIcrm.FixedFee != nil {
		tCost = costIcrm.FixedFee.Big
	}
	// RecurrentFee is configured, used it with increments
	if costIcrm.RecurrentFee.Big.Cmp(decimal.New(-1, 0)) != 0 {
		rcrntCost := utils.MultiplyBig(
			utils.DivideBig(usage.Big, costIcrm.Increment.Big),
			costIcrm.RecurrentFee.Big)
		if tCost == nil {
			tCost = rcrntCost
		} else {
			tCost = utils.SumBig(tCost, rcrntCost)
		}
	}
	clnedUnts := cloneUnitsFromConcretes(cncrtBlncs)
	for _, cB := range cncrtBlncs {
		ev := utils.MapStorage{
			utils.MetaOpts: cgrEv.Opts,
			utils.MetaReq:  cgrEv.Event,
		}
		var dbted *utils.Decimal
		if dbted, _, err = cB.debitUnits(&utils.Decimal{tCost}, cgrEv.Tenant, ev); err != nil {
			restoreUnitsFromClones(cncrtBlncs, clnedUnts)
			return
		}
		tCost = utils.SubstractBig(tCost, dbted.Big)
		if tCost.Cmp(decimal.New(0, 0)) <= 0 {
			return // have debited all, total is smaller or equal to 0
		}
	}
	// we could not debit all, put back what we have debited
	restoreUnitsFromClones(cncrtBlncs, clnedUnts)
	return utils.ErrInsufficientCredit
}
