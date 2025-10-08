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

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
// if the section is empty or *all all config is returned
func (cSv1 *ConfigSv1) GetConfig(ctx *context.Context, section *config.SectionWithAPIOpts, reply *map[string]any) (err error) {
	return cSv1.cfg.V1GetConfig(ctx, section, reply)
}

// ReloadConfig reloads the configuration at path
// if the section is empty or *all all config is reload
func (cSv1 *ConfigSv1) ReloadConfig(ctx *context.Context, args *config.ReloadArgs, reply *string) (err error) {
	return cSv1.cfg.V1ReloadConfig(ctx, args, reply)
}

// SetConfig sets the config received and reloads the received section
func (cSv1 *ConfigSv1) SetConfig(ctx *context.Context, args *config.SetConfigArgs, reply *string) (err error) {
	return cSv1.cfg.V1SetConfig(ctx, args, reply)
}

// SetConfigFromJSON sets the config received as a string but reloads all config sections
func (cSv1 *ConfigSv1) SetConfigFromJSON(ctx *context.Context, args *config.SetConfigFromJSONArgs, reply *string) (err error) {
	return cSv1.cfg.V1SetConfigFromJSON(ctx, args, reply)
}

// GetConfigAsJSON will retrieve from CGRConfig a section as a string
// if the section is empty or *all all config is returned
func (cSv1 *ConfigSv1) GetConfigAsJSON(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	return cSv1.cfg.V1GetConfigAsJSON(ctx, args, reply)
}

// StoreCfgInDB will store the config section in the dataDB
// if the section is empty or *all all config is stored in DB
func (cSv1 *ConfigSv1) StoreCfgInDB(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	return cSv1.cfg.V1StoreCfgInDB(ctx, args, reply)
}

// DumpConfigDB will dump all of configdb from memory to a file
func (cSv1 *ConfigSv1) DumpConfigDB(ctx *context.Context, ignr *string, reply *string) (err error) {
	if err = cSv1.cfg.ConfigDB().DumpConfigDB(); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// Will rewrite every dump file of ConfigDB
func (cSv1 *ConfigSv1) RewriteConfigDB(ctx *context.Context, ignr *string, reply *string) (err error) {
	if err = cSv1.cfg.ConfigDB().RewriteConfigDB(); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// BackupConfigDB will momentarely stop any dumping and rewriting in configDB, until dump folder is backed up in folder path backupFolderPath. Making zip true will create a zip file in the path instead
func (cSv1 *ConfigSv1) BackupConfigDB(ctx *context.Context, params DumpBackupParams, reply *string) (err error) {
	if err = cSv1.cfg.ConfigDB().BackupConfigDB(params.BackupFolderPath, params.Zip); err != nil {
		return
	}
	*reply = utils.OK
	return
}
