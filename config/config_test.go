/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
)

var cfg *CGRConfig

func TestConfigSharing(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	SetCgrConfig(cfg)
	cfgReturn := CgrConfig()
	if !reflect.DeepEqual(cfgReturn, cfg) {
		t.Errorf("Retrieved %v, Expected %v", cfgReturn, cfg)
	}
}

func TestLoadCgrCfgWithDefaults(t *testing.T) {
	JSN_CFG := `
{
"sm_freeswitch": {
	"enabled": true,				// starts SessionManager service: <true|false>
	"connections":[					// instantiate connections to multiple FreeSWITCH servers
		{"server": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3},
		{"server": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5}
	],
},

}`
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.SmFsConfig.Enabled = true
	eCgrCfg.SmFsConfig.Connections = []*FsConnConfig{
		&FsConnConfig{Server: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 3},
		&FsConnConfig{Server: "1.2.3.5:8021", Password: "ClueCon", Reconnects: 5},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.SmFsConfig, cgrCfg.SmFsConfig) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.SmFsConfig, cgrCfg.SmFsConfig)
	}
}
