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
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, fileCPU *os.File, stopChan chan struct{},
	shdWg *sync.WaitGroup, shutdown *utils.SyncedChan) *CoreS {
	var st *engine.CapsStats
	if caps.IsLimited() && cfg.CoreSCfg().CapsStatsInterval != 0 {
		st = engine.NewCapsStats(cfg.CoreSCfg().CapsStatsInterval, caps, stopChan)
	}
	return &CoreS{
		shdWg:     shdWg,
		shutdown:  shutdown,
		cfg:       cfg,
		CapsStats: st,
		fileCPU:   fileCPU,
		caps:      caps,
	}
}

type CoreS struct {
	cfg       *config.CGRConfig
	CapsStats *engine.CapsStats
	shdWg     *sync.WaitGroup
	shutdown  *utils.SyncedChan

	memProfMux   sync.Mutex
	finalMemProf string        // full path of the final memory profile created on stop/shutdown
	stopMemProf  chan struct{} // signal end of memory profiling

	fileCPUMux sync.Mutex
	fileCPU    *os.File

	caps *engine.Caps
}

func (cS *CoreS) ShutdownEngine() {
	cS.shutdown.CloseOnce()
}

// Shutdown is called to shutdown the service
func (cS *CoreS) Shutdown() {
	// safe to ignore errors (irrelevant)
	_ = cS.StopMemoryProfiling()
	_ = cS.StopCPUProfiling()
}

// StartCPUProfiling starts CPU profiling and saves the profile to the specified path.
func (cS *CoreS) StartCPUProfiling(path string) error {
	if path == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("DirPath")
	}
	cS.fileCPUMux.Lock()
	defer cS.fileCPUMux.Unlock()

	if cS.fileCPU != nil {
		// Check if the profiling is already active by calling Stat() on the file handle.
		// If Stat() returns nil, it means profiling is already active.
		if _, err := cS.fileCPU.Stat(); err == nil {
			return errors.New("start CPU profiling: already started")
		}
	}
	file, err := StartCPUProfiling(path)
	if err != nil {
		return err
	}
	cS.fileCPU = file
	return nil
}

// StartCPUProfiling creates a file and passes it to pprof.StartCPUProfile. It returns the file
// to be able to verify the status of profiling and close it after profiling is stopped.
func StartCPUProfiling(path string) (*os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("could not create CPU profile: %v", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		if err := f.Close(); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
		return nil, fmt.Errorf("could not start CPU profile: %v", err)
	}
	return f, nil
}

// StopCPUProfiling stops CPU profiling and closes the profile file.
func (cS *CoreS) StopCPUProfiling() error {
	cS.fileCPUMux.Lock()
	defer cS.fileCPUMux.Unlock()
	pprof.StopCPUProfile()
	if cS.fileCPU == nil {
		return errors.New("stop CPU profiling: not started yet")
	}
	if err := cS.fileCPU.Close(); err != nil {
		if errors.Is(err, os.ErrClosed) {
			return errors.New("stop CPU profiling: already stopped")
		}
		return fmt.Errorf("could not close profile file: %v", err)
	}
	return nil
}

// MemoryProfilingParams represents the parameters for memory profiling.
type MemoryProfilingParams struct {
	Tenant   string
	DirPath  string        // directory path where memory profiles will be saved
	Interval time.Duration // duration between consecutive memory profile captures
	MaxFiles int           // maximum number of profile files to retain

	// UseTimestamp determines if the filename includes a timestamp.
	// The format is 'mem_20060102150405[_<microseconds>].prof'.
	// Microseconds are included if the interval is less than one second to avoid duplicate names.
	// If false, filenames follow an incremental format: 'mem_<n>.prof'.
	UseTimestamp bool

	APIOpts map[string]any
}

// StartMemoryProfiling starts memory profiling in the specified directory.
func (cS *CoreS) StartMemoryProfiling(params MemoryProfilingParams) error {
	if params.Interval <= 0 {
		params.Interval = 15 * time.Second
	}
	if params.MaxFiles < 0 {
		// consider any negative number to mean unlimited files
		params.MaxFiles = 0
	}

	cS.memProfMux.Lock()
	defer cS.memProfMux.Unlock()

	// Check if profiling is already started.
	select {
	case <-cS.stopMemProf: // triggered only on channel closed
	default:
		if cS.stopMemProf != nil {
			// stopMemProf being not closed and different from nil means that the profiling loop is already active.
			return errors.New("start memory profiling: already started")
		}
	}

	utils.Logger.Info(fmt.Sprintf(
		"<%s> starting memory profiling loop, writing to directory %q", utils.CoreS, params.DirPath))
	cS.stopMemProf = make(chan struct{})
	cS.finalMemProf = filepath.Join(params.DirPath, utils.MemProfFinalFile)
	cS.shdWg.Add(1)
	go cS.profileMemory(params)
	return nil
}

// newMemProfNameFunc returns a closure that generates memory profile filenames.
func newMemProfNameFunc(interval time.Duration, useTimestamp bool) func() string {
	if !useTimestamp {
		i := 0
		return func() string {
			i++
			return fmt.Sprintf("mem_%d.prof", i)
		}
	}
	if interval < time.Second {
		return func() string {
			now := time.Now()
			return fmt.Sprintf("mem_%s_%d.prof", now.Format("20060102150405"), now.Nanosecond()/1e3)
		}
	}

	return func() string {
		return fmt.Sprintf("mem_%s.prof", time.Now().Format("20060102150405"))
	}
}

// profileMemory runs the memory profiling loop, writing profiles to files at the specified interval.
func (cS *CoreS) profileMemory(params MemoryProfilingParams) {
	defer cS.shdWg.Done()
	fileName := newMemProfNameFunc(params.Interval, params.UseTimestamp)
	ticker := time.NewTicker(params.Interval)
	defer ticker.Stop()
	files := make([]string, 0, params.MaxFiles)
	for {
		select {
		case <-ticker.C:
			path := filepath.Join(params.DirPath, fileName())
			if err := writeHeapProfile(path); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
				cS.StopMemoryProfiling()
			}
			if params.MaxFiles == 0 {
				// no file limit
				continue
			}
			if len(files) == params.MaxFiles {
				oldest := files[0]
				utils.Logger.Info(fmt.Sprintf("<%s> removing old heap profile file %q", utils.CoreS, oldest))
				files = files[1:] // remove oldest file from the list
				if err := os.Remove(oldest); err != nil {
					utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.CoreS, err))
				}
			}
			files = append(files, path)
		case <-cS.stopMemProf:
			if err := writeHeapProfile(cS.finalMemProf); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
			}
			return
		}
	}
}

// writeHeapProfile writes the heap profile to the specified path.
func writeHeapProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> could not close file %q: %v", utils.CoreS, f.Name(), err))
		}
	}()
	utils.Logger.Info(fmt.Sprintf("<%s> writing heap profile to %q", utils.CoreS, path))
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %v", err)
	}
	return nil
}

// StopMemoryProfiling stops memory profiling.
func (cS *CoreS) StopMemoryProfiling() error {
	cS.memProfMux.Lock()
	defer cS.memProfMux.Unlock()

	// Check if profiling is already stopped to prevent a channel close panic.
	select {
	case <-cS.stopMemProf: // triggered only on channel closed
		return errors.New("stop memory profiling: already stopped")
	default: // prevents blocking
		if cS.stopMemProf == nil {
			// stopMemProf being nil means that StartMemoryProfiling has never been called. There is nothing to stop.
			return errors.New("stop memory profiling: not started yet")
		}
	}

	utils.Logger.Info(fmt.Sprintf("<%s> stopping memory profiling loop", utils.CoreS))
	close(cS.stopMemProf)
	return nil
}

// V1StatusParams contains required parameters for a CoreSv1.Status request.
type V1StatusParams struct {
	Debug    bool
	Timezone string
	Tenant   string
	APIOpts  map[string]any
}

// V1Status returns metrics related to the engine process.
func (cS *CoreS) V1Status(_ *context.Context, params *V1StatusParams, reply *map[string]any) error {
	metrics, err := computeAppMetrics()
	if err != nil {
		return err
	}
	metrics.NodeID = cS.cfg.GeneralCfg().NodeID
	if cS.cfg.CoreSCfg().Caps != 0 {
		metrics.CapsStats = &CapsStats{
			Allocated: cS.caps.Allocated(),
		}
		if cS.cfg.CoreSCfg().CapsStatsInterval != 0 {
			peak := cS.CapsStats.GetPeak()
			metrics.CapsStats.Peak = &peak
		}
	}
	debug := false
	timezone := cS.cfg.GeneralCfg().DefaultTimezone
	if params != nil {
		debug = params.Debug
		timezone = params.Timezone
	}
	metricsMap, err := metrics.toMap(debug, timezone)
	if err != nil {
		return fmt.Errorf("could not convert StatusMetrics to map[string]any: %v", err)
	}
	*reply = metricsMap
	return nil
}

// Sleep is used to test the concurrent requests mechanism.
func (cS *CoreS) V1Sleep(_ *context.Context, arg *utils.DurationArgs, reply *string) error {
	time.Sleep(arg.Duration)
	*reply = utils.OK
	return nil
}

func (cS *CoreS) V1Shutdown(_ *context.Context, _ *utils.CGREvent, reply *string) error {
	cS.ShutdownEngine()
	*reply = utils.OK
	return nil
}

// V1StartCPUProfiling starts CPU profiling and saves the profile to the specified path.
func (cS *CoreS) V1StartCPUProfiling(_ *context.Context, args *utils.DirectoryArgs, reply *string) error {
	if err := cS.StartCPUProfiling(filepath.Join(args.DirPath, utils.CpuPathCgr)); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1StopCPUProfiling stops CPU Profiling.
func (cS *CoreS) V1StopCPUProfiling(_ *context.Context, _ *utils.TenantWithAPIOpts, reply *string) error {
	if err := cS.StopCPUProfiling(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1StartMemoryProfiling starts memory profiling in the specified directory.
func (cS *CoreS) V1StartMemoryProfiling(_ *context.Context, params MemoryProfilingParams, reply *string) error {
	if params.DirPath == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("DirPath")
	}
	if err := cS.StartMemoryProfiling(params); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1StopMemoryProfiling stops memory profiling.
func (cS *CoreS) V1StopMemoryProfiling(_ *context.Context, _ utils.TenantWithAPIOpts, reply *string) error {
	if err := cS.StopMemoryProfiling(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1Panic is used print the Message sent as a panic
func (cS *CoreS) V1Panic(_ *context.Context, args *utils.PanicMessageArgs, _ *string) error {
	panic(args.Message)
}
