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
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
	"sync"
	"time"

	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Caps the structure that allocs requests for API
type Caps struct {
	strategy string
	aReqs    chan struct{}
}

// NewCaps creates a new caps
func NewCaps(reqs int, strategy string) *Caps {
	cR := &Caps{
		strategy: strategy,
		aReqs:    make(chan struct{}, reqs),
	}
	return cR
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
	return
}

type conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

func newCapsGOBCodec(conn conn, caps *Caps, anz *analyzers.AnalyzerService) (r rpc.ServerCodec) {
	r = newCapsServerCodec(newGobServerCodec(conn), caps)
	if anz != nil {
		from := conn.RemoteAddr()
		var fromstr string
		if from != nil {
			fromstr = from.String()
		}
		to := conn.LocalAddr()
		var tostr string
		if to != nil {
			tostr = to.String()
		}
		return analyzers.NewServerCodec(r, anz, utils.MetaGOB, fromstr, tostr)
	}
	return
}

func newCapsJSONCodec(conn conn, caps *Caps, anz *analyzers.AnalyzerService) (r rpc.ServerCodec) {
	r = newCapsServerCodec(jsonrpc.NewServerCodec(conn), caps)
	if anz != nil {
		from := conn.RemoteAddr()
		var fromstr string
		if from != nil {
			fromstr = from.String()
		}
		to := conn.LocalAddr()
		var tostr string
		if to != nil {
			tostr = to.String()
		}
		return analyzers.NewServerCodec(r, anz, utils.MetaJSON, fromstr, tostr)
	}
	return
}

func newCapsServerCodec(sc rpc.ServerCodec, caps *Caps) rpc.ServerCodec {
	if !caps.IsLimited() {
		return sc
	}
	return &capsServerCodec{
		sc:   sc,
		caps: caps,
	}
}

type capsServerCodec struct {
	sc   rpc.ServerCodec
	caps *Caps
}

func (c *capsServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.sc.ReadRequestHeader(r)
}

func (c *capsServerCodec) ReadRequestBody(x interface{}) error {
	if err := c.caps.Allocate(); err != nil {
		return err
	}
	return c.sc.ReadRequestBody(x)
}
func (c *capsServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if r.Error == utils.ErrMaxConcurentRPCExceededNoCaps.Error() {
		r.Error = utils.ErrMaxConcurentRPCExceeded.Error()
	} else {
		defer c.caps.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}
func (c *capsServerCodec) Close() error { return c.sc.Close() }

// NewCapsStats returns the stats for the caps
func NewCapsStats(sampleinterval time.Duration, caps *Caps, exitChan chan bool) (cs *CapsStats) {
	st, _ := engine.NewStatAverage(1, utils.MetaDynReq, nil)
	cs = &CapsStats{st: st}
	go cs.loop(sampleinterval, exitChan, caps)
	return
}

// CapsStats stores the stats for caps
type CapsStats struct {
	sync.RWMutex
	st   engine.StatMetric
	peak int
}

// OnEvict the function that should be called on cache eviction
func (cs *CapsStats) OnEvict(itmID string, value interface{}) {
	cs.st.RemEvent(itmID)
}

func (cs *CapsStats) loop(intr time.Duration, exitChan chan bool, caps *Caps) {
	for {
		select {
		case v := <-exitChan:
			exitChan <- v
			return
		case <-time.After(intr):
			evID := time.Now().String()
			val := caps.Allocated()
			cs.addSample(evID, val)
		}
	}
}

func (cs *CapsStats) addSample(evID string, val int) {
	cs.Lock()
	engine.Cache.SetWithoutReplicate(utils.CacheCapsEvents, evID, val, nil, true, utils.NonTransactional)
	cs.st.AddEvent(evID, floatDP(val))
	if val > cs.peak {
		cs.peak = val
	}
	cs.Unlock()
}

// GetPeak returns the maximum allocated caps
func (cs *CapsStats) GetPeak() (peak int) {
	cs.RLock()
	peak = cs.peak
	cs.RUnlock()
	return
}

// GetAverage returns the average allocated caps
func (cs *CapsStats) GetAverage(roundingDecimals int) (avg float64) {
	cs.RLock()
	avg = cs.st.GetFloat64Value(roundingDecimals)
	cs.RUnlock()
	return
}

// floatDP should be only used by capstats
type floatDP float64

func (f floatDP) String() string                                         { return strconv.FormatFloat(float64(f), 'f', -1, 64) }
func (f floatDP) FieldAsInterface(fldPath []string) (interface{}, error) { return float64(f), nil }
func (f floatDP) FieldAsString(fldPath []string) (string, error)         { return f.String(), nil }
func (f floatDP) RemoteHost() net.Addr                                   { return nil }
