//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

package general_tests

import (
	"net"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	dcCfgPath string
	dcCfg     *config.CGRConfig
	dcRPC     *birpc.Client
	dcConfDIR string // run the tests for specific configuration

	// subtests to be executed for each confDIR
	sTestsDataConverterIT = []func(t *testing.T){
		testDCRemoveFolders,
		testDCCreateFolders,

		testDCInitConfig,
		testDCFlushDBs,
		testDCStartEngine,
		testDCRpcConn,
		testDCWriteCSVs,
		testDCLoaderRun,
		testDCAttributeProcessEvent,

		testDCRemoveFolders,
		testDCKillEngine,
	}
)

// Tests starting here
func TestDataConverterIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		dcConfDIR = "dataconverter_internal"
	case utils.MetaMySQL:
		dcConfDIR = "dataconverter_mysql"
	case utils.MetaMongo:
		dcConfDIR = "dataconverter_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsDataConverterIT {
		t.Run(dcConfDIR, stest)
	}
}

func testDCInitConfig(t *testing.T) {
	var err error
	dcCfgPath = path.Join(*utils.DataDir, "conf", "samples", dcConfDIR)
	if dcCfg, err = config.NewCGRConfigFromPath(context.Background(), dcCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testDCFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(dcCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(dcCfg); err != nil {
		t.Fatal(err)
	}
}

func testDCStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dcCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testDCRpcConn(t *testing.T) {
	var err error
	dcRPC, err = engine.NewRPCClient(dcCfg.ListenCfg(), *utils.Encoding)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testDCCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestDataConverter/in", 0755); err != nil {
		t.Error(err)
	}
}

func testDCRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestDataConverter/in"); err != nil {
		t.Error(err)
	}
}

func testDCWriteCSVs(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join(dcCfg.LoaderCfg()[0].TpInDir, fileName))
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
cgrates.org,ATTR_DC,*string:~*req.Account:1001,;20,,,,,,
cgrates.org,ATTR_DC,,,,,,*req.VariableInSeconds,*variable,~*opts.variable01{*duration_seconds}
cgrates.org,ATTR_DC,,,,,,*req.MultipliedVariable,*variable,~*opts.variable02{*multiply:2.5}
cgrates.org,ATTR_DC,,,,,,*req.DivideVariable,*variable,~*opts.variable03{*divide:1.2}
cgrates.org,ATTR_DC,,,,,,*req.RoundVariable,*variable,~*opts.variable04{*round:2}
cgrates.org,ATTR_DC,,,,,,*req.JSONVariable,*variable,~*opts.variable05{*json}
cgrates.org,ATTR_DC,,,,,,*req.DurationVariable,*variable,~*opts.variable06{*duration}
cgrates.org,ATTR_DC,,,,,,*req.IP2HexVariable,*variable,~*opts.variable07{*ip2hex}
cgrates.org,ATTR_DC,,,,,,*req.String2HexVariable,*variable,~*opts.variable08{*string2hex}
cgrates.org,ATTR_DC,,,,,,*req.UnixTimeVariable,*variable,~*opts.variable12{*unixtime}
cgrates.org,ATTR_DC,,,,,,*req.LengthVariable,*variable,~*opts.variable13{*len}
cgrates.org,ATTR_DC,,,,,,*req.SliceVariable,*variable,~*opts.variable14{*slice}
cgrates.org,ATTR_DC,,,,,,*req.Float64Variable,*variable,~*opts.variable15{*float64}
`); err != nil {
		t.Fatal(err)
	}
}

// cgrates.org,ATTR_DC,,,,,,*req.SIPURIHostVariable,*variable,~*opts.variable09{*sipuri_host}
// cgrates.org,ATTR_DC,,,,,,*req.SIPURIUserVariable,*variable,~*opts.variable10{*sipuri_user}
// cgrates.org,ATTR_DC,,,,,,*req.SIPURIMethodVariable,*variable,~*opts.variable11{*sipuri_method}
// cgrates.org,ATTR_DC,,,,,,*req.E164DomainVariable,*variable,~*opts.variable16{*e164Domain}
// cgrates.org,ATTR_DC,,,,,,*req.E164Variable,*variable,~*opts.variable17{*e164}
// cgrates.org,ATTR_DC,,,,,,*req.LibphonenumberVariable,*variable,~*opts.variable18{*libphonenumber}
// cgrates.org,ATTR_DC,,,,,,*req.TimeStringVariable,*variable,~*opts.variable19{*time_string}
// cgrates.org,ATTR_DC,,,,,,*req.RandomVariable,*variable,~*opts.variable20{*random}
// cgrates.org,ATTR_DC,,,,,,*req.JoinVariable,*variable,~*opts.variable21{*join}
// cgrates.org,ATTR_DC,,,,,,*req.SplitVariable,*variable,~*opts.variable22{*split}

func testDCLoaderRun(t *testing.T) {
	var reply string
	if err := dcRPC.Call(context.Background(), utils.LoaderSv1Run,
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

func testDCAttributeProcessEvent(t *testing.T) {
	expected := engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_DC",
				Fields: []string{"*req.DivideVariable", "*req.DurationVariable", "*req.IP2HexVariable",
					"*req.JSONVariable", "*req.LengthVariable", "*req.MultipliedVariable", "*req.RoundVariable",
					"*req.SliceVariable", "*req.String2HexVariable", "*req.VariableInSeconds", "*req.UnixTimeVariable",
					"*req.Float64Variable"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "DCEvent",
			Event: map[string]any{
				"Account":            1001,
				"DivideVariable":     "2",
				"DurationVariable":   "5s",
				"Float64Variable":    "23.5",
				"IP2HexVariable":     "0x0aff0000",
				"JSONVariable":       "\"123\"",
				"LengthVariable":     "4",
				"MultipliedVariable": "25",
				"RoundVariable":      "0.55",
				"SliceVariable":      "[\"firstitem\",\"seconditem\",\"thirditem\"]",
				"String2HexVariable": "0x3634",
				"UnixTimeVariable":   "1704067200",
				"VariableInSeconds":  "3600",
			},
			APIOpts: map[string]any{
				"variable01": time.Hour,
				"variable02": 10,
				"variable03": 2.4,
				"variable04": 0.54697,
				"variable05": 123,
				"variable06": "5000000000",
				"variable07": net.IPv4(10, 255, 0, 0),
				"variable08": "64",
				"variable12": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05Z07:00"),
				"variable13": "abcd",
				"variable14": `["firstitem", "seconditem", "thirditem"]`,
				"variable15": "23.5",
			},
		},
	}
	var reply engine.AttrSProcessEventReply
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DCEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			"variable01": time.Hour,
			"variable02": 10,
			"variable03": 2.4,
			"variable04": 0.54697,
			"variable05": 123,
			"variable06": "5000000000",
			"variable07": net.IPv4(10, 255, 0, 0),
			"variable08": "64",
			"variable12": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05Z07:00"),
			"variable13": "abcd",
			"variable14": `["firstitem", "seconditem", "thirditem"]`,
			"variable15": "23.5",
		},
	}

	if err := dcRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent, ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		// t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(reply.Event), utils.ToJSON(expected.Event))
	}
	// fmt.Println(utils.ToJSON(reply))
}

func testDCKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
