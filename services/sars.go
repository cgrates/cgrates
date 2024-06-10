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
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type SaRService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (sa *SaRService) Start() error {
	if sa.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	sa.srvDep[utils.DataDB].Add(1)
	<-sa.cacheS.GetPrecacheChannel(utils.CacheStatFilterIndexes)

	filterS := <-sa.filterSChan
	sa.filterSChan <- filterS
	dbchan := sa.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.SaRS))
	srv, err := engine.NewService(v1.NewSArSv1())
	if err != nil {
		return err
	}
	if !sa.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			sa.server.RpcRegister(s)
		}
	}
	sa.connChan <- sa.anz.GetInternalCodec(srv, utils.StatS)
	return nil
}

// Reload handles the change of config
func (sa *SaRService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (sa *SaRService) Shutdown() (err error) {
	defer sa.srvDep[utils.DataDB].Done()
	sa.Lock()
	defer sa.Unlock()
	<-sa.connChan
	return
}

// IsRunning returns if the service is running
func (sa *SaRService) IsRunning() bool {
	sa.RLock()
	defer sa.RUnlock()
	return false
}

// ServiceName returns the service name
func (sa *SaRService) ServiceName() string {
	return utils.SaRS
}

// ShouldRun returns if the service should be running
func (sa *SaRService) ShouldRun() bool {
	return sa.cfg.SarSCfg().Enabled
}
