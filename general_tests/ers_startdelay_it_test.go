//go:build integration

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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestErsStartDelay(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported db type value")
	}
	csvcontent := `1d65221e540dmp55gw,1001,1303535,1727779754,1727779754,60`
	csvFd, procFd := t.TempDir(), t.TempDir()
	filePath := filepath.Join(csvFd, fmt.Sprintf("file1%s", utils.CSVSuffix))
	if err := os.WriteFile(filePath, []byte(csvcontent), 0644); err != nil {
		t.Fatalf("could not write to file %s: %v", filePath, err)
	}
	content := fmt.Sprintf(`{
		"general": {
			"log_level": 7
		},
		"data_db": {
			"db_type": "*internal"
		},
		"stor_db": {
			"db_type": "*internal"
		},
		"cdrs":{
		"enabled":true,
		"rals_conns":["*localhost"]
		},
		"sessions":{
		   "enabled": true,
		   "cdrs_conns":["*localhost"]
        },
		"rals": {
			"enabled": true
		},
		"apiers":{
		"enabled":true,
		},
		"ers": {
			"enabled": true,
			"sessions_conns": ["*localhost"],
			"readers": [
				{
					"id": "file_csv_reader",
					"run_delay":  "-1",
					"start_delay":"1s",
					"type": "*file_csv",
					"source_path": "%s",
					"flags": ["*cdrs"],
					"processed_path": "%s",
					"fields":[
						{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
                        {"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.0", "mandatory": true},
                        {"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*rated", "mandatory": true},
						{"tag":"Category","path":"*cgreq.Category","type":"*constant","value":"call"},
						{"tag":"Subject","path":"*cgreq.Subject","type":"*variable","value":"~*req.1"},
						{"tag":"Destination","path":"*cgreq.Destination","type":"*variable","value":"~*req.2"},
						{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable", "value": "~*req.3"},
                        {"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable", "value": "~*req.4"},
                        {"tag": "Usage", "path": "*cgreq.Usage", "filters": ["*notempty:~*req.5:"],"type": "*variable", "value": "~*req.5;s", "mandatory": true},
					]
				}
			]
            }
		}`, csvFd, procFd)

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	t.Run("CheckForCdrs", func(t *testing.T) {
		newtpFiles := map[string]string{
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1.7,60s,1s,0s`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,2,0,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
		}
		engine.LoadCSVs(t, client, "", newtpFiles)
		time.Sleep(100 * time.Millisecond)

		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{Subjects: []string{"1001"}}}, &cdrs); err == nil {
			t.Error(err)
		}
	})
	time.Sleep(1 * time.Second)
	t.Run("ErsAfterStartDelay", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{Subjects: []string{"1001"}}}, &cdrs); err != nil {
			t.Error(err)
		} else if len(cdrs) != 1 {
			t.Errorf("expected a CDR generated from ers")
		} else if cdrs[0].Cost != 1.7 {
			t.Errorf("expected %f,received %f", 1.7, cdrs[0].Cost)
		}
	})
}
