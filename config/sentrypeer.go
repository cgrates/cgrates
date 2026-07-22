// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

type SentryPeerCfg struct {
	ClientID     string
	ClientSecret string
	TokenUrl     string
	IpsUrl       string
	NumbersUrl   string
	Audience     string
	GrantType    string
}

func (sp *SentryPeerCfg) loadFromJSONCfg(jsnCfg *SentryPeerJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.TokenUrl != nil {
		sp.TokenUrl = *jsnCfg.TokenUrl
	}
	if jsnCfg.IpsUrl != nil {
		sp.IpsUrl = *jsnCfg.IpsUrl
	}
	if jsnCfg.NumbersUrl != nil {
		sp.NumbersUrl = *jsnCfg.NumbersUrl
	}
	if jsnCfg.ClientSecret != nil {
		sp.ClientSecret = *jsnCfg.ClientSecret
	}
	if jsnCfg.ClientID != nil {
		sp.ClientID = *jsnCfg.ClientID
	}
	if jsnCfg.Audience != nil {
		sp.Audience = *jsnCfg.Audience
	}
	if jsnCfg.GrantType != nil {
		sp.GrantType = *jsnCfg.GrantType
	}
	return
}

func (sp *SentryPeerCfg) AsMapInterface() map[string]any {
	return map[string]any{
		"TokenURL":     sp.TokenUrl,
		"ClientSecret": sp.ClientSecret,
		"ClientID":     sp.ClientID,
		"IpUrl":        sp.IpsUrl,
		"NumberUrl":    sp.NumbersUrl,
		"Audience":     sp.Audience,
		"GrantType":    sp.GrantType,
	}
}

func (sp *SentryPeerCfg) Clone() (cln *SentryPeerCfg) {
	cln = &SentryPeerCfg{
		TokenUrl:     sp.TokenUrl,
		ClientSecret: sp.ClientSecret,
		ClientID:     sp.ClientID,
		IpsUrl:       sp.IpsUrl,
		NumbersUrl:   sp.NumbersUrl,
		Audience:     sp.Audience,
		GrantType:    sp.GrantType,
	}
	return
}
