/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"testing"
)

func TestConfigDBOptsInvalid(t *testing.T) {
	cfgJSONStr := `{
	"config_db": {                               
        "opts":{
            "redisClusterSync": "invalid",              
        }
    }}`
	expErr := "time: invalid duration \"invalid\""
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}

}

func TestConfigDBCfgloadFromJSONCfg(t *testing.T) {
	str := "test"
	dbcfg := &ConfigDBCfg{}
	jsnDbCfg := &ConfigDbJsonCfg{
		Db_type: &str,
	}

	err := dbcfg.loadFromJSONCfg(jsnDbCfg)
	if err != nil {
		t.Error(err)
	}

	if dbcfg.Type != "*test" {
		t.Error(dbcfg.Type)
	}
}
