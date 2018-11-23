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
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderSCfgloadFromJsonCfg(t *testing.T) {
	var loadscfg, expected LoaderSCfg
	if err := loadscfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(loadscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, loadscfg)
	}
	if err := loadscfg.loadFromJsonCfg(new(LoaderJsonCfg), utils.INFIELD_SEP); err != nil {
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
		"caches_conns": [
			{"address": "*internal"},						// address where to reach the CacheS for data reload, empty for no reloads  <""|*internal|x.y.z.y:1234>
		],
		"field_separator": ",",								// separator used in case of csv files
		"tp_in_dir": "/var/spool/cgrates/loader/in",		// absolute path towards the directory where the CDRs are stored
		"tp_out_dir": "/var/spool/cgrates/loader/out",		// absolute path towards the directory where processed CDRs will be moved
		"data":[											// data profiles to load
			{
				"type": "*attributes",						// data source type
				"file_name": "Attributes.csv",				// file name in the tp_in_dir
				"fields": [
					{"tag": "TenantID", "field_id": "Tenant", "type": "*composed", "value": "~0", "mandatory": true},
				],
			},]
		}
	]
}`
	val, err := NewRSRParsers("~0", true, utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	ten, err := NewRSRParsers("cgrates.org", true, utils.INFIELD_SEP)
	if err != nil {
		t.Error(err)
	}
	expected = LoaderSCfg{
		Id:             "*default",
		Tenant:         ten,
		LockFileName:   ".cgr.lck",
		CacheSConns:    []*HaPoolConfig{{Address: utils.MetaInternal}},
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
						FieldId:   "Tenant",
						Type:      "*composed",
						Value:     val,
						Mandatory: true,
					},
				},
			},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLoadersCfg, err := jsnCfg.LoaderJsonCfg(); err != nil {
		t.Error(err)
	} else if err = loadscfg.loadFromJsonCfg(jsnLoadersCfg[0], utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, loadscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(loadscfg))
	}
}
