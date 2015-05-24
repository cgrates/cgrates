/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	//"os"
	"path"
	"reflect"
	//"strings"
	"testing"
	"time"

	//"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tutLocalCfgPath string
var tutFsLocalCfg *config.CGRConfig
var tutLocalRpc *rpc.Client

func TestTutLocalInitCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	tutLocalCfgPath = path.Join(*dataDir, "conf", "samples", "tutlocal")
	// Init config first
	var err error
	tutFsLocalCfg, err = config.NewCGRConfigFromFolder(tutLocalCfgPath)
	if err != nil {
		t.Error(err)
	}
	tutFsLocalCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutFsLocalCfg)
}

// Remove data in both rating and accounting db
func TestTutLocalResetDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(tutFsLocalCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutLocalResetStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(tutFsLocalCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutLocalStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := engine.StopStartEngine(tutLocalCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutLocalRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	tutLocalRpc, err = jsonrpc.Dial("tcp", tutFsLocalCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutLocalLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutLocalRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestTutLocalCacheStats(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 3, RatingPlans: 3, RatingProfiles: 8, Actions: 6, SharedGroups: 1, RatingAliases: 1, AccountAliases: 1,
		DerivedChargers: 1, LcrProfiles: 4}
	var args utils.AttrCacheStats
	if err := tutLocalRpc.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

// Check items age
func TestTutLocalGetCachedItemAge(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvAge *utils.CachedItemAge
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "1002", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.Destination > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "RP_RETAIL1", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.RatingPlan > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.RatingProfile > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "LOG_WARNING", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.Action > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "SHARED_A", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:1001:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:*any:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}

	/*
		if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "1006", &rcvAge); err != nil {
			t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
		} else if rcvAge.RatingAlias > time.Duration(2)*time.Second {
			t.Errorf("Cache too old: %d", rcvAge)
		}
		if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "1006", &rcvAge); err != nil {
			t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
		} else if rcvAge.RatingAlias > time.Duration(2)*time.Second || rcvAge.AccountAlias > time.Duration(2)*time.Second {
			t.Errorf("Cache too old: %d", rcvAge)
		}
	*/
}

// Check call costs
func TestTutLocalGetCosts(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:00:20Z")
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	var cc engine.CallCost
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.6 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.6417 { // 0.01 first minute, 0.04 25 seconds with RT_20CNT
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:00:20Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1003",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1003",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1.3 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:00:20Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1004",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1004",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1.3 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
}

// Check call costs
func TestTutLocalMaxDebit(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:00:20Z")
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	var cc engine.CallCost
	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() == 20 {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
}

// Make sure queueids were created
func TestTutFsCallsCdrStats(t *testing.T) {
	if !*testCalls {
		return
	}
	var queueIds []string
	eQueueIds := []string{"*default", "CDRST1", "CDRST_1001", "CDRST_1002", "CDRST_1003", "STATS_SUPPL1", "STATS_SUPPL2"}
	if err := tutLocalRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
	} else if len(eQueueIds) != len(queueIds) {
		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
	}
}

// Check LCR
//FixMe:
/*{"id":16,"result":{"Entry":{"DestinationId":"*any","RPCategory":"lcr_profile1","Strategy":"*static","StrategyParams":"suppl1;suppl2","Weight":10},"SupplierCosts":[{"Supplier":"*out:cgrates.org:lcr_profile1:suppl1","Cost":0,"Duration":0,"Error":{},"QOS":null},{"Supplier":"*out:cgrates.org:lcr_profile1:suppl2","Cost":0,"Duration":0,"Error":{},"QOS":null}]},"error":null}
 */

func TestTutLocalLcrStatic(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_STATIC, StrategyParams: "suppl2;suppl1", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	var lcr engine.LCRCost
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1003",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_STATIC, StrategyParams: "suppl1;suppl2", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
}

func TestTutLocalStopCgrEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
