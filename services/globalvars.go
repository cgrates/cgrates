// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewGlobalVarS .
func NewGlobalVarS(cfg *config.CGRConfig) *GlobalVarS {
	return &GlobalVarS{
		cfg: cfg,
	}
}

// GlobalVarS implements Agent interface
type GlobalVarS struct {
	cfg *config.CGRConfig
}

// Start should handle the sercive start
func (gv *GlobalVarS) Start(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	utils.RoutesDefaultRatio = gv.cfg.RouteSCfg().DefaultRatio
	utils.DecimalContext.MaxScale = gv.cfg.GeneralCfg().DecimalMaxScale
	utils.DecimalContext.MinScale = gv.cfg.GeneralCfg().DecimalMinScale
	utils.DecimalContext.Precision = gv.cfg.GeneralCfg().DecimalPrecision
	utils.DecimalContext.RoundingMode = gv.cfg.GeneralCfg().DecimalRoundingMode
	return nil
}

// Reload handles the change of config
func (gv *GlobalVarS) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	utils.RoutesDefaultRatio = gv.cfg.RouteSCfg().DefaultRatio
	utils.DecimalContext.MaxScale = gv.cfg.GeneralCfg().DecimalMaxScale
	utils.DecimalContext.MinScale = gv.cfg.GeneralCfg().DecimalMinScale
	utils.DecimalContext.Precision = gv.cfg.GeneralCfg().DecimalPrecision
	utils.DecimalContext.RoundingMode = gv.cfg.GeneralCfg().DecimalRoundingMode
	return nil
}

// Shutdown stops the service
func (gv *GlobalVarS) Shutdown(_ *servmanager.Registry) error {
	return nil
}

// ServiceName returns the service name
func (gv *GlobalVarS) ServiceName() string {
	return utils.GlobalVarS
}

// ShouldRun returns if the service should be running
func (gv *GlobalVarS) ShouldRun() bool {
	return true
}
