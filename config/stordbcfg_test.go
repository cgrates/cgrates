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

func TestStoreDbCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &DbJsonCfg{
		Db_type:               utils.StringPointer(utils.MySQL),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(-1),
		Db_name:               utils.StringPointer(utils.CGRateSLwr),
		Db_user:               utils.StringPointer(utils.CGRateSLwr),
		Db_password:           utils.StringPointer("pass123"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1"},
		Remote_conns:          &[]string{"*conn1"},
		Replication_conns:     &[]string{"*conn1"},
		Items: &map[string]*ItemOptJson{
			utils.MetaSessionsCosts: {
				Remote:    utils.BoolPointer(true),
				Replicate: utils.BoolPointer(true),
			},
			utils.MetaCDRs: {
				Remote:    utils.BoolPointer(true),
				Replicate: utils.BoolPointer(false),
			},
		},
		Opts: &DBOptsJson{
			SQLMaxOpenConns:    utils.IntPointer(100),
			SQLMaxIdleConns:    utils.IntPointer(10),
			SQLConnMaxLifetime: utils.StringPointer("0"),
			MySQLDSNParams:     make(map[string]string),
			MySQLLocation:      utils.StringPointer("UTC"),
		},
	}
	expected := &StorDbCfg{
		Type:                utils.MySQL,
		Host:                "127.0.0.1",
		Port:                "-1",
		Name:                utils.CGRateSLwr,
		User:                utils.CGRateSLwr,
		Password:            "pass123",
		StringIndexedFields: []string{"*req.index1"},
		PrefixIndexedFields: []string{"*req.index1"},
		RmtConns:            []string{"*conn1"},
		RplConns:            []string{"*conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaSessionsCosts: {
				Remote:    true,
				Replicate: true,
			},
			utils.MetaCDRs: {
				Remote:    true,
				Replicate: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns:    100,
			SQLMaxIdleConns:    10,
			SQLConnMaxLifetime: 0,
			MongoQueryTimeout:  10 * time.Second,
			PgSSLMode:          "disable",
			MySQLLocation:      "UTC",
			MySQLDSNParams:     make(map[string]string),
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.Items[utils.MetaSessionsCosts], jsonCfg.storDbCfg.Items[utils.MetaSessionsCosts]) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.Items[utils.MetaSessionsCosts]),
			utils.ToJSON(jsonCfg.storDbCfg.Items[utils.MetaSessionsCosts]))
	} else if !reflect.DeepEqual(expected.Opts, jsonCfg.storDbCfg.Opts) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.Opts), utils.ToJSON(jsonCfg.storDbCfg.Opts))
	} else if !reflect.DeepEqual(expected.RplConns, jsonCfg.storDbCfg.RplConns) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.RplConns), utils.ToJSON(jsonCfg.storDbCfg.RplConns))
	} else if !reflect.DeepEqual(expected.RmtConns, jsonCfg.storDbCfg.RmtConns) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.RmtConns), utils.ToJSON(jsonCfg.storDbCfg.RmtConns))
	}

	if err := jsonCfg.storDbCfg.Opts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err := jsonCfg.storDbCfg.Opts.loadFromJSONCfg(&DBOptsJson{
		SQLConnMaxLifetime: utils.StringPointer("test1"),
	}); err == nil {
		t.Error(err)
	} else if err := jsonCfg.storDbCfg.Opts.loadFromJSONCfg(&DBOptsJson{
		MongoQueryTimeout: utils.StringPointer("test2"),
	}); err == nil {
		t.Error(err)
	} else if err := jsonCfg.storDbCfg.loadFromJSONCfg(&DbJsonCfg{
		Items: &map[string]*ItemOptJson{
			utils.MetaSessionsCosts: {
				Ttl: utils.StringPointer("test3"),
			},
		}}); err == nil {

	}
}

func TestStoreDbCfgloadFromJsonCfgCase2(t *testing.T) {
	storDbJSON := &DbJsonCfg{
		Replication_conns: &[]string{utils.MetaInternal},
	}
	expected := "Replication connection ID needs to be different than *internal"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(storDbJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", storDbJSON, expected)
	}
}

func TestStoreDbCfgloadFromJsonCfgCase3(t *testing.T) {
	storDbJSON := &DbJsonCfg{
		Remote_conns: &[]string{utils.MetaInternal},
	}
	expected := "Remote connection ID needs to be different than *internal"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(storDbJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", storDbJSON, expected)
	}
}

func TestStoreDbCfgloadFromJsonCfgCase4(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()
	clonedStoreDb := jsonCfg.storDbCfg.Clone()
	if !reflect.DeepEqual(clonedStoreDb, jsonCfg.storDbCfg) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(clonedStoreDb), utils.ToJSON(jsonCfg.storDbCfg))
	}
}

func TestStoreDbCfgloadFromJsonCfgPort(t *testing.T) {
	var dbcfg StorDbCfg
	cfgJSONStr := `{
"stor_db": {
	"db_type": "mongo",
	}
}`
	dbcfg.Opts = &StorDBOpts{}
	expected := StorDbCfg{
		Type: "mongo",
		Opts: &StorDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"stor_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`
	expected = StorDbCfg{
		Type: "mongo",
		Port: "27017",
		Opts: &StorDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"stor_db": {
	"db_type": "*internal",
	"db_port": -1,
	}
}`
	expected = StorDbCfg{
		Type: "internal",
		Port: "internal",
		Opts: &StorDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDataDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if err = dbcfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
}

func TestStorDbCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"stor_db": {								
			"db_type": "*mysql",					
			"db_host": "127.0.0.1",					
			"db_port": -1,						
			"db_name": "cgrates",					
			"db_user": "cgrates",					
			"db_password": "",						
			"string_indexed_fields": [],			
			"prefix_indexed_fields":[],	
            "remote_conns": ["*conn1"],
            "replication_conns": ["*conn1"],
			"opts": {	
				"sqlMaxOpenConns": 100,					
				"sqlMaxIdleConns": 10,					
				"sqlConnMaxLifetime": "0",
				"mysqlDSNParams": {},  			
				"mongoQueryTimeout":"10s",
				"pgSSLMode":"disable",		
				"mysqlLocation": "UTC",			
			},
			"items":{
				"session_costs": {}, 
				"cdrs": {}, 		
			},
		},
}`

	eMap := map[string]interface{}{
		utils.DataDbTypeCfg:          "*mysql",
		utils.DataDbHostCfg:          "127.0.0.1",
		utils.DataDbPortCfg:          3306,
		utils.DataDbNameCfg:          "cgrates",
		utils.DataDbUserCfg:          "cgrates",
		utils.DataDbPassCfg:          "",
		utils.StringIndexedFieldsCfg: []string{},
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.RemoteConnsCfg:         []string{"*conn1"},
		utils.ReplicationConnsCfg:    []string{"*conn1"},
		utils.OptsCfg: map[string]interface{}{
			utils.SQLMaxOpenConnsCfg:    100,
			utils.SQLMaxIdleConnsCfg:    10,
			utils.SQLConnMaxLifetimeCfg: "0s",
			utils.MYSQLDSNParams:        make(map[string]string),
			utils.MongoQueryTimeoutCfg:  "10s",
			utils.PgSSLModeCfg:          "disable",
			utils.MysqlLocation:         "UTC",
		},
		utils.ItemsCfg: map[string]interface{}{
			utils.SessionCostsTBL: map[string]interface{}{utils.RemoteCfg: false, utils.ReplicateCfg: false},
			utils.CDRsTBL:         map[string]interface{}{utils.RemoteCfg: false, utils.ReplicateCfg: false},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cfgCgr.storDbCfg.AsMapInterface()
		if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg],
			rcv[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg],
				rcv[utils.ItemsCfg].(map[string]interface{})[utils.SessionSConnsCfg])
		} else if !reflect.DeepEqual(eMap[utils.OptsCfg], rcv[utils.OptsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.OptsCfg]), utils.ToJSON(rcv[utils.OptsCfg]))
		} else if !reflect.DeepEqual(eMap[utils.PrefixIndexedFieldsCfg], rcv[utils.PrefixIndexedFieldsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", eMap[utils.PrefixIndexedFieldsCfg], rcv[utils.PrefixIndexedFieldsCfg])
		} else if !reflect.DeepEqual(eMap[utils.RemoteConnsCfg], rcv[utils.RemoteConnsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", eMap[utils.RemoteConnsCfg], rcv[utils.RemoteConnsCfg])
		} else if !reflect.DeepEqual(eMap[utils.ReplicationConnsCfg], rcv[utils.ReplicationConnsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", eMap[utils.ReplicationConnsCfg], rcv[utils.ReplicationConnsCfg])
		}
	}
}

func TestStorDbCfgClone(t *testing.T) {
	ban := &StorDbCfg{
		Type:                utils.MySQL,
		Host:                "127.0.0.1",
		Port:                "-1",
		Name:                utils.CGRateSLwr,
		User:                utils.CGRateSLwr,
		Password:            "pass123",
		StringIndexedFields: []string{"*req.index1"},
		PrefixIndexedFields: []string{"*req.index1"},
		RmtConns:            []string{"*conn1"},
		RplConns:            []string{"*conn1"},
		Items: map[string]*ItemOpt{
			utils.MetaSessionsCosts: {
				Remote:    true,
				Replicate: true,
			},
			utils.MetaCDRs: {
				Remote:    true,
				Replicate: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns:    100,
			SQLMaxIdleConns:    10,
			SQLConnMaxLifetime: 0,
			MySQLDSNParams:     make(map[string]string),
			MySQLLocation:      "UTC",
			PgSSLMode:          "disable",
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.StringIndexedFields[0] = ""; ban.StringIndexedFields[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.PrefixIndexedFields[0] = ""; ban.PrefixIndexedFields[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.RmtConns[0] = ""; ban.RmtConns[0] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.RplConns[0] = ""; ban.RplConns[0] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.Items[utils.MetaCDRs].Remote = false; !ban.Items[utils.MetaCDRs].Remote {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Opts.PgSSLMode = ""; ban.Opts.PgSSLMode != "disable" {
		t.Errorf("Expected clone to not modify the cloned")
	}

}
