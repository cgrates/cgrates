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
	var loadscfg, expected LoaderSCfg
	if err := loadscfg.loadFromJsonCfg(nil, nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	if err := loadscfg.loadFromJsonCfg(new(LoaderJsonCfg), nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	cfgJSONStr := `{
"loaders": [
	{
		"id": "*default",									// identifier of the Loader
		"enabled": false,									// starts as service: <true|false>.
		"tenant": "cgrates.org",							// tenant used in filterS.Pass
		"dry_run": false,									// do not send the CDRs to CDRS, just parse them
		"run_delay": 0,										// sleep interval in seconds between consecutive runs, 0 to use automation via inotify
		"lock_filename": ".cgr.lck",						// Filename containing concurrency lock in case of delayed processing
		"caches_conns": ["*internal"],
		"field_separator": ",",								// separator used in case of csv files
		"tp_in_dir": "/var/spool/cgrates/loader/in",		// absolute path towards the directory where the CDRs are stored
		"tp_out_dir": "/var/spool/cgrates/loader/out",		// absolute path towards the directory where processed CDRs will be moved
		"data":[											// data profiles to load
			{
				"type": "*attributes",						// data source type
				"file_name": "Attributes.csv",				// file name in the tp_in_dir
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*composed", "value": "~0", "mandatory": true},
				],
			},]
		}
	]
}`
	val, err := NewRSRParsers("~0", utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	ten, err := NewRSRParsers("cgrates.org", utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	expected = LoaderSCfg{
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
						Type:      "*composed",
						Value:     val,
						Mandatory: true,
						Layout:    time.RFC3339,
					},
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderJsonCfg(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg[0], nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, loadscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(loadscfg))
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
