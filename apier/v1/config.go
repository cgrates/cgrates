// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
)

// NewConfigSv1 returns a new ConfigSv1
func NewConfigSv1(cfg *config.CGRConfig) *ConfigSv1 {
	return &ConfigSv1{cfg: cfg}
}

// ConfigSv1 exports RPC for config
type ConfigSv1 struct {
	cfg *config.CGRConfig
}

// GetConfig will retrieve from CGRConfig a section
func (cSv1 *ConfigSv1) GetConfig(ctx *context.Context, section *config.SectionWithAPIOpts, reply *map[string]any) (err error) {
	return cSv1.cfg.V1GetConfig(ctx, section, reply)
}

// ReloadConfig reloads the configuration
func (cSv1 *ConfigSv1) ReloadConfig(ctx *context.Context, args *config.ReloadArgs, reply *string) (err error) {
	return cSv1.cfg.V1ReloadConfig(ctx, args, reply)
}

// SetConfig reloads the sections of config
func (cSv1 *ConfigSv1) SetConfig(ctx *context.Context, args *config.SetConfigArgs, reply *string) (err error) {
	return cSv1.cfg.V1SetConfig(ctx, args, reply)
}

// SetConfigFromJSON reloads the sections of config
func (cSv1 *ConfigSv1) SetConfigFromJSON(ctx *context.Context, args *config.SetConfigFromJSONArgs, reply *string) (err error) {
	return cSv1.cfg.V1SetConfigFromJSON(ctx, args, reply)
}

// GetConfigAsJSON will retrieve from CGRConfig a section
func (cSv1 *ConfigSv1) GetConfigAsJSON(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	return cSv1.cfg.V1GetConfigAsJSON(ctx, args, reply)
}
