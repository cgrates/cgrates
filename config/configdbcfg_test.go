// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"testing"
)

func TestConfigDBOptsInvalid(t *testing.T) {
	cfgJSONStr := `{
	"configDB": {                               
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
