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

package actions

import (
	"context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// actTopup will log data to CGRateS logger
type actTopup struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
	tnt     string
	reset   bool
}

func (aL *actTopup) id() string {
	return aL.aCfg.ID
}

func (aL *actTopup) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actTopup) execute(ctx context.Context, data utils.MapStorage, trgID string) (err error) {
	return
	/*
		if len(aL.config.ActionSCfg().AccountSConns) == 0 {
			return fmt.Errorf("no connection with AccountS")
		}
		var valStr string
		if valStr, err = aL.cfg().Value.ParseDataProviderWithInterfaces(data); err != nil {
			return
		}
		var val float64
		if val, err = utils.IfaceAsFloat64(valStr); err != nil {
			return
		}

		// *accounts.1001.Balance.MonetaryBalance
		path := strings.SplitN(aL.cfg().Path, utils.ConcatenatedKeySep, 2)
		if len(path) != 2 {
			err = fmt.Errorf("Unsupported path: %s", aL.cfg().Path)
			return
		}
		args := &utils.ArgsModifyBalance{
			Tenant:    aL.tnt,
			AccountID: trgID,
			BalanceID: aL.cfg().Path,
			Value:     val,
			Reset:     aL.reset,
		}
		var rply string
		return aL.connMgr.Call(aL.config.ActionSCfg().AccountSConns, nil,
			utils.AccountSv1TopupBalance, args, &rply)
	*/
}
