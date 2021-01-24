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
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAbstractBalance constructs an abstractBalanceOperator
func newAbstractBalanceOperator(blnCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &abstractBalance{blnCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns}
}

// abstractBalance is the operator for *abstract balance type
type abstractBalance struct {
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	connMgr    *engine.ConnManager
	attrSConns,
	rateSConns []string
}

// costIncrement finds out the cost increment for the event
func (aB *abstractBalance) costIncrement(tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	for _, cIcrm := range aB.blnCfg.CostIncrements {
		var pass bool
		if pass, err = aB.fltrS.Pass(tnt, cIcrm.FilterIDs, ev); err != nil {
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

// unitFactor returns the unitFactor for the event
func (aB *abstractBalance) unitFactor(tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uF = range aB.blnCfg.UnitFactors {
		var pass bool
		if pass, err = aB.fltrS.Pass(tnt, uF.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		return
	}
	return
}

// balanceLimit returns the balance's limit
func (aB *abstractBalance) balanceLimit() (bL *utils.Decimal) {
	if _, isUnlimited := aB.blnCfg.Opts[utils.MetaBalanceUnlimited]; isUnlimited {
		return
	}
	if lmtIface, has := aB.blnCfg.Opts[utils.MetaBalanceLimit]; has {
		bL = lmtIface.(*utils.Decimal)
	}
	// nothing matched, return default
	bL = utils.NewDecimal(0, 0)
	return
}

// processAttributeS will process the event with AttributeS
func (aB *abstractBalance) processAttributeS(cgrEv *utils.CGREvent) (rplyEv *engine.AttrSProcessEventReply, err error) {
	if len(aB.attrSConns) == 0 {
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
		AttributeIDs: aB.blnCfg.AttributeIDs,
		ProcessRuns:  procRuns,
	}
	err = aB.connMgr.Call(aB.attrSConns, nil, utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv)
	return
}

// rateSCostForEvent will process the event with RateS in order to get the cost
func (aB *abstractBalance) rateSCostForEvent(cgrEv *utils.CGREvent) (rplyCost *engine.RateProfileCost, err error) {
	if len(aB.rateSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	err = aB.connMgr.Call(aB.rateSConns, nil, utils.RateSv1CostForEvent,
		&utils.ArgsCostForEvent{CGREvent: cgrEv}, &rplyCost)
	return
}

// debitUsageFromConcrete attempts to debit the usage out of concrete balances
// returns utils.ErrInsufficientCredit if complete usage cannot be debitted
func (aB *abstractBalance) debitUsageFromConcrete(cBs []*concreteBalance, usage *utils.Decimal,
	costIcrm *utils.CostIncrement, cgrEv *utils.CGREvent) (err error) {
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil {
		var rplyCost *engine.RateProfileCost
		if rplyCost, err = aB.rateSCostForEvent(cgrEv); err != nil {
			return
		}
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
	clnedUnts := cloneUnitsFromConcretes(cBs)
	for _, cB := range cBs {
		ev := utils.MapStorage{
			utils.MetaOpts: cgrEv.Opts,
			utils.MetaReq:  cgrEv.Event,
		}
		var dbted *utils.Decimal
		if dbted, _, err = cB.debitUnits(&utils.Decimal{tCost}, cgrEv.Tenant, ev); err != nil {
			restoreUnitsFromClones(cBs, clnedUnts)
			return
		}
		tCost = utils.SubstractBig(tCost, dbted.Big)
		if tCost.Cmp(decimal.New(0, 0)) <= 0 {
			return // have debited all, total is smaller or equal to 0
		}
	}
	// we could not debit all, put back what we have debited
	restoreUnitsFromClones(cBs, clnedUnts)
	return utils.ErrInsufficientCredit
}

// debitUsage implements the balanceOperator interface
func (aB *abstractBalance) debitUsage(usage *utils.Decimal, startTime time.Time,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = aB.fltrS.Pass(cgrEv.Tenant, aB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = aB.costIncrement(cgrEv.Tenant, evNm); err != nil {
		return
	}
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil &&
		len(aB.blnCfg.AttributeIDs) != 0 { // cost unknown, apply AttributeS to query from RateS
		var rplyAttrS *engine.AttrSProcessEventReply
		if rplyAttrS, err = aB.processAttributeS(cgrEv); err != nil {
			return
		}
		if len(rplyAttrS.AlteredFields) != 0 { // event was altered
			cgrEv = rplyAttrS.CGREvent
		}
	}

	origBlclVal := new(decimal.Big).Copy(aB.blnCfg.Units.Big) // so we can restore on errors

	// balanceLimit
	var hasLmt bool
	blncLmt := aB.balanceLimit()
	if blncLmt != nil && blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, blncLmt.Big)
		hasLmt = true
	}

	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = aB.unitFactor(cgrEv.Tenant, evNm); err != nil {
		return
	}
	var hasUF bool
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		usage.Big = utils.MultiplyBig(usage.Big, uF.Factor.Big)
		hasUF = true
	}

	// balance smaller than usage, correct usage
	if aB.blnCfg.Units.Compare(usage) == -1 {
		// will use special rounding to 0 since otherwise we go negative (ie: 0.05 as increment)
		maxIncrm := decimal.WithContext(
			decimal.Context{RoundingMode: decimal.ToZero}).Quo(aB.blnCfg.Units.Big,
			costIcrm.Increment.Big).RoundToInt()
		usage.Big = utils.MultiplyBig(maxIncrm, costIcrm.Increment.Big) // decrease the usage to match the maximum increments
	}

	// attempt to debit usage with cost
	// fix the maximum number of iterations
	for i := 0; i < 10000; i++ {
		continue
	}

	fmt.Printf("costIcrm: %+v, blncLmt: %+v, hasLmt: %+v, uF: %+v, origBlclVal: %s, hasUF: %v",
		costIcrm, blncLmt, hasLmt, uF, origBlclVal, hasUF)

	return
}
