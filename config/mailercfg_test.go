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
)

func TestMailerCfgloadFromJsonCfg(t *testing.T) {
	var mailcfg, expected MailerCfg
	if err := mailcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mailcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, mailcfg)
	}
	if err := mailcfg.loadFromJsonCfg(new(MailerJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(mailcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, mailcfg)
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
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnMailCfg, err := jsnCfg.MailerJsonCfg(); err != nil {
		t.Error(err)
	} else if err = mailcfg.loadFromJsonCfg(jsnMailCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, mailcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, mailcfg)
	}
}
