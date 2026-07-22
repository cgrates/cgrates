// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewGlobalVarS .
func NewGlobalVarS(cfg *config.CGRConfig,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &GlobalVarS{
		cfg:    cfg,
		srvDep: srvDep,
	}
}

// GlobalVarS implements Agent interface
type GlobalVarS struct {
	cfg    *config.CGRConfig
	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (gv *GlobalVarS) Start() (err error) {
	engine.SetRoundingDecimals(gv.cfg.GeneralCfg().RoundingDecimals)
	ees.InitFailedPostCache(gv.cfg.EEsCfg().FailedPosts.TTL, gv.cfg.EEsCfg().FailedPosts.StaticTTL)
	engine.SetHTTPPstrTransport(gv.cfg.HTTPCfg().ClientOpts)
	return nil
}

// Reload handles the change of config
func (gv *GlobalVarS) Reload() (err error) {
	engine.SetHTTPPstrTransport(gv.cfg.HTTPCfg().ClientOpts)
	return nil
}

// Shutdown stops the service
func (gv *GlobalVarS) Shutdown() (err error) {
	return
}

// IsRunning returns if the service is running
func (gv *GlobalVarS) IsRunning() bool {
	return true
}

// ServiceName returns the service name
func (gv *GlobalVarS) ServiceName() string {
	return utils.GlobalVarS
}

// ShouldRun returns if the service should be running
func (gv *GlobalVarS) ShouldRun() bool {
	return true
}
