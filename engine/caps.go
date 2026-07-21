// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"strconv"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Caps the structure that allocs requests for API
type Caps struct {
	strategy string
	aReqs    chan struct{}
}

// NewCaps creates a new caps
func NewCaps(reqs int, strategy string) *Caps {
	return &Caps{
		strategy: strategy,
		aReqs:    make(chan struct{}, reqs),
	}
}

// IsLimited returns true if the limit is not 0
func (cR *Caps) IsLimited() bool {
	return cap(cR.aReqs) != 0
}

// Allocated returns the number of requests actively serviced
func (cR *Caps) Allocated() int {
	return len(cR.aReqs)
}

// Allocate will reserve a channel for the API call
func (cR *Caps) Allocate() (err error) {
	switch cR.strategy {
	case utils.MetaBusy:
		if len(cR.aReqs) == cap(cR.aReqs) {
			return utils.ErrMaxConcurentRPCExceededNoCaps
		}
		fallthrough
	case utils.MetaQueue:
		cR.aReqs <- struct{}{}
	}
	return
}

// Deallocate will free a channel for the API call
func (cR *Caps) Deallocate() {
	<-cR.aReqs
}

// NewCapsStats returns the stats for the caps
func NewCapsStats(sampleinterval time.Duration, caps *Caps, stopChan chan struct{}) (cs *CapsStats) {
	cs = &CapsStats{st: utils.NewStatAverage(1, utils.MetaDynReq, nil)}
	go cs.loop(sampleinterval, stopChan, caps)
	return
}

// CapsStats stores the stats for caps
type CapsStats struct {
	mu    sync.RWMutex
	cache *CacheS
	st    utils.StatMetric
	peak  int
}

func (cs *CapsStats) SetCache(cache *CacheS) {
	cs.mu.Lock()
	cs.cache = cache
	cs.mu.Unlock()
}

// OnEvict the function that should be called on cache eviction
func (cs *CapsStats) OnEvict(itmID string, value any) {
	cs.st.RemEvent(itmID)
}

func (cs *CapsStats) loop(intr time.Duration, stopChan chan struct{}, caps *Caps) {
	for {
		select {
		case <-stopChan:
			return
		case <-time.After(intr):
			evID := time.Now().String()
			val := caps.Allocated()
			cs.addSample(evID, val)
		}
	}
}

func (cs *CapsStats) addSample(evID string, val int) {
	cs.mu.Lock()
	if cs.cache != nil {
		cs.cache.SetWithoutReplicate(utils.CacheCapsEvents, evID, val, nil, true, utils.NonTransactional)
	}
	cs.st.AddEvent(evID, floatDP(val))
	if val > cs.peak {
		cs.peak = val
	}
	cs.mu.Unlock()
}

// GetPeak returns the maximum allocated caps
func (cs *CapsStats) GetPeak() (peak int) {
	cs.mu.RLock()
	peak = cs.peak
	cs.mu.RUnlock()
	return
}

// GetAverage returns the average allocated caps
func (cs *CapsStats) GetAverage() (avg float64) {
	cs.mu.RLock()
	avg, _ = cs.st.GetValue().Float64()
	cs.mu.RUnlock()
	return
}

// floatDP should be only used by capstats
type floatDP float64

func (f floatDP) String() string                                 { return strconv.FormatFloat(float64(f), 'f', -1, 64) }
func (f floatDP) FieldAsInterface(fldPath []string) (any, error) { return float64(f), nil }
func (f floatDP) FieldAsString(fldPath []string) (string, error) { return f.String(), nil }
