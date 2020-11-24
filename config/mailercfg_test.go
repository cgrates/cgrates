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

func TestMailerCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &MailerJsonCfg{
		Server:        utils.StringPointer("localhost"),
		Auth_user:     utils.StringPointer("cgrates"),
		Auth_password: utils.StringPointer("CGRateS.org"),
		From_address:  utils.StringPointer("cgr-mailer@localhost.localdomain"),
	}
	expected := &MailerCfg{
		MailerServer:   "localhost",
		MailerAuthUser: "cgrates",
		MailerAuthPass: "CGRateS.org",
		MailerFromAddr: "cgr-mailer@localhost.localdomain",
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.mailerCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.mailerCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.mailerCfg))
	}
}

func TestMailerCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"mailer": {
		"server": "",
		"auth_user": "",
		"auth_password": "",
		"from_address": "",
		},
}`
	eMap := map[string]interface{}{
		utils.MailerServerCfg:   "",
		utils.MailerAuthUserCfg: "",
		utils.MailerAuthPassCfg: "",
		utils.MailerFromAddrCfg: "",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.mailerCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestMailerCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"mailer": {},
}`
	eMap := map[string]interface{}{
		utils.MailerServerCfg:   "localhost",
		utils.MailerAuthUserCfg: "cgrates",
		utils.MailerAuthPassCfg: "CGRateS.org",
		utils.MailerFromAddrCfg: "cgr-mailer@localhost.localdomain",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.mailerCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestMailerCfgClone(t *testing.T) {
	cS := &MailerCfg{
		MailerServer:   "localhost",
		MailerAuthUser: "cgrates",
		MailerAuthPass: "CGRateS.org",
		MailerFromAddr: "cgr-mailer@localhost.localdomain",
	}
	rcv := cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.MailerServer = ""; cS.MailerServer != "localhost" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
