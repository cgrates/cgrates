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
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, file io.Closer,
	stopChan chan struct{}, shdEngine *utils.SyncedChan) *CoreService {
	var st *engine.CapsStats
	if caps.IsLimited() && cfg.CoreSCfg().CapsStatsInterval != 0 {
		st = engine.NewCapsStats(cfg.CoreSCfg().CapsStatsInterval, caps, stopChan)
	}
	return &CoreService{
		cfg:       cfg,
		CapsStats: st,
		shdEngine: shdEngine,
		file:      file,
	}
}

type CoreService struct {
	cfg       *config.CGRConfig
	CapsStats *engine.CapsStats
	shdEngine *utils.SyncedChan
	file      io.Closer
	fileMx    sync.Mutex
}

func (cS *CoreService) ShutdownEngine() {
	cS.shdEngine.CloseOnce()
}

// Shutdown is called to shutdown the service
func (cS *CoreService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.CoreS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.CoreS))
}

// Status returns the status of the engine
func (cS *CoreService) Status(_ *utils.TenantWithAPIOpts, reply *map[string]interface{}) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	response[utils.NodeID] = cS.cfg.GeneralCfg().NodeID
	response[utils.MemoryUsage] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	if response[utils.VersionName], err = utils.GetCGRVersion(); err != nil {
		utils.Logger.Err(err.Error())
		err = nil
	}
	response[utils.RunningSince] = utils.GetStartTime()
	response[utils.GoVersion] = runtime.Version()
	*reply = response
	return
}

// StartCPUProfiling is used to start CPUProfiling in the given path
func (cS *CoreService) StartCPUProfiling(argPath string) (err error) {
	cS.fileMx.Lock()
	defer cS.fileMx.Unlock()
	if cS.file != nil {
		return fmt.Errorf("CPU profiling already started")
	}
	if argPath == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("Path")
	}
	cS.file, err = StartCPUProfiling(argPath)
	return
}

// StopCPUProfiling is used to stop CPUProfiling in the given path
func (cS *CoreService) StopCPUProfiling() (err error) {
	cS.fileMx.Lock()
	defer cS.fileMx.Unlock()
	if cS.file != nil {
		pprof.StopCPUProfile()
		err = cS.file.Close()
		cS.file = nil
		return
	}
	return fmt.Errorf(" cannot stop because CPUProfiling is not active")
}

func StartCPUProfiling(path string) (file io.WriteCloser, err error) {
	file, err = os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("could not create CPU profile: %v", err)
	}
	err = pprof.StartCPUProfile(file)
	return
}
