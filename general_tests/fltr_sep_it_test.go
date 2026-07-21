//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"os"
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
	fltrSepCfgPath string
	fltrSepCfg     *config.CGRConfig
	fltrSepRPC     *birpc.Client
	fltrSepDelay   int
	fltrSepConfDIR string //run tests for specific configuration

	sTestsFltrSep = []func(t *testing.T){
		testFltrSepRemoveFolders,
		testFltrSepCreateFolders,

		testFltrSepLoadConfig,
		testFltrSepFlushDBs,
		testFltrSepStartEngine,
		testFltrSepRpcConn,

		testFltrSepWriteCSVs,
		testFltrSepLoadTarrifPlans,
		testFltrSepFilterSeparation,

		testFltrSepStopEngine,
		testFltrSepRemoveFolders,
	}
)

// Test start here
func TestFltrSepIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		fltrSepConfDIR = "fltr_sep_internal"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		fltrSepConfDIR = "fltr_sep_mysql"
	case utils.MetaMongo:
		fltrSepConfDIR = "fltr_sep_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsFltrSep {
		t.Run(fltrSepConfDIR, stest)
	}
}

func testFltrSepLoadConfig(t *testing.T) {
	var err error
	fltrSepCfgPath = path.Join(*utils.DataDir, "conf", "samples", fltrSepConfDIR)
	if fltrSepCfg, err = config.NewCGRConfigFromPath(context.Background(), fltrSepCfgPath); err != nil {
		t.Error(err)
	}
	fltrSepDelay = 1000
}

func testFltrSepFlushDBs(t *testing.T) {
	if err := engine.InitDB(fltrSepCfg); err != nil {
		t.Fatal(err)
	}
}

func testFltrSepStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrSepCfgPath, fltrSepDelay); err != nil {
		t.Fatal(err)
	}
}

func testFltrSepRpcConn(t *testing.T) {
	fltrSepRPC = engine.NewRPCClient(t, fltrSepCfg.ListenCfg(), *utils.Encoding)
}

func testFltrSepWriteCSVs(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join(fltrSepCfg.LoaderCfg()[0].TpInDir, fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ATTR_FLTR_TEST,*string:~*req.Account:1001|1002|1003|1101;*prefix:~*req.Account:10,;20,;false,,,,,
cgrates.org,ATTR_FLTR_TEST,,,,,,*req.TestField,*constant,testValue
`); err != nil {
		t.Fatal(err)
	}

}

func testFltrSepLoadTarrifPlans(t *testing.T) {
	var reply string
	if err := fltrSepRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaCache:       utils.MetaReload,
				utils.MetaStopOnError: false,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testFltrSepFilterSeparation(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "filter_separation_test",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaAny,
		},
	}

	eAttrPrf := &utils.APIAttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_FLTR_TEST",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003|1101", "*prefix:~*req.Account:10"},
		Attributes: []*utils.ExternalAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "TestField",
				Value: "testValue",
				Type:  utils.MetaConstant,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}

	var attrReply *utils.APIAttributeProfile

	// first option of the first filter and the second filter match
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(attrReply.FilterIDs, func(i, j int) bool {
			return attrReply.FilterIDs[i] > attrReply.FilterIDs[j]
		})
		if !reflect.DeepEqual(eAttrPrf, attrReply) {
			t.Errorf("expected: %+v, \nreceived: %+v",
				utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
		}
	}

	// third option of the first filter and the second filter match
	ev.Event[utils.AccountField] = "1003"
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(attrReply.FilterIDs, func(i, j int) bool {
			return attrReply.FilterIDs[i] > attrReply.FilterIDs[j]
		})
		if !reflect.DeepEqual(eAttrPrf, attrReply) {
			t.Errorf("expected: %+v, \nreceived: %+v",
				utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
		}
	}

	// the second filter matches while none of the options from the first filter match
	ev.Event[utils.AccountField] = "1004"
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	// fourth option of the first filter matches while the second filter doesn't
	ev.Event[utils.AccountField] = "1101"
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFltrSepStopEngine(t *testing.T) {
	if err := engine.KillEngine(fltrSepDelay); err != nil {
		t.Error(err)
	}
}

func testFltrSepCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestFltrSepIT/in", 0755); err != nil {
		t.Error(err)
	}
}

func testFltrSepRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestFltrSepIT/in"); err != nil {
		t.Error(err)
	}
}
