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

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
)

// NewConfigSv1 returns a new ConfigSv1
func NewConfigSv1(cfg *config.CGRConfig) *ConfigSv1 {
	return &ConfigSv1{
		cfg: cfg,
	}
}

// ConfigSv1 exports RPC for config
type ConfigSv1 struct {
	ping
	cfg *config.CGRConfig
}

// GetConfig will return the CGRConfig section as a map
// if the section is empty or *all all config is stored in DB
func (cSv1 *ConfigSv1) GetConfig(ctx *context.Context, section *config.SectionWithAPIOpts, reply *map[string]interface{}) (err error) {
	return cSv1.cfg.V1GetConfig(ctx, section, reply)
}

/*
// ReloadConfig reloads the configuration
func (cSv1 *ConfigSv1) ReloadConfig(args *config.ReloadArgs, reply *string) (err error) {
	return cSv1.cfg.V1ReloadConfig(args, reply)
}

// SetConfig reloads the sections of config
func (cSv1 *ConfigSv1) SetConfig(args *config.SetConfigArgs, reply *string) (err error) {
	return cSv1.cfg.V1SetConfig(args, reply)
}

// SetConfigFromJSON reloads the sections of config
func (cSv1 *ConfigSv1) SetConfigFromJSON(args *config.SetConfigFromJSONArgs, reply *string) (err error) {
	return cSv1.cfg.V1SetConfigFromJSON(args, reply)
}

// GetConfigAsJSON will retrieve from CGRConfig a section
func (cSv1 *ConfigSv1) GetConfigAsJSON(args *config.SectionWithAPIOpts, reply *string) (err error) {
	return cSv1.cfg.V1GetConfigAsJSON(args, reply)
}
*/

// StoreCfgInDB will store the config section in the dataDB
// if the section is empty or *all all config is stored in DB
func (cSv1 *ConfigSv1) StoreCfgInDB(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	return cSv1.cfg.V1StoreCfgInDB(ctx, args, reply)
}
