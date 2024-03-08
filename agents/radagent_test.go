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
package agents

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/radigo"
)

func TestRadAgentSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(RadiusAgent))
}

func TestNewRadiusAgentFailDict(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RadiusAgentCfg().ClientDictionaries = map[string][]string{
		"badpath": {"bad/path"},
	}
	exp := "stat bad/path: no such file or directory"
	if _, err := NewRadiusAgent(cfg, nil, nil); err == nil || err.Error() != exp {
		t.Errorf("Expected error <%v>, received <%v>", exp, err)
	}
}

func TestNewRadiusAgentOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	exp := &RadiusAgent{
		cgrCfg: cfg,
	}
	if rcv, err := NewRadiusAgent(cfg, nil, nil); err != nil {
		if err.Error() == "stat /usr/share/cgrates/radius/dict/: no such file or directory" {
			t.SkipNow() // skipping if running in gitactions
		}
		t.Error(err)
	} else if exp.RWMutex != rcv.RWMutex || exp.cgrCfg != rcv.cgrCfg || exp.connMgr != rcv.connMgr || exp.filterS != rcv.filterS || exp.dacCfg != rcv.dacCfg || exp.WaitGroup != rcv.WaitGroup {
		t.Errorf("Expected <%+v>,\nreceived\n<%+v>", exp, rcv)
	}
}

func TestNewRadiusDAClientCfgOK(t *testing.T) {

	radAgCfg := &config.RadiusAgentCfg{
		ClientDaAddresses: map[string]config.DAClientOpts{
			"udp://127.0.0.1:1813": {
				Transport: "udp",
				Host:      "127.0.0.1",
				Port:      1813,
			},
		},
	}
	dicts := &radigo.Dictionaries{}
	secrets := &radigo.Secrets{}
	exp := radiusDAClientCfg{}
	expDicts := make(map[string]*radigo.Dictionary, len(radAgCfg.ClientDaAddresses))
	expSecrets := make(map[string]string, len(radAgCfg.ClientDaAddresses))
	for client := range radAgCfg.ClientDaAddresses {
		expDicts[client] = dicts.GetInstance(client)
		exp.dicts = radigo.NewDictionaries(expDicts)
		expSecrets[client] = secrets.GetSecret(client)
		exp.secrets = radigo.NewSecrets(expSecrets)
	}

	if rcv := newRadiusDAClientCfg(dicts, secrets, radAgCfg); !reflect.DeepEqual(exp, rcv) {
		rcvStr := map[string]string{
			"dicts":   fmt.Sprint(rcv.dicts),
			"secrets": fmt.Sprint(rcv.secrets),
		}
		expStr := map[string]any{
			"dicts":   fmt.Sprint(exp.dicts),
			"secrets": fmt.Sprint(exp.secrets),
		}
		t.Errorf("Expected <%+v>,\nreceived\n<%+v>", expStr, rcvStr)
	}

}
