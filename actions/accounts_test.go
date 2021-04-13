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
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestACExecuteAccountsSetBalance(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	internalChan := make(chan birpc.ClientConnector, 1)
	connMngr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts): internalChan,
	})
	apAction := &engine.APAction{
		ID:   "TestACExecuteAccounts",
		Type: utils.MetaSetBalance,
		Diktats: []*engine.APDiktat{
			{
				Path:  "~*balance.TestBalance.Value",
				Value: "\"constant;`>;q=0.7;expires=3600constant\"",
			},
		},
	}

	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]interface{}{
			utils.Usage: 10 * time.Minute,
		},
	}

	actCdrLG := &actSetBalance{
		config:  cfg,
		connMgr: connMngr,
		aCfg:    apAction,
	}

	expected := "no connection with AccountS"
	if err := actCdrLG.execute(context.Background(), dataStorage, utils.MetaBalanceLimit); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	actCdrLG.config.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	expected = "Closed unspilit syntax "
	if err := actCdrLG.execute(context.Background(), dataStorage, utils.MetaBalanceLimit); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//invalid to parse a value from diktats
	actCdrLG.aCfg.Diktats[0].Value = "10"
	expected = context.DeadlineExceeded.Error()
	if err := actCdrLG.execute(context.Background(), dataStorage, utils.MetaBalanceLimit); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestACExecuteAccountsRemBalance(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	internalChan := make(chan birpc.ClientConnector, 1)
	connMngr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts): internalChan,
	})
	apAction := &engine.APAction{
		ID:   "TestACExecuteAccountsRemBalance",
		Type: utils.MetaSetBalance,
		Diktats: []*engine.APDiktat{
			{
				Path:  "~*balance.TestBalance.Value",
				Value: "10",
			},
		},
	}

	actRemBal := &actRemBalance{
		config:  cfg,
		connMgr: connMngr,
		aCfg:    apAction,
		tnt:     "cgrates.org",
	}

	expected := "no connection with AccountS"
	if err := actRemBal.execute(context.Background(), nil, utils.MetaRemBalance); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	actRemBal.config.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	expected = context.DeadlineExceeded.Error()
	if err := actRemBal.execute(context.Background(), nil, utils.MetaRemBalance); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestACExecuteAccountsParseError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	internalChan := make(chan birpc.ClientConnector, 1)
	connMngr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts): internalChan,
	})
	apAction := &engine.APAction{
		ID:   "TestACExecuteAccountsRemBalance",
		Type: utils.MetaSetBalance,
		Diktats: []*engine.APDiktat{
			{
				Path:  "~*balance.TestBalance.Value",
				Value: "~10",
			},
		},
	}

	actsetBal := &actSetBalance{
		config:  cfg,
		connMgr: connMngr,
		aCfg:    apAction,
		tnt:     "cgrates.org",
	}
	dataStorage := utils.MapStorage{}

	if err := actsetBal.execute(nil, dataStorage, utils.MetaRemBalance); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestACAccountsGetIDs(t *testing.T) {
	apAction := &engine.APAction{
		ID: "TestACExecuteAccountsRemBalance",
	}

	actRemBal := &actRemBalance{
		aCfg: apAction,
	}
	if rcv := actRemBal.id(); rcv != apAction.ID {
		t.Errorf("Expected %+v, received %+v", apAction.ID, rcv)
	}

	actSeTBal := &actSetBalance{
		aCfg: apAction,
	}
	if rcv := actSeTBal.id(); rcv != apAction.ID {
		t.Errorf("Expected %+v, received %+v", apAction.ID, rcv)
	}
}
