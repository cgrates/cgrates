// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package cores

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// serviceRanger exposes the RPC services registered on the server
type serviceRanger interface {
	RangeServices(func(name string, svc *birpc.Service) bool)
}

func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, profiler *Profiler, stopChan chan struct{},
	shutdown *utils.SyncedChan, cl serviceRanger) *CoreS {
	var st *engine.CapsStats
	if caps.IsLimited() && cfg.CoreSCfg().CapsStatsInterval != 0 {
		st = engine.NewCapsStats(cfg.CoreSCfg().CapsStatsInterval, caps, stopChan)
	}
	return &CoreS{
		shutdown:  shutdown,
		cfg:       cfg,
		CapsStats: st,
		profiler:  profiler,
		caps:      caps,
		cl:        cl,
	}
}

type CoreS struct {
	cfg       *config.CGRConfig
	CapsStats *engine.CapsStats
	shutdown  *utils.SyncedChan
	cl        serviceRanger
	profiler  *Profiler
	caps      *engine.Caps
}

func (cS *CoreS) ShutdownEngine() {
	cS.shutdown.CloseOnce()
}

// Profiler controls CPU and memory profiling for the engine process.
type Profiler struct {
	shdWg *sync.WaitGroup

	cpuMu      sync.Mutex
	fileCPU    *os.File
	cpuStarted bool

	stateMu    sync.Mutex
	closing    atomic.Bool
	memoryStop chan struct{}
	memoryDone chan struct{}
}

// NewProfiler uses shdWg to account for memory profiling and wait for process shutdown.
func NewProfiler(shdWg *sync.WaitGroup) *Profiler {
	return &Profiler{shdWg: shdWg}
}

// Shutdown prevents new profiling runs and asks memory profiling to stop.
// It waits for the process WaitGroup until it completes or ctx is done, then
// flushes and closes the CPU profile. The CPU cleanup is not limited by ctx,
// and errors from waiting and closing the file are returned together.
func (p *Profiler) Shutdown(ctx *context.Context) error {
	p.stateMu.Lock()
	p.closing.Store(true)
	if p.memoryStop != nil {
		select {
		case <-p.memoryStop:
		default:
			close(p.memoryStop)
		}
	}
	p.stateMu.Unlock()

	done := make(chan struct{})
	go func() {
		p.shdWg.Wait()
		close(done)
	}()
	var err error
	select {
	case <-done:
	case <-ctx.Done():
		err = fmt.Errorf("waiting for process shutdown: %w", ctx.Err())
	}

	p.cpuMu.Lock()
	defer p.cpuMu.Unlock()
	return errors.Join(err, p.stopCPUProfiling())
}

// StartCPUProfiling starts CPU profiling and writes the profile to path.
func (p *Profiler) StartCPUProfiling(path string) error {
	p.cpuMu.Lock()
	defer p.cpuMu.Unlock()
	if p.closing.Load() {
		return errors.New("start CPU profiling: shutting down")
	}
	if p.fileCPU != nil {
		return errors.New("start CPU profiling: already started")
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		if err := f.Close(); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
		return fmt.Errorf("could not start CPU profile: %v", err)
	}
	p.fileCPU = f
	p.cpuStarted = true
	return nil
}

// StopCPUProfiling flushes and closes the CPU profile.
func (p *Profiler) StopCPUProfiling() error {
	p.cpuMu.Lock()
	defer p.cpuMu.Unlock()
	if p.fileCPU == nil {
		if p.cpuStarted {
			return errors.New("stop CPU profiling: already stopped")
		}
		return errors.New("stop CPU profiling: not started yet")
	}
	return p.stopCPUProfiling()
}

func (p *Profiler) stopCPUProfiling() error {
	if p.fileCPU == nil {
		return nil
	}
	pprof.StopCPUProfile()
	err := p.fileCPU.Close()
	p.fileCPU = nil
	if err != nil {
		return fmt.Errorf("could not close profile file: %v", err)
	}
	return nil
}

// MemoryProfilingParams configures memory profiling.
type MemoryProfilingParams struct {
	Tenant   string
	DirPath  string        // heap profile directory, used by only this engine
	Interval time.Duration // time between captures, defaults to 15 seconds for values <= 0

	// MaxFiles limits periodic profiles in DirPath, including earlier runs.
	// The final profile is excluded. Values <= 0 keep all files.
	MaxFiles int

	// UseTimestamp adds the capture's start time to filenames, using
	// mem_20060102150405_<suffix>.prof. Without timestamps, filenames use
	// mem_<n>_<suffix>.prof. The random suffix prevents overwriting existing files.
	UseTimestamp bool

	APIOpts map[string]any
}

// StartMemoryProfiling starts periodic heap profiles in params.DirPath.
func (p *Profiler) StartMemoryProfiling(params MemoryProfilingParams) error {
	if params.Interval <= 0 {
		params.Interval = 15 * time.Second
	}

	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	if p.closing.Load() {
		return errors.New("start memory profiling: shutting down")
	}
	if p.memoryDone != nil {
		select {
		case <-p.memoryDone:
		default:
			return errors.New("start memory profiling: already started")
		}
	}

	utils.Logger.Info(fmt.Sprintf(
		"<%s> starting memory profiling loop, writing to directory %q", utils.CoreS, params.DirPath))
	p.memoryStop = make(chan struct{})
	p.memoryDone = make(chan struct{})
	p.shdWg.Add(1)
	go p.profileMemory(params, p.memoryStop, p.memoryDone)
	return nil
}

// Close done after the final write finishes so the next run cannot overlap it.
func (p *Profiler) profileMemory(params MemoryProfilingParams, stop <-chan struct{}, done chan<- struct{}) {
	defer p.shdWg.Done()
	defer close(done)
	defer func() {
		f, err := os.Create(filepath.Join(params.DirPath, utils.MemProfFinalFile))
		if err != nil {
			err = fmt.Errorf("could not create memory profile: %v", err)
		} else {
			err = writeHeapProfileFile(f)
		}
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
	}()
	var files []string
	if params.MaxFiles > 0 {
		var err error
		if files, err = listMemoryProfiles(params.DirPath); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
			return
		}
	}
	i := 0
	ticker := time.NewTicker(params.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			i++
			pattern := fmt.Sprintf("mem_%d_*.prof", i)
			if params.UseTimestamp {
				pattern = fmt.Sprintf("mem_%s_*.prof", time.Now().Format("20060102150405"))
			}
			f, err := os.CreateTemp(filepath.Clean(params.DirPath), pattern)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> could not create memory profile: %v", utils.CoreS, err))
				return
			}
			path := f.Name()
			if err := writeHeapProfileFile(f); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
				return
			}
			if params.MaxFiles <= 0 {
				continue
			}
			for len(files) >= params.MaxFiles {
				oldest := files[0]
				utils.Logger.Info(fmt.Sprintf("<%s> removing old heap profile file %q", utils.CoreS, oldest))
				files = slices.Delete(files, 0, 1)
				if err := os.Remove(oldest); err != nil {
					utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.CoreS, err))
				}
			}
			files = append(files, path)
		case <-stop:
			return
		}
	}
}

func listMemoryProfiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Clean(dir))
	if err != nil {
		return nil, fmt.Errorf("could not list memory profiles: %w", err)
	}
	var profiles []os.FileInfo
	for _, entry := range entries {
		name := entry.Name()
		if name == utils.MemProfFinalFile ||
			!strings.HasPrefix(name, "mem_") || !strings.HasSuffix(name, ".prof") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("could not stat memory profile %q: %w", name, err)
		}
		if info.Mode().IsRegular() {
			profiles = append(profiles, info)
		}
	}
	slices.SortStableFunc(profiles, func(a, b os.FileInfo) int {
		return a.ModTime().Compare(b.ModTime())
	})
	files := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		files = append(files, filepath.Join(dir, profile.Name()))
	}
	return files, nil
}

func writeHeapProfileFile(f *os.File) error {
	defer func() {
		if err := f.Close(); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> could not close file %q: %v", utils.CoreS, f.Name(), err))
		}
	}()
	utils.Logger.Info(fmt.Sprintf("<%s> writing heap profile to %q", utils.CoreS, f.Name()))
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %v", err)
	}
	return nil
}

// StopMemoryProfiling stops memory profiling and waits for the final write.
func (p *Profiler) StopMemoryProfiling() error {
	p.stateMu.Lock()
	stop, done := p.memoryStop, p.memoryDone
	if done == nil {
		p.stateMu.Unlock()
		return errors.New("stop memory profiling: not started yet")
	}
	var err error
	select {
	case <-done:
		err = errors.New("stop memory profiling: already stopped")
	case <-stop:
		err = errors.New("stop memory profiling: already stopped")
	default:
		utils.Logger.Info(fmt.Sprintf("<%s> stopping memory profiling loop", utils.CoreS))
		close(stop)
	}
	p.stateMu.Unlock()
	<-done
	return err
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

// V1StartCPUProfiling starts CPU profiling in args.DirPath.
func (cS *CoreS) V1StartCPUProfiling(_ *context.Context, args *utils.DirectoryArgs, reply *string) error {
	if err := cS.profiler.StartCPUProfiling(filepath.Join(args.DirPath, utils.CpuPathCgr)); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1StopCPUProfiling stops CPU profiling.
func (cS *CoreS) V1StopCPUProfiling(_ *context.Context, _ *utils.TenantWithAPIOpts, reply *string) error {
	if err := cS.profiler.StopCPUProfiling(); err != nil {
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
	if err := cS.profiler.StartMemoryProfiling(params); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1StopMemoryProfiling stops memory profiling.
func (cS *CoreS) V1StopMemoryProfiling(_ *context.Context, _ utils.TenantWithAPIOpts, reply *string) error {
	if err := cS.profiler.StopMemoryProfiling(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// V1Panic is used print the Message sent as a panic
func (cS *CoreS) V1Panic(_ *context.Context, args *utils.PanicMessageArgs, _ *string) error {
	panic(args.Message)
}

// V1DescribeMethods returns every registered method's argument and reply types
// as field descriptors. Services in args.SkipServices are dropped; empty args
// returns the full set, _goRPC_ included.
func (cS *CoreS) V1DescribeMethods(_ *context.Context, args *utils.DescribeMethodsArgs, reply *[]utils.MethodDescriptor) error {
	var skip []string
	if args != nil {
		skip = args.SkipServices
	}
	var mds []utils.MethodDescriptor
	cS.cl.RangeServices(func(name string, svc *birpc.Service) bool {
		if slices.Contains(skip, name) {
			return true
		}
		for method, mt := range svc.Methods {
			mds = append(mds, utils.MethodDescriptor{
				Method: name + "." + method,
				Args:   utils.DescribeType(mt.ArgType),
				Result: utils.DescribeType(mt.ReplyType),
			})
		}
		return true
	})
	slices.SortFunc(mds, func(a, b utils.MethodDescriptor) int {
		return strings.Compare(a.Method, b.Method)
	})
	*reply = mds
	return nil
}
