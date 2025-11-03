//go:build flaky

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

package apis

import (
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	idxLoadCfgPath   string
	idxLoadCfg       *config.CGRConfig
	idxLoadBiRPC     *birpc.Client
	idxLoadConfigDIR string //run tests for specific configuration

	idxLoadAccPrf = []func(t *testing.T){
		testIdxLoadInitCfg,
		testIdxLoadInitDataDb,

		testIdxLoadStartEngine,
		testIdxLoadRPCConn,
		testIdxLoadTariffPlan,
		testIdxLoadCheckIndexes,
		testIdxLoadKillEngine,
	}
)

func TestIdxCheckAfterLoad(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		idxLoadConfigDIR = "session_volume_discount_internal"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		idxLoadConfigDIR = "session_volume_discount_mysql"
	case utils.MetaMongo:
		idxLoadConfigDIR = "session_volume_discount_mongo"
	case utils.MetaPostgres:
		t.Skip()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range idxLoadAccPrf {
		t.Run(idxLoadConfigDIR+"config", stest)
	}
}

func testIdxLoadInitCfg(t *testing.T) {
	var err error
	idxLoadCfgPath = path.Join(*utils.DataDir, "conf", "samples", idxLoadConfigDIR)
	idxLoadCfg, err = config.NewCGRConfigFromPath(context.Background(), idxLoadCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testIdxLoadInitDataDb(t *testing.T) {
	if err := engine.InitDB(idxLoadCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testIdxLoadStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(idxLoadCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testIdxLoadRPCConn(t *testing.T) {
	idxLoadBiRPC = engine.NewRPCClient(t, idxLoadCfg.ListenCfg(), *utils.Encoding)
}

func testIdxLoadTariffPlan(t *testing.T) {
	var reply string
	if err := idxLoadBiRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			// StopOnError: true,
			APIOpts: map[string]any{utils.MetaCache: utils.MetaReload}, // after laod, we got CacheIDs and it will be called Cachesv1.Clear, so indexes will be removed
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testIdxLoadCheckIndexes(t *testing.T) {
	expected := []string{
		"*string:*opts.*runID:balanceonly:RP_ABS_VOLUME2",
		"*string:*opts.*runID:customers:RP_CUSTOMER1",
		"*string:*opts.*runID:suppliers:RP_SUPPLIER1",
		"*string:*opts.*runID:suppliers:RP_SUPPLIER2",
		"*string:*opts.*runID:suppliers:RP_SUPPLIER3",
		"*string:*opts.*runID:suppliers:RP_SUPPLIER4",
		"*string:*req.RouteID:supplier1:RP_SUPPLIER1",
		"*string:*req.RouteID:supplier2:RP_SUPPLIER2",
		"*string:*req.RouteID:supplier3:RP_SUPPLIER3",
		"*string:*req.RouteID:supplier4:RP_SUPPLIER4",
	}
	//get indexes *rate_profiles
	var reply []string
	if err := idxLoadBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType: utils.MetaRateProfiles,
		}, &reply); (err == nil || err.Error() != utils.ErrNotFound.Error()) && idxLoadCfg.DbCfg().DBConns[utils.MetaDefault].Type == utils.MetaInternal {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(expected, reply) && idxLoadConfigDIR != "session_volume_discount_internal" {
			t.Errorf("Expected %+v \n, received %+v", expected, reply)
		}
	}

	expected = []string{
		"*none:*any:*any:CHRG_SUPPLIER",
		"*none:*any:*any:CHRG_CUSTOMER",
	}
	//get indexes *chargers
	if err := idxLoadBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType: utils.MetaChargers,
		}, &reply); (err == nil || err.Error() != utils.ErrNotFound.Error()) && idxLoadCfg.DbCfg().DBConns[utils.MetaDefault].Type == utils.MetaInternal {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(expected, reply) && idxLoadConfigDIR != "session_volume_discount_internal" {
			t.Errorf("Expected %+v, received %+v", expected, reply)
		}
	}

	expected = []string{
		"*none:*any:*any:ATTR_RATES",
		"*string:*req.Account:ACCOUNT1:ATTR_ACCOUNTS",
	}
	//get indexes *attributes
	if err := idxLoadBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType: utils.MetaAttributes,
		}, &reply); (err == nil || err.Error() != utils.ErrNotFound.Error()) && idxLoadCfg.DbCfg().DBConns[utils.MetaDefault].Type == utils.MetaInternal {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(expected, reply) && idxLoadConfigDIR != "session_volume_discount_internal" {
			t.Errorf("Expected %+v, received %+v", expected, reply)
		}
	}
}

// Kill the engine when it is about to be finished
func testIdxLoadKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
