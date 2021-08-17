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

package services

import (
	"fmt"
	"net/http"
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
	ees.SetFailedPostCacheTTL(gv.cfg.GeneralCfg().FailedPostsTTL)
	return gv.initHTTPTransport()
}

// Reload handles the change of config
func (gv *GlobalVarS) Reload() (err error) {
	return gv.initHTTPTransport()
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

func (gv *GlobalVarS) initHTTPTransport() (err error) {
	var trsp *http.Transport
	if trsp, err = engine.NewHTTPTransport(gv.cfg.HTTPCfg().ClientOpts); err != nil {
		utils.Logger.Crit(fmt.Sprintf("Could not configure the http transport: %s exiting!", err))
		return
	}
	engine.SetHTTPPstrTransport(trsp)
	return
}
