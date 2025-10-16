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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestErsStartDelay(t *testing.T) {
	csvcontent := ``
	csvFd, csvFd2, procFd := t.TempDir(), t.TempDir(), t.TempDir()
	filePath := filepath.Join(csvFd, fmt.Sprintf("file1%s", utils.CSVSuffix))
	if err := os.WriteFile(filePath, []byte(csvcontent), 0644); err != nil {
		t.Fatalf("could not write to file %s: %v", filePath, err)
	}
	content := fmt.Sprintf(`{
		"general": {
			"log_level": 7
		},
		"db": {
			"db_conns": {
				"*default": {
					"db_type": "*internal"
				}
			},
			"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
			}
		},
		"cdrs":{
		"enabled":true,
		"rates_conns": ["*localhost"],	
		"opts":{
		"*rates":[
		 {
			"Value":true
		 }
		  ]
		      }
		},	
		"sessions":{
		   "enabled": true,
		   "cdrs_conns":["*localhost"]
        },
		"logger": {
     	"level": 7,								},
		"rates": {
			"enabled": true,
			"opts": {
			 "*usage": [			
			{
				"Value": "~*req.Usage",
	       	}
				]
			},
		},
		"admins":{
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
					"type": "*fileCSV",
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
				},
				{
					"id": "file_csv_reader2",
					"run_delay":  "-1",
					"start_delay":"2s",
					"type": "*fileCSV",
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
				},
			]
            }
		}`, csvFd, procFd, csvFd2, procFd)
	var buf bytes.Buffer
	ng := engine.TestEngine{
		ConfigJSON: content,
		LogBuffer:  &buf,
		Encoding:   *utils.Encoding,
	}

	fileIdx := 0
	createFile := func(t *testing.T, dir, ext, content string) {
		fileIdx++
		filePath := filepath.Join(dir, fmt.Sprintf("file%d%s", fileIdx, ext))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("could not write to file %s: %v", filePath, err)
		}
	}
	client, _ := ng.Run(t)
	createFile(t, csvFd, utils.CSVSuffix, "csvfile1,1001,1303535,1727779754,1727779754,60")
	createFile(t, csvFd2, utils.CSVSuffix, "csvfile2,1001,1303535,1727779754,1727779754,120")
	t.Run("ReaderBeforeStartDelay", func(t *testing.T) {
		var cdrs []*engine.CDRsql
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org"}, &cdrs); err == nil || strings.Contains(utils.ErrNotFound.Error(), err.Error()) {
			t.Errorf("expected %v, received %v", utils.ErrNotFound, err)
		}
	})

	t.Run("ReaderAfterStartDelay1s", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile, &utils.RateProfile{Tenant: "cgrates.org", ID: "1001", FilterIDs: []string{"*string:~*req.Subject:1001"}, Rates: map[string]*utils.Rate{
			"RT_ANY": {ID: "RT_ANY", IntervalRates: []*utils.IntervalRate{{IntervalStart: utils.NewDecimal(0, 0), Unit: utils.NewDecimalFromUsageIgnoreErr("60s"), RecurrentFee: utils.NewDecimalFromFloat64(1.7), Increment: utils.NewDecimalFromUsageIgnoreErr("1s")}}},
		}}, &reply); err != nil {
			t.Fatalf("Failed to set rate profile: %v", err)
		}

		time.Sleep(1200 * time.Millisecond)
		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org", FilterIDs: []string{"*string:~*req.OriginID:csvfile1"}}, &cdrs); err != nil {
			t.Error(err)
		} else if len(cdrs) != 1 {
			t.Errorf("expected a CDR generated from ers")
		} else if cdrs[0].Opts["*cost"] != 1.7 {
			t.Errorf("expected %f,received %f", 1.7, cdrs[0].Opts["*cost"])
		}
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org", FilterIDs: []string{"*string:~*req.OriginID:csvfile2"}}, &cdrs); err == nil || strings.Contains(utils.ErrNotFound.Error(), err.Error()) {
			t.Errorf("expected %v, received %v", utils.ErrNotFound, err)
		}

	})
	time.Sleep(2200 * time.Millisecond)
	t.Run("ReaderAfterStartDelay2s", func(t *testing.T) {
		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org", FilterIDs: []string{"*string:~*req.OriginID:csvfile2"}}, &cdrs); err != nil {
			t.Error(err)
		} else if len(cdrs) != 1 {
			t.Errorf("expected a CDR generated from ers")
		} else if cdrs[0].Opts["*cost"] != 3.4 {
			t.Errorf("expected %f,received %f", 3.4, cdrs[0].Opts["*cost"])
		}
	})

}
