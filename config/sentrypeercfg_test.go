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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestSentryPeerCfgloadFromJsonCfg(t *testing.T) {
	var sp, expected SentryPeerCfg
	if err := sp.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sp, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, sp)
	}

	cfgJSONStr := `{
		"sentrypeer":{
			"client_id":"SecretID",
			"client_secret":"SecretKey",
			"token_url":"https://authz.sentrypeer.com/oauth/token",
			"ips_url":"https://sentrypeer.com/api/ip-addresses",
			"numbers_url":"https://sentrypeer.com/api/phone-numbers",
			"audience":"https://sentrypeer.com/api",
			"grant_type":"client_credentials",
		}
	}`
	expected = SentryPeerCfg{
		ClientID:     "SecretID",
		ClientSecret: "SecretKey",
		TokenUrl:     "https://authz.sentrypeer.com/oauth/token",
		IpsUrl:       "https://sentrypeer.com/api/ip-addresses",
		NumbersUrl:   "https://sentrypeer.com/api/phone-numbers",
		Audience:     "https://sentrypeer.com/api",
		GrantType:    "client_credentials",
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = sp.Load(context.Background(), jsnCfg, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, sp) {
		t.Errorf("Expected: %+v ,received: %+v", expected, sp)
	}
}

func TestSentryPeerCfgAsMapInterface(t *testing.T) {
	var sp SentryPeerCfg
	cfgJSONStr := `{
		"sentrypeer":{
			"client_id":"SecretID",
			"client_secret":"Secretkey",
			"token_url":"https://authz.sentrypeer.com/oauth/token",
			"ips_url":"https://sentrypeer.com/api/ip-addresses",
			"numbers_url":"https://sentrypeer.com/api/phone-numbers",
			"audience":"https://sentrypeer.com/api",
			"grant_type":"client_credentials",
		}
	}`
	eMap := map[string]any{
		"client_id":     "SecretID",
		"client_secret": "Secretkey",
		"token_url":     "https://authz.sentrypeer.com/oauth/token",
		"ips_url":       "https://sentrypeer.com/api/ip-addresses",
		"numbers_url":   "https://sentrypeer.com/api/phone-numbers",
		"audience":      "https://sentrypeer.com/api",
		"grant_type":    "client_credentials",
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err = sp.Load(context.Background(), jsnCfg, nil); err != nil {
		t.Error(err)
	} else if rcv := sp.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v ,received: %+v", eMap, rcv)
	}
}

/*func TestSentryPeerCfgClone(t *testing.T) {
	sentryP := SentryPeerCfg{
		IpsUrl:     "https://sentrypeer.com/api/ip-addresses",
		NumbersUrl: "https://sentrypeer.com/api/phone-numbers",
	}
	rcv := sentryP.Clone()
	if !reflect.DeepEqual(sentryP, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sentryP), utils.ToJSON(rcv))
	}
	if rcv.IpsUrl = ""; sentryP.IpsUrl != "https://sentrypeer.com/api/ip-addresses" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}*/

func TestDiffSentryPeerJsonCfg(t *testing.T) {
	var d *SentryPeerJsonCfg

	v1 := &SentryPeerCfg{
		ClientID:     "SecretID1",
		ClientSecret: "SecretKey1",
	}

	v2 := &SentryPeerCfg{
		ClientID:     "SecretID2",
		ClientSecret: "SecretKey2",
	}

	expected := &SentryPeerJsonCfg{
		Client_id:     utils.StringPointer("SecretID2"),
		Client_secret: utils.StringPointer("SecretKey2"),
	}

	rcv := diffSentryPeerJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	v2 = v1
	expected2 := &SentryPeerJsonCfg{}
	rcv = diffSentryPeerJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestSentryPeerCloneSection(t *testing.T) {
	apbCfg := &SentryPeerCfg{
		ClientID:     "SecretID1",
		ClientSecret: "SecretKey1",
	}

	exp := &SentryPeerCfg{
		ClientID:     "SecretID1",
		ClientSecret: "SecretKey1",
	}
	rcv := apbCfg.CloneSection()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestSentryPeerdiffSentryPeerJsonCfg(t *testing.T) {
	str := "test"
	str2 := "test2"
	d := &SentryPeerJsonCfg{}
	v2 := &SentryPeerCfg{
		ClientID:     str,
		ClientSecret: str,
		TokenUrl:     str,
		IpsUrl:       str,
		NumbersUrl:   str,
		Audience:     str,
		GrantType:    str,
	}
	v1 := &SentryPeerCfg{
		ClientID:     str2,
		ClientSecret: str2,
		TokenUrl:     str2,
		IpsUrl:       str2,
		NumbersUrl:   str2,
		Audience:     str2,
		GrantType:    str2,
	}
	exp := &SentryPeerJsonCfg{
		Client_id:     &str,
		Client_secret: &str,
		Token_url:     &str,
		Ips_url:       &str,
		Numbers_url:   &str,
		Audience:      &str,
		Grant_type:    &str,
	}

	rcv := diffSentryPeerJsonCfg(d, v1, v2)

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
