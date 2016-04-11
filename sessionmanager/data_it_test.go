package sessionmanager

import (
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

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
	eAcntVal := 50000000000.000000
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998945280.000000
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998924800.000000
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
	eAcntVal = 49999979520.000000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
}

func TestSMGDataLastUsedMultipleData(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1010"}
	eAcntVal := 49999979520.000000
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998924800.000000 // 1054720
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998904320.000000 // 20480
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998883840.000000 // 20480
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998863360.000000 // 20480
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998842880.000000 // 20480
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
	eAcntVal = 49999897600.000000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
}

func TestSMGDataDerivedChargingNoCredit(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1011"}
	eAcntVal := 50000.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12349",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1011",
		utils.SUBJECT:     "1011",
		utils.DESTINATION: "+49",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "100",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.SessionStart", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 0 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 50000.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
}

func TestSMGDataTTLExpired(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1010"}
	eAcntVal := 49999897600.000000
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
	if maxUsage != 1.048576e+06 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 49998842880.000000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	time.Sleep(50 * time.Millisecond)
	eAcntVal = 49999887360.000000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
}
