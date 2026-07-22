// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSentryPeerCfgloadFromJsonCfg(t *testing.T) {
	var sp, expected SentryPeerCfg

	if err := sp.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sp, expected) {
		t.Errorf("Expected %+v,Received %+v", expected, sp)
	}

	cfgJsonStr := ` 
{
	"sentrypeer":{
		"client_id":"sd5h7rhyrtudjj",
		"client_secret":"w9buwnbw9ewbegsaovnmn9wr",
		"token_url":"https://authz.sentrypeer.com/oauth/token",
		"ips_url":"https://sentrypeer.com/api/ip-addresses",
		"numbers_url":"https://sentrypeer.com/api/phone-numbers",
		"audience":"https://sentrypeer.com/api",
		"grant_type":"client_credentials",
	 }
	}
	`
	expected = SentryPeerCfg{
		ClientID:     "sd5h7rhyrtudjj",
		ClientSecret: "w9buwnbw9ewbegsaovnmn9wr",
		TokenUrl:     "https://authz.sentrypeer.com/oauth/token",
		IpsUrl:       "https://sentrypeer.com/api/ip-addresses",
		NumbersUrl:   "https://sentrypeer.com/api/phone-numbers",
		Audience:     "https://sentrypeer.com/api",
		GrantType:    "client_credentials",
	}
	cfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJsonStr))
	if err != nil {
		t.Error(err)
	}
	sentryPeerJson, err := cfgJson.SentryPeerJson()
	if err != nil {
		t.Error(err)
	}
	if err := sp.loadFromJSONCfg(sentryPeerJson); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sp, expected) {
		t.Errorf("Expected %+v,Received %+v", expected, sp)
	}
}

func TestSentryPeerAsMapInterface(t *testing.T) {
	var sp SentryPeerCfg
	cfgJsonStr := `
{
	"sentrypeer":{
		"client_id":     "sd5h7rhyrtudjj",
		"client_secret": "w9buwnbw9ewbegsaovnmn9wr",
		"token_url":     "https://authz.sentrypeer.com/oauth/token",
		"ips_url":       "https://sentrypeer.com/api/ip-addresses",
		"numbers_url":   "https://sentrypeer.com/api/phone-numbers",
		"audience":      "https://sentrypeer.com/api",
		"grant_type":    "client_credentials",
	 }
}
`
	expected := map[string]any{"ClientID": "sd5h7rhyrtudjj",
		"ClientSecret": "w9buwnbw9ewbegsaovnmn9wr",
		"TokenURL":     "https://authz.sentrypeer.com/oauth/token",
		"IpUrl":        "https://sentrypeer.com/api/ip-addresses",
		"NumberUrl":    "https://sentrypeer.com/api/phone-numbers",
		"Audience":     "https://sentrypeer.com/api",
		"GrantType":    "client_credentials",
	}

	cfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJsonStr))
	if err != nil {
		t.Error(err)
	}

	sentryPeerJson, err := cfgJson.SentryPeerJson()
	if err != nil {
		t.Error(err)
	}
	if err := sp.loadFromJSONCfg(sentryPeerJson); err != nil {
		t.Error(err)
	} else if rcv := sp.AsMapInterface(); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v,Received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestSentryPeerCfgClone(t *testing.T) {

	sp := &SentryPeerCfg{
		Audience:  "https://sentrypeer.com/api",
		GrantType: "client_credentials",
	}
	rcv := sp.Clone()
	if !reflect.DeepEqual(rcv, sp) {
		t.Errorf("Expected %+v,Received %+v", rcv, sp)
	}
	if rcv.GrantType = ""; sp.GrantType != "client_credentials" {
		t.Error("Expected clone to not modify the cloned")
	}
}
