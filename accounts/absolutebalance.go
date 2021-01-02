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

// debit implements the balanceOperator interface
func (ab *abstractBalance) debit(cgrEv *utils.CGREventWithOpts,
	startTime time.Time, usage *decimal.Big) (ec *utils.EventCharges, err error) {
	//var uF *utils.UsageFactor
	if _, _, err = usageWithFactor(ab.blnCfg, ab.fltrS, cgrEv, usage); err != nil {
		return
	}
	return
}
