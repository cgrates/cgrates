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

// newConcreteBalance constructs a concreteBalanceOperator
func newConcreteBalanceOperator(blnCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &concreteBalance{blnCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	connMgr    *engine.ConnManager
	attrSConns,
	rateSConns []string
}

// costIncrement finds out the cost increment for the event
func (cB *concreteBalance) costIncrement(tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	for _, cIcrm := range cB.blnCfg.CostIncrements {
		var pass bool
		if pass, err = cB.fltrS.Pass(tnt, cIcrm.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		costIcrm = cIcrm.Clone() // need clone since we might modify
		return
	}
	// nothing matched, return default
	costIcrm = &utils.CostIncrement{
		Increment:    &utils.Decimal{decimal.New(1, 0)},
		RecurrentFee: &utils.Decimal{decimal.New(-1, 0)}}

	return
}

// unitFactor returns the unitFactor for the event
func (cB *concreteBalance) unitFactor(tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uF = range cB.blnCfg.UnitFactors {
		var pass bool
		if pass, err = cB.fltrS.Pass(tnt, uF.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		return
	}
	return
}

// balanceLimit returns the balance's limit
func (cB *concreteBalance) balanceLimit() (bL *utils.Decimal) {
	if _, isUnlimited := cB.blnCfg.Opts[utils.MetaBalanceUnlimited]; isUnlimited {
		return
	}
	if lmtIface, has := cB.blnCfg.Opts[utils.MetaBalanceLimit]; has {
		bL = lmtIface.(*utils.Decimal)
		return
	}
	// nothing matched, return default
	bL = utils.NewDecimal(0, 0)
	return
}

// debit implements the balanceOperator interface
func (cB *concreteBalance) debitUsage(usage *utils.Decimal, startTime time.Time,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(cgrEv.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	return
}

// debitUnits is a direct debit of balance units
func (cB *concreteBalance) debitUnits(dUnts *utils.Decimal, incrm *utils.Decimal,
	cgrEv *utils.CGREvent) (dbted *utils.Decimal, uF *utils.UnitFactor, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(cgrEv.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, nil, utils.ErrFilterNotPassingNoCaps
	}

	// unitFactor
	if uF, err = cB.unitFactor(cgrEv.Tenant, evNm); err != nil {
		return
	}

	var hasUF bool
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		dUnts = &utils.Decimal{utils.MultiplyBig(dUnts.Big, uF.Factor.Big)}
		incrm = &utils.Decimal{utils.MultiplyBig(incrm.Big, uF.Factor.Big)}
		hasUF = true
	}

	blcVal := cB.blnCfg.Units

	// balanceLimit
	var hasLmt bool
	blncLmt := cB.balanceLimit()
	if blncLmt != nil && blncLmt.Big.Cmp(decimal.New(0, 0)) != 0 {
		blcVal = &utils.Decimal{utils.SubstractBig(blcVal.Big, blncLmt.Big)}
		hasLmt = true
	}
	if blcVal.Compare(dUnts) == -1 && blncLmt != nil { // balance smaller than debit
		// will use special rounding to 0 since otherwise we go negative (ie: 0.05 as increment)
		maxIncrm := &utils.Decimal{
			decimal.WithContext(
				decimal.Context{RoundingMode: decimal.ToZero}).Quo(blcVal.Big,
				incrm.Big).RoundToInt()}
		dUnts = utils.MultiplyDecimal(incrm, maxIncrm)
	}
	rmain := &utils.Decimal{utils.SubstractBig(blcVal.Big, dUnts.Big)}
	if hasLmt {
		rmain = &utils.Decimal{utils.AddBig(rmain.Big, blncLmt.Big)}
	}
	if hasUF {
		dbted = &utils.Decimal{utils.DivideBig(dUnts.Big, uF.Factor.Big)}
	} else {
		dbted = dUnts
	}
	rmainFlt64, ok := rmain.Float64()
	if !ok {
		return nil, nil, fmt.Errorf("failed representing decimal <%s> as float64", rmain)
	}
	cB.blnCfg.Units = utils.NewDecimalFromFloat64(rmainFlt64)
	return
}
