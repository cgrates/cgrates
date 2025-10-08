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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherHCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &RegistrarCJsonCfgs{
		RPC: &RegistrarCJsonCfg{
			Registrars_conns: &[]string{"*conn1", "*conn2"},
			Hosts: []*RemoteHostJsonWithTenant{
				{

					Tenant: utils.StringPointer(utils.MetaDefault),
					RemoteHostJson: &RemoteHostJson{
						Id:        utils.StringPointer("Host1"),
						Transport: utils.StringPointer(utils.MetaJSON),
					},
				},
				{
					Tenant: utils.StringPointer(utils.MetaDefault),
					RemoteHostJson: &RemoteHostJson{
						Id:        utils.StringPointer("Host2"),
						Transport: utils.StringPointer(utils.MetaGOB),
					},
				},
				{
					Tenant: utils.StringPointer("cgrates.net"),
					RemoteHostJson: &RemoteHostJson{
						Id:        utils.StringPointer("Host1"),
						Transport: utils.StringPointer(utils.MetaJSON),
						Tls:       utils.BoolPointer(true),
					},
				},
				{
					Tenant: utils.StringPointer("cgrates.net"),
					RemoteHostJson: &RemoteHostJson{
						Id:        utils.StringPointer("Host2"),
						Transport: utils.StringPointer(utils.MetaGOB),
						Tls:       utils.BoolPointer(true),
					},
				},
			},
			Refresh_interval: utils.StringPointer("5"),
		},
	}
	expected := &RegistrarCCfgs{
		RPC: &RegistrarCCfg{
			RegistrarSConns: []string{"*conn1", "*conn2"},
			Hosts: map[string][]*RemoteHost{
				utils.MetaDefault: {
					{
						ID:        "Host1",
						Transport: utils.MetaJSON,
					},
					{
						ID:        "Host2",
						Transport: utils.MetaGOB,
					},
				},
				"cgrates.net": {
					{
						ID:        "Host1",
						Transport: utils.MetaJSON,
						TLS:       true,
					},
					{
						ID:        "Host2",
						Transport: utils.MetaGOB,
						TLS:       true,
					},
				},
			},
			RefreshInterval: 5,
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.registrarCCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.registrarCCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.registrarCCfg))
	}

	jsonCfg.RPC.Hosts = append(jsonCfg.RPC.Hosts, &RemoteHostJsonWithTenant{
		Tenant: utils.StringPointer(""),
		RemoteHostJson: &RemoteHostJson{
			Id:        utils.StringPointer("Host1"),
			Transport: utils.StringPointer(utils.MetaJSON),
		},
	})
	if err := jsnCfg.registrarCCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}

	jsonCfg = nil
	if err := jsnCfg.registrarCCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
}

func TestDiffRegistrarCJsonCfg(t *testing.T) {
	var d *RegistrarCJsonCfg

	v1 := &RegistrarCCfg{
		RegistrarSConns: []string{"*localhost"},
		Hosts: map[string][]*RemoteHost{
			"HOST_1": {
				{
					ID:        "host1_ID",
					Address:   "127.0.0.1:8080",
					Transport: "tcp",
					TLS:       false,
				},
			},
		},
		RefreshInterval: 2 * time.Second,
	}

	v2 := &RegistrarCCfg{
		RegistrarSConns: []string{"*birpc"},
		Hosts: map[string][]*RemoteHost{
			"HOST_1": {
				{
					ID:        "host2_ID",
					Address:   "0.0.0.0:8080",
					Transport: "udp",
					TLS:       true,
				},
			},
		},
		RefreshInterval: 4 * time.Second,
	}

	expected := &RegistrarCJsonCfg{
		Registrars_conns: &[]string{"*birpc"},
		Hosts: []*RemoteHostJsonWithTenant{
			{
				Tenant: utils.StringPointer("HOST_1"),
				RemoteHostJson: &RemoteHostJson{
					Id:        utils.StringPointer("host2_ID"),
					Address:   utils.StringPointer("0.0.0.0:8080"),
					Transport: utils.StringPointer("udp"),
					Tls:       utils.BoolPointer(true),
				},
			},
		},
		Refresh_interval: utils.StringPointer("4s"),
	}

	rcv := diffRegistrarCJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &RegistrarCJsonCfg{
		Hosts: []*RemoteHostJsonWithTenant{
			{
				Tenant: utils.StringPointer("HOST_1"),
				RemoteHostJson: &RemoteHostJson{
					Id:        utils.StringPointer("host2_ID"),
					Address:   utils.StringPointer("0.0.0.0:8080"),
					Transport: utils.StringPointer("udp"),
					Tls:       utils.BoolPointer(true),
				},
			},
		},
	}
	rcv = diffRegistrarCJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
