//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrWDcCfgPath string
	attrWDcCfg     *config.CGRConfig
	attrWDcRpc     *birpc.Client
	attrWDcConfDIR string //run tests for specific configuration

	sTestsAttrWDc = []func(t *testing.T){
		testAttrWDcInitCfg,
		testAttrWDcInitDataDb,
		testAttrWDcResetStorDb,
		testAttrWDcStartEngine,
		testAttrWDcRPCConn,
		testAttrWDcLoadFromFolder,
		testAttrWDcProcessEvent,
		testAttrWDcProcessEventWithStat,
		testAttrWDcStripConverter,
		testAttrWDcStopEngine,
	}
)

func TestAttrWDcIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		attrWDcConfDIR = "tutinternal"
	case utils.MetaMySQL:
		attrWDcConfDIR = "tutmysql"
	case utils.MetaMongo:
		attrWDcConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknow Database type")
	}
	for _, stest := range sTestsAttrWDc {
		t.Run(attrWDcConfDIR, stest)
	}

}

func testAttrWDcInitCfg(t *testing.T) {
	var err error
	attrWDcCfgPath = path.Join(*utils.DataDir, "conf", "samples", attrWDcConfDIR)
	attrWDcCfg, err = config.NewCGRConfigFromPath(attrWDcCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAttrWDcInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(attrWDcCfg); err != nil {
		t.Fatal(err)
	}
}

func testAttrWDcResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(attrWDcCfg); err != nil {
		t.Fatal(err)
	}
}

func testAttrWDcStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrWDcCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAttrWDcRPCConn(t *testing.T) {
	attrWDcRpc = engine.NewRPCClient(t, attrWDcCfg.ListenCfg())
}

func testAttrWDcLoadFromFolder(t *testing.T) {
	var reply string
	attrs := utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "dataconverters")}
	if err := attrWDcRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)

}

func testAttrWDcProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttrWDcProcessEvent",
		Event: map[string]any{
			utils.Cost: "10.252",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_SEC"},
		AlteredFields:   []string{"*req.Cost"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttrWDcProcessEvent",
			Event: map[string]any{
				utils.Cost: "10.26",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply

	if err := attrWDcRpc.Call(context.Background(), utils.AttributeSv1ProcessEvent, ev, &rplyEv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttrWDcProcessEventWithStat(t *testing.T) {
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.AnswerTime:   time.Date(2023, 9, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        15 * time.Second,
			utils.Cost:         10.0,
		},
	}
	if err := attrWDcRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.AnswerTime:   time.Date(2023, 9, 10, 9, 0, 0, 0, time.UTC),
			utils.Usage:        50 * time.Second,
			utils.Cost:         23.5,
		},
	}
	if err := attrWDcRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttrWDcProcessEventWithStat",
		Event: map[string]any{
			"EventName": "StatsTest",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_STAT"},
		AlteredFields:   []string{"*req.AcdMetric"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttrWDcProcessEventWithStat",
			Event: map[string]any{
				"EventName": "StatsTest",
				"AcdMetric": "33",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}

	var replyEv engine.AttrSProcessEventReply

	if err := attrWDcRpc.Call(context.Background(), utils.AttributeSv1ProcessEvent, ev, &replyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, replyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(replyEv))
	}
}
func testAttrWDcStripConverter(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttrWDcStripConverter",
		Event: map[string]any{
			"EventName":        "CallTest",
			utils.AccountField: "1001",
			utils.AnswerTime:   time.Date(2023, 9, 10, 9, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_VARIABLE"},
		AlteredFields:   []string{"*req.AnswerTime", "*req.Category"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttrWDcStripConverter",
			Event: map[string]any{
				"EventName":        "CallTest",
				utils.AccountField: "1001",
				utils.Category:     "Call",
				utils.AnswerTime:   "2023-09-10 09:00:00 +0000 UTC",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}

	var replyEv engine.AttrSProcessEventReply

	if err := attrWDcRpc.Call(context.Background(), utils.AttributeSv1ProcessEvent, ev, &replyEv); err != nil {
		t.Error(err)
	} else if sort.Slice(replyEv.AlteredFields, func(i, j int) bool {
		return replyEv.AlteredFields[i] < replyEv.AlteredFields[j]
	}); !reflect.DeepEqual(replyEv, eRply) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(eRply), utils.ToJSON(replyEv))
	}

}

func testAttrWDcStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
