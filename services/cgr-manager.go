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
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
)

// NewCGRManager .
func NewCGRManager(cfg *config.CGRConfig, cM *engine.ConnManager,
	iConnCh chan birpc.ClientConnector,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &CGRManager{
		cfg:     cfg,
		srvDep:  srvDep,
		cM:      cM,
		iConnCh: iConnCh,
	}
}

// CGRManager implements Agent interface
type CGRManager struct {
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup
	cM      *engine.ConnManager
	iConnCh chan birpc.ClientConnector
}

// Start should handle the sercive start
func (gv *CGRManager) Start(*context.Context, context.CancelFunc) error {
	return nil
}

// Reload handles the change of config
func (gv *CGRManager) Reload(*context.Context, context.CancelFunc) error {
	return nil
}

// Shutdown stops the service
func (gv *CGRManager) Shutdown() error {
	return nil
}

// IsRunning returns if the service is running
func (gv *CGRManager) IsRunning() bool {
	return true
}

// ServiceName returns the service name
func (gv *CGRManager) ServiceName() string {
	return "CGRManager"
}

// ShouldRun returns if the service should be running
func (gv *CGRManager) ShouldRun() bool {
	return true
}
