//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestRadiusChargingPool tests a pooled plan: several SIMs share one account and
// one data balance, while sim3 also keeps its own account for when the pool empties.
func TestRadiusChargingPool(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	dictDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dictDir, "dictionary.test"), []byte(radiusDict), 0644); err != nil {
		t.Fatal(err)
	}
	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "radius_charging"),
		ConfigJSON: fmt.Sprintf(`{
"radiusAgent": {"clientDictionaries": {"*default": [%q]}}
}`, dictDir+"/"),
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, cfg := ng.Run(t)

	const usULIHex = "8213f010000113f01000000101" // United States
	const (
		sim1 = "310150000000001"
		sim2 = "310150000000002"
		sim3 = "310150000000003"
	)

	setFilter(t, client, "FLTR_POOL_US", []*engine.FilterRule{
		{
			Type:    utils.MetaString,
			Element: "~*req.IMSI",
			Values:  []string{sim1, sim2, sim3},
		},
	})
	setPoolAccount(t, client, "POOL_US", []string{"FLTR_POOL_US"}, 20, gb)
	// sim3 also has its own account, weighted below the pool
	setPoolAccount(t, client, sim3, []string{"*string:~*req.IMSI:" + sim3}, 10, gb)

	sendAcct(t, cfg, radiusDict, sim1, "pool-1", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, sim1, "pool-1", "Stop", usULIHex, 600*mb)
	checkUnits(t, balanceUnits(t, client, "POOL_US", "data"), utils.NewDecimal(gb-600*mb, 0), "pool after sim1")

	// sim2 gets the 0.4 GB left, the pool is now empty
	sendAcct(t, cfg, radiusDict, sim2, "pool-2", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, sim2, "pool-2", "Stop", usULIHex, 600*mb)
	checkUnits(t, balanceUnits(t, client, "POOL_US", "data"), utils.NewDecimal(0, 0), "pool after sim2")

	// pool empty, sim3 falls back to its own account
	sendAcct(t, cfg, radiusDict, sim3, "pool-3", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, sim3, "pool-3", "Stop", usULIHex, 300*mb)
	checkUnits(t, balanceUnits(t, client, sim3, "data"), utils.NewDecimal(gb-300*mb, 0), "sim3 own balance after pool empty")
}

func setPoolAccount(t *testing.T, c *birpc.Client, id string, filterIDs []string, weight float64, units int64) {
	t.Helper()
	var reply string
	if err := c.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        id,
				FilterIDs: filterIDs,
				Weights:   utils.DynamicWeights{{Weight: weight}},
				Balances: map[string]*utils.Balance{
					"data": {
						ID:      "data",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 10}},
						Units:   utils.NewDecimal(units, 0),
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetAccount %s: %v", id, err)
	}
}
