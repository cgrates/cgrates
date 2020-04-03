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

func TestRadiusAgentCfgloadFromJsonCfg(t *testing.T) {
	var racfg, expected RadiusAgentCfg
	if err := racfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(racfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, racfg)
	}
	if err := racfg.loadFromJsonCfg(new(RadiusAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(racfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, racfg)
	}
	cfgJSONStr := `{
"radius_agent": {
	"enabled": false,											// enables the radius agent: <true|false>
	"listen_net": "udp",										// network to listen on <udp|tcp>
	"listen_auth": "127.0.0.1:1812",							// address where to listen for radius authentication requests <x.y.z.y:1234>
	"listen_acct": "127.0.0.1:1813",							// address where to listen for radius accounting requests <x.y.z.y:1234>
	"client_secrets": {											// hash containing secrets for clients connecting here <*default|$client_ip>
		"*default": "CGRateS.org"
	},
	"client_dictionaries": {									// per client path towards directory holding additional dictionaries to load (extra to RFC)
		"*default": "/usr/share/cgrates/radius/dict/",			// key represents the client IP or catch-all <*default|$client_ip>
	},
	"sessions_conns": ["*internal"],
	"request_processors": [],
},
}`
	expected = RadiusAgentCfg{
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string]string{utils.MetaDefault: "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRaCfg, err := jsnCfg.RadiusAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = racfg.loadFromJsonCfg(jsnRaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, racfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(racfg))
	}
}
