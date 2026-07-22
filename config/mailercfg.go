// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

// MailerCfg stores Mailer config section
type MailerCfg struct {
	MailerServer   string // The server to use when sending emails out
	MailerAuthUser string // Authenticate to email server using this user
	MailerAuthPass string // Authenticate to email server with this password
	MailerFromAddr string // From address used when sending emails out
}

// loadFromJSONCfg loads Database config from JsonCfg
func (mailcfg *MailerCfg) loadFromJSONCfg(jsnMailerCfg *MailerJsonCfg) (err error) {
	if jsnMailerCfg == nil {
		return nil
	}
	if jsnMailerCfg.Server != nil {
		mailcfg.MailerServer = *jsnMailerCfg.Server
	}
	if jsnMailerCfg.Auth_user != nil {
		mailcfg.MailerAuthUser = *jsnMailerCfg.Auth_user
	}
	if jsnMailerCfg.Auth_password != nil {
		mailcfg.MailerAuthPass = *jsnMailerCfg.Auth_password
	}
	if jsnMailerCfg.From_address != nil {
		mailcfg.MailerFromAddr = *jsnMailerCfg.From_address
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (mailcfg *MailerCfg) AsMapInterface() (initialMP map[string]any) {
	return map[string]any{
		utils.MailerServerCfg:   mailcfg.MailerServer,
		utils.MailerAuthUserCfg: mailcfg.MailerAuthUser,
		utils.MailerAuthPassCfg: mailcfg.MailerAuthPass,
		utils.MailerFromAddrCfg: mailcfg.MailerFromAddr,
	}
}

// Clone returns a deep copy of MailerCfg
func (mailcfg *MailerCfg) Clone() *MailerCfg {
	if mailcfg == nil {
		return nil
	}
	return &MailerCfg{
		MailerServer:   mailcfg.MailerServer,
		MailerAuthUser: mailcfg.MailerAuthUser,
		MailerAuthPass: mailcfg.MailerAuthPass,
		MailerFromAddr: mailcfg.MailerFromAddr,
	}
}
