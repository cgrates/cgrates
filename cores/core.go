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

package cores

import (
	"fmt"
	"runtime"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreService(cfg *config.CGRConfig, caps *Caps, stopChan chan struct{}) *CoreService {
	var st *CapsStats
	if caps.IsLimited() && cfg.CoreSCfg().CapsStatsInterval != 0 {
		st = NewCapsStats(cfg.CoreSCfg().CapsStatsInterval, caps, stopChan)
	}
	return &CoreService{
		cfg:       cfg,
		capsStats: st,
	}
}

type CoreService struct {
	cfg       *config.CGRConfig
	capsStats *CapsStats
}

// Shutdown is called to shutdown the service
func (cS *CoreService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.CoreS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.CoreS))
	return
}

func (cS *CoreService) Status(arg *utils.TenantWithOpts, reply *map[string]interface{}) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	response[utils.NodeID] = cS.cfg.GeneralCfg().NodeID
	response[utils.MemoryUsage] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	if response[utils.Version], err = utils.GetCGRVersion(); err != nil {
		utils.Logger.Err(err.Error())
		err = nil
	}
	response[utils.RunningSince] = utils.GetStartTime()
	response[utils.GoVersion] = runtime.Version()
	*reply = response
	return
}
