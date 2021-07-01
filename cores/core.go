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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, file io.Closer, stopChan chan struct{},
	shdWg *sync.WaitGroup, stopMemPrf chan struct{}, shdChan *utils.SyncedChan) *CoreService {
	var st *engine.CapsStats
	if caps.IsLimited() && cfg.CoreSCfg().CapsStatsInterval != 0 {
		st = engine.NewCapsStats(cfg.CoreSCfg().CapsStatsInterval, caps, stopChan)
	}
	return &CoreService{
		shdWg:      shdWg,
		stopMemPrf: stopMemPrf,
		shdChan:    shdChan,
		cfg:        cfg,
		CapsStats:  st,
		fileCPU:    file,
	}
}

type CoreService struct {
	cfg        *config.CGRConfig
	CapsStats  *engine.CapsStats
	shdWg      *sync.WaitGroup
	stopMemPrf chan struct{}
	shdChan    *utils.SyncedChan
	fileCPU    io.Closer
	fileMx     sync.Mutex
}

// Shutdown is called to shutdown the service
func (cS *CoreService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.CoreS))
	cS.StopChanMemProf()
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.CoreS))
	return
}

// StopChanMemProf will stop the MemoryProfiling Channel in order to create
// the final MemoryProfiling when CoreS subsystem will stop.
func (cS *CoreService) StopChanMemProf() bool {
	if cS.stopMemPrf != nil {
		close(cS.stopMemPrf)
		cS.stopMemPrf = nil
		return true
	}
	return false
}

func StartCPUProfiling(path string) (file io.WriteCloser, err error) {
	file, err = os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("could not create CPU profile: %v", err)
	}
	err = pprof.StartCPUProfile(file)
	return
}

func MemProfFile(memProfPath string) bool {
	f, err := os.Create(memProfPath)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<memProfile>could not create memory profile file: %s", err))
		return false
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<memProfile>could not write memory profile: %s", err))
		f.Close()
		return false
	}
	f.Close()
	return true
}

func MemProfiling(memProfDir string, interval time.Duration, nrFiles int, shdWg *sync.WaitGroup, stopChan chan struct{}, shdChan *utils.SyncedChan) {
	tm := time.NewTimer(interval)
	for i := 1; ; i++ {
		select {
		case <-stopChan:
			tm.Stop()
			shdWg.Done()
			return
		case <-tm.C:
		}
		if !MemProfFile(path.Join(memProfDir, fmt.Sprintf("mem%v.prof", i))) {
			shdChan.CloseOnce()
			shdWg.Done()
			return
		}
		if i%nrFiles == 0 {
			i = 0 // reset the counting
		}
		tm.Reset(interval)
	}
}

// Status returns the status of the engine
func (cS *CoreService) Status(arg *utils.TenantWithAPIOpts, reply *map[string]interface{}) (err error) {
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
	if cS.fileCPU != nil {
		return fmt.Errorf("CPU profiling already started")
	}
	if argPath == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("Path")
	}
	cS.fileCPU, err = StartCPUProfiling(argPath)
	return
}

// StopCPUProfiling is used to stop CPUProfiling in the given path
func (cS *CoreService) StopCPUProfiling() (err error) {
	cS.fileMx.Lock()
	defer cS.fileMx.Unlock()
	if cS.fileCPU != nil {
		pprof.StopCPUProfile()
		err = cS.fileCPU.Close()
		cS.fileCPU = nil
		return
	}
	return fmt.Errorf(" cannot stop because CPUProfiling is not active")
}

// StartMemoryProfiling is used to start MemoryProfiling in the given path
func (cS *CoreService) StartMemoryProfiling(args *utils.MemoryPrf) (err error) {
	if args.DirPath == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("Path")
	}
	if cS.stopMemPrf != nil {
		return errors.New("Memory Profiling already started")
	}
	if args.Interval <= 0 {
		args.Interval = 5 * time.Second
	}
	if args.NrFiles == 0 {
		args.NrFiles = 1
	}
	cS.shdWg.Add(1)
	cS.stopMemPrf = make(chan struct{})
	go MemProfiling(args.DirPath, args.Interval, args.NrFiles, cS.shdWg, cS.stopMemPrf, cS.shdChan)
	return
}

// StopMemoryProfiling is used to stop MemoryProfiling
func (cS *CoreService) StopMemoryProfiling() (err error) {
	if cS.stopMemPrf == nil {
		return errors.New(" Memory Profiling is not started")
	}
	close(cS.stopMemPrf)
	cS.stopMemPrf = nil
	return
}
