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

package engine

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCaps(t *testing.T) {
	exp := &Caps{
		strategy: utils.MetaBusy,
		aReqs:    make(chan struct{}),
	}
	cs := NewCaps(0, utils.MetaBusy)

	// only check the strategy
	if !reflect.DeepEqual(exp.strategy, cs.strategy) {
		t.Errorf("Expected: %v ,received: %v", exp, cs)
	}

	if cs.IsLimited() {
		t.Errorf("Expected to not be limited")
	}

	if al := cs.Allocated(); al != 0 {
		t.Errorf("Expected: %v ,received: %v", 0, al)
	}
	if err := cs.Allocate(); err != utils.ErrMaxConcurrentRPCExceededNoCaps {
		t.Errorf("Expected: %v ,received: %v", utils.ErrMaxConcurrentRPCExceededNoCaps, err)
	}
	cs = NewCaps(1, utils.MetaBusy)
	if err := cs.Allocate(); err != nil {
		t.Error(err)
	}
	cs.Deallocate()
}

func TestCapsStats(t *testing.T) {
	st, err := NewStatAverage(1, utils.MetaDynReq, nil)
	if err != nil {
		t.Error(err)
	}
	exp := &CapsStats{st: st}
	cr := NewCaps(0, utils.MetaBusy)
	stopChan := make(chan struct{}, 1)
	close(stopChan)
	cs := NewCapsStats(1, cr, stopChan)
	if !reflect.DeepEqual(exp, cs) {
		t.Errorf("Expected: %v ,received: %v", exp, cs)
	}
	<-stopChan
	stopChan = make(chan struct{}, 1)
	go func() {
		runtime.Gosched()
		time.Sleep(100 * time.Nanosecond)
		close(stopChan)
	}()
	cr = NewCaps(10, utils.MetaBusy)
	cr.Allocate()
	cr.Allocate()
	cs.loop(1, stopChan, cr)
	if avg := cs.GetAverage(2); avg <= 0 {
		t.Errorf("Expected at least an event to be processed: %v", avg)
	}
	if pk := cs.GetPeak(); pk != 2 {
		t.Errorf("Expected the peak to be 2 received: %v", pk)
	}
	<-stopChan
}

func TestCapsStatsGetAverage(t *testing.T) {
	st, err := NewStatAverage(1, utils.MetaDynReq, nil)
	if err != nil {
		t.Error(err)
	}
	cs := &CapsStats{st: st}
	cs.addSample("1", 10)
	expAvg := 10.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk := 10
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	cs.addSample("2", 16)
	expAvg = 13.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk = 16
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	cs.OnEvict("2", nil)
	expAvg = 10.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
}

func TestFloatDP(t *testing.T) {
	f := floatDP(10.)
	expStr := "10"
	if s := f.String(); s != expStr {
		t.Errorf("Expected: %v ,received:%v", expStr, s)
	}
	if s, err := f.FieldAsString(nil); err != nil {
		t.Error(err)
	} else if s != expStr {
		t.Errorf("Expected: %v ,received:%v", expStr, s)
	}
	exp := 10.
	if s, err := f.FieldAsInterface(nil); err != nil {
		t.Error(err)
	} else if s != exp {
		t.Errorf("Expected: %v ,received:%v", exp, s)
	}
}

func TestCapsStatsGetAverageOnEvict(t *testing.T) {
	st, err := NewStatAverage(1, utils.MetaDynReq, nil)
	if err != nil {
		t.Error(err)
	}
	cs := &CapsStats{st: st}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheCapsEvents] = &config.CacheParamCfg{Limit: 2}
	tmp := Cache
	Cache = NewCacheS(cfg, nil, cs)

	cs.addSample("1", 10)
	expAvg := 10.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk := 10
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	cs.addSample("2", 16)
	expAvg = 13.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk = 16
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	cs.addSample("3", 18)
	expAvg = 17.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk = 18
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	Cache = tmp
}

func BenchmarkCaps(b *testing.B) {
	caps := NewCaps(10, utils.MetaBusy)
	for i := 0; i < b.N; i++ {
		if caps.IsLimited() {
			caps.Allocate()
			caps.Deallocate()
		}
	}
}
