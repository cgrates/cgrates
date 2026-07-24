// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestMailerCfgloadFromJsonCfg(t *testing.T) {
	var mailcfg, expected MailerCfg
	if err := mailcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mailcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, mailcfg)
	}
	if err := mailcfg.loadFromJsonCfg(new(MailerJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mailcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, mailcfg)
	}
	cfgJSONStr := `{
"mailer": {
	"server": "localhost",								// the server to use when sending emails out
	"auth_user": "cgrates",								// authenticate to email server using this user
	"auth_password": "CGRateS.org",						// authenticate to email server with this password
	"from_address": "cgr-mailer@localhost.localdomain"	// from address used when sending emails out
	},
}`
	expected = MailerCfg{
		MailerServer:   "localhost",
		MailerAuthUser: "cgrates",
		MailerAuthPass: "CGRateS.org",
		MailerFromAddr: "cgr-mailer@localhost.localdomain",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnMailCfg, err := jsnCfg.MailerJsonCfg(); err != nil {
		t.Error(err)
	} else if err = mailcfg.loadFromJsonCfg(jsnMailCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, mailcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, mailcfg)
	}
}

func TestMailerCfgAsMapInterface(t *testing.T) {
	var mailcfg MailerCfg
	cfgJSONStr := `{
	"mailer": {
		"server": "",
		"auth_user": "",
		"auth_password": "",
		"from_address": "",
		},
}`
	eMap := map[string]any{
		"server":        "",
		"auth_user":     "",
		"auth_password": "",
		"from_address":  "",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnMailCfg, err := jsnCfg.MailerJsonCfg(); err != nil {
		t.Error(err)
	} else if err = mailcfg.loadFromJsonCfg(jsnMailCfg); err != nil {
		t.Error(err)
	} else if rcv := mailcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"mailer": {
			"server": "localhost",
			"auth_user": "cgrates",
			"auth_password": "CGRateS.org",
			"from_address": "cgr-mailer@localhost.localdomain",
			},
	}`
	eMap = map[string]any{
		"server":        "localhost",
		"auth_user":     "cgrates",
		"auth_password": "CGRateS.org",
		"from_address":  "cgr-mailer@localhost.localdomain",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnMailCfg, err := jsnCfg.MailerJsonCfg(); err != nil {
		t.Error(err)
	} else if err = mailcfg.loadFromJsonCfg(jsnMailCfg); err != nil {
		t.Error(err)
	} else if rcv := mailcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
