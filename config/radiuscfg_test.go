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

func TestRadiusAgentCfgloadFromJsonCfg(t *testing.T) {
	var racfg, expected RadiusAgentCfg
	if err := racfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(racfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, racfg)
	}
	if err := racfg.loadFromJsonCfg(new(RadiusAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(racfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, racfg)
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
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(racfg))
	}
}

func TestRadiusAgentCfgAsMapInterface(t *testing.T) {
	var racfg RadiusAgentCfg
	cfgJSONStr := `{
	"radius_agent": {
		"enabled": false,
		"listen_net": "udp",
		"listen_auth": "127.0.0.1:1812",
		"listen_acct": "127.0.0.1:1813",
		"client_secrets": {
			"*default": "CGRateS.org"
		},
		"client_dictionaries": {
			"*default": "/usr/share/cgrates/radius/dict/",
		},
		"sessions_conns": ["*internal"],
		"request_processors": [
		],
	},
}`
	eMap := map[string]any{
		"enabled":     false,
		"listen_net":  "udp",
		"listen_auth": "127.0.0.1:1812",
		"listen_acct": "127.0.0.1:1813",
		"client_secrets": map[string]any{
			"*default": "CGRateS.org",
		},
		"client_dictionaries": map[string]any{
			"*default": "/usr/share/cgrates/radius/dict/",
		},
		"sessions_conns":     []string{"*internal"},
		"request_processors": []map[string]any{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRaCfg, err := jsnCfg.RadiusAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = racfg.loadFromJsonCfg(jsnRaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if rcv := racfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRadiusAgentCfgloadFromJsonCfg2(t *testing.T) {
	str := "test"
	str2 := "test)"
	self := &RadiusAgentCfg{
		RequestProcessors: []*RequestProcessor{{
			ID: str,
		}},
	}
	jsnCfg := &RadiusAgentJsonCfg{
		Sessions_conns: &[]string{str},
		Request_processors: &[]*ReqProcessorJsnCfg{{
			ID:       &str,
			Filters:  &[]string{str},
			Tenant:   &str,
			Timezone: &str,
			Flags:    &[]string{str},
			Request_fields: &[]*FcTemplateJsonCfg{{
				Tag:          &str,
				Type:         &str,
				Path:         &str,
				Attribute_id: &str,
				Filters:      &[]string{str},
			}},
			Reply_fields: &[]*FcTemplateJsonCfg{{
				Tag:          &str,
				Type:         &str,
				Path:         &str,
				Attribute_id: &str,
				Filters:      &[]string{str},
			}},
		}, {
			ID: &str2,
		}},
	}

	err := self.loadFromJsonCfg(jsnCfg, "")
	if err != nil {
		t.Error(err)
	}

	jsnCfg2 := &RadiusAgentJsonCfg{
		Sessions_conns: &[]string{str},
		Request_processors: &[]*ReqProcessorJsnCfg{{
			ID:       &str2,
			Filters:  &[]string{str2},
			Tenant:   &str2,
			Timezone: &str2,
			Flags:    &[]string{str2},
			Request_fields: &[]*FcTemplateJsonCfg{{
				Tag:          &str2,
				Type:         &str2,
				Path:         &str2,
				Attribute_id: &str2,
				Filters:      &[]string{"test)"},
			}},
			Reply_fields: &[]*FcTemplateJsonCfg{{
				Tag:          &str2,
				Type:         &str2,
				Path:         &str2,
				Attribute_id: &str2,
				Filters:      &[]string{"test)"},
			}},
		}},
	}

	err = self.loadFromJsonCfg(jsnCfg2, "")
	if err != nil {
		t.Error(err)
	}
}

func TestRadiusAgentCfgAsMapInterface2(t *testing.T) {
	str := "test"
	ra := &RadiusAgentCfg{
		Enabled:            true,
		ListenNet:          str,
		ListenAuth:         str,
		ListenAcct:         str,
		ClientSecrets:      map[string]string{str: str},
		ClientDictionaries: map[string]string{str: str},
		SessionSConns:      []string{str},
		RequestProcessors: []*RequestProcessor{
			{
				ID: str,
				Tenant: RSRParsers{{
					Rules:           "t",
					AllFiltersMatch: true,
				},
					{
						Rules:           "e",
						AllFiltersMatch: true,
					},
					{
						Rules:           "s",
						AllFiltersMatch: true,
					},
					{
						Rules:           "t",
						AllFiltersMatch: true,
					}},
				Filters:  []string{str},
				Flags:    utils.FlagsWithParams{str: {}},
				Timezone: str,
				RequestFields: []*FCTemplate{
					{
						AttributeID: str,
						Tag:         str,
						Type:        str,
						Path:        str,
						Filters:     []string{str},
					},
				},
				ReplyFields: []*FCTemplate{
					{
						AttributeID: str,
						Tag:         str,
						Type:        str,
						Path:        str,
						Filters:     []string{str},
					},
				},
			},
		},
	}

	exp := map[string]any{
		utils.EnabledCfg:            ra.Enabled,
		utils.ListenNetCfg:          ra.ListenNet,
		utils.ListenAuthCfg:         ra.ListenAuth,
		utils.ListenAcctCfg:         ra.ListenAcct,
		utils.ClientSecretsCfg:      map[string]any{str: str},
		utils.ClientDictionariesCfg: map[string]any{str: str},
		utils.SessionSConnsCfg:      []string{str},
		utils.RequestProcessorsCfg:  []map[string]any{ra.RequestProcessors[0].AsMapInterface("")},
	}

	rcv := ra.AsMapInterface("")

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected %s: \nreceived %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
