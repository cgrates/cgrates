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

// import (
// 	"io/ioutil"
// 	"reflect"
// 	"testing"

// 	"github.com/cgrates/cgrates/utils"
// 	//"github.com/dlintw/goconf"
// )

// func TestCgrLoaderCfgLoad(t *testing.T) {
// 	cfgPath := "/go/src/github.com/cgrates/cgrates/config/config.go"
// 	c, err := ioutil.ReadFile(cfgPath)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	rcv := &CGRConfig{}
// 	if err := rcv.loadConfig(c); err != nil {
// 		t.Error(err)
// 	}
// 	expected := &CGRConfig{
// 		DataDbType:     "redis",
// 		DataDbHost:     "127.0.0.1",
// 		DataDbPort:     "6379",
// 		DataDbName:     "10",
// 		DataDbUser:     "cgrates",
// 		DataDbPass:     "CGRateS",
// 		StorDBType:     "mysql",
// 		StorDBHost:     "127.0.0.1",
// 		StorDBPort:     "3306",
// 		StorDBName:     "cgrates",
// 		StorDBUser:     "cgrates",
// 		StorDBPass:     "CGRateS",
// 		DBDataEncoding: "json",
// 		//Tpid:            "",
// 		//DataPath:        "./",
// 		RpcEncoding:     "json",
// 		RalsAddress:     "127.0.0.1:2012",
// 		RunId:           "",
// 		LoadHistorySize: 10,
// 		Timezone:        "Local",
// 		DisableReverse:  false,
// 	}
// 	if !reflect.DeepEqual(expected, rcv) {
// 		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
// 	}
// }
