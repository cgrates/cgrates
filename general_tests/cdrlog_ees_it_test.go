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
	"os"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCdrLogEes(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	dir := "/tmp/testComposed"

	content := `{
		// Sample CGRateS Configuration file for EEs
		//
		// Copyright (C) ITsysCOM GmbH
		
		"general": {
			"log_level": 7,
		},
		
		"listen": {
			"rpc_json": ":2012",
			"rpc_gob": ":2013",
			"http": ":2080",
		},
		
	
       "data_db": {
	     "db_type": "*internal"
        },


        "stor_db": {
	       "db_type": "*internal"
        },
		
		"rals": {
			"enabled": true,
		},
		
		
		"schedulers": {
			"enabled": true,
			"cdrs_conns":["*internal"],
		},
		
		
		"cdrs": {
			"enabled": true,
			"chargers_conns": ["*localhost"],
			"rals_conns": ["*internal"],
			"session_cost_retries": 0,
			"ees_conns":["*internal"],
		},
		
		
		"chargers": {
			"enabled": true,
			"attributes_conns": ["*internal"],
		},
		
		
		"attributes": {
			"enabled": true,
			"stats_conns": ["*localhost"],
	         "apiers_conns": ["*localhost"]
		},
		
		
		"ees": {
			"enabled": true,
			"attributes_conns":["*internal"],
			"exporters": [
				{
					"id": "CSVExporter",
					"type": "*file_csv",
					"export_path": "/tmp/testComposed",
					"flags": ["*attributes","*log"],
					"attribute_context": "customContext",
					"attempts": 1,
					"field_separator": ",",
					"fields":[
		
					{"tag": "Number", "path": "*exp.Number", "type": "*variable", "value": "~*dc.NumberOfEvents"},
						{"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"},
						{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
						{"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR"},
						{"tag": "OriginID1", "path": "*exp.OriginID", "type": "*composed", "value": "~*req.ComposedOriginID1"},
						{"tag": "OriginID2", "path": "*exp.OriginID", "type": "*composed", "value": "~*req.ComposedOriginID2"},
						{"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType"},
						{"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant"},
						{"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category"},
						{"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account"},
						{"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject"},
						{"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination"},
						{"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime", "layout": "2006-01-02T15:04:05Z07:00"},
						{"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime", "layout": "2006-01-02T15:04:05Z07:00"},
						{"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage"},
						{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
						{"tag": "RatingPlan", "path": "*exp.RatingPlan", "type": "*variable", "value": "~*ec.Charges[0].Rating.RatingFilter.RatingPlanID"},
						{"tag": "RatingPlanSubject", "path": "*exp.RatingPlanSubject", "type": "*variable", "value": "~*ec.Charges[0].Rating.RatingFilter.Subject"}, 
		 
						{"tag": "NumberOfEvents", "path": "*trl.NumberOfEvents", "type": "*variable", "value": "~*dc.NumberOfEvents"},
						{"tag": "TotalDuration", "path": "*trl.TotalDuration", "type": "*variable", "value": "~*dc.TotalDuration"},
						{"tag": "TotalSMSUsage", "path": "*trl.TotalSMSUsage", "type": "*variable", "value": "~*dc.TotalSMSUsage"},
						{"tag": "TotalCost", "path": "*trl.TotalCost", "type": "*variable", "value": "~*dc.TotalCost{*round:4}"}, 
					],
				},
		
		
			]
		},
		
		
		"apiers": {
			"enabled": true,
			"scheduler_conns": ["*internal"],
		},
		
		"templates": {
			"requiredFields": [
				{"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"},
				{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
				{"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR"},
				{"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID"},
				{"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType"},
				{"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant"},
				{"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category"},
				{"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account"},
				{"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject"},
				{"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination"},
				{"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime", "layout": "2006-01-02T15:04:05Z07:00"},
				{"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime", "layout": "2006-01-02T15:04:05Z07:00"},
				{"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage"},
				{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"}
			],
		},
		
		
		}
		`

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal("Error removing folder: ", dir, err)
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal("Error creating folder: ", dir, err)

	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    map[string]string{},
	}
	client, _ := ng.Run(t)

	t.Run("AddBalanceCdrLog", func(t *testing.T) {
		var reply string
		attrs := &v1.AttrAddBalance{
			Account:     "testAccAddBalance",
			BalanceType: utils.MetaMonetary,
			Value:       1.5,
			Cdrlog:      true,
		}
		if err := client.Call(context.Background(), utils.APIerSv1AddBalance, &attrs, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
		}
		time.Sleep(50 * time.Millisecond)
		var cdrs []*engine.ExternalCDR
		req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}}
		if err := client.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
		} else if cdrs[0].Cost != 1.5 {
			t.Errorf("Expected cost to be %v,Received %v", 1.5, cdrs[0].Cost)
		}
	})

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal("Error removing folder: ", dir, err)
	}

}
