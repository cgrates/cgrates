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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSchedulerCfgloadFromJsonCfg(t *testing.T) {
	var schdcfg, expected SchedulerCfg
	if err := schdcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(schdcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, schdcfg)
	}
	if err := schdcfg.loadFromJsonCfg(new(SchedulerJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(schdcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, schdcfg)
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
		t.Errorf("Expected: %+v , received: %+v", expected, schdcfg)
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
	eMap := map[string]any{
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSchedulerCfgloadFromJsonCfg2(t *testing.T) {
	s := SchedulerCfg{}

	tests := []struct {
		name string
		arg  *SchedulerJsonCfg
		err  string
	}{
		{
			name: "cdrs conns diff from *internal",
			arg:  &SchedulerJsonCfg{Cdrs_conns: &[]string{"test"}},
			err:  "",
		},
		{
			name: "cdrs conns equal to *internal",
			arg:  &SchedulerJsonCfg{Cdrs_conns: &[]string{"*internal"}},
			err:  "",
		},
		{
			name: "filers with value",
			arg:  &SchedulerJsonCfg{Filters: &[]string{"test"}},
			err:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.loadFromJsonCfg(tt.arg)

			if err != nil {
				if err.Error() != tt.err {
					t.Errorf("\nExpected: %+v\nReceived: %+v", tt.err, err)
				}
			}
		})
	}
}
