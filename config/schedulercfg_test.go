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
)

func TestSchedulerCfgloadFromJsonCfg(t *testing.T) {
	var schdcfg, expected SchedulerCfg
	if err := schdcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(schdcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, schdcfg)
	}
	if err := schdcfg.loadFromJsonCfg(new(SchedulerJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(schdcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, schdcfg)
	}
	cfgJSONStr := `{
"schedulers": {
	"enabled": true,				// start Scheduler service: <true|false>
	"cdrs_conns": [],				// address where to reach CDR Server, empty to disable CDR capturing <*internal|x.y.z.y:1234>
	},
}`
	expected = SchedulerCfg{
		Enabled:   true,
		CDRsConns: []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSchCfg, err := jsnCfg.SchedulerJsonCfg(); err != nil {
		t.Error(err)
	} else if err = schdcfg.loadFromJsonCfg(jsnSchCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, schdcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, schdcfg)
	}
}

func TestSchedulerCfgAsMapInterface(t *testing.T) {
	var schdcfg SchedulerCfg
	cfgJSONStr := `{
	"schedulers": {
		"enabled": true,				
		"cdrs_conns": [],				
		"filters": [],
	},
}`
	eMap := map[string]interface{}{
		"enabled":    true,
		"cdrs_conns": []string{},
		"filters":    []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSchCfg, err := jsnCfg.SchedulerJsonCfg(); err != nil {
		t.Error(err)
	} else if err = schdcfg.loadFromJsonCfg(jsnSchCfg); err != nil {
		t.Error(err)
	} else if rcv := schdcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
