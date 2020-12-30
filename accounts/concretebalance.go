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
func newConcreteBalanceOperator(blnCfg *utils.Balance,
	fltrS *engine.FilterS, ralsConns []string) balanceOperator {
	return &concreteBalance{blnCfg, fltrS, ralsConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	blnCfg    *utils.Balance
	fltrS     *engine.FilterS
	ralsConns []string
}

// debit implements the balanceOperator interface
func (cb *concreteBalance) debit(cgrEv *utils.CGREventWithOpts,
	startTime time.Time, usage *decimal.Big) (ec *utils.EventCharges, err error) {
	//var uF *utils.UsageFactor
	if _, err = usageWithFactor(usage, cb.blnCfg, cb.fltrS, cgrEv); err != nil {
		return
	}
	return
}

// debitUnits is a direct debit of balance units
func (cb *concreteBalance) debitUnits(unts *decimal.Big, incrm *decimal.Big) (dbted *decimal.Big, err error) {
	// toDo: handle *balanceLimit
	dbted = unts
	if cb.blnCfg.DecimalValue().Cmp(unts) == -1 {
		dbted = utils.DivideBig(cb.blnCfg.DecimalValue(), incrm).RoundToInt()
	}
	rmain := utils.SubstractBig(cb.blnCfg.DecimalValue(), dbted)
	flt64, ok := rmain.Float64()
	if !ok {
		return nil, fmt.Errorf("failed representing decimal <%s> as float64", rmain)
	}
	cb.blnCfg.Value = flt64
	cb.blnCfg.SetDecimalValue(rmain)
	return
}
