// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewConfigSv1 returns a new ConfigSv1
func NewConfigSv1(cfg *config.CGRConfig) *ConfigSv1 {
	return &ConfigSv1{cfg: cfg}
}

// ConfigSv1 exports RPC for config
type ConfigSv1 struct {
	cfg *config.CGRConfig
}

// GetJSONSection will retrieve from CGRConfig a section
func (cSv1 *ConfigSv1) GetJSONSection(section *config.StringWithArgDispatcher, reply *map[string]any) (err error) {
	return cSv1.cfg.V1GetConfigSection(section, reply)
}

// ReloadConfigFromPath reloads the configuration
func (cSv1 *ConfigSv1) ReloadConfigFromPath(args *config.ConfigReloadWithArgDispatcher, reply *string) (err error) {
	return cSv1.cfg.V1ReloadConfigFromPath(args, reply)
}

// ReloadConfigFromJSON reloads the sections of configz
func (cSv1 *ConfigSv1) ReloadConfigFromJSON(args *config.JSONReloadWithArgDispatcher, reply *string) (err error) {
	return cSv1.cfg.V1ReloadConfigFromJSON(args, reply)
}

// Call implements birpc.ClientConnector interface for internal RPC
func (cSv1 *ConfigSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(cSv1, serviceMethod, args, reply)
}
