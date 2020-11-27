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
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestConfigsloadFromJsonCfg(t *testing.T) {
	jsonCfgs := &ConfigSCfgJson{
		Enabled:  utils.BoolPointer(true),
		Url:      utils.StringPointer("/randomURL/"),
		Root_dir: utils.StringPointer("/randomPath/"),
	}
	expectedCfg := &ConfigSCfg{
		Enabled: true,
		URL:     "/randomURL/",
		RootDir: "/randomPath/",
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.configSCfg.loadFromJSONCfg(jsonCfgs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cgrCfg.configSCfg, expectedCfg) {
		t.Errorf("Expected %+v, received %+v", expectedCfg, cgrCfg.configSCfg)
	}
}

func TestConfigsAsMapInterface(t *testing.T) {
	cfgsJSONStr := `{
      "configs": {
          "enabled": true,
          "url": "",
          "root_dir": "/var/spool/cgrates/configs"
      },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg: true,
		utils.URLCfg:     "",
		utils.RootDirCfg: "/var/spool/cgrates/configs",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgsJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.configSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestConfigsAsMapInterface2(t *testing.T) {
	cfgsJSONStr := `{
      "configs":{}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg: false,
		utils.URLCfg:     "/configs/",
		utils.RootDirCfg: "/var/spool/cgrates/configs",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgsJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.configSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestNewCGRConfigFromPathWithoutEnv(t *testing.T) {
	cfgsJSONStr := `{
		"general": {
			"node_id": "*env:NODE_ID",
		},
  }`
	cfg := NewDefaultCGRConfig()

	if err = cfg.loadConfigFromReader(strings.NewReader(cfgsJSONStr), []func(*CgrJsonCfg) error{cfg.loadFromJSONCfg}, true); err != nil {
		t.Fatal(err)
	}
	exp := "*env:NODE_ID"
	if cfg.GeneralCfg().NodeID != exp {
		t.Errorf("Expected %+v, received %+v", exp, cfg.GeneralCfg().NodeID)
	}
}

func TestConfigSCfgClone(t *testing.T) {
	cS := &ConfigSCfg{
		Enabled: true,
		URL:     "/randomURL/",
		RootDir: "/randomPath/",
	}
	rcv := cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.URL = ""; cS.URL != "/randomURL/" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
