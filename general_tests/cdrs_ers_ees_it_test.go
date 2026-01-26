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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRsERsEEs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	csvcontent := `csvfile1,1001,1303535,1727779754,1727779754,199
csvfile2,1002,1303535,1727779754,1727779754,201
csvfile3,1003,1303535,1727779754,1727779754,300
csvfile4,1004,1303535,1727779754,1727779754,201
csvfile5,1004,1303535,1727779754,1727779754,100`
	if err := os.MkdirAll("/tmp/CDRstoRead", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("/tmp/CDRsProcessed", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal(err)
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf/samples/cdrs_ers_ees"),
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_2Bals,,,
cgrates.org,1002,PACKAGE_2Bals,,,
cgrates.org,1003,PACKAGE_2Bals,,,
cgrates.org,1004,PACKAGE_2Bals,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_2Bals,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_200internat,*voice,,,,,*unlimited,,200m,20,,,
ACT_TOPUP,*topup_reset,,,balance_PAYG,*voice,,,accSubject,,*unlimited,,9999m,10,,,`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,DEFAULT,*none,20`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,accSubject,,RP_ANY,`,
		},
		// LogBuffer: new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/CDRstoRead"); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll("/tmp/CDRsProcessed"); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal(err)
		}
	})
	client, _ := ng.Run(t)

	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile("/tmp/CDRstoRead/file1.csv", []byte(csvcontent), 0644); err != nil {
		t.Fatalf("could not write to file /tmp/CDRstoRead/file1.csv: %v", err)
	}
	t.Run("TestERsRunReader", func(t *testing.T) {
		var rply *string
		if err := client.Call(context.Background(), utils.ErSv1RunReader, &ers.V1RunReaderParams{ReaderID: "file_csv_reader"}, &rply); err != nil {
			t.Fatal(err)
		}
	})
	time.Sleep(1700 * time.Millisecond)

	// print the cdrs to see the account details
	// t.Run("TestGetCDRs", func(t *testing.T) {
	// 	var cdrs []*engine.CDR
	// 	if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	t.Log(utils.ToIJSON(cdrs))
	// })

	// read the PayAsYouGo usage used from the generated CDRs
	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCnt := "RunID,ToR,OriginID,RequestType,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage,Cost,PAYGUsage\n" +
			"*default,*voice,csvfile1,*postpaid,call,1001,accSubject,1303535,2024-10-01T10:49:14Z,2024-10-01T10:49:14Z,0,11940,0\n" + // expected no PAYG usage to be used
			"*default,*voice,csvfile2,*postpaid,call,1002,accSubject,1303535,2024-10-01T10:49:14Z,2024-10-01T10:49:14Z,60,12060,60000000000\n" + // expected 1m PAYG usage to be used
			"*default,*voice,csvfile3,*postpaid,call,1003,accSubject,1303535,2024-10-01T10:49:14Z,2024-10-01T10:49:14Z,6000,18000,6000000000000\n" + // expected 100m PAYG usage to be used
			"*default,*voice,csvfile4,*postpaid,call,1004,accSubject,1303535,2024-10-01T10:49:14Z,2024-10-01T10:49:14Z,60,12060,60000000000\n" + // expected 1m PAYG usage to be used, free minute account is emptied
			"*default,*voice,csvfile5,*postpaid,call,1004,accSubject,1303535,2024-10-01T10:49:14Z,2024-10-01T10:49:14Z,6000,6000,6000000000000\n" // expected 100m PAYG usage to be used, it took only from the PAYG account
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCnt) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
		}
	})

}
