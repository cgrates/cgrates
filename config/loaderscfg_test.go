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

func TestLoaderSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"tenant": "cgrates.org",
		"lock_filename": ".cgr.lck",
		"caches_conns": ["*internal"],
		"field_separator": ",",
		"tp_in_dir": "/var/spool/cgrates/loader/in",
		"tp_out_dir": "/var/spool/cgrates/loader/out",
		"data":[
			{
				"type": "*attributes",
				"file_name": "Attributes.csv",
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*composed", "value": "~req.0", "mandatory": true,"layout": "2006-01-02T15:04:05Z07:00"},
					],
				},
			],
		},
	],
}`
	val, err := NewRSRParsers("~req.0", utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	ten, err := NewRSRParsers("cgrates.org", utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	expected := LoaderSCfgs{
		{
			Enabled:        true,
			Id:             utils.MetaDefault,
			Tenant:         ten,
			LockFileName:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Data: []*LoaderDataType{
				{
					Type:     "*attributes",
					Filename: "Attributes.csv",
					Fields: []*FCTemplate{
						{
							Tag:       "TenantID",
							Path:      "Tenant",
							pathSlice: []string{"Tenant"},
							pathItems: utils.PathItems{{Field: "Tenant"}},
							Type:      utils.META_COMPOSED,
							Value:     val,
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
				},
			},
		},
	}
	newCfg := new(CGRConfig)
	newCfg.generalCfg = new(GeneralCfg)
	newCfg.generalCfg.RSRSep = ";"
	if jsonCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = newCfg.loadLoaderSCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, newCfg.loaderCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newCfg.loaderCfg))
	}
}

func TestLoaderCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "",										
		"dry_run": false,									
		"run_delay": 0,										
		"lock_filename": ".cgr.lck",						
		"caches_conns": ["*internal"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]interface{}{
		{
			utils.IdCfg:             "*default",
			utils.EnabledCfg:        false,
			utils.TenantCfg:         "",
			utils.DryRunCfg:         false,
			utils.RunDelayCfg:       "0",
			utils.LockFileNameCfg:   ".cgr.lck",
			utils.CachesConnsCfg:    []string{"*internal"},
			utils.FieldSeparatorCfg: ",",
			utils.TpInDirCfg:        "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:       "/var/spool/cgrates/loader/out",
			utils.DataCfg: []map[string]interface{}{
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FieldsCfg: []map[string]interface{}{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
			},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.loaderCfg.AsMapInterface(cfgCgr.generalCfg.RSRSep)
		if !reflect.DeepEqual(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0],
			rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0]),
				utils.ToJSON(rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0]))
		} else if !reflect.DeepEqual(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1],
			rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1]),
				utils.ToJSON(rcv[0][utils.DataCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[1]))
		} else if !reflect.DeepEqual(eMap[0][utils.CachesConnsCfg], rcv[0][utils.CachesConnsCfg]) {
			t.Errorf("Expecetd %+v, received %+v", eMap[0][utils.CachesConnsCfg], rcv[0][utils.CachesConnsCfg])
		} else if !reflect.DeepEqual(eMap[0][utils.TpInDirCfg], rcv[0][utils.TpInDirCfg]) {
			t.Errorf("Expecetd %+v, received %+v", eMap[0][utils.TpInDirCfg], rcv[0][utils.TpInDirCfg])
		} else if !reflect.DeepEqual(eMap[0][utils.LockFileNameCfg], rcv[0][utils.LockFileNameCfg]) {
			t.Errorf("Expecetd %+v, received %+v", eMap[0][utils.LockFileNameCfg], rcv[0][utils.LockFileNameCfg])
		}
	}
}
