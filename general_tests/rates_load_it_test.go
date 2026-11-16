//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func TestLoadRatesOverwriteAndAlter(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unkwown database type")
	}
	tpPath := t.TempDir()
	content := fmt.Sprintf(`{
"admins": {
	"enabled": true,
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
			"opts":{
			"internalDBRewriteInterval": "0s",
			"internalDBDumpInterval": "0s"
		}
    	}
	},
},
"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"tenant": "cgrates.org",
		"lockfilePath": ".cgr.lck",
		"tpInPath": "%s",
		"tpOutPath": ""
	},
	{
		"id": "*alterRatesLoader",				
		"enabled": true,				
		"tenant": "",					
		"runDelay": "0",				
		"lockfilePath": ".cgr.lck",			
		"conns": {
			"*caches": [{"connIDs": ["*internal"]}]
		},
		"fieldSeparator": ",",			
		"tpInPath": "%s",	
		"tpOutPath": "",	
		"action": "*store", 				
		"opts": {
			"*withIndex": true
		},
		"data":[					
			{
				"type": "*rateProfiles",	
				"fileName": "Rates.csv",
				"flags":["*partialRates"],	
				"fields": [
					{"tag": "Tenant", "path": "Tenant", "type": "*variable", "value": "~*req.0", "mandatory": true},
					{"tag": "ID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					{"tag": "FilterIDs", "path": "FilterIDs", "type": "*variable", "value": "~*req.2"},
					{"tag": "Weights", "path": "Weights", "type": "*variable", "value": "~*req.3"},
					{"tag": "MinCost", "path": "MinCost", "type": "*variable", "value": "~*req.4"},
					{"tag": "MaxCost", "path": "MaxCost", "type": "*variable", "value": "~*req.5"},
					{"tag": "MaxCostStrategy", "path": "MaxCostStrategy", "type": "*variable", "value": "~*req.6"},
					// {"tag": "RateID", "path": "RateID", "type": "*variable", "value": "~*req.7"},
					{"tag": "RateFilterIDs", "path": "Rates[<~*req.7>].FilterIDs", "type": "*variable", "value": "~*req.8", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateActivationTimes", "path": "Rates[<~*req.7>].ActivationTimes", "type": "*variable", "value": "~*req.9", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateWeights", "path": "Rates[<~*req.7>].Weights", "type": "*variable", "value": "~*req.10", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateBlocker", "path": "Rates[<~*req.7>].Blocker", "type": "*variable", "value": "~*req.11", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateIntervalStart", "path": "Rates[<~*req.7>].IntervalRates.IntervalStart", "type": "*variable", "value": "~*req.12","newBranch":true, "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateFixedFee", "path": "Rates[<~*req.7>].IntervalRates.FixedFee", "type": "*variable", "value": "~*req.13", "filters": ["*notempty:~*req.7:"]},
                    {"tag": "RateRecurrentFee", "path": "Rates[<~*req.7>].IntervalRates.RecurrentFee", "type": "*variable", "value": "~*req.14", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateUnit", "path": "Rates[<~*req.7>].IntervalRates.Unit", "type": "*variable", "value": "~*req.15", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateIncrement", "path": "Rates[<~*req.7>].IntervalRates.Increment", "type": "*variable", "value": "~*req.16", "filters": ["*notempty:~*req.7:"]}
				]
			},
		]
	},
	{
		"id": "*alterRatesLoaderRem",				
		"enabled": true,				
		"tenant": "",					
		"runDelay": "0",				
		"lockfilePath": ".cgr.lck",			
		"conns": {
			"*caches": [{"connIDs": ["*internal"]}]
		},
		"fieldSeparator": ",",			
		"tpInPath": "%s",	
		"tpOutPath": "",	
		"action": "*remove", 				
		"opts": {
			"*withIndex": true
		},
		"data":[					
			{
				"type": "*rateProfiles",	
				"fileName": "Rates.csv",
				"flags":["*partialRates"],	
				"fields": [
					{"tag": "Tenant", "path": "Tenant", "type": "*variable", "value": "~*req.0", "mandatory": true},
					{"tag": "ID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					{"tag": "FilterIDs", "path": "FilterIDs", "type": "*variable", "value": "~*req.2"},
					{"tag": "Weights", "path": "Weights", "type": "*variable", "value": "~*req.3"},
					{"tag": "MinCost", "path": "MinCost", "type": "*variable", "value": "~*req.4"},
					{"tag": "MaxCost", "path": "MaxCost", "type": "*variable", "value": "~*req.5"},
					{"tag": "MaxCostStrategy", "path": "MaxCostStrategy", "type": "*variable", "value": "~*req.6"},
					// {"tag": "RateID", "path": "RateID", "type": "*variable", "value": "~*req.7"},
					{"tag": "RateFilterIDs", "path": "Rates[<~*req.7>].FilterIDs", "type": "*variable", "value": "~*req.8", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateActivationTimes", "path": "Rates[<~*req.7>].ActivationTimes", "type": "*variable", "value": "~*req.9", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateWeights", "path": "Rates[<~*req.7>].Weights", "type": "*variable", "value": "~*req.10", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateBlocker", "path": "Rates[<~*req.7>].Blocker", "type": "*variable", "value": "~*req.11", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateIntervalStart", "path": "Rates[<~*req.7>].IntervalRates.IntervalStart", "type": "*variable", "value": "~*req.12","newBranch":true, "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateFixedFee", "path": "Rates[<~*req.7>].IntervalRates.FixedFee", "type": "*variable", "value": "~*req.13", "filters": ["*notempty:~*req.7:"]},
                    {"tag": "RateRecurrentFee", "path": "Rates[<~*req.7>].IntervalRates.RecurrentFee", "type": "*variable", "value": "~*req.14", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateUnit", "path": "Rates[<~*req.7>].IntervalRates.Unit", "type": "*variable", "value": "~*req.15", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateIncrement", "path": "Rates[<~*req.7>].IntervalRates.Increment", "type": "*variable", "value": "~*req.16", "filters": ["*notempty:~*req.7:"]}
				]
			},
		]
	},
	{
		"id": "*loaderRemoveRateProfile",				
		"enabled": true,				
		"tenant": "",					
		"runDelay": "0",				
		"lockfilePath": ".cgr.lck",			
		"conns": {
			"*caches": [{"connIDs": ["*internal"]}]
		},
		"fieldSeparator": ",",			
		"tpInPath": "%s",	
		"tpOutPath": "",	
		"action": "*remove", 				
		"opts": {
			"*withIndex": true
		},
		"data":[					
			{
				"type": "*rateProfiles",	
				"fileName": "Rates.csv",
				"fields": [
					{"tag": "Tenant", "path": "Tenant", "type": "*variable", "value": "~*req.0", "mandatory": true},
					{"tag": "ID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					{"tag": "FilterIDs", "path": "FilterIDs", "type": "*variable", "value": "~*req.2"},
					{"tag": "Weights", "path": "Weights", "type": "*variable", "value": "~*req.3"},
					{"tag": "MinCost", "path": "MinCost", "type": "*variable", "value": "~*req.4"},
					{"tag": "MaxCost", "path": "MaxCost", "type": "*variable", "value": "~*req.5"},
					{"tag": "MaxCostStrategy", "path": "MaxCostStrategy", "type": "*variable", "value": "~*req.6"},
					// {"tag": "RateID", "path": "RateID", "type": "*variable", "value": "~*req.7"},
					{"tag": "RateFilterIDs", "path": "Rates[<~*req.7>].FilterIDs", "type": "*variable", "value": "~*req.8", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateActivationTimes", "path": "Rates[<~*req.7>].ActivationTimes", "type": "*variable", "value": "~*req.9", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateWeights", "path": "Rates[<~*req.7>].Weights", "type": "*variable", "value": "~*req.10", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateBlocker", "path": "Rates[<~*req.7>].Blocker", "type": "*variable", "value": "~*req.11", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateIntervalStart", "path": "Rates[<~*req.7>].IntervalRates.IntervalStart", "type": "*variable", "value": "~*req.12","newBranch":true, "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateFixedFee", "path": "Rates[<~*req.7>].IntervalRates.FixedFee", "type": "*variable", "value": "~*req.13", "filters": ["*notempty:~*req.7:"]},
                    {"tag": "RateRecurrentFee", "path": "Rates[<~*req.7>].IntervalRates.RecurrentFee", "type": "*variable", "value": "~*req.14", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateUnit", "path": "Rates[<~*req.7>].IntervalRates.Unit", "type": "*variable", "value": "~*req.15", "filters": ["*notempty:~*req.7:"]},
					{"tag": "RateIncrement", "path": "Rates[<~*req.7>].IntervalRates.Increment", "type": "*variable", "value": "~*req.16", "filters": ["*notempty:~*req.7:"]}
				]
			},
		]
	},
],
}	
`, tpPath, t.TempDir(), t.TempDir(), t.TempDir())

	ratesTp := `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s`

	if err := os.WriteFile(filepath.Join(tpPath, "Rates.csv"), []byte(ratesTp), 0644); err != nil {
		t.Fatal(err)
	}
	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	t.Run("LoaderSetRateProfile", func(t *testing.T) {
		var reply map[string]int
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				Path: tpPath,
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		}
		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP1",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 3 {
			t.Errorf("expected len of rates added 3, got %d", len(replyRts))
		}
	})
	t.Run("LoaderOverwriteRateProfile", func(t *testing.T) {
		ratesTp = `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s`
		if err := os.WriteFile(filepath.Join(tpPath, "Rates.csv"), []byte(ratesTp), 0644); err != nil {
			t.Fatal(err)
		}
		var reply map[string]int
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				Path: tpPath,
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP1",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 2 {
			t.Errorf("expected len of rates after overwrite 2, got %d", len(replyRts))
		}
	})
	t.Run("LoaderAlterRateProfile", func(t *testing.T) {
		ratesTp = `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s`
		if err := os.WriteFile(filepath.Join(tpPath, "Rates.csv"), []byte(ratesTp), 0644); err != nil {
			t.Fatal(err)
		}
		var reply map[string]int
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				LoaderID: "*alterRatesLoader",
				Path:     tpPath,
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		}
		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP1",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 3 {
			t.Errorf("expected len of rates after alter 3, got %d", len(replyRts))
		}
	})
	t.Run("LoaderAlterRemoveRates", func(t *testing.T) {
		ratesTp = `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s`
		if err := os.WriteFile(filepath.Join(tpPath, "Rates.csv"), []byte(ratesTp), 0644); err != nil {
			t.Fatal(err)
		}
		var reply map[string]int
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				LoaderID: "*alterRatesLoaderRem",
				Path:     tpPath,
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP1",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 2 {
			t.Errorf("expected len of rates after remove 2, got %d", len(replyRts))
		}
	})
	t.Run("LoaderRemoveRateProfile", func(t *testing.T) {
		ratesTp = `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s`
		var reply map[string]int
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				LoaderID: "*loaderRemoveRateProfile",
				Path:     tpPath,
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		var rpf *utils.RateProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "RP1",
				},
			}, &rpf); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})
	t.Run("LoaderZipSetRateProfile", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		fw, err := zipWriter.Create("Rates.csv")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(`#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RT_VENDOR,,;0,,,,RT_RO,*prefix:~*req.Destination:+40,"* * * * *",;0,false,0s,0,0.02,1s,1s
cgrates.org,RT_VENDOR,,,,,,RT_DE,*prefix:~*req.Destination:+49,"* * * * *",;0,false,0s,0,0.05,1s,1s
cgrates.org,RT_VENDOR,,,,,,RT_UK,*prefix:~*req.Destination:+44,"* * * * *",;0,false,0s,0,0.06,1s,1s`)); err != nil {
			t.Fatal(err)
		}
		if err := zipWriter.Close(); err != nil {
			t.Fatal(err)
		}

		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1ImportZip,
			&loaders.ArgsProcessZip{
				Data: buf.Bytes(),
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RT_VENDOR",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 3 {
			t.Errorf("expected len of rates added 3, got %d", len(replyRts))
		}
	})
	t.Run("LoaderZipOverwriteRateProfile", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		fw, err := zipWriter.Create("Rates.csv")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(`#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RT_VENDOR,,;0,,,,RT_RO,*prefix:~*req.Destination:+40,"* * * * *",;0,false,0s,0,0.02,1s,1s
cgrates.org,RT_VENDOR,,,,,,RT_UK,*prefix:~*req.Destination:+44,"* * * * *",;0,false,0s,0,0.06,1s,1s`)); err != nil {
			t.Fatal(err)
		}
		if err := zipWriter.Close(); err != nil {
			t.Fatal(err)
		}

		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1ImportZip,
			&loaders.ArgsProcessZip{
				Data: buf.Bytes(),
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RT_VENDOR",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 2 {
			t.Errorf("expected len of rates after overwrite 2, got %d", len(replyRts))
		}
	})

	t.Run("LoaderZipAlterRates", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		fw, err := zipWriter.Create("Rates.csv")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(`#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RT_VENDOR,,;0,,,,RT_DE,*prefix:~*req.Destination:+49,"* * * * *",;0,false,0s,0,0.05,1s,1s`)); err != nil {
			t.Fatal(err)
		}
		if err := zipWriter.Close(); err != nil {
			t.Fatal(err)
		}

		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1ImportZip,
			&loaders.ArgsProcessZip{
				LoaderID: "*alterRatesLoader",
				Data:     buf.Bytes(),
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RT_VENDOR",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 3 {
			t.Errorf("expected len of rates after alter 3, got %d", len(replyRts))
		}
	})
	t.Run("LoaderZipAlterRemoveRates", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		fw, err := zipWriter.Create("Rates.csv")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(`#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RT_VENDOR,,;0,,,,RT_RO,*prefix:~*req.Destination:+40,"* * * * *",;0,false,0s,0,0.02,1s,1s`)); err != nil {
			t.Fatal(err)
		}
		if err := zipWriter.Close(); err != nil {
			t.Fatal(err)
		}

		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1ImportZip,
			&loaders.ArgsProcessZip{
				LoaderID: "*alterRatesLoaderRem",
				Data:     buf.Bytes(),
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RT_VENDOR",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 2 {
			t.Errorf("expected len of rates after overwrite 2, got %d", len(replyRts))
		}
	})
	t.Run("LoaderZipRemoveRateProfile", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		fw, err := zipWriter.Create("Rates.csv")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(`#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RT_VENDOR,,;0,,,,RT_UK,*prefix:~*req.Destination:+44,"* * * * *",;0,false,0s,0,0.06,1s,1s`)); err != nil {
			t.Fatal(err)
		}
		if err := zipWriter.Close(); err != nil {
			t.Fatal(err)
		}

		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1ImportZip,
			&loaders.ArgsProcessZip{
				LoaderID: "*loaderRemoveRateProfile",
				Data:     buf.Bytes(),
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var rpf *utils.RateProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "RT_VENDOR",
				},
			}, &rpf); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})
	t.Run("SetRateProfileAPI", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
			&utils.APIRateProfile{
				RateProfile: &utils.RateProfile{
					Tenant:  "cgrates.org",
					ID:      "RP_STANDARD",
					Weights: utils.DynamicWeights{{Weight: 0}},
					Rates: map[string]*utils.Rate{
						"RT_1CSEC": {
							ID:              "RT_1CSEC",
							ActivationTimes: "* * * * *",
							Weights:         utils.DynamicWeights{{Weight: 0}},
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(1, 2),
									Unit:          utils.NewDecimal(int64(time.Second), 0),
									Increment:     utils.NewDecimal(int64(time.Minute), 0),
								},
							},
						},
						"RT_10C": {
							ID:              "RT_10C",
							FilterIDs:       []string{"*prefix:~*req.Destination:1002"},
							ActivationTimes: "* * * * *",
							Weights:         utils.DynamicWeights{{Weight: 10}},
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(1, 1),
									Unit:          utils.NewDecimal(int64(time.Minute), 0),
									Increment:     utils.NewDecimal(int64(time.Second), 0),
								},
							},
						},
						"RT_5C": {
							ID:              "RT_5C",
							FilterIDs:       []string{"*prefix:~*req.Destination:1003"},
							ActivationTimes: "* * * * *",
							Weights:         utils.DynamicWeights{{Weight: 10}},
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(5, 2),
									Unit:          utils.NewDecimal(int64(time.Minute), 0),
									Increment:     utils.NewDecimal(int64(time.Second), 0),
								},
							},
						},
					},
				},
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP_STANDARD",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 3 {
			t.Errorf("expected len of rates added 3, got %d", len(replyRts))
		}
	})
	t.Run("OverwriteRatesAPI", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
			&utils.APIRateProfile{
				RateProfile: &utils.RateProfile{
					Tenant:  "cgrates.org",
					ID:      "RP_STANDARD",
					Weights: utils.DynamicWeights{{Weight: 0}},
					Rates: map[string]*utils.Rate{
						"RT_1CSEC": {
							ID:              "RT_1CSEC",
							ActivationTimes: "* * * * *",
							Weights:         utils.DynamicWeights{{Weight: 0}},
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(1, 2),
									Unit:          utils.NewDecimal(int64(time.Second), 0),
									Increment:     utils.NewDecimal(int64(time.Minute), 0),
								},
								{
									IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
									RecurrentFee:  utils.NewDecimal(1, 2),
									Unit:          utils.NewDecimal(int64(time.Second), 0),
									Increment:     utils.NewDecimal(int64(time.Second), 0),
								},
							},
						},
						"RT_10C": {
							ID:              "RT_10C",
							FilterIDs:       []string{"*prefix:~*req.Destination:1002"},
							ActivationTimes: "* * * * *",
							Weights:         utils.DynamicWeights{{Weight: 10}},
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(1, 1),
									Unit:          utils.NewDecimal(int64(time.Minute), 0),
									Increment:     utils.NewDecimal(int64(time.Second), 0),
								},
							},
						},
					},
				},
				APIOpts: map[string]any{
					utils.MetaCache:          utils.MetaNone,
					utils.MetaRateSOverwrite: true,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP_STANDARD",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 2 {
			t.Errorf("expected len of rates after overwrite 2, got %d", len(replyRts))
		}
	})
	t.Run("AlterRatesAPI", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
			&utils.APIRateProfile{
				RateProfile: &utils.RateProfile{
					Tenant:  "cgrates.org",
					ID:      "RP_STANDARD",
					Weights: utils.DynamicWeights{{Weight: 0}},
					Rates: map[string]*utils.Rate{
						"RT_5C": {
							ID:              "RT_5C",
							FilterIDs:       []string{"*prefix:~*req.Destination:1003"},
							ActivationTimes: "* * * * *",
							Weights:         utils.DynamicWeights{{Weight: 10}},
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(5, 2),
									Unit:          utils.NewDecimal(int64(time.Minute), 0),
									Increment:     utils.NewDecimal(int64(time.Second), 0),
								},
							},
						},
					},
				},
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP_STANDARD",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 3 {
			t.Errorf("expected len of rates after alter 3, got %d", len(replyRts))
		}
	})
	t.Run("RemoveRatesAPI", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1RemoveRateProfileRates,
			&utils.RemoveRPrfRates{
				Tenant:  "cgrates.org",
				ID:      "RP_STANDARD",
				RateIDs: []string{"RT_10C"},
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var replyRts []*utils.Rate
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfileRates,
			&utils.ArgsSubItemIDs{
				Tenant:      "cgrates.org",
				ProfileID:   "RP_STANDARD",
				ItemsPrefix: "",
			}, &replyRts); err != nil {
			t.Error(err)
		}
		if len(replyRts) != 2 {
			t.Errorf("expected len of rates after remove 2, got %d", len(replyRts))
		}
	})
	t.Run("RemoveRateProfileAPI", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "RP_STANDARD",
				},
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected %s, got %s", utils.OK, reply)
		}

		var rpf *utils.RateProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetRateProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "RP_STANDARD",
				},
			}, &rpf); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})
}
