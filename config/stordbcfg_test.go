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

	"github.com/cgrates/birpc/context"
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
		Items: map[string]*ItemOptsJson{
			utils.MetaCDRs: {
				Remote:    utils.BoolPointer(true),
				Replicate: utils.BoolPointer(false),
			},
		},
		Opts: &DBOptsJson{
			SQLMaxOpenConns:    utils.IntPointer(100),
			SQLMaxIdleConns:    utils.IntPointer(10),
			SQLConnMaxLifetime: utils.StringPointer("0"),
			MYSQLDSNParams:     make(map[string]string),
			MySQLLocation:      utils.StringPointer("UTC"),
			MongoConnScheme:    utils.StringPointer("mongodb+srv"),
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
		Items: map[string]*ItemOpts{
			utils.MetaCDRs: {
				Remote:    true,
				Replicate: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns:   100,
			SQLMaxIdleConns:   10,
			SQLDSNParams:      make(map[string]string),
			MongoQueryTimeout: 10 * time.Second,
			MongoConnScheme:   "mongodb+srv",
			PgSSLMode:         "disable",
			MySQLLocation:     "UTC",
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.Opts, jsonCfg.storDbCfg.Opts) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.Opts), utils.ToJSON(jsonCfg.storDbCfg.Opts))
	} else if !reflect.DeepEqual(expected.RplConns, jsonCfg.storDbCfg.RplConns) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.RplConns), utils.ToJSON(jsonCfg.storDbCfg.RplConns))
	} else if !reflect.DeepEqual(expected.RmtConns, jsonCfg.storDbCfg.RmtConns) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(expected.RmtConns), utils.ToJSON(jsonCfg.storDbCfg.RmtConns))
	}

	newCfgJSON := cfgJSON

	cfgJSON = nil
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}

	newCfgJSON.Opts = &DBOptsJson{
		SQLConnMaxLifetime: utils.StringPointer("error"),
	}

	experr := `time: invalid duration "error"`
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(newCfgJSON); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	newCfgJSON.Opts = &DBOptsJson{
		MongoQueryTimeout: utils.StringPointer("error"),
	}

	if err = jsonCfg.storDbCfg.loadFromJSONCfg(newCfgJSON); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}

func TestStoreDbloadFromJsonOpts(t *testing.T) {
	strDbOpts := &StorDBOpts{
		SQLMaxOpenConns: 1,
	}

	exp := &StorDBOpts{
		SQLMaxOpenConns: 1,
	}

	strDbOpts.loadFromJSONCfg(nil)
	if !reflect.DeepEqual(exp, strDbOpts) {
		t.Errorf("Expected %v \n but received \n %v", exp, strDbOpts)
	}
}

func TestStoreDbCfgloadFromJsonCfgCase2(t *testing.T) {
	storDbJSON := &DbJsonCfg{
		Replication_conns: &[]string{utils.MetaInternal},
	}
	expected := "Replication connection ID needs to be different than *internal "
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.storDbCfg.loadFromJSONCfg(storDbJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", storDbJSON, expected)
	}
}

func TestStoreDbCfgloadFromJsonCfgCase3(t *testing.T) {
	storDbJSON := &DbJsonCfg{
		Remote_conns: &[]string{utils.MetaInternal},
	}
	expected := "Remote connection ID needs to be different than *internal "
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
	"db_type": "*mongo",
	}
}`
	dbcfg.Opts = &StorDBOpts{}
	expected := StorDbCfg{
		Type: "*mongo",
		Opts: &StorDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = dbcfg.Load(context.Background(), jsnCfg, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dbcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, dbcfg)
	}
	cfgJSONStr = `{
"stor_db": {
	"db_type": "*mongo",
	"db_port": -1,
	}
}`
	expected = StorDbCfg{
		Type: "*mongo",
		Port: "27017",
		Opts: &StorDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = dbcfg.Load(context.Background(), jsnCfg, nil); err != nil {
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
		Type: "*internal",
		Port: "internal",
		Opts: &StorDBOpts{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = dbcfg.Load(context.Background(), jsnCfg, nil); err != nil {
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
				"mongoConnScheme":"mongodb+srv",
				"PgSSLMode":"disable",		
				"mysqlLocation": "UTC",			
			},
			"items":{
				"session_costs": {}, 
				"cdrs": {}, 		
			},
		},
}`

	eMap := map[string]any{
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
		utils.OptsCfg: map[string]any{
			utils.SQLMaxOpenConnsCfg:    100,
			utils.SQLMaxIdleConnsCfg:    10,
			utils.SQLConnMaxLifetimeCfg: "0s",
			utils.MYSQLDSNParams:        make(map[string]string),
			utils.MongoQueryTimeoutCfg:  "10s",
			utils.MongoConnSchemeCfg:    "mongodb+srv",
			utils.PgSSLModeCfg:          "disable",
			utils.MysqlLocation:         "UTC",
		},
		utils.ItemsCfg: map[string]any{
			utils.SessionCostsTBL: map[string]any{utils.RemoteCfg: false, utils.ReplicateCfg: false},
			utils.CDRsTBL:         map[string]any{utils.RemoteCfg: false, utils.ReplicateCfg: false},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cfgCgr.storDbCfg.AsMapInterface("").(map[string]any)
		if !reflect.DeepEqual(eMap[utils.ItemsCfg].(map[string]any)[utils.SessionSConnsCfg],
			rcv[utils.ItemsCfg].(map[string]any)[utils.SessionSConnsCfg]) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap[utils.ItemsCfg].(map[string]any)[utils.SessionSConnsCfg]),
				utils.ToJSON(rcv[utils.ItemsCfg].(map[string]any)[utils.SessionSConnsCfg]))
		} else if !reflect.DeepEqual(eMap[utils.OptsCfg], rcv[utils.OptsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.OptsCfg]), utils.ToJSON(rcv[utils.OptsCfg]))
		} else if !reflect.DeepEqual(eMap[utils.PrefixIndexedFieldsCfg], rcv[utils.PrefixIndexedFieldsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.PrefixIndexedFieldsCfg]), utils.ToJSON(rcv[utils.PrefixIndexedFieldsCfg]))
		} else if !reflect.DeepEqual(eMap[utils.RemoteConnsCfg], rcv[utils.RemoteConnsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.RemoteConnsCfg]), utils.ToJSON(rcv[utils.RemoteConnsCfg]))
		} else if !reflect.DeepEqual(eMap[utils.ReplicationConnsCfg], rcv[utils.ReplicationConnsCfg]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.ReplicationConnsCfg]), utils.ToJSON(rcv[utils.ReplicationConnsCfg]))
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
		Items: map[string]*ItemOpts{
			utils.MetaCDRs: {
				Remote:    true,
				Replicate: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns:   100,
			SQLMaxIdleConns:   10,
			MongoQueryTimeout: 10 * time.Second,
			PgSSLMode:         "disable",
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

func TestDiffStorDBJsonCfg(t *testing.T) {
	var d *DbJsonCfg

	v1 := &StorDbCfg{
		Type:                "*mysql",
		Host:                "localhost",
		Port:                "8080",
		Name:                "cgrates",
		User:                "cgrates_user",
		Password:            "cgrates_password",
		StringIndexedFields: []string{"*req.index1"},
		PrefixIndexedFields: []string{"*req.index2"},
		RmtConns:            []string{"*rmt_conn"},
		RplConns:            []string{"*rpl_conns"},
		Items: map[string]*ItemOpts{
			"ITEM_1": {
				Remote: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns: 50,
		},
	}

	v2 := &StorDbCfg{
		Type:                "*postgres",
		Host:                "0.0.0.0",
		Port:                "8037",
		Name:                "itsyscom",
		User:                "itsyscom_user",
		Password:            "itsyscom_password",
		StringIndexedFields: []string{"*req.index11"},
		PrefixIndexedFields: []string{"*req.index22"},
		RmtConns:            []string{"*rmt_conn2"},
		RplConns:            []string{"*rpl_conn2"},
		Items: map[string]*ItemOpts{
			"ITEM_1": {
				Remote: true,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns: 100,
		},
	}

	expected := &DbJsonCfg{
		Db_type:               utils.StringPointer("*postgres"),
		Db_host:               utils.StringPointer("0.0.0.0"),
		Db_port:               utils.IntPointer(8037),
		Db_name:               utils.StringPointer("itsyscom"),
		Db_user:               utils.StringPointer("itsyscom_user"),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Db_password:           utils.StringPointer("itsyscom_password"),
		Remote_conns:          &[]string{"*rmt_conn2"},
		Replication_conns:     &[]string{"*rpl_conn2"},
		Items: map[string]*ItemOptsJson{
			"ITEM_1": {
				Remote: utils.BoolPointer(true),
			},
		},
		Opts: &DBOptsJson{
			SQLMaxOpenConns: utils.IntPointer(100),
		},
	}

	rcv := diffStorDBJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &DbJsonCfg{
		Items: map[string]*ItemOptsJson{},
		Opts:  &DBOptsJson{},
	}
	rcv = diffStorDBJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffStorDBOptsJsonCfg(t *testing.T) {
	var d *DBOptsJson

	v1 := &StorDBOpts{
		SQLConnMaxLifetime: 10 * time.Second,
		SQLMaxOpenConns:    100,
		SQLMaxIdleConns:    10,
		MongoQueryTimeout:  10 * time.Second,
		PgSSLMode:          "disable",
		MySQLLocation:      "UTC",
	}

	v2 := &StorDBOpts{
		SQLConnMaxLifetime: 11 * time.Second,
		SQLMaxOpenConns:    101,
		SQLMaxIdleConns:    11,
		MongoQueryTimeout:  11 * time.Second,
		PgSSLMode:          "enable",
		MySQLLocation:      "/usr/share/db",
	}

	exp := &DBOptsJson{
		SQLConnMaxLifetime: utils.StringPointer("11s"),
		SQLMaxOpenConns:    utils.IntPointer(101),
		SQLMaxIdleConns:    utils.IntPointer(11),
		MongoQueryTimeout:  utils.StringPointer("11s"),
		PgSSLMode:          utils.StringPointer("enable"),
		MySQLLocation:      utils.StringPointer("/usr/share/db"),
	}

	rcv := diffStorDBOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestStorDbCloneSection(t *testing.T) {
	storDbCfg := &StorDbCfg{
		Type:                "*mysql",
		Host:                "localhost",
		Port:                "8080",
		Name:                "cgrates",
		User:                "cgrates_user",
		Password:            "cgrates_password",
		StringIndexedFields: []string{"*req.index1"},
		PrefixIndexedFields: []string{"*req.index2"},
		RmtConns:            []string{"*rmt_conn"},
		RplConns:            []string{"*rpl_conns"},
		Items: map[string]*ItemOpts{
			"ITEM_1": {
				Remote: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns: 50,
		},
	}

	exp := &StorDbCfg{
		Type:                "*mysql",
		Host:                "localhost",
		Port:                "8080",
		Name:                "cgrates",
		User:                "cgrates_user",
		Password:            "cgrates_password",
		StringIndexedFields: []string{"*req.index1"},
		PrefixIndexedFields: []string{"*req.index2"},
		RmtConns:            []string{"*rmt_conn"},
		RplConns:            []string{"*rpl_conns"},
		Items: map[string]*ItemOpts{
			"ITEM_1": {
				Remote: false,
			},
		},
		Opts: &StorDBOpts{
			SQLMaxOpenConns: 50,
		},
	}

	rcv := storDbCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
