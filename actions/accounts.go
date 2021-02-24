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
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// actSetBalance will update the account
type actSetBalance struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
	tnt     string
	reset   bool
}

func (aL *actSetBalance) id() string {
	return aL.aCfg.ID
}

func (aL *actSetBalance) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actSetBalance) execute(ctx context.Context, data utils.MapStorage, trgID string) (err error) {
	if len(aL.config.ActionSCfg().AccountSConns) == 0 {
		return fmt.Errorf("no connection with AccountS")
	}

	args := &utils.ArgsActSetBalance{
		Tenant:    aL.tnt,
		AccountID: trgID,
		Reset:     aL.reset,
		Diktats:   make([]*utils.BalDiktat, len(aL.cfg().Diktats)),
		Opts:      aL.cfg().Opts,
	}
	for i, actD := range aL.cfg().Diktats {
		var val string
		if val, err = actD.Value.ParseDataProvider(data); err != nil {
			return
		}
		args.Diktats[i] = &utils.BalDiktat{
			Path:  actD.Path,
			Value: val,
		}
	}
	var rply string
	return aL.connMgr.Call(aL.config.ActionSCfg().AccountSConns, nil,
		utils.AccountSv1UpdateBalance, args, &rply)
}

// actRemBalance will remove multiple balances from account
type actRemBalance struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
	tnt     string
}

func (aL *actRemBalance) id() string {
	return aL.aCfg.ID
}

func (aL *actRemBalance) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actRemBalance) execute(ctx context.Context, data utils.MapStorage, trgID string) (err error) {
	if len(aL.config.ActionSCfg().AccountSConns) == 0 {
		return fmt.Errorf("no connection with AccountS")
	}

	args := &utils.ArgsActRemoveBalances{
		Tenant:     aL.tnt,
		AccountID:  trgID,
		BalanceIDs: make([]string, len(aL.cfg().Diktats)),
		Opts:       aL.cfg().Opts,
	}
	for i, actD := range aL.cfg().Diktats {
		args.BalanceIDs[i] = actD.Path
	}
	var rply string
	return aL.connMgr.Call(aL.config.ActionSCfg().AccountSConns, nil,
		utils.AccountSv1UpdateBalance, args, &rply)
}
