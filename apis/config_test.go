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

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
)

func TestConfigNewConfigSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &ConfigSv1{
		cfg: cfg,
	}
	result := NewConfigSv1(cfg)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestConfigSetGetConfig(t *testing.T) {
	cfgJSONStr := `{
"attributes": {								
	"enabled": true,	
	"stats_conns": ["*internal"],			
	"resources_conns": ["*internal"],		
	"accounts_conns": ["*internal"],			
	"prefix_indexed_fields": ["index1","index2"],		
	"opts": {
		"*processRuns": [
				{
					"Value": 3,
				},
			],
		},					
	},		
}`
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	rlcCfg := NewConfigSv1(cfg)
	args := &config.SetConfigArgs{}
	var reply string
	err = rlcCfg.SetConfig(context.Background(), args, &reply)
	expected := `OK`
	if err != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expected, reply)
	}
	argsGet := &config.SectionWithAPIOpts{
		Sections: []string{"attributes"},
	}
	var replyGet map[string]any
	err = rlcCfg.GetConfig(context.Background(), argsGet, &replyGet)
	expectedGet := map[string]any{
		"attributes": map[string]any{
			"accounts_conns":           []string{"*internal"},
			"enabled":                  true,
			"indexed_selects":          true,
			"nested_fields":            false,
			"prefix_indexed_fields":    []string{"index1", "index2"},
			"resources_conns":          []string{"*internal"},
			"stats_conns":              []string{"*internal"},
			"suffix_indexed_fields":    []string{},
			"exists_indexed_fields":    []string{},
			"notexists_indexed_fields": []string{},
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs: []*config.DynamicStringSliceOpt{},
				utils.MetaProcessRunsCfg: []*config.DynamicIntOpt{
					{
						Value: 3,
					},
				},
				utils.MetaProfileRunsCfg:       []*config.DynamicIntOpt{},
				utils.MetaProfileIgnoreFilters: []*config.DynamicBoolOpt{},
			},
		},
	}
	if err != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expectedGet, replyGet) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(replyGet))
	}
}

func TestConfigSetGetReloadConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rlcCfg := NewConfigSv1(cfg)
	args := &config.SetConfigArgs{
		Tenant: utils.CGRateSorg,
		Config: map[string]any{
			"attributes": map[string]any{
				"accounts_conns":           []string{"*internal"},
				"enabled":                  true,
				"indexed_selects":          false,
				"nested_fields":            false,
				"prefix_indexed_fields":    []string{},
				"resources_conns":          []string{"*internal"},
				"stats_conns":              []string{"*internal"},
				"suffix_indexed_fields":    []string{},
				"exists_indexed_fields":    []string{},
				"notexists_indexed_fields": []string{},
				utils.OptsCfg: map[string]any{
					utils.MetaProcessRunsCfg: []*config.DynamicIntOpt{
						{
							Value: 2,
						},
					},
				},
			},
		},
		DryRun: true,
	}
	var reply string
	if err := rlcCfg.SetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Unexpected reply: <%s>", reply)
	}
	argsGet := &config.SectionWithAPIOpts{
		Sections: []string{"attributes"},
	}
	expectedGet := map[string]any{
		"attributes": map[string]any{
			"accounts_conns":           []string{},
			"enabled":                  false,
			"indexed_selects":          true,
			"nested_fields":            false,
			"prefix_indexed_fields":    []string{},
			"resources_conns":          []string{},
			"stats_conns":              []string{},
			"suffix_indexed_fields":    []string{},
			"exists_indexed_fields":    []string{},
			"notexists_indexed_fields": []string{},
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*config.DynamicStringSliceOpt{},
				utils.MetaProcessRunsCfg:       []*config.DynamicIntOpt{},
				utils.MetaProfileRunsCfg:       []*config.DynamicIntOpt{},
				utils.MetaProfileIgnoreFilters: []*config.DynamicBoolOpt{},
			},
		},
	}
	var replyGet map[string]any
	if err := rlcCfg.GetConfig(context.Background(), argsGet, &replyGet); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedGet, replyGet) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>",
			utils.ToJSON(expectedGet), utils.ToJSON(replyGet))
	}
	argsRld := &config.ReloadArgs{
		DryRun:  true,
		Section: "attributes",
	}
	experr := `path:"" is not reachable`
	var replyRld string
	if err := rlcCfg.ReloadConfig(context.Background(), argsRld, &replyRld); err == nil ||
		err.Error() != experr { // path is required for ReloadConfig api, but it is not provided in this unit test
		t.Errorf("expected: <%s>, \nreceived: <%s>", experr, err.Error())
	}
	argsGetRld := &config.SectionWithAPIOpts{
		Sections: []string{"attributes"},
	}
	expectedGetRld := map[string]any{
		"attributes": map[string]any{
			"accounts_conns":           []string{},
			"enabled":                  false,
			"indexed_selects":          true,
			"nested_fields":            false,
			"prefix_indexed_fields":    []string{},
			"resources_conns":          []string{},
			"stats_conns":              []string{},
			"suffix_indexed_fields":    []string{},
			"exists_indexed_fields":    []string{},
			"notexists_indexed_fields": []string{},
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*config.DynamicStringSliceOpt{},
				utils.MetaProcessRunsCfg:       []*config.DynamicIntOpt{},
				utils.MetaProfileRunsCfg:       []*config.DynamicIntOpt{},
				utils.MetaProfileIgnoreFilters: []*config.DynamicBoolOpt{},
			},
		},
	}
	var replyGetRld map[string]any
	if err := rlcCfg.GetConfig(context.Background(), argsGetRld, &replyGetRld); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedGetRld, replyGetRld) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGetRld), utils.ToJSON(replyGetRld))
	}
}

func TestConfigGetSetConfigFromJSONErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rlcCfg := NewConfigSv1(cfg)
	args := &config.SetConfigFromJSONArgs{
		APIOpts: nil,
		Tenant:  utils.CGRateSorg,
		Config: `{
"attributes":{
	"accounts_conns":["*localhost"],
	"enabled":true,
	"indexed_selects":true,
	"nested_fields":false,
	"opts":{
		"*profileIDs": [],
		"*processRuns": [
			{
				"Value": 2,
			},
		],
		"*profileRuns": [
			{
				"Value": 0,
			},
		],
	},
	"prefix_indexed_fields":[],
	"resources_conns":["*localhost"],
	"stats_conns":["*localhost"],
	"suffix_indexed_fields":[],
	"exists_indexed_fields":[],
	"notexists_indexed_fields":[],
	},
}`,
		DryRun: true,
	}

	var reply string
	if err := rlcCfg.SetConfigFromJSON(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Unexpected reply <%s>", reply)
	}

	argsGet := &config.SectionWithAPIOpts{
		Tenant:   utils.CGRateSorg,
		Sections: []string{"attributes"},
	}
	expectedGet := `{"attributes":{"accounts_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*processRuns":[],"*profileIDs":[],"*profileIgnoreFilters":[],"*profileRuns":[]},"prefix_indexed_fields":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]}}`
	var replyGet string
	if err := rlcCfg.GetConfigAsJSON(context.Background(), argsGet, &replyGet); err != nil {
		t.Error(err)
	} else if replyGet != expectedGet {
		t.Errorf("Expected <%s>, \nReceived <%s>", expectedGet, replyGet)
	}
}

func TestConfigStoreCfgInDBErr(t *testing.T) {
	//for coverage purposes only
	cfg := config.NewDefaultCGRConfig()
	rlcCfg := NewConfigSv1(cfg)
	args := &config.SectionWithAPIOpts{}
	var reply string
	err := rlcCfg.StoreCfgInDB(context.Background(), args, &reply)
	expected := "no DB connection for config"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expected, err)
	}
}
