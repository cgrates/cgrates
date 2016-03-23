package sessionmanager

import (
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var testIntegration = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args
var waitRater = flag.Int("wait_rater", 150, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

var daCfgPath string
var daCfg *config.CGRConfig
var smgRPC *rpc.Client
var err error

func TestSMGDataInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	daCfgPath = path.Join(*dataDir, "conf", "samples", "smg")
	// Init config first
	var err error
	daCfg, err = config.NewCGRConfigFromFolder(daCfgPath)
	if err != nil {
		t.Error(err)
	}
	daCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(daCfg)
}

// Remove data in both rating and accounting db
func TestSMGDataResetDataDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGDataResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGDataStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGDataApierRpcConn(t *testing.T) {
	if !*testIntegration {
		return
	}
	var err error
	smgRPC, err = jsonrpc.Dial("tcp", daCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGDataTPFromFolder(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	var loadInst engine.LoadInstance
	if err := smgRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGDataLastUsedData(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1010"}
	eAcntVal := 6.59000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.ACCID:       "12349",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1010",
		utils.SUBJECT:     "1010",
		utils.DESTINATION: "222",
		utils.CATEGORY:    "data",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1048576",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.SessionStart", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 5.190020
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.ACCID:       "12349",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1010",
		utils.SUBJECT:     "1010",
		utils.DESTINATION: "222",
		utils.CATEGORY:    "data",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "1048576",
		utils.LastUsed:    "20000",
	}
	if err := smgRPC.Call("SMGenericV1.SessionUpdate", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 5.123360
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.ACCID:       "12349",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1010",
		utils.SUBJECT:     "1010",
		utils.DESTINATION: "222",
		utils.CATEGORY:    "data",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.LastUsed:    "0",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.SessionEnd", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 5.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
}
