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

package v1

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewConfigSv1(cfg *config.CGRConfig) *ConfigSv1 {
	return &ConfigSv1{cfg: cfg}
}

// Exports RPC from ConfigSv1
type ConfigSv1 struct {
	cfg *config.CGRConfig
}

func (cSv1 *ConfigSv1) GetJSONSection(section *config.StringWithArgDispatcher, reply *map[string]interface{}) (err error) {
	return cSv1.cfg.V1GetConfigSection(section, reply)
}

func (cSv1 *ConfigSv1) ReloadConfig(args *config.ConfigReloadWithArgDispatcher, reply *string) (err error) {
	return cSv1.cfg.V1ReloadConfig(args, reply)
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (cSv1 *ConfigSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(cSv1, serviceMethod, args, reply)
}
