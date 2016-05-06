package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var destCfgPath string
var destCfg *config.CGRConfig
var destRPC *rpc.Client
var destLoadInst engine.LoadInstance // Share load information between tests

func TestDestManagInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	destCfgPath = path.Join(*dataDir, "conf", "samples", "tutlocal")
	// Init config first
	var err error
	destCfg, err = config.NewCGRConfigFromFolder(destCfgPath)
	if err != nil {
		t.Error(err)
	}
	destCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(destCfg)
}

// Remove data in both rating and accounting db
func TestDestManagResetDataDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(destCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestDestManagResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(destCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestDestManagStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(destCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestDestManagRpcConn(t *testing.T) {
	if !*testIntegration {
		return
	}
	var err error
	destRPC, err = jsonrpc.Dial("tcp", destCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestDestManagLoadTariffPlanFromFolderAll(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "destinations", "alldests")}
	if err := destRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &destLoadInst); err != nil {
		t.Error(err)
	} else if destLoadInst.LoadId == "" {
		t.Error("Empty loadId received, loadInstance: ", destLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestDestManagAllDestinationLoaded(t *testing.T) {
	if !*testIntegration {
		return
	}
	dests := make([]*engine.Destination, 0)
	if err := destRPC.Call("ApierV2.GetDestinations", v2.AttrGetDestinations{DestinationIDs: []string{}}, &dests); err != nil {
		t.Error("Got error on ApierV2.GetDestinations: ", err.Error())
	} else if len(dests) != 6 {
		t.Errorf("Calling ApierV2.GetDestinations got reply: %v", utils.ToIJSON(dests))
	}

	var rcvStats utils.CacheStats
	if err := destRPC.Call("ApierV1.GetCacheStats", utils.AttrCacheStats{}, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if rcvStats.Destinations != 9 {
		t.Errorf("Calling ApierV1.GetCacheStats received: %+v", rcvStats)
	}
}

func TestDestManagLoadTariffPlanFromFolderRemoveSome(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "destinations", "removesome")}
	if err := destRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &destLoadInst); err != nil {
		t.Error(err)
	} else if destLoadInst.LoadId == "" {
		t.Error("Empty loadId received, loadInstance: ", destLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestDestManagRemoveSomeDestinationLoaded(t *testing.T) {
	if !*testIntegration {
		return
	}
	dests := make([]*engine.Destination, 0)
	if err := destRPC.Call("ApierV2.GetDestinations", v2.AttrGetDestinations{DestinationIDs: []string{}}, &dests); err != nil {
		t.Error("Got error on ApierV2.GetDestinations: ", err.Error())
	} else if len(dests) != 6 {
		t.Errorf("Calling ApierV2.GetDestinations got reply: %v", utils.ToIJSON(dests))
	}

	var rcvStats utils.CacheStats
	if err := destRPC.Call("ApierV1.GetCacheStats", utils.AttrCacheStats{}, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if rcvStats.Destinations != 9 {
		t.Errorf("Calling ApierV1.GetCacheStats received: %+v", rcvStats)
	}
}

func TestDestManagLoadTariffPlanFromFolderAddBack(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "destinations", "addback")}
	if err := destRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &destLoadInst); err != nil {
		t.Error(err)
	} else if destLoadInst.LoadId == "" {
		t.Error("Empty loadId received, loadInstance: ", destLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestDestManagAddBackDestinationLoaded(t *testing.T) {
	if !*testIntegration {
		return
	}
	dests := make([]*engine.Destination, 0)
	if err := destRPC.Call("ApierV2.GetDestinations", v2.AttrGetDestinations{DestinationIDs: []string{}}, &dests); err != nil {
		t.Error("Got error on ApierV2.GetDestinations: ", err.Error())
	} else if len(dests) != 6 {
		t.Errorf("Calling ApierV2.GetDestinations got reply: %v", utils.ToIJSON(dests))
	}

	var rcvStats utils.CacheStats
	if err := destRPC.Call("ApierV1.GetCacheStats", utils.AttrCacheStats{}, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if rcvStats.Destinations != 9 {
		t.Errorf("Calling ApierV1.GetCacheStats received: %+v", rcvStats)
	}
}

func TestDestManagLoadTariffPlanFromFolderAddOne(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "destinations", "addone")}
	if err := destRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &destLoadInst); err != nil {
		t.Error(err)
	} else if destLoadInst.LoadId == "" {
		t.Error("Empty loadId received, loadInstance: ", destLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestDestManagAddOneDestinationLoaded(t *testing.T) {
	if !*testIntegration {
		return
	}
	dests := make([]*engine.Destination, 0)
	if err := destRPC.Call("ApierV2.GetDestinations", v2.AttrGetDestinations{DestinationIDs: []string{}}, &dests); err != nil {
		t.Error("Got error on ApierV2.GetDestinations: ", err.Error())
	} else if len(dests) != 7 {
		t.Errorf("Calling ApierV2.GetDestinations got reply: %v", utils.ToIJSON(dests))
	}

	var rcvStats utils.CacheStats
	if err := destRPC.Call("ApierV1.GetCacheStats", utils.AttrCacheStats{}, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if rcvStats.Destinations != 10 {
		t.Errorf("Calling ApierV1.GetCacheStats received: %+v", rcvStats)
	}
}
