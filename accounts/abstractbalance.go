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
	fltrS *engine.FilterS, ralsConns []string) balanceOperator {
	return &abstractBalance{blnCfg, cncrtBlncs, fltrS, ralsConns}
}

// abstractBalance is the operator for *abstract balance type
type abstractBalance struct {
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	ralsConns  []string
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
		return
	}
	// nothing matched, return default
	costIcrm = &utils.CostIncrement{
		Increment:    &utils.Decimal{decimal.New(1, 0)},
		RecurrentFee: &utils.Decimal{decimal.New(-1, 0)}}
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
func (aB *abstractBalance) balanceLimit() (bL *decimal.Big) {
	if _, isUnlimited := aB.blnCfg.Opts[utils.MetaBalanceUnlimited]; isUnlimited {
		return
	}
	if lmtIface, has := aB.blnCfg.Opts[utils.MetaBalanceLimit]; has {
		bL = lmtIface.(*decimal.Big)
	}
	// nothing matched, return default
	bL = decimal.New(0, 0)
	return
}

// debitUsage implements the balanceOperator interface
func (aB *abstractBalance) debitUsage(usage *decimal.Big, startTime time.Time,
	cgrEv *utils.CGREventWithOpts) (ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = aB.fltrS.Pass(cgrEv.CGREvent.Tenant, aB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	blcVal := aB.blnCfg.Units.Big
	// balanceLimit
	var hasLmt bool
	blncLmt := aB.balanceLimit()
	if blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		blcVal = utils.SubstractBig(blcVal, blncLmt)
		hasLmt = true
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = aB.costIncrement(cgrEv.CGREvent.Tenant, evNm); err != nil {
		return
	}

	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = aB.unitFactor(cgrEv.CGREvent.Tenant, evNm); err != nil {
		return
	}

	fmt.Printf("costIcrm: %+v, blncLmt: %+v, hasLmt: %+v, uF: %+v", costIcrm, blncLmt, hasLmt, uF)

	return
}
