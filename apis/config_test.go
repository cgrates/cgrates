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
	"path"
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
	//for coverage purposes only
	cfgTestPath := path.Join(*dataDir, "conf", "samples", "tutinternal")
	cfg, err := config.NewCGRConfigFromPath(cfgTestPath)
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
	var replyGet map[string]interface{}
	err = rlcCfg.GetConfig(context.Background(), argsGet, &replyGet)
	expectedGet := map[string]interface{}{
		"attributes": map[string]interface{}{
			"accounts_conns":        []string{"*localhost"},
			"enabled":               true,
			"indexed_selects":       true,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*localhost"},
			"stats_conns":           []string{"*localhost"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.OptsAttributesProcessRuns: float64(1),
			},
			utils.ProfileRunsCfg: 0,
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
	//for coverage purposes only
	var err error
	cfgTestPath := path.Join(*dataDir, "conf", "samples", "tutinternal")
	cfg, err := config.NewCGRConfigFromPath(cfgTestPath)
	if err != nil {
		t.Error(err)
	}
	rlcCfg := NewConfigSv1(cfg)
	args := &config.SetConfigArgs{
		APIOpts: nil,
		Tenant:  utils.CGRateSorg,
		Config: map[string]interface{}{
			"attributes": map[string]interface{}{
				"accounts_conns":        []string{"*internal"},
				"enabled":               true,
				"indexed_selects":       false,
				"nested_fields":         false,
				"prefix_indexed_fields": []string{},
				"resources_conns":       []string{"*internal"},
				"stats_conns":           []string{"*internal"},
				"suffix_indexed_fields": []string{},
				utils.OptsCfg: map[string]interface{}{
					utils.OptsAttributesProcessRuns: 2,
				},
			},
		},
		DryRun: true,
	}
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
	var replyGet map[string]interface{}
	errGet := rlcCfg.GetConfig(context.Background(), argsGet, &replyGet)
	expectedGet := map[string]interface{}{
		"attributes": map[string]interface{}{
			"accounts_conns":        []string{"*localhost"},
			"enabled":               true,
			"indexed_selects":       true,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*localhost"},
			"stats_conns":           []string{"*localhost"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.OptsAttributesProcessRuns: float64(1),
			},
			utils.ProfileRunsCfg: 0,
		},
	}
	if errGet != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, errGet)
	}
	if !reflect.DeepEqual(expectedGet, replyGet) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expectedGet, replyGet)
	}
	argsRld := &config.ReloadArgs{
		DryRun:  true,
		Section: "attributes",
	}

	var replyRld string
	errRld := rlcCfg.ReloadConfig(context.Background(), argsRld, &replyRld)
	expectedRld := `OK`
	if err != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, errRld)
	}
	if !reflect.DeepEqual(expectedRld, replyRld) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expectedRld, replyRld)
	}
	argsGetRld := &config.SectionWithAPIOpts{
		Sections: []string{"attributes"},
	}
	var replyGetRld map[string]interface{}
	errGetRld := rlcCfg.GetConfig(context.Background(), argsGetRld, &replyGetRld)
	expectedGetRld := map[string]interface{}{
		"attributes": map[string]interface{}{
			"accounts_conns":        []string{"*localhost"},
			"enabled":               true,
			"indexed_selects":       true,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*localhost"},
			"stats_conns":           []string{"*localhost"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.OptsAttributesProcessRuns: float64(1),
			},
			utils.ProfileRunsCfg: 0,
		},
	}
	if errGetRld != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, errGetRld)
	}
	if !reflect.DeepEqual(expectedGetRld, replyGetRld) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expectedGetRld, replyGetRld)
	}
}

func TestConfigGetSetConfigFromJSONErr(t *testing.T) {
	//for coverage purposes only
	var err error
	cfgTestPath := path.Join(*dataDir, "conf", "samples", "tutinternal")
	cfg, err := config.NewCGRConfigFromPath(cfgTestPath)
	if err != nil {
		t.Error(err)
	}
	rlcCfg := NewConfigSv1(cfg)
	args := &config.SetConfigFromJSONArgs{
		APIOpts: nil,
		Tenant:  utils.CGRateSorg,
		Config:  "{\"attributes\":{\"accounts_conns\":[\"*localhost\"],\"opts\":{\"*processRuns\":2},\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"resources_conns\":[\"*localhost\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}",
		DryRun:  true,
	}
	var reply string
	err = rlcCfg.SetConfigFromJSON(context.Background(), args, &reply)
	expected := "OK"
	if err != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expected, reply)
	}

	argsGet := &config.SectionWithAPIOpts{
		APIOpts:  nil,
		Tenant:   utils.CGRateSorg,
		Sections: []string{"attributes"},
	}
	var replyGet string
	errGet := rlcCfg.GetConfigAsJSON(context.Background(), argsGet, &replyGet)
<<<<<<< HEAD
	expectedGet := `{"attributes":{"accounts_conns":["*localhost"],"default_opts":{"*processRuns":1},"enabled":true,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"profile_runs":0,"resources_conns":["*localhost"],"stats_conns":["*localhost"],"suffix_indexed_fields":[]}}`
=======
	expectedGet := "{\"attributes\":{\"accounts_conns\":[\"*localhost\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*processRuns\":1},\"prefix_indexed_fields\":[],\"resources_conns\":[\"*localhost\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}"
>>>>>>> 3df2de22e (Rename default_opts to opts)
	if err != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", nil, errGet)
	}
	if !reflect.DeepEqual(expectedGet, replyGet) {
		t.Errorf("Expected <%+v>, \nReceived <%+v>", expectedGet, replyGet)
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
