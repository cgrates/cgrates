/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
)

type Scheduler struct {
	queue       engine.ActionTimingPriotityList
	timer       *time.Timer
	restartLoop chan bool
	sync.Mutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{restartLoop: make(chan bool)}
}

func (s *Scheduler) Loop() {
	for {
		for len(s.queue) == 0 { //hang here if empty
			<-s.restartLoop
		}
		s.Lock()
		a0 := s.queue[0]
		now := time.Now()
		if a0.GetNextStartTime(now).Equal(now) || a0.GetNextStartTime(now).Before(now) {
			engine.Logger.Debug(fmt.Sprintf("%v - %v", a0.Id, a0.Timing))
			go a0.Execute()
			// if after execute the next start time is in the past then
			// do not add it to the queue
			now = time.Now()
			if !a0.GetNextStartTime(now).Before(now) {
				s.queue = append(s.queue, a0)
				s.queue = s.queue[1:]
				sort.Sort(s.queue)
			}
			s.Unlock()
		} else {
			s.Unlock()
			d := a0.GetNextStartTime(now).Sub(now)
			// engine.Logger.Info(fmt.Sprintf("Timer set to wait for %v", d))
			s.timer = time.NewTimer(d)
			select {
			case <-s.timer.C:
				// timer has expired
				engine.Logger.Info(fmt.Sprintf("Time for action on %v", a0))
			case <-s.restartLoop:
				// nothing to do, just continue the loop
			}
		}
	}
}

func (s *Scheduler) LoadActionTimings(storage engine.AccountingStorage) {
	actionTimings, err := storage.GetAllActionTimings()
	if err != nil {
		engine.Logger.Warning(fmt.Sprintf("Cannot get action timings: %v", err))
	}
	// recreate the queue
	s.Lock()
	s.queue = engine.ActionTimingPriotityList{}
	for key, ats := range actionTimings {
		toBeSaved := false
		isAsap := false
		newAts := make([]*engine.ActionTiming, 0) // will remove the one time runs from the database
		for _, at := range ats {
			isAsap = at.IsASAP()
			toBeSaved = toBeSaved || isAsap
			if isAsap {
				engine.Logger.Info(fmt.Sprintf("Time for one time action on %v", key))
				go at.Execute()
				// do not append it to the newAts list to be saved
			} else {
				now := time.Now()
				if at.GetNextStartTime(now).Before(now) {
					// the task is obsolete, do not add it to the queue
					continue
				}
				s.queue = append(s.queue, at)
				newAts = append(newAts, at)
			}
		}
		if toBeSaved {
			engine.AccLock.Guard(engine.ACTION_TIMING_PREFIX, func() (float64, error) {
				storage.SetActionTimings(key, newAts)
				return 0, nil
			})
		}
	}
	sort.Sort(s.queue)
	s.Unlock()
}

func (s *Scheduler) Restart() {
	s.restartLoop <- true
	if s.timer != nil {
		s.timer.Stop()
	}
}

func (s *Scheduler) GetQueue() engine.ActionTimingPriotityList {
	return s.queue
}
