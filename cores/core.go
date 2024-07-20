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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, fileCPU io.Closer, fileMem string, stopChan chan struct{},
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
		fileCPU:    fileCPU,
		fileMEM:    fileMem,
		caps:       caps,
	}
}

type CoreService struct {
	cfg        *config.CGRConfig
	CapsStats  *engine.CapsStats
	shdWg      *sync.WaitGroup
	stopMemPrf chan struct{}
	shdChan    *utils.SyncedChan
	fileMEM    string

	fileMux sync.Mutex
	fileCPU io.Closer

	caps *engine.Caps
}

// Shutdown is called to shutdown the service
func (cS *CoreService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.CoreS))
	cS.StopChanMemProf()
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.CoreS))
}

// StopChanMemProf will stop the MemoryProfiling Channel in order to create
// the final MemoryProfiling when CoreS subsystem will stop.
func (cS *CoreService) StopChanMemProf() {
	if cS.stopMemPrf != nil {
		MemProfFile(cS.fileMEM)
		close(cS.stopMemPrf)
		cS.stopMemPrf = nil
	}
}

// StartCPUProfiling creates a file and passes it to pprof.StartCPUProfile. It returns the file
// as an io.Closer to be able to close it later when stopping the CPU profiling.
func StartCPUProfiling(path string) (io.Closer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("could not create CPU profile: %v", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		if err := f.Close(); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> could not close file %q: %v", utils.CoreS, f.Name(), err))
		}
		return nil, fmt.Errorf("could not start CPU profile: %v", err)
	}
	return f, nil
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

// V1Status returns the status of the engine
func (cS *CoreService) V1Status(_ *context.Context, _ *utils.TenantWithAPIOpts, reply *map[string]any) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]any)
	response[utils.NodeID] = cS.cfg.GeneralCfg().NodeID
	response[utils.MemoryUsage] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	if response[utils.VersionName], err = utils.GetCGRVersion(); err != nil {
		utils.Logger.Err(err.Error())
		err = nil
	}
	response[utils.RunningSince] = utils.GetStartTime()
	response[utils.GoVersion] = runtime.Version()
	if cS.cfg.CoreSCfg().Caps != 0 {
		response[utils.CAPSAllocated] = cS.caps.Allocated()
		if cS.cfg.CoreSCfg().CapsStatsInterval != 0 {
			response[utils.CAPSPeak] = cS.CapsStats.GetPeak()
		}
	}
	*reply = response
	return
}

// StartCPUProfiling starts CPU profiling and saves the profile to the specified path.
func (cS *CoreService) StartCPUProfiling(path string) (err error) {
	if path == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("DirPath")
	}
	cS.fileMux.Lock()
	defer cS.fileMux.Unlock()
	cS.fileCPU, err = StartCPUProfiling(path)
	return
}

// StopCPUProfiling stops CPU profiling and closes the profile file.
func (cS *CoreService) StopCPUProfiling() error {
	cS.fileMux.Lock()
	defer cS.fileMux.Unlock()
	pprof.StopCPUProfile()
	if cS.fileCPU == nil {
		return errors.New("CPU profiling has not been started")
	}
	if err := cS.fileCPU.Close(); err != nil {
		if errors.Is(err, os.ErrClosed) {
			return errors.New("CPU profiling has already been stopped")
		}
		return fmt.Errorf("could not close profile file: %v", err)
	}
	return nil
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
	cS.fileMEM = args.DirPath
	go MemProfiling(args.DirPath, args.Interval, args.NrFiles, cS.shdWg, cS.stopMemPrf, cS.shdChan)
	return
}

// StopMemoryProfiling is used to stop MemoryProfiling
func (cS *CoreService) StopMemoryProfiling() (err error) {
	if cS.stopMemPrf == nil {
		return errors.New(" Memory Profiling is not started")
	}
	cS.fileMEM = path.Join(cS.fileMEM, utils.MemProfFileCgr)
	cS.StopChanMemProf()
	return
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreService) V1Sleep(_ *context.Context, arg *utils.DurationArgs, reply *string) error {
	time.Sleep(arg.Duration)
	*reply = utils.OK
	return nil
}

// StartCPUProfiling is used to start CPUProfiling in the given path
func (cS *CoreService) V1StartCPUProfiling(_ *context.Context, args *utils.DirectoryArgs, reply *string) error {
	if err := cS.StartCPUProfiling(path.Join(args.DirPath, utils.CpuPathCgr)); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// StopCPUProfiling is used to stop CPUProfiling. The file should be written on the path
// where the CPUProfiling already started
func (cS *CoreService) V1StopCPUProfiling(_ *context.Context, _ *utils.TenantWithAPIOpts, reply *string) error {
	if err := cS.StopCPUProfiling(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// StartMemoryProfiling is used to start MemoryProfiling in the given path
func (cS *CoreService) V1StartMemoryProfiling(_ *context.Context, args *utils.MemoryPrf, reply *string) error {
	if err := cS.StartMemoryProfiling(args); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1StopMemoryProfiling is used to stop MemoryProfiling. The file should be written on the path
// where the MemoryProfiling already started
func (cS *CoreService) V1StopMemoryProfiling(_ *context.Context, _ *utils.TenantWithAPIOpts, reply *string) error {
	if err := cS.StopMemoryProfiling(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1Panic is used print the Message sent as a panic
func (cS *CoreService) V1Panic(_ *context.Context, args *utils.PanicMessageArgs, _ *string) error {
	panic(args.Message)
}
