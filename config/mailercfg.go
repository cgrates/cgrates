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

// AsMapInterface returns the config as a map[string]interface{}
func (mailcfg *MailerCfg) AsMapInterface() (initialMP map[string]interface{}) {
	return map[string]interface{}{
		utils.MailerServerCfg:   mailcfg.MailerServer,
		utils.MailerAuthUserCfg: mailcfg.MailerAuthUser,
		utils.MailerAuthPassCfg: mailcfg.MailerAuthPass,
		utils.MailerFromAddrCfg: mailcfg.MailerFromAddr,
	}
}

// Clone returns a deep copy of MailerCfg
func (mailcfg MailerCfg) Clone() *MailerCfg {
	return &MailerCfg{
		MailerServer:   mailcfg.MailerServer,
		MailerAuthUser: mailcfg.MailerAuthUser,
		MailerAuthPass: mailcfg.MailerAuthPass,
		MailerFromAddr: mailcfg.MailerFromAddr,
	}
}
