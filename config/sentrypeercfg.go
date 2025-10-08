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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type SentryPeerCfg struct {
	ClientID     string
	ClientSecret string
	TokenUrl     string
	IpsUrl       string
	NumbersUrl   string
	Audience     string
	GrantType    string
}

func (sentrypeer *SentryPeerCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnSentryPeerCfg := new(SentryPeerJsonCfg)
	if err = jsnCfg.GetSection(ctx, SentryPeerJSON, jsnSentryPeerCfg); err != nil {
		return
	}
	return sentrypeer.loadFromJSONCfg(jsnSentryPeerCfg)
}

func (sentrypeer *SentryPeerCfg) loadFromJSONCfg(jsnCfg *SentryPeerJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Client_id != nil {
		sentrypeer.ClientID = *jsnCfg.Client_id
	}
	if jsnCfg.Client_secret != nil {
		sentrypeer.ClientSecret = *jsnCfg.Client_secret
	}
	if jsnCfg.Token_url != nil {
		sentrypeer.TokenUrl = *jsnCfg.Token_url
	}
	if jsnCfg.Ips_url != nil {
		sentrypeer.IpsUrl = *jsnCfg.Ips_url
	}
	if jsnCfg.Numbers_url != nil {
		sentrypeer.NumbersUrl = *jsnCfg.Numbers_url
	}
	if jsnCfg.Audience != nil {
		sentrypeer.Audience = *jsnCfg.Audience
	}
	if jsnCfg.Grant_type != nil {
		sentrypeer.GrantType = *jsnCfg.Grant_type
	}
	return
}
func (sentrypeer SentryPeerCfg) AsMapInterface() any {
	return map[string]any{
		utils.ClientIDCfg:     sentrypeer.ClientID,
		utils.ClientSecretCfg: sentrypeer.ClientSecret,
		utils.TokenUrlCfg:     sentrypeer.TokenUrl,
		utils.IpsUrlCfg:       sentrypeer.IpsUrl,
		utils.NumbersUrlCfg:   sentrypeer.NumbersUrl,
		utils.AudienceCfg:     sentrypeer.Audience,
		utils.GrantTypeCfg:    sentrypeer.GrantType,
	}
}

func (SentryPeerCfg) SName() string                    { return SentryPeerJSON }
func (sentrypeer SentryPeerCfg) CloneSection() Section { return sentrypeer.Clone() }

func (sentrypeer SentryPeerCfg) Clone() (cln *SentryPeerCfg) {

	return &SentryPeerCfg{
		ClientID:     sentrypeer.ClientID,
		ClientSecret: sentrypeer.ClientSecret,
		TokenUrl:     sentrypeer.TokenUrl,
		IpsUrl:       sentrypeer.IpsUrl,
		NumbersUrl:   sentrypeer.NumbersUrl,
		Audience:     sentrypeer.Audience,
		GrantType:    sentrypeer.GrantType,
	}
}

type SentryPeerJsonCfg struct {
	Client_id     *string
	Client_secret *string
	Token_url     *string
	Ips_url       *string
	Numbers_url   *string
	Audience      *string
	Grant_type    *string
}

func diffSentryPeerJsonCfg(d *SentryPeerJsonCfg, v1, v2 *SentryPeerCfg) *SentryPeerJsonCfg {
	if d == nil {
		d = new(SentryPeerJsonCfg)
	}
	if v1.ClientID != v2.ClientID {
		d.Client_id = utils.StringPointer(v2.ClientID)
	}
	if v1.ClientSecret != v2.ClientSecret {
		d.Client_secret = utils.StringPointer(v2.ClientSecret)
	}
	if v1.TokenUrl != v2.TokenUrl {
		d.Token_url = utils.StringPointer(v2.TokenUrl)
	}
	if v1.IpsUrl != v2.IpsUrl {
		d.Ips_url = utils.StringPointer(v2.IpsUrl)
	}
	if v1.NumbersUrl != v2.NumbersUrl {
		d.Numbers_url = utils.StringPointer(v2.NumbersUrl)
	}
	if v1.Audience != v2.Audience {
		d.Audience = utils.StringPointer(v2.Audience)
	}
	if v1.GrantType != v2.GrantType {
		d.Grant_type = utils.StringPointer(v2.GrantType)
	}

	return d
}
