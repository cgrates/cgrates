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
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
	"github.com/dlintw/goconf"
)

func TestCgrLoaderCfgLoad(t *testing.T) {
	cfgPath := "/usr/share/cgrates/conf/samples/cgrloaderconfig/cgr-loader.cfg"
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
		DataDBPass:      "CGRateS",
		StorDBType:      "mysql",
		StorDBHost:      "127.0.0.1",
		StorDBPort:      "3306",
		StorDBName:      "cgrates",
		StorDBUser:      "cgrates",
		StorDBPass:      "CGRateS",
		Tpid:            "",
		DataPath:        "./",
		RpcEncoding:     "json",
		RalsAddress:     "127.0.0.1:2012",
		CdrstatsAddress: "127.0.0.1:2012",
		UsersAddress:    "127.0.0.1:2012",
		RunId:           "",
		LoadHistorySize: 10,
		Timezone:        "Local",
		DisableReverse:  false,
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
