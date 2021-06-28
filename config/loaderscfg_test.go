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
package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"tenant": "cgrates.org",
		"lock_filename": ".cgr.lck",
		"caches_conns": ["*internal","*conn1"],
		"field_separator": ",",
		"tp_in_dir": "/var/spool/cgrates/loader/in",
		"tp_out_dir": "/var/spool/cgrates/loader/out",
		"data":[
			{
				"type": "*attributes",
				"file_name": "Attributes.csv",
                "flags": [],
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*composed", "value": "~req.0", "mandatory": true,"layout": "2006-01-02T15:04:05Z07:00"},
					],
				},
			],
		},
	],
}`
	val, err := NewRSRParsers("~req.0", utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ten, err := NewRSRParsers("cgrates.org", utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	expected := LoaderSCfgs{
		{
			Enabled:        true,
			ID:             utils.MetaDefault,
			Tenant:         ten,
			LockFileName:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Data: []*LoaderDataType{
				{
					Type:     "*attributes",
					Filename: "Attributes.csv",
					Flags:    utils.FlagsWithParams{},
					Fields: []*FCTemplate{
						{
							Tag:       "TenantID",
							Path:      "Tenant",
							pathSlice: []string{"Tenant"},
							Type:      utils.MetaComposed,
							Value:     val,
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
				},
			},
		},
	}
	newCfg := new(CGRConfig)
	newCfg.generalCfg = new(GeneralCfg)
	newCfg.generalCfg.RSRSep = ";"
	if jsonCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = newCfg.loadLoaderSCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, newCfg.loaderCfg) {
		t.Errorf("Expected %+v,\n received %+v", utils.ToJSON(expected), utils.ToJSON(newCfg.loaderCfg))
	}
}

func TestLoaderSCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &LoaderJsonCfg{
		Tenant: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase3(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase4(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Type: utils.StringPointer(utils.MetaTemplate),
					},
				},
			},
		},
	}
	expected := "no template with id: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase5(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer("randomTag"),
						Path:  utils.StringPointer("randomPath"),
						Type:  utils.StringPointer(utils.MetaTemplate),
						Value: utils.StringPointer("randomTemplate"),
					},
				},
			},
		},
	}
	expectedFields := LoaderSCfgs{
		{
			Data: []*LoaderDataType{
				{
					Fields: []*FCTemplate{
						{
							Tag:       "TenantID",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
						},
					},
				},
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:       "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaVariable,
				Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, msgTemplates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.loaderCfg[0].Data[0].Fields[0], expectedFields[0].Data[0].Fields[0]) {
		t.Errorf("Expected %+v,\n received %+v", utils.ToJSON(expectedFields[0].Data[0].Fields[0]), utils.ToJSON(jsonCfg.loaderCfg[0].Data[0].Fields[0]))
	}
}

func TestLoaderSCfgloadFromJsonCfgCase6(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{nil},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestEnabledCase1(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	if enabled := jsonCfg.loaderCfg.Enabled(); enabled {
		t.Errorf("Expected %+v", enabled)
	}
}
func TestEnabledCase2(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"enabled": true,
		},
	],	
}`
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if enabled := jsonCfg.loaderCfg.Enabled(); !enabled {
		t.Errorf("Expected %+v", enabled)
	}
}

func TestLoaderCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"enabled": true,
		"run_delay": "1sa",										
	},
	],	
}`
	expected := "time: unknown unit \"sa\" in duration \"1sa\""
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expected {
		t.Errorf("Expected error: %s ,received: %v", expected, err)
	}
}

func TestLoaderCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "cgrates.org",										
		"dry_run": false,									
		"run_delay": "0",										
		"lock_filename": ".cgr.lck",						
		"caches_conns": ["*internal:*caches"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~*req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]interface{}{
		{
			utils.IDCfg:           "*default",
			utils.EnabledCfg:      false,
			utils.TenantCfg:       "cgrates.org",
			utils.DryRunCfg:       false,
			utils.RunDelayCfg:     "0",
			utils.LockFileNameCfg: ".cgr.lck",
			utils.CachesConnsCfg:  []string{utils.MetaInternal},
			utils.FieldSepCfg:     ",",
			utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
			utils.DataCfg: []map[string]interface{}{
				{
					utils.TypeCfg:     "*filters",
					utils.FilenameCfg: "Filters.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "Type",
							utils.PathCfg:  "Type",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Element",
							utils.PathCfg:  "Element",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Values",
							utils.PathCfg:  "Values",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
					},
				},
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
				{
					utils.TypeCfg:     "*resources",
					utils.FilenameCfg: "Resources.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "TTL",
							utils.PathCfg:  "UsageTTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Limit",
							utils.PathCfg:  "Limit",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "AllocationMessage",
							utils.PathCfg:  "AllocationMessage",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "Blocker",
							utils.PathCfg:  "Blocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
					},
				},
				{
					utils.TypeCfg:     "*stats",
					utils.FilenameCfg: "Stats.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "QueueLength",
							utils.PathCfg:  "QueueLength",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "TTL",
							utils.PathCfg:  "TTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "MinItems",
							utils.PathCfg:  "MinItems",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "MetricIDs",
							utils.PathCfg:  "MetricIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "MetricFilterIDs",
							utils.PathCfg:  "MetricFilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "Blocker",
							utils.PathCfg:  "Blocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
					},
				},
				{
					utils.TypeCfg:     "*thresholds",
					utils.FilenameCfg: "Thresholds.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "MaxHits",
							utils.PathCfg:  "MaxHits",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "MinHits",
							utils.PathCfg:  "MinHits",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "MinSleep",
							utils.PathCfg:  "MinSleep",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "Blocker",
							utils.PathCfg:  "Blocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "ActionProfileIDs",
							utils.PathCfg:  "ActionProfileIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "Async",
							utils.PathCfg:  "Async",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
					},
				},
				{
					utils.TypeCfg:     "*routes",
					utils.FilenameCfg: "Routes.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Sorting",
							utils.PathCfg:  "Sorting",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "SortingParameters",
							utils.PathCfg:  "SortingParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "RouteID",
							utils.PathCfg:  "RouteID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "RouteFilterIDs",
							utils.PathCfg:  "RouteFilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "RouteAccountIDs",
							utils.PathCfg:  "RouteAccountIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "RouteRatingPlanIDs",
							utils.PathCfg:  "RouteRatingPlanIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "RouteResourceIDs",
							utils.PathCfg:  "RouteResourceIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
						{
							utils.TagCfg:   "RouteStatIDs",
							utils.PathCfg:  "RouteStatIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
						{
							utils.TagCfg:   "RouteWeight",
							utils.PathCfg:  "RouteWeight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.12",
						},
						{
							utils.TagCfg:   "RouteBlocker",
							utils.PathCfg:  "RouteBlocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.13",
						},
						{
							utils.TagCfg:   "RouteParameters",
							utils.PathCfg:  "RouteParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.14",
						},
					},
				},
				{
					utils.TypeCfg:     "*chargers",
					utils.FilenameCfg: "Chargers.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "RunID",
							utils.PathCfg:  "RunID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "AttributeIDs",
							utils.PathCfg:  "AttributeIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
					},
				},
				{
					utils.TypeCfg:     "*dispatchers",
					utils.FilenameCfg: "DispatcherProfiles.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Strategy",
							utils.PathCfg:  "Strategy",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "StrategyParameters",
							utils.PathCfg:  "StrategyParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "ConnID",
							utils.PathCfg:  "ConnID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "ConnFilterIDs",
							utils.PathCfg:  "ConnFilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "ConnWeight",
							utils.PathCfg:  "ConnWeight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "ConnBlocker",
							utils.PathCfg:  "ConnBlocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "ConnParameters",
							utils.PathCfg:  "ConnParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
					},
				},
				{
					utils.TypeCfg:     "*dispatcher_hosts",
					utils.FilenameCfg: "DispatcherHosts.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "Address",
							utils.PathCfg:  "Address",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Transport",
							utils.PathCfg:  "Transport",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "TLS",
							utils.PathCfg:  "TLS",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
					},
				},
				{
					utils.TypeCfg:     "*rate_profiles",
					utils.FilenameCfg: "RateProfiles.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "MinCost",
							utils.PathCfg:  "MinCost",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "MaxCost",
							utils.PathCfg:  "MaxCost",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "MaxCostStrategy",
							utils.PathCfg:  "MaxCostStrategy",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "RateID",
							utils.PathCfg:  "RateID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "RateFilterIDs",
							utils.PathCfg:  "RateFilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "RateActivationTimes",
							utils.PathCfg:  "RateActivationTimes",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "RateWeight",
							utils.PathCfg:  "RateWeight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
						{
							utils.TagCfg:   "RateBlocker",
							utils.PathCfg:  "RateBlocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
						{
							utils.TagCfg:   "RateIntervalStart",
							utils.PathCfg:  "RateIntervalStart",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.12",
						},
						{
							utils.TagCfg:   "RateFixedFee",
							utils.PathCfg:  "RateFixedFee",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.13",
						},
						{
							utils.TagCfg:   "RateRecurrentFee",
							utils.PathCfg:  "RateRecurrentFee",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.14",
						},
						{
							utils.TagCfg:   "RateUnit",
							utils.PathCfg:  "RateUnit",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.15",
						},
						{
							utils.TagCfg:   "RateIncrement",
							utils.PathCfg:  "RateIncrement",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.16",
						},
					},
				},
				{
					utils.TypeCfg:     "*action_profiles",
					utils.FilenameCfg: "ActionProfiles.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Schedule",
							utils.PathCfg:  "Schedule",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "TargetType",
							utils.PathCfg:  "TargetType",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "TargetIDs",
							utils.PathCfg:  "TargetIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "ActionID",
							utils.PathCfg:  "ActionID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "ActionFilterIDs",
							utils.PathCfg:  "ActionFilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "ActionBlocker",
							utils.PathCfg:  "ActionBlocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "ActionTTL",
							utils.PathCfg:  "ActionTTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
						{
							utils.TagCfg:   "ActionType",
							utils.PathCfg:  "ActionType",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
						{
							utils.TagCfg:   "ActionOpts",
							utils.PathCfg:  "ActionOpts",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.12",
						},
						{
							utils.TagCfg:   "ActionPath",
							utils.PathCfg:  "ActionPath",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.13",
						},
						{
							utils.TagCfg:   "ActionValue",
							utils.PathCfg:  "ActionValue",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.14",
						},
					},
				},
				{
					utils.TypeCfg:     "*accounts",
					utils.FilenameCfg: "Accounts.csv",
					utils.FlagsCfg:    nil,
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weight",
							utils.PathCfg:  "Weight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "BalanceID",
							utils.PathCfg:  "BalanceID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "BalanceFilterIDs",
							utils.PathCfg:  "BalanceFilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "BalanceWeight",
							utils.PathCfg:  "BalanceWeight",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "BalanceBlocker",
							utils.PathCfg:  "BalanceBlocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "BalanceType",
							utils.PathCfg:  "BalanceType",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "BalanceOpts",
							utils.PathCfg:  "BalanceOpts",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "BalanceCostIncrements",
							utils.PathCfg:  "BalanceCostIncrements",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
						{
							utils.TagCfg:   "BalanceAttributeIDs",
							utils.PathCfg:  "BalanceAttributeIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
						{
							utils.TagCfg:   "BalanceRateProfileIDs",
							utils.PathCfg:  "BalanceRateProfileIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.12",
						},
						{
							utils.TagCfg:   "BalanceUnitFactors",
							utils.PathCfg:  "BalanceUnitFactors",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.13",
						},
						{
							utils.TagCfg:   "BalanceUnits",
							utils.PathCfg:  "BalanceUnits",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.14",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.15",
						},
					},
				},
			},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cfgCgr.loaderCfg.AsMapInterface(cfgCgr.generalCfg.RSRSep)
		if len(cfgCgr.loaderCfg) != 1 {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", 1, len(cfgCgr.loaderCfg))
		} else if !reflect.DeepEqual(rcv, eMap) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
	}
}

func TestLoaderCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "~*req.Destination1",										
		"dry_run": false,									
		"run_delay": "1",										
		"lock_filename": ".cgr.lck",						
		"caches_conns": ["*conn1"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]interface{}{
		{
			utils.IDCfg:           "*default",
			utils.EnabledCfg:      false,
			utils.TenantCfg:       "~*req.Destination1",
			utils.DryRunCfg:       false,
			utils.RunDelayCfg:     "0",
			utils.LockFileNameCfg: ".cgr.lck",
			utils.CachesConnsCfg:  []string{"*conn1"},
			utils.FieldSepCfg:     ",",
			utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
			utils.DataCfg: []map[string]interface{}{
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := jsonCfg.loaderCfg.AsMapInterface(jsonCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv[0][utils.Tenant], eMap[0][utils.Tenant]) {
		t.Errorf("Expected %+v, received %+v", rcv[0][utils.Tenant], eMap[0][utils.Tenant])
	}
}

func TestLoaderSCfgsClone(t *testing.T) {
	ban := LoaderSCfgs{{
		Enabled:        true,
		ID:             utils.MetaDefault,
		Tenant:         NewRSRParsersMustCompile("cgrate.org", utils.InfieldSep),
		LockFileName:   ".cgr.lck",
		CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		FieldSeparator: ",",
		TpInDir:        "/var/spool/cgrates/loader/in",
		TpOutDir:       "/var/spool/cgrates/loader/out",
		Data: []*LoaderDataType{{
			Type:     "*attributes",
			Filename: "Attributes.csv",
			Flags:    utils.FlagsWithParams{},
			Fields: []*FCTemplate{
				{
					Tag:       "TenantID",
					Path:      "Tenant",
					pathSlice: []string{"Tenant"},
					Type:      utils.MetaComposed,
					Value:     NewRSRParsersMustCompile("cgrate.org", utils.InfieldSep),
					Mandatory: true,
					Layout:    time.RFC3339,
				},
			}},
		},
	}}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv[0].CacheSConns[1] = ""; ban[0].CacheSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv[0].Data[0].Type = ""; ban[0].Data[0].Type != "*attributes" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestEqualsLoaderDatasType(t *testing.T) {
	v1 := []*LoaderDataType{
		{
			Type:     "*json",
			Filename: "file.json",
			Flags: utils.FlagsWithParams{
				"FLAG_1": {
					"PARAM_1": []string{"param1"},
				},
			},
			Fields: []*FCTemplate{
				{
					Type: "Type",
					Tag:  "Tag",
				},
			},
		},
	}

	v2 := []*LoaderDataType{
		{
			Type:     "*xml",
			Filename: "file.xml",
			Flags: utils.FlagsWithParams{
				"FLAG_2": {
					"PARAM_2": []string{"param2"},
				},
			},
			Fields: []*FCTemplate{
				{
					Type: "Type2",
					Tag:  "Tag2",
				},
			},
		},
	}

	if equalsLoaderDatasType(v1, v2) {
		t.Error("Loaders should not match")
	}

	v1 = v2
	if !equalsLoaderDatasType(v1, v2) {
		t.Error("Loaders should match")
	}

	v2 = []*LoaderDataType{}
	if equalsLoaderDatasType(v1, v2) {
		t.Error("Loaders should not match")
	}
}

func TestDiffLoaderJsonCfg(t *testing.T) {

	v1 := &LoaderSCfg{
		ID:      "LoaderID",
		Enabled: true,
		Tenant: RSRParsers{
			{
				Rules: "cgrates.org",
			},
		},
		DryRun:         false,
		RunDelay:       1 * time.Millisecond,
		LockFileName:   "lockFileName",
		CacheSConns:    []string{"*localhost"},
		FieldSeparator: ";",
		TpInDir:        "/tp/in/dir",
		TpOutDir:       "/tp/out/dir",
		Data:           nil,
	}

	v2 := &LoaderSCfg{
		ID:      "LoaderID2",
		Enabled: false,
		Tenant: RSRParsers{
			{
				Rules: "itsyscom.com",
			},
		},
		DryRun:         true,
		RunDelay:       2 * time.Millisecond,
		LockFileName:   "lockFileName2",
		CacheSConns:    []string{"*birpc"},
		FieldSeparator: ":",
		TpInDir:        "/tp/in/dir/2",
		TpOutDir:       "/tp/out/dir/2",
		Data: []*LoaderDataType{
			{
				Type:     "*xml",
				Filename: "file.xml",
				Flags: utils.FlagsWithParams{
					"FLAG_2": {
						"PARAM_2": []string{"param2"},
					},
				},
				Fields: []*FCTemplate{
					{
						Type: "Type2",
						Tag:  "Tag2",
					},
				},
			},
		},
	}

	expected := &LoaderJsonCfg{
		ID:              utils.StringPointer("LoaderID2"),
		Enabled:         utils.BoolPointer(false),
		Tenant:          utils.StringPointer("itsyscom.com"),
		Dry_run:         utils.BoolPointer(true),
		Run_delay:       utils.StringPointer("2ms"),
		Lock_filename:   utils.StringPointer("lockFileName2"),
		Caches_conns:    &[]string{"*birpc"},
		Field_separator: utils.StringPointer(":"),
		Tp_in_dir:       utils.StringPointer("/tp/in/dir/2"),
		Tp_out_dir:      utils.StringPointer("/tp/out/dir/2"),
		Data: &[]*LoaderJsonDataType{
			{
				Type:      utils.StringPointer("*xml"),
				File_name: utils.StringPointer("file.xml"),
				Flags:     &[]string{"FLAG_2:PARAM_2:param2"},
				Fields: &[]*FcTemplateJsonCfg{
					{
						Type:   utils.StringPointer("Type2"),
						Tag:    utils.StringPointer("Tag2"),
						Layout: utils.StringPointer(""),
					},
				},
			},
		},
	}

	rcv := diffLoaderJsonCfg(v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &LoaderJsonCfg{}
	rcv = diffLoaderJsonCfg(v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}

func TestEqualsLoadersJsonCfg(t *testing.T) {
	v1 := LoaderSCfgs{
		{
			ID:      "LoaderID",
			Enabled: true,
			Tenant: RSRParsers{
				{
					Rules: "cgrates.org",
				},
			},
			DryRun:         false,
			RunDelay:       1 * time.Millisecond,
			LockFileName:   "lockFileName",
			CacheSConns:    []string{"*localhost"},
			FieldSeparator: ";",
			TpInDir:        "/tp/in/dir",
			TpOutDir:       "/tp/out/dir",
			Data:           nil,
		},
	}

	v2 := LoaderSCfgs{
		{
			ID:      "LoaderID2",
			Enabled: false,
			Tenant: RSRParsers{
				{
					Rules: "itsyscom.com",
				},
			},
			DryRun:         true,
			RunDelay:       2 * time.Millisecond,
			LockFileName:   "lockFileName2",
			CacheSConns:    []string{"*birpc"},
			FieldSeparator: ":",
			TpInDir:        "/tp/in/dir/2",
			TpOutDir:       "/tp/out/dir/2",
			Data: []*LoaderDataType{
				{
					Type:     "*xml",
					Filename: "file.xml",
					Flags: utils.FlagsWithParams{
						"FLAG_2": {
							"PARAM_2": []string{"param2"},
						},
					},
					Fields: []*FCTemplate{
						{
							Type: "Type2",
							Tag:  "Tag2",
						},
					},
				},
			},
		},
	}

	if equalsLoadersJsonCfg(v1, v2) {
		t.Error("Loaders shouldn't match")
	}

	v2 = v1
	if !equalsLoadersJsonCfg(v1, v2) {
		t.Error("Loaders shouldn't match")
	}

	v2 = LoaderSCfgs{}
	if equalsLoadersJsonCfg(v1, v2) {
		t.Error("Loaders shouldn't match")
	}
}

func TestDiffLoadersJsonCfg(t *testing.T) {
	var d []*LoaderJsonCfg

	v1 := LoaderSCfgs{
		{
			ID:      "LoaderID",
			Enabled: false,
			Tenant: RSRParsers{
				{
					Rules: "cgrates.org",
				},
			},
			DryRun:         false,
			RunDelay:       1 * time.Millisecond,
			LockFileName:   "lockFileName",
			CacheSConns:    []string{"*localhost"},
			FieldSeparator: ";",
			TpInDir:        "/tp/in/dir",
			TpOutDir:       "/tp/out/dir",
			Data:           nil,
		},
	}

	v2 := LoaderSCfgs{
		{
			ID:      "LoaderID2",
			Enabled: true,
			Tenant: RSRParsers{
				{
					Rules: "itsyscom.com",
				},
			},
			DryRun:         true,
			RunDelay:       2 * time.Millisecond,
			LockFileName:   "lockFileName2",
			CacheSConns:    []string{"*birpc"},
			FieldSeparator: ":",
			TpInDir:        "/tp/in/dir/2",
			TpOutDir:       "/tp/out/dir/2",
			Data: []*LoaderDataType{
				{
					Type:     "*xml",
					Filename: "file.xml",
					Flags: utils.FlagsWithParams{
						"FLAG_2": {
							"PARAM_2": []string{"param2"},
						},
					},
					Fields: []*FCTemplate{
						{
							Type: "Type2",
							Tag:  "Tag2",
						},
					},
				},
			},
		},
	}

	expected := []*LoaderJsonCfg{
		{
			ID:              utils.StringPointer("LoaderID2"),
			Enabled:         utils.BoolPointer(true),
			Tenant:          utils.StringPointer("itsyscom.com"),
			Dry_run:         utils.BoolPointer(true),
			Run_delay:       utils.StringPointer("2ms"),
			Lock_filename:   utils.StringPointer("lockFileName2"),
			Caches_conns:    &[]string{"*birpc"},
			Field_separator: utils.StringPointer(":"),
			Tp_in_dir:       utils.StringPointer("/tp/in/dir/2"),
			Tp_out_dir:      utils.StringPointer("/tp/out/dir/2"),
			Data: &[]*LoaderJsonDataType{
				{
					Type:      utils.StringPointer("*xml"),
					File_name: utils.StringPointer("file.xml"),
					Flags:     &[]string{"FLAG_2:PARAM_2:param2"},
					Fields: &[]*FcTemplateJsonCfg{
						{
							Type:   utils.StringPointer("Type2"),
							Tag:    utils.StringPointer("Tag2"),
							Layout: utils.StringPointer(""),
						},
					},
				},
			},
		},
	}

	rcv := diffLoadersJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = nil
	rcv = diffLoadersJsonCfg(d, v1, v2, ";")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
