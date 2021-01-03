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
	fltrS *engine.FilterS, ralsConns []string) balanceOperator {
	return &concreteBalance{blnCfg, cncrtBlncs, fltrS, ralsConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	ralsConns  []string
}

// debit implements the balanceOperator interface
func (cB *concreteBalance) debitUsage(usage *decimal.Big, startTime time.Time,
	cgrEv *utils.CGREventWithOpts) (ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(cgrEv.CGREvent.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	return
}

// debitUnits is a direct debit of balance units
func (cB *concreteBalance) debitUnits(dUnts *decimal.Big, incrm *decimal.Big,
	cgrEv *utils.CGREventWithOpts) (dbted *decimal.Big, mtchedUF *utils.UnitFactor, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// dynamic unit factor
	fctr := decimal.New(1, 0)
	for _, uF := range cB.blnCfg.UnitFactors {
		var pass bool
		if pass, err = cB.fltrS.Pass(cgrEv.CGREvent.Tenant, uF.FilterIDs, evNm); err != nil {
			return nil, nil, err
		} else if !pass {
			continue
		}
		fctr = uF.Factor
		mtchedUF = uF
		break
	}
	var hasUF bool
	if mtchedUF != nil && fctr.Cmp(decimal.New(1, 0)) != 0 {
		dUnts = utils.MultiplyBig(dUnts, fctr)
		incrm = utils.MultiplyBig(incrm, fctr)
		hasUF = true
	}

	blcVal := new(decimal.Big).SetFloat64(cB.blnCfg.Value) // FixMe without float64

	// *balanceLimit
	var hasLmt bool
	blncLmt := decimal.New(0, 0)
	if lmt, has := cB.blnCfg.Opts[utils.MetaBalanceLimit].(*decimal.Big); has {
		blncLmt = lmt
	}
	if blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		blcVal = utils.SubstractBig(blcVal, blncLmt)
		hasLmt = true
	}
	if blcVal.Cmp(dUnts) == -1 { // balance smaller than debit
		maxIncrm := utils.DivideBig(blcVal, incrm).RoundToInt()
		dUnts = utils.MultiplyBig(incrm, maxIncrm)
	}
	rmain := utils.SubstractBig(blcVal, dUnts)
	if hasLmt {
		rmain = utils.AddBig(rmain, blncLmt)
	}
	if hasUF {
		dbted = utils.DivideBig(dUnts, fctr)
	} else {
		dbted = dUnts
	}
	rmainFlt64, ok := rmain.Float64()
	if !ok {
		return nil, nil, fmt.Errorf("failed representing decimal <%s> as float64", rmain)
	}
	cB.blnCfg.Value = rmainFlt64
	return
}
