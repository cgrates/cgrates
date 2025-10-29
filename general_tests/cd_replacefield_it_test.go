//go:build integration
// +build integration

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
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdRFCfgPath string
	cdRFCfg     *config.CGRConfig
	cdRFRPC     *rpc.Client
	cdRFDelay   int

	sTestsCDReplaceField = []func(t *testing.T){
		testCDReplaceFieldLoadConfig,
		testCDReplaceFieldInitDataDb,
		testCDReplaceFieldResetStorDb,
		testCDReplaceFieldStartEngine,
		testCDReplaceFieldRpcConn,

		testCDReplaceFieldStoreCDR,
		testCDReplaceFieldExportCDR,

		testCDReplaceFieldStopEngine,
	}
)

func TestCDReplaceFieldIT(t *testing.T) {
	for _, stest := range sTestsCDReplaceField {
		t.Run("cost details replace field", stest)
	}
}

func testCDReplaceFieldLoadConfig(t *testing.T) {
	content := `{
"general": {
	"log_level": 7,
	"node_id": "cd_replace_field",
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"cdrs": {
	"enabled": true,
},
"apiers": {
	"enabled": true
},
"cdre": {
	"csv_exporter": {
		"export_format": "*file_csv",
		"export_path": "/tmp/",
		"fields": [
			{"path": "*exp.CGRID", "type": "*composed", "value": "~*req.CGRID"},
			// {
			// 	"path": "*exp.TariffClass",
			// 	"type": "*composed",
			// 	"value": "~*req.CostDetails"
			// },
			{
				"path": "*exp.TariffClass",
				"type": "*composed",
				"filters": [
					"*rsr::~*req.CostDetails(~^.+\"DestinationID\":\"\\w+_(\\w{5})\".+$)"
				],
				"value": "~*req.CostDetails:s/^.+\"DestinationID\":\"\\w+_(\\w{5})\".+$/${1}/:s/\"DestinationID\":\"INTERNAL\"/ON010/"
			},
		],
	},
},
}`
	folderNameSuffix, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		t.Fatalf("could not generate random number for folder name suffix, err: %s", err.Error())
	}
	cdRFCfgPath = fmt.Sprintf("/tmp/config%d", folderNameSuffix)
	err = os.MkdirAll(cdRFCfgPath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(cdRFCfgPath, "cgrates.json")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if cdRFCfg, err = config.NewCGRConfigFromPath(cdRFCfgPath); err != nil {
		t.Error(err)
	}
	cdRFDelay = 100
}

func testCDReplaceFieldInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdRFCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDReplaceFieldResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(cdRFCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDReplaceFieldStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdRFCfgPath, cdRFDelay); err != nil {
		t.Fatal(err)
	}
}

func testCDReplaceFieldRpcConn(t *testing.T) {
	var err error
	cdRFRPC, err = newRPCClient(cdRFCfg.ListenCfg())
	if err != nil {
		t.Fatal("Could not connect to engine: ", err.Error())
	}
}

func testCDReplaceFieldStoreCDR(t *testing.T) {
	sampleCD := &engine.EventCost{}
	err := json.Unmarshal(
		[]byte(`{"CGRID":"31ee543fdc15e011332e0a0952c0b37b161e3a59","RunID":"*default","StartTime":"2023-06-19T11:15:46-04:00","Usage":10000000000,"Cost":0.1,"Charges":[{"RatingID":"efd715b","Increments":[{"Usage":1000000000,"Cost":0.01,"AccountingID":"c946243","CompressFactor":5}],"CompressFactor":2}],"AccountSummary":{"Tenant":"cgrates.org","ID":"1001","BalanceSummaries":[{"UUID":"e4bf588f-583b-44d9-9b5d-cd12d9b7331c","ID":"test","Type":"*monetary","Value":8.5682,"Disabled":false}],"AllowNegative":false,"Disabled":false},"Rating":{"efd715b":{"ConnectFee":0,"RoundingMethod":"*up","RoundingDecimals":4,"MaxCost":0.12,"MaxCostStrategy":"*disconnect","TimingID":"240ecf8","RatesID":"8042384","RatingFiltersID":"ef07765"}},"Accounting":{"c946243":{"AccountID":"cgrates.org:1001","BalanceUUID":"e4bf588f-583b-44d9-9b5d-cd12d9b7331c","RatingID":"","Units":0.01,"ExtraChargeID":""}},"RatingFilters":{"ef07765":{"DestinationID":"CST_316_NL002","DestinationPrefix":"1003","RatingPlanID":"RP_1001","Subject":"*out:cgrates.org:call:1001"}},"Rates":{"8042384":[{"GroupIntervalStart":0,"Value":0.01,"RateIncrement":1000000000,"RateUnit":1000000000}]},"Timings":{"240ecf8":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00"}}}`),
		sampleCD)
	if err != nil {
		t.Fatal(err)
	}
	setupTime := time.Now()
	cdr := &engine.CDR{
		CGRID:       "CDR_1",
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "1",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.META_PREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   setupTime,
		AnswerTime:  setupTime.Add(2 * time.Second),
		RunID:       utils.MetaDefault,
		Usage:       30 * time.Second,
		Cost:        1.01,
		CostDetails: sampleCD,
	}

	var reply string
	if err := cdRFRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testCDReplaceFieldExportCDR(t *testing.T) {
	attr := v1.ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "csv_exporter",
		},
		Verbose: true,
	}
	var reply *v1.RplExportedCDRs
	err := cdRFRPC.Call(utils.APIerSv1ExportCDRs, attr, &reply)
	if err != nil {
		t.Error(err)
	}
	exportedCDR, err := os.ReadFile(reply.ExportedPath)
	if err != nil {
		t.Fatal(err)
	}
	exp := "CDR_1,NL002\n"
	if rcv := string(exportedCDR); rcv != exp {
		t.Errorf("expected <%s>, received <%s>", exp, rcv)
	}

}

func testCDReplaceFieldStopEngine(t *testing.T) {
	err := engine.KillEngine(cdRFDelay)
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(cdRFCfgPath)
	if err != nil {
		t.Error(err)
	}
}
