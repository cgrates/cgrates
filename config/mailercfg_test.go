// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.mailerCfg.loadFromJSONCfg(cfgJSON); err != nil {
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
	eMap := map[string]any{
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
	eMap := map[string]any{
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

	cS = nil
	rcv = cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
}
