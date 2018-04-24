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

	"github.com/cgrates/cgrates/utils"
	"github.com/dlintw/goconf"
)

func TestCgrLoaderCfgSetDefault(t *testing.T) {
	rcv := &LoaderCfg{}
	rcv.setDefaults()
	expected := &LoaderCfg{
		DataDBType:      cgrCfg.DataDbType,
		DataDBHost:      utils.MetaDynamic,
		DataDBPort:      utils.MetaDynamic,
		DataDBName:      utils.MetaDynamic,
		DataDBUser:      utils.MetaDynamic,
		DataDBPass:      utils.MetaDynamic,
		StorDBType:      cgrCfg.StorDBType,
		StorDBHost:      utils.MetaDynamic,
		StorDBPort:      utils.MetaDynamic,
		StorDBName:      utils.MetaDynamic,
		StorDBUser:      utils.MetaDynamic,
		StorDBPass:      utils.MetaDynamic,
		Flush:           false,
		Tpid:            "",
		DataPath:        "./",
		Version:         false,
		Verbose:         false,
		DryRun:          false,
		Validate:        false,
		Stats:           false,
		FromStorDB:      false,
		ToStorDB:        false,
		RpcEncoding:     "json",
		RalsAddress:     cgrCfg.RPCJSONListen,
		CdrstatsAddress: cgrCfg.RPCJSONListen,
		UsersAddress:    cgrCfg.RPCJSONListen,
		RunId:           "",
		LoadHistorySize: cgrCfg.LoadHistorySize,
		Timezone:        cgrCfg.DefaultTimezone,
		DisableReverse:  false,
		FlushStorDB:     false,
		Remove:          false,
	}

	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", expected, rcv)
	}
}

func TestCgrLoaderCfgNewDefault(t *testing.T) {
	rcv := NewDefaultLoaderConfig()
	expected := &LoaderCfg{
		DataDBType:      cgrCfg.DataDbType,
		DataDBHost:      utils.MetaDynamic,
		DataDBPort:      utils.MetaDynamic,
		DataDBName:      utils.MetaDynamic,
		DataDBUser:      utils.MetaDynamic,
		DataDBPass:      utils.MetaDynamic,
		StorDBType:      cgrCfg.StorDBType,
		StorDBHost:      utils.MetaDynamic,
		StorDBPort:      utils.MetaDynamic,
		StorDBName:      utils.MetaDynamic,
		StorDBUser:      utils.MetaDynamic,
		StorDBPass:      utils.MetaDynamic,
		Flush:           false,
		Tpid:            "",
		DataPath:        "./",
		Version:         false,
		Verbose:         false,
		DryRun:          false,
		Validate:        false,
		Stats:           false,
		FromStorDB:      false,
		ToStorDB:        false,
		RpcEncoding:     "json",
		RalsAddress:     cgrCfg.RPCJSONListen,
		CdrstatsAddress: cgrCfg.RPCJSONListen,
		UsersAddress:    cgrCfg.RPCJSONListen,
		RunId:           "",
		LoadHistorySize: cgrCfg.LoadHistorySize,
		Timezone:        cgrCfg.DefaultTimezone,
		DisableReverse:  false,
		FlushStorDB:     false,
		Remove:          false,
	}

	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", expected, rcv)
	}
}

func TestCgrLoaderCfgLoad(t *testing.T) {
	cfgPath := "/usr/share/cgrates/conf/cgrates/cgr-loader.cfg"
	c, err := goconf.ReadConfigFile(cfgPath)
	if err != nil {
		t.Error(err)
	}
	rcv := &LoaderCfg{}
	if err := rcv.loadConfig(c); err != nil {
		t.Error(err)
	}
	expected := &LoaderCfg{
		DataDBType:      "redis",
		DataDBHost:      "127.0.0.1",
		DataDBPort:      "6379",
		DataDBName:      "10",
		DataDBUser:      "cgrates",
		DataDBPass:      "testdatapw",
		StorDBType:      "mysql",
		StorDBHost:      "127.0.0.1",
		StorDBPort:      "3306",
		StorDBName:      "cgrates",
		StorDBUser:      "cgrates",
		StorDBPass:      "teststorpw",
		Tpid:            "testtpid",
		DataPath:        "./",
		RpcEncoding:     "json",
		RalsAddress:     "testRALsAddress",
		CdrstatsAddress: "testcdrstatsaddress",
		UsersAddress:    "testuseraddress",
		RunId:           "testrunId",
		LoadHistorySize: 10,
		Timezone:        "Local",
		DisableReverse:  false,
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
