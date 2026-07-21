//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package registrarc

import (
	"os/exec"
	"path"
	"reflect"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	rpcDir     string
	rpcCMD     *exec.Cmd
	rpcCfgPath string

	rpcsDir     string
	rpcsCfgPath string
	rpcsCfg     *config.CGRConfig
	rpcsRPC     *birpc.Client

	rpchTest = []func(t *testing.T){
		testRPCInitCfg,
		testRPCInitDB,
		testRPCStartEngine,
		testRPCLoadData,
		testRPCChargerSNoAttr,
		testRPCStartRegc,
		testRPCChargerSWithAttr,
		testRPCStopEngines,
		testRPCChargerSNoAttr,
		testRPCStopRegs,
	}
)

func TestRPCHosts(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		rpcDir = "registrarc_rpc_mysql"
		rpcsDir = "registrars_rpc_mysql"
	case utils.MetaMongo:
		rpcDir = "registrarc_rpc_mongo"
		rpcsDir = "registrars_rpc_mongo"
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range rpchTest {
		t.Run(rpcDir, stest)
	}
}

func testRPCInitCfg(t *testing.T) {
	rpcCfgPath = path.Join(*utils.DataDir, "conf", "samples", "registrarc", rpcDir)
	rpcsCfgPath = path.Join(*utils.DataDir, "conf", "samples", "registrarc", rpcsDir)
	var err error
	if rpcsCfg, err = config.NewCGRConfigFromPath(context.Background(), rpcsCfgPath); err != nil {
		t.Error(err)
	}
}

func testRPCInitDB(t *testing.T) {
	if err := engine.InitDB(rpcsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRPCStartEngine(t *testing.T) {
	var err error
	if _, err = engine.StopStartEngine(rpcsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	rpcsRPC = engine.NewRPCClient(t, rpcsCfg.ListenCfg(), *utils.Encoding)
}

func testRPCLoadData(t *testing.T) {
	var reply string
	if err := rpcsRPC.Call(context.Background(), utils.LoaderSv1Run, &loaders.ArgsProcessFolder{
		APIOpts: map[string]any{utils.MetaCache: utils.MetaReload},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testRPCChargerSNoAttr(t *testing.T) {
	cgrEv := &utils.CGREvent{ // matching Charger1
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1010",
		},
		APIOpts: map[string]any{utils.OptsAttributesProcessRuns: 1.},
	}
	expErr := utils.NewErrServerError(rpcclient.ErrDisconnected).Error()
	var rply []*chargers.ChrgSProcessEventReply
	if err := rpcsRPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, cgrEv, &rply); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
}

func testRPCStartRegc(t *testing.T) {
	var err error
	if rpcCMD, err = engine.StartEngine(rpcCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}

func testRPCChargerSWithAttr(t *testing.T) {
	cgrEv := &utils.CGREvent{ // matching Charger1
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1010",
		},
		APIOpts: map[string]any{utils.OptsAttributesProcessRuns: 1.},
	}

	processedEv := []*chargers.ChrgSProcessEventReply{
		{
			ChargerSProfile: "CustomerCharges",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					"Account": "1010",
				},
				APIOpts: map[string]any{
					"*chargeID":              "908bd346b2203977a829c917ba25d1cd784842be",
					"*attributesProcessRuns": 1.,
					"*subsys":                "*chargers",
					"*runID":                 "CustomerCharges",
				},
			},
		}, {
			ChargerSProfile: "Raw",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
				{
					MatchedProfileID: "*constant:*req.RequestType:*none",
					Fields:           []string{"*req.RequestType"},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					"Account":     "1010",
					"RequestType": "*none",
				},
				APIOpts: map[string]any{
					"*chargeID":              "ce15802a8c5e8e9db0ffaf10130ef265296e9ea4",
					"*attributesProcessRuns": 1.,
					"*subsys":                "*chargers",
					"*runID":                 "raw",
					"*attributesProfileIDs":  []any{"*constant:*req.RequestType:*none"},
					"*context":               "*chargers",
				},
			},
		}, {
			ChargerSProfile: "SupplierCharges", AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
				{
					MatchedProfileID: "cgrates.org:ATTR_SUPPLIER1",
					Fields:           []string{"*req.Subject"},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					"Account": "1010",
					"Subject": "SUPPLIER1",
				},
				APIOpts: map[string]any{
					"*chargeID":              "c0766c230f77b0ee496629be7efa0db24e208cfe",
					"*context":               "*chargers",
					"*attributesProcessRuns": 1.,
					"*subsys":                "*chargers",
					"*runID":                 "SupplierCharges",
					"*attributesProfileIDs":  []any{"ATTR_SUPPLIER1"},
				},
			},
		},
	}
	var rply []*chargers.ChrgSProcessEventReply
	if err := rpcsRPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, cgrEv, &rply); err != nil {
		t.Fatal(err)
	}
	sort.Slice(rply, func(i, j int) bool {
		return rply[i].ChargerSProfile < rply[j].ChargerSProfile
	})
	if !reflect.DeepEqual(rply, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(rply))
	}
}

func testRPCStopEngines(t *testing.T) {
	if err := rpcCMD.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
}

func testRPCStopRegs(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
